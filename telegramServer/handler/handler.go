package handler

import (
	"gopkg.in/telebot.v3"
	pb "kursach/proto"
	"log"
)

type Service interface {
	SendAudio(text string, audioData []byte) (*pb.ProcessingResponse, error)
	SaveModel()
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) SendAudio(c telebot.Context) error {
	log.Println("gg")
	return nil
}

func (h *Handler) SaveAudio(c telebot.Context) error {
	return nil
}
