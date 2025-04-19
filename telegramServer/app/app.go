package app

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"kursach/client"
	"kursach/config"
	"kursach/handler"
	"kursach/service"
	"log"
	"time"
)

type App struct {
	Bot     *telebot.Bot
	Handler *handler.Handler
}

func NewApp() *App {
	return &App{}
}

func (a *App) Init() error {
	log.Println("Init App")
	cfg := config.LoadConfig()
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  cfg.TelegramToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	a.Bot = bot
	if err != nil {
		return fmt.Errorf("ошибка инициализации бота: %w", err)
	}

	audioClient, err := client.NewAudioProcessorClient(cfg.GRPCHost, cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("ошибка в создании клиента: %w", err)
	}

	svc := service.NewService(audioClient)
	controller := handler.NewHandler(svc)

	a.Handler = controller
	log.Println("Init success")
	return nil
}

func (a *App) Start() {
	log.Println("Start App")
	a.Bot.Handle(telebot.OnText, a.Handler.SendAudio)
	a.Bot.Handle(telebot.OnVoice, a.Handler.SaveModel)
	a.Bot.Start()
}
