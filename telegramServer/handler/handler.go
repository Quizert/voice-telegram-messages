package handler

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/prometheus/common/log"
	"gopkg.in/telebot.v3"
	pb "kursach/proto"
	"os"
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
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) HandleText(c telebot.Context) error {
	userID := c.Sender().ID

	state, err := h.service.GetUserState(userID)
	if err != nil {
		_ = c.Send("Возникла ошибка, повторите попытку позже.")
		return err
	}

	switch state {
	case "waiting_model_name":
		modelName := c.Text()

		//fileInfo, err := c.Bot().FileByID(c.Message().Voice.FileID)
		err = h.service.SetPendingModel(userID, modelName)
		if err != nil {
			_ = c.Send("Ошибка при сохранении имени модели.")
			return err
		}

		err = h.service.SetUserState(userID, "waiting_voice")
		if err != nil {
			_ = c.Send("Ошибка при сохранении имени модели.")
		}

		return c.Send("Пришлите голосовое сообщение для создания модели.")

	default:
		text := c.Text()
		audioRes, err := h.service.SendAudio(userID, text)
		if err != nil {
			return err
		}

		// Конвертация в OGG + Opus
		ogg, dur, err := EnsureVoiceNOTE(audioRes.Result.ProcessedAudio)
		if err != nil {
			return c.Send("Ошибка подготовки voice-файла.")
		}

		voiceMsg := &telebot.Voice{
			File: telebot.File{
				FileReader: bytes.NewReader(ogg),
			},
			MIME:     "audio/ogg",
			Duration: dur,
		}

		return c.Send(voiceMsg)
	}
}

func EnsureVoiceNOTE(in []byte) (outBytes []byte, durationSec int, err error) {
	const ffprobeFmt = "default=noprint_wrappers=1:nokey=1"

	// ── 1. Проверяем кодек / частоту ─────────────────────────────────────────────
	infoCmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_name,channels,sample_rate",
		"-of", ffprobeFmt,
		"pipe:0",
	)
	infoCmd.Stdin = bytes.NewReader(in)
	infoOut, _ := infoCmd.Output() // ошибки не критичны – просто перекодируем

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

	// ── 2. Либо используем bytes как есть, либо прогоняем через ffmpeg ───────────
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

	// ── 3. Считаем длительность ─────────────────────────────────────────────────
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

func (h *Handler) HandleVoice(c telebot.Context) error {
	userID := c.Sender().ID
	state, err := h.service.GetUserState(userID)
	if err != nil {
		_ = c.Send("Ошибка получения состояния.")
		return err
	}

	if state != "waiting_voice" {
		return c.Send("Я не жду голосовое сообщение. Сначала введите имя модели.")
	}

	modelName, err := h.service.GetModelName(userID)
	log.Info(modelName, userID)
	if err != nil {
		_ = c.Send("Ошибка: имя модели не найдено.")
		return err
	}

	fileInfo, err := c.Bot().FileByID(c.Message().Voice.FileID)
	if err != nil {
		_ = c.Send("Ошибка при получении файла.")
		return err
	}

	err = h.service.SaveModel(userID, fileInfo.FilePath, c.Bot().Token, modelName)
	if err != nil {
		_ = c.Send("Не удалось сохранить модель.")
		return err
	}

	//h.service.ClearUserState(userID)
	return c.Send("Модель успешно сохранена.")
}

// SendAudio Тут парсинг
func (h *Handler) SendAudio(c telebot.Context) error {
	text := c.Text()
	userID := c.Sender().ID
	audio, err := h.service.SendAudio(userID, text)
	if err != nil {
		return err
	}
	file, _ := os.Create("alen.ogg")

	_, err = file.Write(audio.Result.ProcessedAudio)
	if err != nil {
		return err
	}

	// Отправка результата
	voice := &telebot.Voice{
		File:     telebot.FromDisk("alen.ogg"),
		MIME:     "audio/ogg",
		Duration: 5,
	}

	return c.Send(voice)
}

func (h *Handler) GetModelName(c telebot.Context) error {
	userID := c.Sender().ID
	err := h.service.SetUserState(userID, "waiting_model_name")
	if err != nil {
		return err
	}
	return c.Send("Введите имя модели:")
}
