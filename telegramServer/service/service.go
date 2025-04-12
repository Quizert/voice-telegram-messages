package service

import (
	pb "kursach/proto"
)

type AudioProcessorClient interface {
	SendAudio(text string, audioData []byte) (*pb.ProcessingResponse, error)
}

type Service struct {
	audioProcessorClient AudioProcessorClient
}

func (s *Service) SaveModel() {
	//TODO implement me
	panic("implement me")
}

func NewService(client AudioProcessorClient) *Service {
	return &Service{
		audioProcessorClient: client,
	}
}

func (s *Service) SendAudio(text string, audioData []byte) (*pb.ProcessingResponse, error) {
	return s.audioProcessorClient.SendAudio(text, audioData)
}
