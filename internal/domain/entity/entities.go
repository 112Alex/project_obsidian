package entity

import (
	"time"
)

// User представляет собой сущность пользователя
type User struct {
	ID              int64     `json:"id" db:"id"`
	TelegramID      int64     `json:"telegram_id" db:"telegram_id"`
	Username        string    `json:"username" db:"username"`
	FirstName       string    `json:"first_name" db:"first_name"`
	LastName        string    `json:"last_name" db:"last_name"`
	NotionToken     string    `json:"notion_token" db:"notion_token"`
	NotionDatabaseID string    `json:"notion_database_id" db:"notion_database_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Job представляет собой сущность задачи обработки аудио
type Job struct {
	ID              int64     `json:"id" db:"id"`
	UserID          int64     `json:"user_id" db:"user_id"`
	Type            JobType   `json:"type" db:"type"`
	Status          JobStatus `json:"status" db:"status"`
	AudioFilePath   string    `json:"audio_file_path" db:"audio_file_path"`
	FileName        string    `json:"file_name" db:"file_name"`
	Duration        float64   `json:"duration" db:"duration"`
	Transcription   string    `json:"transcription" db:"transcription"`
	Summary         string    `json:"summary" db:"summary"`
	NotionPageID    string    `json:"notion_page_id" db:"notion_page_id"`
	NotionDatabaseID string   `json:"notion_database_id" db:"notion_database_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at" db:"completed_at"`
	ErrorMessage    string    `json:"error_message" db:"error_message"`
}

// JobStatus представляет статус задачи
type JobStatus string

// Константы для статусов задач
const (
	JobStatusCreated     JobStatus = "created"      // Задача создана
	JobStatusProcessing  JobStatus = "processing"   // Задача в процессе обработки
	JobStatusTranscribed JobStatus = "transcribed"  // Задача транскрибирована
	JobStatusSummarized  JobStatus = "summarized"   // Задача суммаризирована
	JobStatusCompleted   JobStatus = "completed"    // Задача завершена
	JobStatusFailed      JobStatus = "failed"       // Задача завершена с ошибкой
)

// Дополнительные константы для статусов задач
const (
	JobStatusQueued  JobStatus = "queued"  // Задача добавлена в очередь
	JobStatusPending JobStatus = "pending" // Задача ожидает обработки
)

// QueueJob представляет собой задачу для очереди Redis
type QueueJob struct {
	ID        int64     `json:"id"`        // ID задачи в базе данных
	JobID     int64     `json:"job_id"`     // ID связанной задачи
	UserID    int64     `json:"user_id"`    // ID пользователя
	JobType   JobType   `json:"job_type"`   // Тип задачи
	CreatedAt time.Time `json:"created_at"` // Время создания задачи
	Payload   any       `json:"payload"`    // Дополнительные данные для задачи
}

// JobType представляет собой тип задачи для очереди
type JobType string

// Константы для типов задач
const (
	JobTypeTranscription               JobType = "transcription"                // Транскрибация аудио
	JobTypeTranscriptionWithTimestamps JobType = "transcription_with_timestamps" // Транскрибация аудио с временными метками
	JobTypeSummarization               JobType = "summarization"                // Суммаризация текста
	JobTypeSummarizationWithBulletPoints JobType = "summarization_with_bullets" // Суммаризация текста с маркированным списком
	JobTypeNotionSync                  JobType = "notion_sync"                  // Синхронизация с Notion
	JobTypeNotion                      JobType = "notion"                       // Интеграция с Notion
	JobTypeNotification                JobType = "notification"                 // Уведомление о завершении задачи
)