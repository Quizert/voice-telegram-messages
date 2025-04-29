package service

import (
	"fmt"
	"go.uber.org/zap"
	"io"
	"kursach/defs"
	pb "kursach/proto"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type AudioProcessorClient interface {
	SendAudio(text string, audioData []byte) (*pb.ProcessingResponse, error)
}

type Storage interface {
	IsUserExists(userId int64) (bool, error)
	AddUser(userId int64) error
	SaveModel(userID int64, modelName string) error
	GetUserModels(userID int64) ([]string, error)
	CountModels(userID int64) (int, error)
	DeleteModel(userID int64, modelName string) error
}

type Service struct {
	audioProcessorClient AudioProcessorClient
	storage              Storage
	log                  *zap.Logger
	userStates           map[int64]string
	pendingModels        map[int64]string
}

func NewService(client AudioProcessorClient, storage Storage, logger *zap.Logger) *Service {
	return &Service{
		audioProcessorClient: client,
		storage:              storage,
		log:                  logger,
		userStates:           make(map[int64]string),
		pendingModels:        make(map[int64]string),
	}
}

func (s *Service) SaveModel(userID int64, fileInfo string, token string, modelName string) error {
	s.log.Info("Сохранение новой модели", zap.Int64("userID", userID), zap.String("modelName", modelName))

	userDir := filepath.Join("voices", fmt.Sprintf("%d", userID))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		s.log.Error("Ошибка создания директории для модели", zap.String("path", userDir), zap.Error(err))
		return fmt.Errorf("не удалось создать директорию: %w", err)
	}

	filePath := filepath.Join(userDir, fmt.Sprintf("%s.ogg", modelName))
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", token, fileInfo)

	resp, err := http.Get(fileURL)
	if err != nil {
		s.log.Error("Ошибка загрузки файла с сервера Telegram", zap.Error(err))
		return fmt.Errorf("ошибка загрузки файла: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.log.Error("Сервер Telegram вернул ошибку", zap.Int("status_code", resp.StatusCode))
		return fmt.Errorf("telegram API вернул статус %d", resp.StatusCode)
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		s.log.Error("Ошибка создания локального файла модели", zap.String("path", filePath), zap.Error(err))
		return fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		s.log.Error("Ошибка записи содержимого файла", zap.String("path", filePath), zap.Error(err))
		return fmt.Errorf("ошибка записи файла: %w", err)
	}

	ok, err := s.storage.IsUserExists(userID)
	if err != nil {
		s.log.Error("Ошибка проверки существования пользователя", zap.Error(err))
		return err
	}
	if !ok {
		err = s.storage.AddUser(userID)
		if err != nil {
			s.log.Error("Ошибка добавления нового пользователя в БД", zap.Error(err))
			return err
		}
		s.log.Info("Пользователь успешно добавлен в БД", zap.Int64("userID", userID))
	}

	err = s.storage.SaveModel(userID, modelName)
	if err != nil {
		s.log.Error("Ошибка сохранения модели в БД", zap.Error(err))
		return err
	}

	s.log.Info("Модель успешно сохранена", zap.String("modelPath", filePath))
	return nil
}

func (s *Service) GetUserState(userID int64) (string, error) {
	state := s.userStates[userID]
	s.log.Debug("Получение состояния пользователя", zap.Int64("userID", userID), zap.String("state", state))
	return state, nil
}

func (s *Service) SetUserState(userID int64, state string) error {
	s.userStates[userID] = state
	s.log.Debug("Установка состояния пользователя", zap.Int64("userID", userID), zap.String("newState", state))
	return nil
}

func (s *Service) SetPendingModel(userID int64, name string) error {
	s.pendingModels[userID] = name
	s.log.Debug("Установка pending модели", zap.Int64("userID", userID), zap.String("modelName", name))
	return nil
}

func (s *Service) GetModelName(userID int64) (string, error) {
	modelName := s.pendingModels[userID]
	if modelName == "" {
		s.log.Warn("Модель не найдена у пользователя", zap.Int64("userID", userID))
		return "", fmt.Errorf("no model names: %w", defs.ErrNoModel{})
	}
	s.log.Debug("Получение имени pending модели", zap.Int64("userID", userID), zap.String("modelName", modelName))
	return modelName, nil
}

func (s *Service) SendAudio(userID int64, text string) (*pb.ProcessingResponse, error) {
	modelName, err := s.GetModelName(userID)
	if err != nil {
		s.log.Error("Ошибка получения модели для отправки аудио", zap.Error(err))
		return nil, err
	}

	modelPath := filepath.Join("voices", strconv.FormatInt(userID, 10), modelName+".ogg")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		s.log.Error("Файл модели не найден", zap.String("modelPath", modelPath))
		return nil, err
	}

	modelBytes, err := os.ReadFile(modelPath)
	if err != nil {
		s.log.Error("Ошибка чтения файла модели", zap.String("modelPath", modelPath), zap.Error(err))
		return nil, err
	}

	audio, err := s.audioProcessorClient.SendAudio(text, modelBytes)
	if err != nil {
		s.log.Error("Ошибка отправки аудио в AudioProcessor", zap.Error(err))
		return nil, err
	}

	s.log.Info("Успешно отправлено аудио на обработку", zap.Int64("userID", userID))
	return audio, nil
}

func (s *Service) GetUserModels(userID int64) ([]string, error) {
	models, err := s.storage.GetUserModels(userID)
	if err != nil {
		s.log.Error("Ошибка получения списка моделей пользователя", zap.Error(err))
		return nil, err
	}
	return models, nil
}

func (s *Service) CountModels(userID int64) (int, error) {
	count, err := s.storage.CountModels(userID)
	if err != nil {
		s.log.Error("Ошибка подсчёта моделей пользователя", zap.Error(err))
		return 0, err
	}
	return count, nil
}

func (s *Service) DeleteModel(userID int64, modelName string) error {
	return s.storage.DeleteModel(userID, modelName)
}
