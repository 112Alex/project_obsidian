package usecase

import (
	"context"

	"github.com/112Alex/project_obsidian/internal/config"
	"github.com/112Alex/project_obsidian/internal/domain/repository"
	"github.com/112Alex/project_obsidian/internal/domain/service"
	"github.com/112Alex/project_obsidian/pkg/logger"
)

// App представляет собой приложение
type App struct {
	Config                         *config.Config
	Logger                         *logger.Logger
	UserRepo                       repository.UserRepository
	JobRepo                        repository.JobRepository
	QueueRepo                      repository.QueueRepository
	AudioService                   service.AudioService
	TranscriptionService           service.TranscriptionService
	SummarizationService           service.SummarizationService
	NotionService                  service.NotionService
	QueueService                   service.QueueService
	AudioProcessingUseCase         *AudioProcessingUseCase
	TranscriptionProcessingUseCase *TranscriptionProcessingUseCase
	SummarizationProcessingUseCase *SummarizationProcessingUseCase
	NotionProcessingUseCase        *NotionProcessingUseCase
	TelegramHandlersUseCase        *TelegramHandlersUseCase
	QueueHandlersUseCase           *QueueHandlersUseCase
}

// NewApp создает новое приложение
func NewApp(
	config *config.Config,
	logger *logger.Logger,
	userRepo repository.UserRepository,
	jobRepo repository.JobRepository,
	queueRepo repository.QueueRepository,
	audioService service.AudioService,
	transcriptionService service.TranscriptionService,
	summarizationService service.SummarizationService,
	notionService service.NotionService,
	queueService service.QueueService,
) *App {
	// Создание сценария обработки аудио
	audioProcessingUseCase := NewAudioProcessingUseCase(
		userRepo,
		jobRepo,
		queueService,
		audioService,
		logger,
	)

	// Создание сценария обработки транскрибации
	transcriptionProcessingUseCase := NewTranscriptionProcessingUseCase(
		jobRepo,
		queueService,
		audioService,
		transcriptionService,
		logger,
	)

	// Создание сценария обработки суммаризации
	summarizationProcessingUseCase := NewSummarizationProcessingUseCase(
		jobRepo,
		queueService,
		summarizationService,
		notionService,
		logger,
	)

	// Создание сценария обработки интеграции с Notion
	notionProcessingUseCase := NewNotionProcessingUseCase(
		jobRepo,
		userRepo,
		notionService,
		logger,
	)

	// Создание сценария обработки команд Telegram бота
	telegramHandlersUseCase := NewTelegramHandlersUseCase(
		userRepo,
		jobRepo,
		audioProcessingUseCase,
		notionProcessingUseCase,
		logger,
	)

	// Создание сценария регистрации обработчиков задач в очереди
	queueHandlersUseCase := NewQueueHandlersUseCase(
		queueService,
		transcriptionProcessingUseCase,
		summarizationProcessingUseCase,
		notionProcessingUseCase,
		telegramHandlersUseCase,
		logger,
	)

	return &App{
		Config:                         config,
		Logger:                         logger,
		UserRepo:                       userRepo,
		JobRepo:                        jobRepo,
		QueueRepo:                      queueRepo,
		AudioService:                   audioService,
		TranscriptionService:           transcriptionService,
		SummarizationService:           summarizationService,
		NotionService:                  notionService,
		QueueService:                   queueService,
		AudioProcessingUseCase:         audioProcessingUseCase,
		TranscriptionProcessingUseCase: transcriptionProcessingUseCase,
		SummarizationProcessingUseCase: summarizationProcessingUseCase,
		NotionProcessingUseCase:        notionProcessingUseCase,
		TelegramHandlersUseCase:        telegramHandlersUseCase,
		QueueHandlersUseCase:           queueHandlersUseCase,
	}
}

// Start запускает приложение
func (a *App) Start(ctx context.Context) error {
	// Регистрируем обработчики задач в очереди
	if err := a.QueueHandlersUseCase.RegisterHandlers(ctx); err != nil {
		return err
	}

	// Запускаем воркер очереди
	if err := a.QueueHandlersUseCase.StartWorker(ctx); err != nil {
		return err
	}

	return nil
}

// Stop останавливает приложение
func (a *App) Stop(ctx context.Context) error {
	// Логирование начала остановки приложения
	a.Logger.Info("Stopping application")

	// Здесь должна быть логика остановки приложения
	// Например, остановка обработчика задач из очереди

	// Логирование успешной остановки приложения
	a.Logger.Info("Application stopped successfully")

	return nil
}
