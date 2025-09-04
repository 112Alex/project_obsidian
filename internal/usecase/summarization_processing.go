package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/112Alex/project_obsidian/internal/domain/entity"
	"github.com/112Alex/project_obsidian/internal/domain/repository"
	"github.com/112Alex/project_obsidian/internal/domain/service"
	"github.com/112Alex/project_obsidian/pkg/logger"
)

// SummarizationProcessingUseCase представляет собой сценарий обработки суммаризации
type SummarizationProcessingUseCase struct {
	jobRepo             repository.JobRepository
	queueService        service.QueueService
	summarizationService service.SummarizationService
	telegramHandlers    *TelegramHandlersUseCase
	logger              *logger.Logger
}

// NewSummarizationProcessingUseCase создает новый сценарий обработки суммаризации
func NewSummarizationProcessingUseCase(
	jobRepo repository.JobRepository,
	queueService service.QueueService,
	summarizationService service.SummarizationService,
	telegramHandlers *TelegramHandlersUseCase,
	logger *logger.Logger,
) *SummarizationProcessingUseCase {
	return &SummarizationProcessingUseCase{
		jobRepo:             jobRepo,
		queueService:        queueService,
		summarizationService: summarizationService,
		telegramHandlers:    telegramHandlers,
		logger:              logger,
	}
}

// ProcessSummarization обрабатывает суммаризацию текста
func (uc *SummarizationProcessingUseCase) ProcessSummarization(ctx context.Context, job entity.QueueJob) error {
	// Получение данных из задачи
	payload, ok := job.Payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payload type in job")
	}

	transcription, ok := payload["transcription"].(string)
	if !ok {
		return fmt.Errorf("transcription not found in job payload or has invalid type")
	}

	// Логирование начала обработки суммаризации
	uc.logger.Info("Processing summarization",
		"job_id", job.JobID,
		"transcription_length", len(transcription),
	)

	// Суммаризация текста с использованием маркдаун форматирования
	summary, err := uc.summarizationService.SummarizeText(ctx, transcription)
	if err != nil {
		uc.logger.Error("Failed to summarize text",
			"error", err,
		)
		return fmt.Errorf("failed to summarize text: %w", err)
	}

	// Обновление задачи в базе данных
	err = uc.jobRepo.SetSummary(ctx, job.JobID, summary)
	if err != nil {
		uc.logger.Error("Failed to update job summary",
			"error", err,
		)
		return fmt.Errorf("failed to update job summary: %w", err)
	}

	// Суммаризация транскрипции
	summary, err := uc.summarizationService.Summarize(ctx, transcription)
	if err != nil {
		uc.logger.Error("Failed to summarize text",
			"error", err,
		)
		return fmt.Errorf("failed to summarize text: %w", err)
	}

	// Отправка обновления прогресса после суммаризации
	telegramID, message, err := uc.telegramHandlers.SendProgressUpdate(ctx, job.JobID, entity.JobStatusSummarized)
	if err == nil {
		uc.telegramHandlers.SendMessage(telegramID, message)
	}

	// Обновление статуса задачи
	err = uc.jobRepo.UpdateStatus(ctx, job.JobID, entity.JobStatusSummarized, "")
	if err != nil {
		uc.logger.Error("Failed to update job status",
			"error", err,
		)
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Логирование успешной обработки суммаризации
	uc.logger.Info("Summarization processed successfully",
		"job_id", job.JobID,
		"summary_length", len(summary),
	)

	return nil
}

// ProcessSummarizationWithBulletPoints обрабатывает суммаризацию текста с маркированным списком
func (uc *SummarizationProcessingUseCase) ProcessSummarizationWithBulletPoints(ctx context.Context, job entity.QueueJob) error {
	// Получение данных из задачи
	payload, ok := job.Payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payload type in job")
	}

	transcription, ok := payload["transcription"].(string)
	if !ok {
		return fmt.Errorf("transcription not found in job payload or has invalid type")
	}

	// Логирование начала обработки суммаризации с маркированным списком
	uc.logger.Info("Processing summarization with bullet points",
		"job_id", job.JobID,
		"transcription_length", len(transcription),
	)

	// Суммаризация текста с использованием маркированного списка
	summary, err := uc.summarizationService.SummarizeText(ctx, transcription)
	if err != nil {
		uc.logger.Error("Failed to summarize text with bullet points",
			"error", err,
		)
		return fmt.Errorf("failed to summarize text with bullet points: %w", err)
	}

	// Обновление задачи в базе данных
	err = uc.jobRepo.SetSummary(ctx, job.JobID, summary)
	if err != nil {
		uc.logger.Error("Failed to update job summary with bullet points",
			"error", err,
		)
		return fmt.Errorf("failed to update job summary with bullet points: %w", err)
	}

	// Отправка обновления прогресса перед интеграцией с Notion
	telegramID, message, err = uc.telegramHandlers.SendProgressUpdate(ctx, job.JobID, entity.JobStatusIntegrating)  // Предполагая, что есть статус для интеграции
	if err == nil {
		uc.telegramHandlers.SendMessage(telegramID, message)
	}

	// Обновление статуса задачи
	err = uc.jobRepo.UpdateStatus(ctx, job.JobID, entity.JobStatusSummarized, "")
	if err != nil {
		uc.logger.Error("Failed to update job status",
			"error", err,
		)
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Логирование успешной обработки суммаризации с маркированным списком
	uc.logger.Info("Summarization with bullet points processed successfully",
		"job_id", job.JobID,
		"summary_length", len(summary),
	)

	return nil
}
