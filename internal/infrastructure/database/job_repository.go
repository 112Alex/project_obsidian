package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/112Alex/project_obsidian/internal/domain/entity"
	"github.com/112Alex/project_obsidian/internal/domain/repository"
	"github.com/jackc/pgx/v5"
)

// JobRepositoryPG реализует интерфейс JobRepository для PostgreSQL
type JobRepositoryPG struct {
	db *PostgresDB
}

// NewJobRepository создает новый репозиторий для работы с задачами
func NewJobRepository(db *PostgresDB) repository.JobRepository {
	return &JobRepositoryPG{db: db}
}

// Create создает новую задачу
func (r *JobRepositoryPG) Create(ctx context.Context, job *entity.Job) error {
	now := time.Now()
	job.CreatedAt = now
	job.UpdatedAt = now
	job.Status = entity.JobStatusPending

	query := `
		INSERT INTO jobs (
			user_id, status, audio_file_path, transcription, summary,
			notion_page_id, notion_database_id, created_at, updated_at, completed_at, error_message
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	err := r.db.QueryRow(
		ctx,
		query,
		job.UserID,
		job.Status,
		job.AudioFilePath,
		job.Transcription,
		job.Summary,
		job.NotionPageID,
		job.NotionDatabaseID,
		job.CreatedAt,
		job.UpdatedAt,
		job.CompletedAt,
		job.ErrorMessage,
	).Scan(&job.ID)

	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return nil
}

// GetByID возвращает задачу по её ID
func (r *JobRepositoryPG) GetByID(ctx context.Context, id int64) (*entity.Job, error) {
	query := `
		SELECT 
			id, user_id, status, audio_file_path, transcription, summary,
			notion_page_id, notion_database_id, created_at, updated_at, completed_at, error_message
		FROM jobs
		WHERE id = $1
	`

	job := &entity.Job{}
	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&job.ID,
		&job.UserID,
		&job.Status,
		&job.AudioFilePath,
		&job.Transcription,
		&job.Summary,
		&job.NotionPageID,
		&job.NotionDatabaseID,
		&job.CreatedAt,
		&job.UpdatedAt,
		&job.CompletedAt,
		&job.ErrorMessage,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return job, nil
}

// GetByUserID возвращает задачи пользователя
func (r *JobRepositoryPG) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entity.Job, error) {
	query := `
		SELECT 
			id, user_id, status, audio_file_path, transcription, summary,
			notion_page_id, notion_database_id, created_at, updated_at, completed_at, error_message
		FROM jobs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*entity.Job
	for rows.Next() {
		job := &entity.Job{}
		err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.Status,
			&job.AudioFilePath,
			&job.Transcription,
			&job.Summary,
			&job.NotionPageID,
			&job.NotionDatabaseID,
			&job.CreatedAt,
			&job.UpdatedAt,
			&job.CompletedAt,
			&job.ErrorMessage,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating jobs: %w", err)
	}

	return jobs, nil
}

// Update обновляет информацию о задаче
func (r *JobRepositoryPG) Update(ctx context.Context, job *entity.Job) error {
	job.UpdatedAt = time.Now()

	query := `
		UPDATE jobs
		SET 
			status = $1, 
			audio_file_path = $2, 
			transcription = $3, 
			summary = $4,
			notion_page_id = $5, 
			notion_database_id = $6, 
			updated_at = $7, 
			completed_at = $8, 
			error_message = $9
		WHERE id = $10
	`

	_, err := r.db.Exec(
		ctx,
		query,
		job.Status,
		job.AudioFilePath,
		job.Transcription,
		job.Summary,
		job.NotionPageID,
		job.NotionDatabaseID,
		job.UpdatedAt,
		job.CompletedAt,
		job.ErrorMessage,
		job.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	return nil
}

// UpdateStatus обновляет статус задачи
func (r *JobRepositoryPG) UpdateStatus(ctx context.Context, id int64, status entity.JobStatus, errorMessage string) error {
	now := time.Now()
	var completedAt *time.Time

	if status == entity.JobStatusCompleted || status == entity.JobStatusFailed {
		completedAt = &now
	}

	query := `
		UPDATE jobs
		SET status = $1, updated_at = $2, completed_at = $3, error_message = $4
		WHERE id = $5
	`

	_, err := r.db.Exec(ctx, query, status, now, completedAt, errorMessage, id)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// SetTranscription устанавливает транскрипцию для задачи
func (r *JobRepositoryPG) SetTranscription(ctx context.Context, id int64, transcription string) error {
	query := `
		UPDATE jobs
		SET transcription = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, transcription, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to set transcription: %w", err)
	}

	return nil
}

// SetSummary устанавливает суммаризацию для задачи
func (r *JobRepositoryPG) SetSummary(ctx context.Context, id int64, summary string) error {
	query := `
		UPDATE jobs
		SET summary = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, summary, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to set summary: %w", err)
	}

	return nil
}

// SetNotionIDs устанавливает ID страницы и базы данных Notion для задачи
func (r *JobRepositoryPG) SetNotionIDs(ctx context.Context, id int64, pageID, databaseID string) error {
	query := `
		UPDATE jobs
		SET notion_page_id = $1, notion_database_id = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := r.db.Exec(ctx, query, pageID, databaseID, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to set notion IDs: %w", err)
	}

	return nil
}
