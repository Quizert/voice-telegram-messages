package app

import (
	"fmt"
	"gopkg.in/telebot.v3"
	"kursach/client"
	"kursach/config"
	"kursach/service"
	"time"
)

type App struct {
	Bot     *telebot.Bot
	Service service.Service
}

func Init() (*App, error) {
	cfg := config.LoadConfig()
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  cfg.TelegramToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации бота: %w", err)
	}

	audioClient, err := client.NewAudioProcessorClient(cfg.GRPCHost, cfg.GRPCPort)
	if err != nil {
		return nil, fmt.Errorf("ошибка в создании клиента: %w", err)
	}

	svc := service.NewService(audioClient)
	app := &App{
		Bot:     bot,
		Service: *svc,
	}
	return app, nil
}

func (app *App) Start() {
	app.Bot.Handle(telebot.OnText, app.handleText())
	app.Bot.Handle(telebot.OnVoice, app.handleVoice())

	app.Bot.Start()
}
