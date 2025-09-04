package infrastructure

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/112Alex/project_obsidian/internal/config"
	"github.com/112Alex/project_obsidian/internal/infrastructure/database"
	"github.com/112Alex/project_obsidian/internal/infrastructure/deepseek"
	"github.com/112Alex/project_obsidian/internal/infrastructure/ffmpeg"
	"github.com/112Alex/project_obsidian/internal/infrastructure/notion"
	"github.com/112Alex/project_obsidian/internal/infrastructure/openai"
	"github.com/112Alex/project_obsidian/internal/infrastructure/queue"
	"github.com/112Alex/project_obsidian/internal/infrastructure/telegram"
	"github.com/112Alex/project_obsidian/internal/usecase"
	"github.com/112Alex/project_obsidian/pkg/logger"
)

// App представляет собой приложение
type App struct {
	Config      *config.Config
	Logger      *logger.Logger
	PostgresDB  *database.PostgresDB
	RedisClient *database.RedisClient
	Bot         *telegram.Bot
	UseCase     *usecase.App
}

// NewApp создает новое приложение
func NewApp(config *config.Config, logger *logger.Logger) (*App, error) {
	// Инициализация PostgreSQL
	postgresDB, err := database.NewPostgresDB(context.Background(), config.Postgres)
	if err != nil {
		logger.Error("Failed to initialize PostgreSQL",
			"error", err,
		)
		return nil, err
	}

	// Инициализация Redis
	redisClient, err := database.NewRedisClient(context.Background(), config.Redis)
	if err != nil {
		logger.Error("Failed to initialize Redis",
			"error", err,
		)
		return nil, err
	}

	// Инициализация репозиториев
	userRepo := database.NewUserRepository(postgresDB)
	jobRepo := database.NewJobRepository(postgresDB)
	queueRepo := database.NewQueueRepository(redisClient)

	// Инициализация сервисов
	audioService := ffmpeg.NewAudioService(config.FFmpeg.BinaryPath, logger)
	transcriptionService := openai.NewTranscriptionService(config.OpenAI.APIKey, config.OpenAI.WhisperModel, logger)
	summarizationService := deepseek.NewSummarizationService(config.DeepSeek.APIKey, "", config.DeepSeek.Model, logger)
	notionService := notion.NewNotionService(config.Notion.APIKey, logger)
	queueService := queue.NewQueueService(queueRepo, jobRepo, logger)

	// Инициализация слоя usecase
	useCaseApp := usecase.NewApp(
		config,
		logger,
		userRepo,
		jobRepo,
		queueRepo,
		audioService,
		transcriptionService,
		summarizationService,
		notionService,
		queueService,
	)

	// Инициализация Telegram бота
	// Инициализация Telegram бота
	bot, err := telegram.NewBot(config.Telegram.Token, logger)
	if err != nil {
		logger.Error("Failed to initialize Telegram bot",
			"error", err,
		)
		return nil, err
	}

	return &App{
		Config:      config,
		Logger:      logger,
		PostgresDB:  postgresDB,
		RedisClient: redisClient,
		Bot:         bot,
		UseCase:     useCaseApp,
	}, nil
}

// Start запускает приложение
func (a *App) Start(ctx context.Context) error {
	// Логирование начала запуска приложения
	a.Logger.Info("Starting application")

	// Запуск слоя usecase
	err := a.UseCase.Start(ctx)
	if err != nil {
		a.Logger.Error("Failed to start usecase layer",
			"error", err,
		)
		return err
	}

	// Регистрация обработчиков команд Telegram
	a.Bot.RegisterCommandHandler("start", func(ctx context.Context, m *tgbotapi.Message) error {
		resp, err := a.UseCase.TelegramHandlersUseCase.HandleStart(ctx, m.Chat.ID, m.From.UserName)
		if err != nil {
			return err
		}
		_, err = a.Bot.SendMarkdownMessage(m.Chat.ID, resp)
		return err
	})

	a.Bot.RegisterCommandHandler("help", func(ctx context.Context, m *tgbotapi.Message) error {
		resp, err := a.UseCase.TelegramHandlersUseCase.HandleHelp(ctx, m.Chat.ID)
		if err != nil {
			return err
		}
		_, err = a.Bot.SendMarkdownMessage(m.Chat.ID, resp)
		return err
	})

	a.Bot.RegisterCommandHandler("notion", func(ctx context.Context, m *tgbotapi.Message) error {
		args := strings.TrimSpace(m.CommandArguments())
		resp, err := a.UseCase.TelegramHandlersUseCase.HandleNotion(ctx, m.Chat.ID, args)
		if err != nil {
			return err
		}
		_, err = a.Bot.SendMarkdownMessage(m.Chat.ID, resp)
		return err
	})

	a.Bot.RegisterCommandHandler("jobs", func(ctx context.Context, m *tgbotapi.Message) error {
		resp, err := a.UseCase.TelegramHandlersUseCase.HandleJobs(ctx, m.Chat.ID)
		if err != nil {
			return err
		}
		_, err = a.Bot.SendMarkdownMessage(m.Chat.ID, resp)
		return err
	})

	// Регистрация обработчика аудио и голосовых сообщений
	a.Bot.RegisterAudioHandler(func(ctx context.Context, m *tgbotapi.Message, filePath string, fileName string) error {
		// Определяем тип сообщения и вызываем соответствующий usecase
		var err error
		if m.Voice != nil {
			_, err = a.UseCase.TelegramHandlersUseCase.HandleVoiceMessage(ctx, m.Chat.ID, m.From.UserName, m.Voice.FileID, filePath, fileName)
		} else if m.Audio != nil {
			_, err = a.UseCase.TelegramHandlersUseCase.HandleAudioFile(ctx, m.Chat.ID, m.From.UserName, m.Audio.FileID, filePath, fileName)
		}
		return err
	})

	// Запуск Telegram бота
	err = a.Bot.Start()
	if err != nil {
		a.Logger.Error("Failed to start Telegram bot",
			"error", err,
		)
		return err
	}
	if err != nil {
		a.Logger.Error("Failed to start Telegram bot",
			"error", err,
		)
		return err
	}

	// Логирование успешного запуска приложения
	a.Logger.Info("Application started successfully")

	return nil
}

// Stop останавливает приложение
func (a *App) Stop(ctx context.Context) error {
	// Логирование начала остановки приложения
	a.Logger.Info("Stopping application")

	// Остановка Telegram бота
	a.Bot.Stop()

	// Остановка слоя usecase
	err := a.UseCase.Stop(ctx)
	if err != nil {
		a.Logger.Error("Failed to stop usecase layer",
			"error", err,
		)
		return err
	}

	// Закрытие соединения с Redis
	a.RedisClient.Close()

	// Закрытие соединения с PostgreSQL
	a.PostgresDB.Close()

	// Логирование успешной остановки приложения
	a.Logger.Info("Application stopped successfully")

	return nil
}
