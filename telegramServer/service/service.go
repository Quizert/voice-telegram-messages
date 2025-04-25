package service

import (
	"fmt"
	"io"
	pb "kursach/proto"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type AudioProcessorClient interface {
	SendAudio(text string, audioData []byte) (*pb.ProcessingResponse, error)
}

type Service struct {
	audioProcessorClient AudioProcessorClient
	userStates           map[int64]string
	pendingModels        map[int64]string
}

func NewService(client AudioProcessorClient) *Service {
	return &Service{
		audioProcessorClient: client,
		userStates:           make(map[int64]string),
		pendingModels:        make(map[int64]string),
	}
}

func (s *Service) SaveModel(userID int64, fileInfo string, token string, modelName string) error {
	userDir := filepath.Join("voices", fmt.Sprintf("%d", userID))

	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию: %w", err)
	}

	filePath := filepath.Join(userDir, fmt.Sprintf("%s.ogg", modelName))

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", token, fileInfo)
	log.Println(filePath, "\n", userDir, modelName)
	resp, err := http.Get(fileURL)
	if err != nil {
		return fmt.Errorf("ошибка загрузки файла: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API вернул статус %d", resp.StatusCode)
	}

	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer outFile.Close()

	n, err := io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка записи файла: %w", err)
	}
	log.Printf("Записано байт: %d\n", n)
	return nil
}

func (s *Service) GetUserState(userID int64) (string, error) {
	state := s.userStates[userID]
	return state, nil
}

func (s *Service) SetUserState(userID int64, state string) error {
	s.userStates[userID] = state
	return nil
}

func (s *Service) SetPendingModel(userID int64, name string) error {
	s.pendingModels[userID] = name
	return nil
}

func (s *Service) GetModelName(userID int64) (string, error) {
	modelName, ok := s.pendingModels[userID]
	if !ok {
		return "", fmt.Errorf("no model named: %d", userID)
	}
	return modelName, nil
}

func (s *Service) SendAudio(userID int64, text string) (*pb.ProcessingResponse, error) {
	modelName, err := s.GetModelName(userID)
	fmt.Println(modelName)
	if err != nil {
		return nil, err
	}
	modelPath := filepath.Join("voices", strconv.FormatInt(userID, 10), modelName+".ogg")
	fmt.Println(modelPath)
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return nil, err
	}

	modelBytes, err := os.ReadFile(modelPath)
	audio, err := s.audioProcessorClient.SendAudio(text, modelBytes)
	if err != nil {
		return nil, err
	}

	return audio, nil
}
