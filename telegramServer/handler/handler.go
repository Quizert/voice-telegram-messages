package handler

import (
	"bytes"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
	"kursach/defs"
	pb "kursach/proto"
	"os/exec"
	"strconv"
	"strings"
)

type Service interface {
	SendAudio(userID int64, text string) (*pb.ProcessingResponse, error)
	SaveModel(userID int64, fileInfo string, token string, modelName string) error

	SetPendingModel(userID int64, name string) error
	GetModelName(userID int64) (string, error)

	GetUserState(userID int64) (string, error)
	SetUserState(userID int64, state string) error

	GetUserModels(userID int64) ([]string, error)

	CountModels(userID int64) (int, error)

	DeleteModel(userID int64, modelName string) error
}

type Handler struct {
	service Service
	log     *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		log:     logger,
	}
}

func (h *Handler) HandleText(c telebot.Context) error {
	userID := c.Sender().ID
	h.log.Info("HandleText called", zap.Int64("userID", userID))

	state, err := h.service.GetUserState(userID)
	if err != nil {
		h.log.Error("Ошибка получения состояния пользователя", zap.Error(err))
		return c.Send("Возникла ошибка, повтори попытку позже.")
	}
	h.log.Info("User state retrieved", zap.String("state", state))

	switch state {
	case defs.WaitingModelName:
		modelName := c.Text()
		if modelName == "" {
			return c.Send("Имя модели не может быть пустым.")
		}

		models, err := h.service.GetUserModels(userID)
		if err != nil {
			h.log.Error("Ошибка получения моделей пользователя", zap.Error(err))
			return c.Send("Возникла ошибка, повтори попытку позже.")
		}

		for _, model := range models {
			if modelName == model {
				return c.Send("У тебя уже есть модель с таким именем.")
			}
		}

		err = h.service.SetPendingModel(userID, modelName)
		if err != nil {
			h.log.Error("Ошибка установки PendingModel", zap.Error(err))
			return c.Send("Возникла ошибка, повтори попытку позже.")
		}

		err = h.service.SetUserState(userID, defs.WaitingVoice)
		if err != nil {
			h.log.Error("Ошибка обновления состояния пользователя", zap.Error(err))
			return c.Send("Возникла ошибка, повтори попытку позже.")
		}

		h.log.Info("Пользователь ввёл имя новой модели", zap.String("modelName", modelName))
		return c.Send("Пришли голосовое сообщение для создания модели.")

	case defs.WaitingDeleteModelName:
		modelName := c.Text()
		if modelName == "" {
			return c.Send("Имя модели не может быть пустым.")
		}

		models, err := h.service.GetUserModels(userID)
		if err != nil {
			h.log.Error("Ошибка получения моделей пользователя для удаления", zap.Error(err))
			return c.Send("Возникла ошибка, повтори попытку позже.")
		}

		found := false
		for _, model := range models {
			if modelName == model {
				found = true
				break
			}
		}

		if !found {
			return c.Send("Модель с таким именем не найдена.")
		}

		err = h.service.DeleteModel(userID, modelName)
		if err != nil {
			h.log.Error("Ошибка удаления модели", zap.Error(err))
			return c.Send("Ошибка при удалении модели.")
		}

		h.log.Info("Модель успешно удалена", zap.String("modelName", modelName))
		_ = h.service.SetUserState(userID, defs.FreeState)
		return c.Send("Модель успешно удалена.")

	default:
		text := c.Text()
		audioRes, err := h.service.SendAudio(userID, text)
		if err != nil {
			if errors.Is(err, defs.ErrNoModel{}) {
				return c.Send("Создай модель /save_model или выбери из сохранённых /choose_model.")
			}
			h.log.Error("Ошибка генерации аудио", zap.Error(err))
			return c.Send("Возникла ошибка, повтори попытку позже.")
		}

		ogg, dur, err := EnsureVoiceNOTE(audioRes.Result.ProcessedAudio)
		if err != nil {
			h.log.Error("Ошибка перекодировки аудио", zap.Error(err))
			return c.Send("Возникла ошибка, повтори попытку позже.")
		}

		voiceMsg := &telebot.Voice{
			File: telebot.File{
				FileReader: bytes.NewReader(ogg),
			},
			MIME:     "audio/ogg",
			Duration: dur,
		}

		h.log.Info("Отправка голосового сообщения пользователю", zap.Int64("userID", userID))
		return c.Send(voiceMsg)
	}
}

func (h *Handler) HandleVoice(c telebot.Context) error {
	userID := c.Sender().ID
	h.log.Info("HandleVoice called", zap.Int64("userID", userID))

	state, err := h.service.GetUserState(userID)
	if err != nil {
		h.log.Error("Ошибка получения состояния пользователя", zap.Error(err))
		return c.Send("Возникла ошибка, повтори попытку позже.")
	}

	if state != defs.WaitingVoice {
		return c.Send("Я не жду голосовое сообщение. Сохрани модель через /save_model")
	}

	modelName, err := h.service.GetModelName(userID)
	if err != nil {
		h.log.Error("Ошибка получения имени модели", zap.Error(err))
		return c.Send("Имя модели не найдено.")
	}
	h.log.Info("Получено имя модели для сохранения", zap.String("modelName", modelName))

	fileInfo, err := c.Bot().FileByID(c.Message().Voice.FileID)
	if err != nil {
		h.log.Error("Ошибка получения файла по ID", zap.Error(err))
		return c.Send("Возникла ошибка, повтори попытку позже.")
	}

	err = h.service.SaveModel(userID, fileInfo.FilePath, c.Bot().Token, modelName)
	if err != nil {
		h.log.Error("Ошибка сохранения модели", zap.Error(err))
		return c.Send("Возникла ошибка, повтори попытку позже.")
	}

	err = h.service.SetUserState(userID, defs.FreeState)
	if err != nil {
		h.log.Error("Ошибка сброса состояния пользователя", zap.Error(err))
		return c.Send("Возникла ошибка, повтори попытку позже.")
	}

	h.log.Info("Модель успешно сохранена", zap.Int64("userID", userID))
	return c.Send("Модель успешно сохранена.")
}

