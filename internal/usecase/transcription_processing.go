package usecase

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/112Alex/project_obsidian/internal/domain/entity"
	"github.com/112Alex/project_obsidian/internal/domain/repository"
	"github.com/112Alex/project_obsidian/internal/domain/service"
	"github.com/112Alex/project_obsidian/pkg/logger"
)

// TranscriptionProcessingUseCase представляет собой сценарий обработки транскрибации
type TranscriptionProcessingUseCase struct {
	jobRepo              repository.JobRepository
	queueService         service.QueueService
	audioService         service.AudioService
	transcriptionService service.TranscriptionService
	telegramHandlers     *TelegramHandlersUseCase
	logger               *logger.Logger
}

// NewTranscriptionProcessingUseCase создает новый сценарий обработки транскрибации
func NewTranscriptionProcessingUseCase(
	jobRepo repository.JobRepository,
	queueService service.QueueService,
	audioService service.AudioService,
	transcriptionService service.TranscriptionService,
	telegramHandlers *TelegramHandlersUseCase,
	logger *logger.Logger,
) *TranscriptionProcessingUseCase {
	return &TranscriptionProcessingUseCase{
		jobRepo:              jobRepo,
		queueService:         queueService,
		audioService:         audioService,
		transcriptionService: transcriptionService,
		telegramHandlers:     telegramHandlers,
		logger:               logger,
	}
}

// ProcessTranscription обрабатывает транскрибацию аудио файла
func (uc *TranscriptionProcessingUseCase) ProcessTranscription(ctx context.Context, job entity.QueueJob) error {
	// Получение данных из задачи
	payload, ok := job.Payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payload type in job")
	}

	audioPath, ok := payload["audio_path"].(string)
	if !ok {
		return fmt.Errorf("audio_path not found in job payload or has invalid type")
	}

	// Логирование начала обработки транскрибации
	uc.logger.Info("Processing transcription",
		"job_id", job.JobID,
		"audio_path", audioPath,
	)

	// Обработка аудио файла для транскрибации
	processedAudioPath, err := uc.audioService.ProcessAudio(ctx, audioPath, filepath.Base(audioPath))
	if err != nil {
		uc.logger.Error("Failed to process audio for transcription",
			"error", err,
		)
		return fmt.Errorf("failed to process audio for transcription: %w", err)
	}

	// Отправка обновления прогресса после обработки аудио
	telegramID, message, err := uc.telegramHandlers.SendProgressUpdate(ctx, job.JobID, entity.JobStatusProcessing)
	if err == nil {
		uc.telegramHandlers.SendMessage(telegramID, message)
	}

	// Транскрибация аудио файла
	transcription, err := uc.transcriptionService.Transcribe(ctx, processedAudioPath)
	if err != nil {
		uc.logger.Error("Failed to transcribe audio",
			"error", err,
		)
		return fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Отправка обновления прогресса после транскрипции
	telegramID, message, err = uc.telegramHandlers.SendProgressUpdate(ctx, job.JobID, entity.JobStatusTranscribed)
	if err == nil {
		uc.telegramHandlers.SendMessage(telegramID, message)
	}

	// Обновление задачи в базе данных
	err = uc.jobRepo.SetTranscription(ctx, job.JobID, transcription)
	if err != nil {
		uc.logger.Error("Failed to update job transcription",
			"error", err,
		)
		return fmt.Errorf("failed to update job transcription: %w", err)
	}

	// Создание задачи для суммаризации
	// Получаем user_id из payload
	payloadMap, _ := job.Payload.(map[string]interface{})
	userID, _ := payloadMap["user_id"].(int64)

	summarizationJob := entity.QueueJob{
		JobID:   job.JobID,
		JobType: entity.JobTypeSummarization,
		Payload: map[string]interface{}{
			"transcription": transcription,
			"user_id":       userID,
		},
	}

	// Добавление задачи в очередь
	err = uc.queueService.PushJob(ctx, summarizationJob)
	if err != nil {
		uc.logger.Error("Failed to push summarization job to queue",
			"error", err,
		)
		return fmt.Errorf("failed to push summarization job to queue: %w", err)
	}

	// Обновление статуса задачи
	err = uc.jobRepo.UpdateStatus(ctx, job.JobID, entity.JobStatusTranscribed, "")
	if err != nil {
		uc.logger.Error("Failed to update job status",
			"error", err,
		)
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Логирование успешной обработки транскрибации
	uc.logger.Info("Transcription processed successfully",
		"job_id", job.JobID,
		"transcription_length", len(transcription),
	)

	return nil
}

// ProcessTranscriptionWithTimestamps обрабатывает транскрибацию аудио файла с временными метками
func (uc *TranscriptionProcessingUseCase) ProcessTranscriptionWithTimestamps(ctx context.Context, job entity.QueueJob) error {
	// Получение данных из задачи
	payload, ok := job.Payload.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payload type in job")
	}

	audioPath, ok := payload["audio_path"].(string)
	if !ok {
		return fmt.Errorf("audio_path not found in job payload or has invalid type")
	}

	// Логирование начала обработки транскрибации с временными метками
	uc.logger.Info("Processing transcription with timestamps",
		"job_id", job.JobID,
		"audio_path", audioPath,
	)

	// Обработка аудио файла для транскрибации
	processedAudioPath, err := uc.audioService.ProcessAudio(ctx, audioPath, filepath.Base(audioPath))
	if err != nil {
		uc.logger.Error("Failed to process audio for transcription with timestamps",
			"error", err,
		)
		return fmt.Errorf("failed to process audio for transcription with timestamps: %w", err)
	}

	// Транскрибация аудио файла с временными метками
	// Используем обычный метод Transcribe, так как метод с временными метками не реализован
	transcription, err := uc.transcriptionService.Transcribe(ctx, processedAudioPath)
	if err != nil {
		uc.logger.Error("Failed to transcribe audio with timestamps",
			"error", err,
		)
		return fmt.Errorf("failed to transcribe audio with timestamps: %w", err)
	}

	// Обновление задачи в базе данных
	err = uc.jobRepo.SetTranscription(ctx, job.JobID, transcription)
	if err != nil {
		uc.logger.Error("Failed to update job transcription with timestamps",
			"error", err,
		)
		return fmt.Errorf("failed to update job transcription with timestamps: %w", err)
	}

	// Обновление статуса задачи
	err = uc.jobRepo.UpdateStatus(ctx, job.JobID, entity.JobStatusTranscribed, "")
	if err != nil {
		uc.logger.Error("Failed to update job status",
			"error", err,
		)
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Логирование успешной обработки транскрибации с временными метками
	uc.logger.Info("Transcription with timestamps processed successfully",
		"job_id", job.JobID,
		"transcription_length", len(transcription),
	)

	return nil
}
