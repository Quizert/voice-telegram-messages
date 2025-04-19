package handler

import (
	"gopkg.in/telebot.v3"
	pb "kursach/proto"
)

type Service interface {
	SendAudio(text string, audioData []byte) (*pb.ProcessingResponse, error)
	SaveModel(userID int64) error
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// SendAudio Тут парсинг
func (h *Handler) SendAudio(c telebot.Context) error {
	git
	text := c.Text()
	h.service.SendAudio(text, modelBytes)
	return nil
}

func (h *Handler) SaveModel(c telebot.Context) error {
	id := c.Sender().ID
	err := h.service.SaveModel(id)
	if err != nil {
		return err
	}

	err = c.Send("Введите имя модели:")
	if err != nil {
		return err
	}

	//msg, err := c.Bot().
	//if err != nil {
	//	return c.Send("Время ожидания истекло.")
	//}
	//
	//modelName := msg.Text

	err = c.Send("Модель успешно сохранена")
	if err != nil {
		return err
	}
	return nil
}
