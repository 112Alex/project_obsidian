package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/112Alex/project_obsidian/internal/domain/entity"
	"github.com/112Alex/project_obsidian/internal/domain/repository"
	"github.com/112Alex/project_obsidian/pkg/logger"
)

// DefaultQueueName - имя очереди по умолчанию
const DefaultQueueName = "default"

// QueueService представляет собой сервис для работы с очередью задач
type QueueService struct {
	queueRepo repository.QueueRepository
	jobRepo   repository.JobRepository
	logger    *logger.Logger
	worker    *Worker
}

// NewQueueService создает новый сервис для работы с очередью задач
func NewQueueService(
	queueRepo repository.QueueRepository,
	jobRepo repository.JobRepository,
	logger *logger.Logger,
) *QueueService {
	s := &QueueService{
		queueRepo: queueRepo,
		jobRepo:   jobRepo,
		logger:    logger,
	}
	s.worker = NewWorker(s, logger)
	return s
}

// PushJob добавляет задачу в очередь
func (s *QueueService) PushJob(ctx context.Context, job entity.QueueJob) error {
	// Логирование начала добавления задачи
	s.logger.Info("Pushing job to queue",
		"job_id", job.JobID,
		"job_type", job.JobType,
	)

	// Определение имени очереди на основе типа задачи
	queueName := string(job.JobType)

	// Добавление задачи в очередь
	err := s.queueRepo.Push(ctx, queueName, &job)
	if err != nil {
		s.logger.Error("Failed to push job to queue",
			"error", err,
		)
		return fmt.Errorf("failed to push job to queue: %w", err)
	}

	// Обновление статуса задачи в базе данных
	err = s.jobRepo.UpdateStatus(ctx, job.JobID, entity.JobStatusQueued, "")
	if err != nil {
		s.logger.Error("Failed to update job status",
			"error", err,
		)
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Логирование успешного добавления задачи
	s.logger.Info("Job pushed to queue successfully",
		"job_id", job.JobID,
	)

	return nil
}

// PopJob извлекает задачу из очереди
func (s *QueueService) PopJob(ctx context.Context, queueName string) (*entity.QueueJob, error) {
	// Извлечение задачи из очереди
	job, err := s.queueRepo.Pop(ctx, queueName)
	if err != nil {
		return nil, fmt.Errorf("failed to pop job from queue: %w", err)
	}

	// Если очередь пуста, возвращаем nil
	if job == nil {
		return nil, nil
	}

	// Логирование извлечения задачи
	s.logger.Info("Popped job from queue",
		"job_id", job.JobID,
		"job_type", job.JobType,
	)

	// Обновление статуса задачи в базе данных
	err = s.jobRepo.UpdateStatus(ctx, job.JobID, entity.JobStatusProcessing, "")
	if err != nil {
		s.logger.Error("Failed to update job status",
			"error", err,
		)
		return job, fmt.Errorf("failed to update job status: %w", err)
	}

	return job, nil
}

// GetQueueSize возвращает размер очереди
func (s *QueueService) GetQueueSize(ctx context.Context) (int64, error) {
	// Получение размера очереди для очереди по умолчанию
	size, err := s.queueRepo.Size(ctx, DefaultQueueName)
	if err != nil {
		s.logger.Error("Failed to get queue size",
			"error", err,
		)
		return 0, fmt.Errorf("failed to get queue size: %w", err)
	}

	return size, nil
}

// EnqueueTranscriptionJob добавляет задачу транскрибации в очередь
func (s *QueueService) EnqueueTranscriptionJob(ctx context.Context, jobID, userID int64, audioFilePath string) error {
	job := entity.QueueJob{
		JobID:     jobID,
		UserID:    userID,
		JobType:   entity.JobTypeTranscription,
		CreatedAt: time.Now(),
		Payload:   audioFilePath,
	}
	return s.PushJob(ctx, job)
}

// EnqueueSummarizationJob добавляет задачу суммаризации в очередь
func (s *QueueService) EnqueueSummarizationJob(ctx context.Context, jobID, userID int64, transcription string) error {
	job := entity.QueueJob{
		JobID:     jobID,
		UserID:    userID,
		JobType:   entity.JobTypeSummarization,
		CreatedAt: time.Now(),
		Payload:   transcription,
	}
	return s.PushJob(ctx, job)
}

// EnqueueNotionSyncJob добавляет задачу синхронизации с Notion в очередь
func (s *QueueService) EnqueueNotionSyncJob(ctx context.Context, jobID, userID int64, title, content string) error {
	payload := map[string]string{"title": title, "content": content}
	job := entity.QueueJob{
		JobID:     jobID,
		UserID:    userID,
		JobType:   entity.JobTypeNotionSync,
		CreatedAt: time.Now(),
		Payload:   payload,
	}
	return s.PushJob(ctx, job)
}

// RegisterHandler регистрирует обработчик для определенного типа задач
func (s *QueueService) RegisterHandler(jobType entity.JobType, handler func(ctx context.Context, job entity.QueueJob) error) {
	if s.worker == nil {
		s.worker = NewWorker(s, s.logger)
	}
	s.worker.RegisterHandler(jobType, handler)
}

// StartWorker запускает обработчик задач из очереди
func (s *QueueService) StartWorker(ctx context.Context) error {
	if s.worker == nil {
		s.worker = NewWorker(s, s.logger)
	}
	s.worker.Start(ctx)
	return nil
}

// Worker представляет собой воркер для обработки задач из очереди
type Worker struct {
	queueService *QueueService
	handlers     map[entity.JobType]JobHandler
	logger       *logger.Logger
	shutdown     chan struct{}
}

// JobHandler представляет собой обработчик задачи
type JobHandler func(ctx context.Context, job entity.QueueJob) error

// NewWorker создает нового воркера для обработки задач
func NewWorker(queueService *QueueService, logger *logger.Logger) *Worker {
	return &Worker{
		queueService: queueService,
		handlers:     make(map[entity.JobType]JobHandler),
		logger:       logger,
		shutdown:     make(chan struct{}),
	}
}

// RegisterHandler регистрирует обработчик для типа задачи
func (w *Worker) RegisterHandler(jobType entity.JobType, handler JobHandler) {
	w.handlers[jobType] = handler
}

// Start запускает воркер
func (w *Worker) Start(ctx context.Context) {
	w.logger.Info("Starting worker")

	go func() {
		for {
			select {
			case <-ctx.Done():
				w.logger.Info("Worker stopped due to context cancellation")
				return
			case <-w.shutdown:
				w.logger.Info("Worker stopped due to shutdown signal")
				return
			default:
				// Извлечение задачи из очереди для очереди по умолчанию
				job, err := w.queueService.PopJob(ctx, DefaultQueueName)
				if err != nil {
					w.logger.Error("Failed to pop job from queue",
						"error", err,
					)
					time.Sleep(1 * time.Second)
					continue
				}

				// Если очередь пуста, ждем некоторое время
				if job == nil {
					time.Sleep(1 * time.Second)
					continue
				}

				// Обработка задачи
				w.processJob(ctx, *job)
			}
		}
	}()
}

// Stop останавливает воркер
func (w *Worker) Stop() {
	w.logger.Info("Stopping worker")
	close(w.shutdown)
}

// processJob обрабатывает задачу
func (w *Worker) processJob(ctx context.Context, job entity.QueueJob) {
	// Логирование начала обработки задачи
	w.logger.Info("Processing job",
		"job_id", job.JobID,
		"job_type", job.JobType,
	)

	// Поиск обработчика для типа задачи
	handler, ok := w.handlers[job.JobType]
	if !ok {
		w.logger.Error("No handler registered for job type",
			"job_type", job.JobType,
		)
		return
	}

	// Вызов обработчика
	err := handler(ctx, job)
	if err != nil {
		w.logger.Error("Failed to process job",
			"error", err,
		)
		return
	}

	// Логирование успешной обработки задачи
	w.logger.Info("Job processed successfully",
		"job_id", job.JobID,
	)
}
