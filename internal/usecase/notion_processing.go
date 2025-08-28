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

// NotionProcessingUseCase представляет собой сценарий обработки интеграции с Notion
type NotionProcessingUseCase struct {
	jobRepo       repository.JobRepository
	userRepo      repository.UserRepository
	notionService service.NotionService
	logger        *logger.Logger
}

// NewNotionProcessingUseCase создает новый сценарий обработки интеграции с Notion
func NewNotionProcessingUseCase(
	jobRepo repository.JobRepository,
	userRepo repository.UserRepository,
	notionService service.NotionService,
	logger *logger.Logger,
) *NotionProcessingUseCase {
	return &NotionProcessingUseCase{
		jobRepo:       jobRepo,
		userRepo:      userRepo,
		notionService: notionService,
		logger:        logger,
	}
}

// ProcessNotionIntegration обрабатывает интеграцию с Notion
func (uc *NotionProcessingUseCase) ProcessNotionIntegration(ctx context.Context, job entity.QueueJob) error {
	// Получение данных из задачи
	payload := job.Payload.(map[string]interface{})
	transcription, ok := payload["transcription"].(string)
	if !ok {
		return fmt.Errorf("transcription not found in job payload or has invalid type")
	}

	summary, ok := payload["summary"].(string)
	if !ok {
		return fmt.Errorf("summary not found in job payload or has invalid type")
	}

	userID := job.UserID

	// Логирование начала обработки интеграции с Notion
	uc.logger.Info("Processing Notion integration",
		"job_id", job.JobID,
		"user_id", userID,
	)

	// Получение пользователя
	user, err := uc.userRepo.GetByTelegramID(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to get user",
			"error", err,
			"user_id", userID,
		)
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Проверка наличия Notion интеграции у пользователя
	if user.NotionToken == "" || user.NotionDatabaseID == "" {
		uc.logger.Warn("User has no Notion integration",
			"user_id", userID,
		)
		// Обновление статуса задачи
		err = uc.jobRepo.UpdateStatus(ctx, job.JobID, entity.JobStatusCompleted, "")
		if err != nil {
			uc.logger.Error("Failed to update job status",
				"error", err,
			)
			return fmt.Errorf("failed to update job status: %w", err)
		}
		return nil
	}

	// Создание страницы в Notion
	pageTitle := fmt.Sprintf("Транскрипция от %s", time.Now().Format("02.01.2006 15:04"))
	// Формируем содержимое страницы, включая транскрипцию и суммаризацию
	content := fmt.Sprintf("## Суммаризация\n\n%s\n\n## Полная транскрипция\n\n%s", summary, transcription)
	pageID, err := uc.notionService.CreatePage(
		ctx,
		user.NotionDatabaseID,
		pageTitle,
		content,
	)
	if err != nil {
		uc.logger.Error("Failed to create Notion page",
			"error", err,
		)
		return fmt.Errorf("failed to create Notion page: %w", err)
	}

	// Обновление задачи в базе данных
	err = uc.jobRepo.SetNotionIDs(ctx, job.JobID, pageID, user.NotionDatabaseID)
	if err != nil {
		uc.logger.Error("Failed to update job Notion IDs",
			"error", err,
		)
		return fmt.Errorf("failed to update job Notion IDs: %w", err)
	}

	// Обновление статуса задачи
	err = uc.jobRepo.UpdateStatus(ctx, job.JobID, entity.JobStatusCompleted, "")
	if err != nil {
		uc.logger.Error("Failed to update job status",
			"error", err,
		)
		return fmt.Errorf("failed to update job status: %w", err)
	}

	// Логирование успешной обработки интеграции с Notion
	uc.logger.Info("Notion integration processed successfully",
		"job_id", job.JobID,
		"notion_page_id", pageID,
	)

	return nil
}

// SetupNotionIntegration настраивает интеграцию с Notion для пользователя
func (uc *NotionProcessingUseCase) SetupNotionIntegration(ctx context.Context, userID int64, notionToken string) error {
	// Логирование начала настройки интеграции с Notion
	uc.logger.Info("Setting up Notion integration",
		"user_id", userID,
	)

	// Получение пользователя из базы данных
	user, err := uc.userRepo.GetByTelegramID(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to get user",
			"error", err,
		)
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Создание базы данных в Notion
	databaseID, err := uc.notionService.CreateDatabase(
		ctx,
		user.ID,
		"Транскрипции аудио",
	)
	if err != nil {
		uc.logger.Error("Failed to create Notion database",
			"error", err,
		)
		return fmt.Errorf("failed to create Notion database: %w", err)
	}

	// Обновление пользователя в базе данных
	user.NotionToken = notionToken
	user.NotionDatabaseID = databaseID

	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		uc.logger.Error("Failed to update user",
			"error", err,
		)
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Логирование успешной настройки интеграции с Notion
	uc.logger.Info("Notion integration set up successfully",
		"user_id", userID,
		"notion_database_id", databaseID,
	)

	return nil
}
