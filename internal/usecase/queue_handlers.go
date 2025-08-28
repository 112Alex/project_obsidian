package usecase

import (
	"context"
	"fmt"

	"github.com/112Alex/project_obsidian/internal/domain/entity"
	"github.com/112Alex/project_obsidian/internal/domain/service"
	"github.com/112Alex/project_obsidian/pkg/logger"
)

// QueueHandlersUseCase представляет собой сценарий регистрации обработчиков задач в очереди
type QueueHandlersUseCase struct {
	queueService                   service.QueueService
	transcriptionProcessingUseCase *TranscriptionProcessingUseCase
	summarizationProcessingUseCase *SummarizationProcessingUseCase
	notionProcessingUseCase        *NotionProcessingUseCase
	telegramHandlersUseCase        *TelegramHandlersUseCase
	logger                         *logger.Logger
}

// NewQueueHandlersUseCase создает новый сценарий регистрации обработчиков задач в очереди
func NewQueueHandlersUseCase(
	queueService service.QueueService,
	transcriptionProcessingUseCase *TranscriptionProcessingUseCase,
	summarizationProcessingUseCase *SummarizationProcessingUseCase,
	notionProcessingUseCase *NotionProcessingUseCase,
	telegramHandlersUseCase *TelegramHandlersUseCase,
	logger *logger.Logger,
) *QueueHandlersUseCase {
	return &QueueHandlersUseCase{
		queueService:                   queueService,
		transcriptionProcessingUseCase: transcriptionProcessingUseCase,
		summarizationProcessingUseCase: summarizationProcessingUseCase,
		notionProcessingUseCase:        notionProcessingUseCase,
		telegramHandlersUseCase:        telegramHandlersUseCase,
		logger:                         logger,
	}
}

// RegisterHandlers регистрирует обработчики задач в очереди
func (uc *QueueHandlersUseCase) RegisterHandlers(ctx context.Context) error {
	// Логирование начала регистрации обработчиков
	uc.logger.Info("Registering queue handlers")

	// Регистрация обработчика для задач транскрибации
	uc.queueService.RegisterHandler(entity.JobTypeTranscription, func(ctx context.Context, job entity.QueueJob) error {
		return uc.transcriptionProcessingUseCase.ProcessTranscription(ctx, job)
	})

	// Регистрация обработчика для задач транскрибации с временными метками
	uc.queueService.RegisterHandler(entity.JobTypeTranscriptionWithTimestamps, func(ctx context.Context, job entity.QueueJob) error {
		return uc.transcriptionProcessingUseCase.ProcessTranscriptionWithTimestamps(ctx, job)
	})

	// Регистрация обработчика для задач суммаризации
	uc.queueService.RegisterHandler(entity.JobTypeSummarization, func(ctx context.Context, job entity.QueueJob) error {
		return uc.summarizationProcessingUseCase.ProcessSummarization(ctx, job)
	})

	// Регистрация обработчика для задач суммаризации с маркированным списком
	uc.queueService.RegisterHandler(entity.JobTypeSummarizationWithBulletPoints, func(ctx context.Context, job entity.QueueJob) error {
		return uc.summarizationProcessingUseCase.ProcessSummarizationWithBulletPoints(ctx, job)
	})

	// Регистрация обработчика для задач интеграции с Notion
	uc.queueService.RegisterHandler(entity.JobTypeNotion, func(ctx context.Context, job entity.QueueJob) error {
		return uc.notionProcessingUseCase.ProcessNotionIntegration(ctx, job)
	})

	// Регистрация обработчика для задач уведомления о завершении
	uc.queueService.RegisterHandler(entity.JobTypeNotification, func(ctx context.Context, job entity.QueueJob) error {
		// Отправка уведомления о завершении задачи
		jobIDStr := fmt.Sprintf("%d", job.JobID)
		telegramID, _, err := uc.telegramHandlersUseCase.SendJobCompletionNotification(ctx, jobIDStr)
		if err != nil {
			uc.logger.Error("Failed to send job completion notification",
				"error", err,
			)
			return err
		}

		// Здесь должна быть логика отправки сообщения пользователю через Telegram бота
		// Но так как у нас нет прямого доступа к боту из этого слоя, мы можем использовать
		// канал для отправки сообщений или другой механизм

		// Логирование успешной отправки уведомления
		uc.logger.Info("Successfully sent job completion notification",
			"job_id", job.JobID,
			"telegram_id", telegramID,
		)

		return nil
	})

	// Логирование успешной регистрации обработчиков
	uc.logger.Info("Successfully registered queue handlers")

	return nil
}

// StartWorker запускает обработчик задач из очереди
func (uc *QueueHandlersUseCase) StartWorker(ctx context.Context) error {
	// Логирование начала запуска обработчика задач
	uc.logger.Info("Starting queue worker")

	// Запуск обработчика задач
	err := uc.queueService.StartWorker(ctx)
	if err != nil {
		uc.logger.Error("Failed to start queue worker",
			"error", err,
		)
		return err
	}

	// Логирование успешного запуска обработчика задач
	uc.logger.Info("Successfully started queue worker")

	return nil
}
