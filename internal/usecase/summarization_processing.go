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
	jobRepo              repository.JobRepository
	queueService         service.QueueService
	summarizationService service.SummarizationService
	notionService        service.NotionService
	logger               *logger.Logger
}

// NewSummarizationProcessingUseCase создает новый сценарий обработки суммаризации
func NewSummarizationProcessingUseCase(
	jobRepo repository.JobRepository,
	queueService service.QueueService,
	summarizationService service.SummarizationService,
	notionService service.NotionService,
	logger *logger.Logger,
) *SummarizationProcessingUseCase {
	return &SummarizationProcessingUseCase{
		jobRepo:              jobRepo,
		queueService:         queueService,
		summarizationService: summarizationService,
		notionService:        notionService,
		logger:               logger,
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

	// Создание задачи для интеграции с Notion
	notionJob := entity.QueueJob{
		JobID:     job.JobID,
		UserID:    job.UserID,
		JobType:   entity.JobTypeNotion,
		CreatedAt: time.Now(),
		Payload: map[string]interface{}{
			"transcription": transcription,
			"summary":       summary,
		},
	}

	// Добавление задачи в очередь
	err = uc.queueService.PushJob(ctx, notionJob)
	if err != nil {
		uc.logger.Error("Failed to push notion job to queue",
			"error", err,
		)
		return fmt.Errorf("failed to push notion job to queue: %w", err)
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
