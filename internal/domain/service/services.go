package service

import (
	"context"
	"io"

	"github.com/112Alex/project_obsidian/internal/domain/entity"
)

// UserService определяет интерфейс для работы с пользователями
type UserService interface {
	// CreateOrUpdate создает или обновляет пользователя
	CreateOrUpdate(ctx context.Context, telegramID int64, username, firstName, lastName string) (*entity.User, error)
	// GetByTelegramID возвращает пользователя по его Telegram ID
	GetByTelegramID(ctx context.Context, telegramID int64) (*entity.User, error)
}

// JobService определяет интерфейс для работы с задачами
type JobService interface {
	// Create создает новую задачу
	Create(ctx context.Context, userID int64, audioFilePath string) (*entity.Job, error)
	// GetByID возвращает задачу по её ID
	GetByID(ctx context.Context, id int64) (*entity.Job, error)
	// GetByUserID возвращает задачи пользователя
	GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entity.Job, error)
	// UpdateStatus обновляет статус задачи
	UpdateStatus(ctx context.Context, id int64, status entity.JobStatus, errorMessage string) error
}

// AudioService определяет интерфейс для работы с аудиофайлами
type AudioService interface {
	// SaveAudio сохраняет аудиофайл
	SaveAudio(ctx context.Context, userID int64, audioData io.Reader, filename string) (string, error)
	// ConvertToWAV конвертирует аудиофайл в формат WAV
	ConvertToWAV(ctx context.Context, inputPath string) (string, error)
	// GetAudioDuration возвращает длительность аудиофайла в секундах
	GetAudioDuration(ctx context.Context, audioPath string) (float64, error)
	// ProcessAudio обрабатывает аудиофайл для дальнейшего использования
	ProcessAudio(ctx context.Context, audioPath string, fileName string) (string, error)
}

// TranscriptionService определяет интерфейс для транскрибации аудио
type TranscriptionService interface {
	// Transcribe выполняет транскрибацию аудиофайла
	Transcribe(ctx context.Context, audioFilePath string) (string, error)
}

// SummarizationService определяет интерфейс для суммаризации текста
type SummarizationService interface {
	// Summarize выполняет суммаризацию текста
	Summarize(ctx context.Context, text string) (string, error)
	// SummarizeText выполняет суммаризацию текста с форматированием
	SummarizeText(ctx context.Context, text string) (string, error)
}

// NotionService определяет интерфейс для работы с Notion
type NotionService interface {
	// CreateDatabase создает базу данных в Notion
	CreateDatabase(ctx context.Context, userID int64, title string) (string, error)
	// CreatePage создает страницу в Notion
	CreatePage(ctx context.Context, databaseID, title, content string) (string, error)
	// ConvertMarkdownToBlocks конвертирует Markdown в блоки Notion
	ConvertMarkdownToBlocks(ctx context.Context, markdown string) (interface{}, error)
}

// QueueService определяет интерфейс для работы с очередью задач
type QueueService interface {
	// EnqueueTranscriptionJob добавляет задачу транскрибации в очередь
	EnqueueTranscriptionJob(ctx context.Context, jobID, userID int64, audioFilePath string) error
	// EnqueueSummarizationJob добавляет задачу суммаризации в очередь
	EnqueueSummarizationJob(ctx context.Context, jobID, userID int64, transcription string) error
	// EnqueueNotionSyncJob добавляет задачу синхронизации с Notion в очередь
	EnqueueNotionSyncJob(ctx context.Context, jobID, userID int64, title, content string) error
	// RegisterHandler регистрирует обработчик для определенного типа задач
	RegisterHandler(jobType entity.JobType, handler func(ctx context.Context, job entity.QueueJob) error)
	// StartWorker запускает обработчик задач из очереди
	StartWorker(ctx context.Context) error
	// PushJob добавляет задачу в очередь
	PushJob(ctx context.Context, job entity.QueueJob) error
}
