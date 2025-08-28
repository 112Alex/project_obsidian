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

// AudioProcessingUseCase представляет собой сценарий обработки аудио
type AudioProcessingUseCase struct {
	userRepo     repository.UserRepository
	jobRepo      repository.JobRepository
	queueService service.QueueService
	audioService service.AudioService
	logger       *logger.Logger
}

// NewAudioProcessingUseCase создает новый сценарий обработки аудио
func NewAudioProcessingUseCase(
	userRepo repository.UserRepository,
	jobRepo repository.JobRepository,
	queueService service.QueueService,
	audioService service.AudioService,
	logger *logger.Logger,
) *AudioProcessingUseCase {
	return &AudioProcessingUseCase{
		userRepo:     userRepo,
		jobRepo:      jobRepo,
		queueService: queueService,
		audioService: audioService,
		logger:       logger,
	}
}

// ProcessAudio обрабатывает аудио файл
func (uc *AudioProcessingUseCase) ProcessAudio(ctx context.Context, userID int64, audioPath string, fileName string) (int64, error) {
	// Логирование начала обработки аудио
	uc.logger.Info("Processing audio",
		"user_id", userID,
		"audio_path", audioPath,
		"file_name", fileName,
	)

	// Получение пользователя
	user, err := uc.userRepo.GetByTelegramID(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to get user",
			"error", err,
		)
		return 0, fmt.Errorf("failed to get user: %w", err)
	}

	// Если пользователь не найден, создаем нового
	if user == nil {
		user = &entity.User{
			TelegramID: userID,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		err := uc.userRepo.Create(ctx, user)
		if err != nil {
			uc.logger.Error("Failed to create user",
				"error", err,
			)
			return 0, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Получение длительности аудио
	duration, err := uc.audioService.GetAudioDuration(ctx, audioPath)
	if err != nil {
		uc.logger.Error("Failed to get audio duration",
			"error", err,
		)
		return 0, fmt.Errorf("failed to get audio duration: %w", err)
	}

	// Создание задачи
	job := entity.Job{
		UserID:        user.ID,
		Type:          entity.JobTypeTranscription,
		Status:        entity.JobStatusCreated,
		AudioFilePath: audioPath,
		FileName:      fileName,
		Duration:      duration,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Сохранение задачи в базе данных
	job.ID = 0 // Убедимся, что ID не задан
	err = uc.jobRepo.Create(ctx, &job)
	if err != nil {
		uc.logger.Error("Failed to create job",
			"error", err,
		)
		return 0, fmt.Errorf("failed to create job: %w", err)
	}
	jobID := job.ID

	// Создание задачи для очереди - используем напрямую EnqueueTranscriptionJob

	// Добавление задачи в очередь
	err = uc.queueService.EnqueueTranscriptionJob(ctx, jobID, user.ID, audioPath)
	if err != nil {
		uc.logger.Error("Failed to push job to queue",
			"error", err,
		)
		return 0, fmt.Errorf("failed to push job to queue: %w", err)
	}

	// Логирование успешной обработки аудио
	uc.logger.Info("Audio processed successfully",
		"job_id", jobID,
	)

	return jobID, nil
}

// GetJobStatus возвращает статус задачи
func (uc *AudioProcessingUseCase) GetJobStatus(ctx context.Context, jobID int64) (entity.JobStatus, error) {
	// Получение задачи
	job, err := uc.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		uc.logger.Error("Failed to get job",
			"error", err,
		)
		return "", fmt.Errorf("failed to get job: %w", err)
	}

	// Если задача не найдена, возвращаем ошибку
	if job == nil {
		return "", fmt.Errorf("job not found")
	}

	return job.Status, nil
}

// GetJobResult возвращает результат задачи
func (uc *AudioProcessingUseCase) GetJobResult(ctx context.Context, jobID int64) (*entity.Job, error) {
	// Получение задачи
	job, err := uc.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		uc.logger.Error("Failed to get job",
			"error", err,
		)
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Если задача не найдена, возвращаем ошибку
	if job == nil {
		return nil, fmt.Errorf("job not found")
	}

	return job, nil
}

// GetUserJobs возвращает задачи пользователя
func (uc *AudioProcessingUseCase) GetUserJobs(ctx context.Context, telegramID int64) ([]*entity.Job, error) {
	// Получение пользователя
	user, err := uc.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		uc.logger.Error("Failed to get user",
			"error", err,
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Если пользователь не найден, возвращаем ошибку
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Получение задач пользователя
	jobs, err := uc.jobRepo.GetByUserID(ctx, user.ID, 100, 0) // Лимит 100, смещение 0
	if err != nil {
		uc.logger.Error("Failed to get user jobs",
			"error", err,
		)
		return nil, fmt.Errorf("failed to get user jobs: %w", err)
	}

	return jobs, nil
}
