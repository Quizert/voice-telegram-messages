package app

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
	"kursach/client"
	"kursach/config"
	"kursach/handler"
	"kursach/service"
	"kursach/storage"
	"time"
)

type App struct {
	Bot     *telebot.Bot
	Handler *handler.Handler
	dbPool  *pgxpool.Pool
	log     *zap.Logger
}

func setCommands(b *telebot.Bot, logger *zap.Logger) {
	commands := []telebot.Command{
		{Text: "/save_model", Description: "Отправить голосовое сообщения для генерации модели."},
		{Text: "/delete_model", Description: "Удалить модель"},
		{Text: "/choose_model", Description: "Выбрать модель для генерации."},
		{Text: "/start", Description: "Старт"},
	}

	err := b.SetCommands(commands)
	if err != nil {
		logger.Warn("Ошибка при установке команд", zap.Error(err))
	}
}

func NewApp() *App {
	return &App{}
}

func (a *App) Init() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("ошибка создания логгера: %w", err)
	}
	a.log = logger
	a.log.Info("Инициализация приложения...")

	cfg := config.LoadConfig()

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  cfg.TelegramToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return fmt.Errorf("ошибка инициализации бота: %w", err)
	}
	a.Bot = bot
	a.log.Info("Бот успешно создан")

	audioClient, err := client.NewAudioProcessorClient(cfg.GRPCHost, cfg.GRPCPort)
	if err != nil {
		return fmt.Errorf("ошибка создания gRPC клиента: %w", err)
	}
	a.log.Info("gRPC клиент успешно подключён")

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	a.log.Info("Подключение к базе данных...", zap.String("connString", connString))

	dbPool, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %w", err)
	}
	a.dbPool = dbPool
	a.log.Info("Подключение к базе данных успешно")

	postgres := storage.NewPostgresStorage(dbPool, logger)
	svc := service.NewService(audioClient, postgres, logger)
	controller := handler.NewHandler(svc, logger)

	a.Handler = controller
	a.log.Info("Инициализация компонентов приложения завершена")

	return nil
}

func (a *App) Start() {
	a.log.Info("Запуск приложения...")
	setCommands(a.Bot, a.log)

	a.Bot.Handle("/save_model", a.Handler.GetModelName)
	a.Bot.Handle("/delete_model", a.Handler.GetModelName)
	a.Bot.Handle("/choose_model", a.Handler.GetUserModels)
	a.Bot.Handle("/start", a.Handler.Start)

	a.Bot.Handle(telebot.OnText, a.Handler.HandleText)
	a.Bot.Handle(telebot.OnVoice, a.Handler.HandleVoice)
	a.Bot.Handle(telebot.OnCallback, a.Handler.OnChooseModel)

	a.log.Info("Бот готов к работе")
	a.Bot.Start()
}