func (h *Handler) GetModelName(c telebot.Context) error {
	userID := c.Sender().ID
	h.log.Info("GetModelName called", zap.Int64("userID", userID))

	command := c.Message().Text

	switch command {
	case "/save_model":
		count, err := h.service.CountModels(userID)
		if err != nil {
			h.log.Error("Ошибка подсчёта моделей", zap.Error(err))
			return c.Send("Возникла ошибка, повтори попытку позже.")
		}
		if count >= defs.MaxModels {
			return c.Send("Превышен лимит количества моделей.")
		}
		err = h.service.SetUserState(userID, defs.WaitingModelName)
		if err != nil {
			h.log.Error("Ошибка установки состояния ожидания имени модели", zap.Error(err))
			return err
		}
	case "/delete_model":
		err := h.service.SetUserState(userID, defs.WaitingDeleteModelName)
		if err != nil {
			h.log.Error("Ошибка установки состояния ожидания удаления модели", zap.Error(err))
			return err
		}
	default:
		h.log.Warn("Неизвестная команда в GetModelName", zap.String("command", command))
		return c.Send("Неизвестная команда.")
	}

	return c.Send("Введи имя модели:")
}

func (h *Handler) GetUserModels(c telebot.Context) error {
	userID := c.Sender().ID
	h.log.Info("GetUserModels called", zap.Int64("userID", userID))

	models, err := h.service.GetUserModels(userID)
	if err != nil {
		h.log.Error("Ошибка получения списка моделей", zap.Error(err))
		_ = c.Send("Ошибка при получении моделей.")
		return err
	}
	if len(models) == 0 {
		return c.Send("Пока нет сохранённых моделей.")
	}

	markup := &telebot.ReplyMarkup{}
	rows := make([]telebot.Row, 0, len(models))

	for _, name := range models {
		btn := markup.Data(name, "choose_model", name)
		rows = append(rows, markup.Row(btn))
	}
	markup.Inline(rows...)

	h.log.Info("Отправка списка моделей пользователю", zap.Int64("userID", userID))
	return c.Send("Выбери модель:", markup)
}

func (h *Handler) OnChooseModel(c telebot.Context) error {
	userID := c.Sender().ID
	data := strings.Split(c.Callback().Data, "|")
	if len(data) < 2 {
		h.log.Warn("Неверный формат callback данных", zap.String("data", c.Callback().Data))
		return c.Send("Некорректный выбор модели.")
	}
	modelName := data[1]

	err := h.service.SetPendingModel(userID, modelName)
	if err != nil {
		h.log.Error("Ошибка выбора модели", zap.Error(err))
		return c.Send("Возникла ошибка, повтори попытку позже.")
	}

	err = c.Respond(&telebot.CallbackResponse{
		Text:      "Выбрана модель: " + modelName,
		ShowAlert: false,
	})
	if err != nil {
		h.log.Error("Ошибка ответа на callback", zap.Error(err))
		return err
	}

	h.log.Info("Пользователь выбрал модель", zap.String("modelName", modelName), zap.Int64("userID", userID))
	return c.Send("Модель \"" + modelName + "\" выбрана.")
}

func (h *Handler) Start(c telebot.Context) error {
	text := `Привет! 👋

Я помогу тебе создавать голосовые модели и генерировать аудио.

Вот что ты можешь сделать:
  
/save_model — создать новую голосовую модель
/choose_model — выбрать одну из сохранённых моделей
/start — показать эту инструкцию ещё раз

⚡ Просто напиши мне текст — я озвучу его выбранной моделью!`
	return c.Send(text)
}

func EnsureVoiceNOTE(in []byte) (outBytes []byte, durationSec int, err error) {
	const ffprobeFmt = "default=noprint_wrappers=1:nokey=1"

	infoCmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_name,channels,sample_rate",
		"-of", ffprobeFmt,
		"pipe:0",
	)
	infoCmd.Stdin = bytes.NewReader(in)
	infoOut, _ := infoCmd.Output()

	needRecode := true
	if len(infoOut) > 0 {
		fields := strings.Split(strings.TrimSpace(string(infoOut)), "\n")
		if len(fields) == 3 &&
			fields[0] == "opus" &&
			fields[1] == "1" &&
			fields[2] == "48000" {
			needRecode = false
		}
	}

	var ogg bytes.Buffer
	if !needRecode {
		ogg.Write(in)
	} else {
		ff := exec.Command("ffmpeg",
			"-loglevel", "error",
			"-i", "pipe:0",
			"-c:a", "libopus",
			"-application", "voip",
			"-b:a", "64k",
			"-ac", "1",
			"-ar", "48000",
			"-f", "ogg",
			"pipe:1",
		)
		ff.Stdin = bytes.NewReader(in)
		ff.Stdout = &ogg

		if err = ff.Run(); err != nil {
			return nil, 0, fmt.Errorf("ffmpeg recode: %w", err)
		}
	}

	pr := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", ffprobeFmt,
		"pipe:0",
	)
	pr.Stdin = bytes.NewReader(ogg.Bytes())

	durOut, err := pr.Output()
	if err != nil {
		return nil, 0, errors.New("ffprobe: " + err.Error())
	}
	dur, _ := strconv.ParseFloat(strings.TrimSpace(string(durOut)), 64)

	return ogg.Bytes(), int(dur + 0.5), nil
}
