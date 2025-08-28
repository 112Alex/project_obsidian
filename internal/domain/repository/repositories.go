package repository

import (
	"context"

	"github.com/112Alex/project_obsidian/internal/domain/entity"
)

// UserRepository определяет интерфейс для работы с пользователями
type UserRepository interface {
	// Create создает нового пользователя
	Create(ctx context.Context, user *entity.User) error
	// GetByTelegramID возвращает пользователя по его Telegram ID
	GetByTelegramID(ctx context.Context, telegramID int64) (*entity.User, error)
	// Update обновляет информацию о пользователе
	Update(ctx context.Context, user *entity.User) error
}

// JobRepository определяет интерфейс для работы с задачами
type JobRepository interface {
	// Create создает новую задачу
	Create(ctx context.Context, job *entity.Job) error
	// GetByID возвращает задачу по её ID
	GetByID(ctx context.Context, id int64) (*entity.Job, error)
	// GetByUserID возвращает задачи пользователя
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entity.Job, error)
	// Update обновляет информацию о задаче
	Update(ctx context.Context, job *entity.Job) error
	// UpdateStatus обновляет статус задачи
	UpdateStatus(ctx context.Context, id int64, status entity.JobStatus, errorMessage string) error
	// SetTranscription устанавливает транскрипцию для задачи
	SetTranscription(ctx context.Context, id int64, transcription string) error
	// SetSummary устанавливает суммаризацию для задачи
	SetSummary(ctx context.Context, id int64, summary string) error
	// SetNotionIDs устанавливает ID страницы и базы данных Notion для задачи
	SetNotionIDs(ctx context.Context, id int64, pageID, databaseID string) error
}

// QueueRepository определяет интерфейс для работы с очередью задач
type QueueRepository interface {
	// Push добавляет задачу в очередь
	Push(ctx context.Context, queueName string, job *entity.QueueJob) error
	// Pop извлекает задачу из очереди
	Pop(ctx context.Context, queueName string) (*entity.QueueJob, error)
	// Size возвращает размер очереди
	Size(ctx context.Context, queueName string) (int64, error)
}
