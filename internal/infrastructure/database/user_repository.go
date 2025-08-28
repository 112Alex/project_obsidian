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

// UserRepositoryPG реализует интерфейс UserRepository для PostgreSQL
type UserRepositoryPG struct {
	db *PostgresDB
}

// NewUserRepository создает новый репозиторий для работы с пользователями
func NewUserRepository(db *PostgresDB) repository.UserRepository {
	return &UserRepositoryPG{db: db}
}

// Create создает нового пользователя
func (r *UserRepositoryPG) Create(ctx context.Context, user *entity.User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
		INSERT INTO users (telegram_id, username, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRow(
		ctx,
		query,
		user.TelegramID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByTelegramID возвращает пользователя по его Telegram ID
func (r *UserRepositoryPG) GetByTelegramID(ctx context.Context, telegramID int64) (*entity.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, created_at, updated_at
		FROM users
		WHERE telegram_id = $1
	`

	user := &entity.User{}
	err := r.db.QueryRow(
		ctx,
		query,
		telegramID,
	).Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Update обновляет информацию о пользователе
func (r *UserRepositoryPG) Update(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET username = $1, first_name = $2, last_name = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.Exec(
		ctx,
		query,
		user.Username,
		user.FirstName,
		user.LastName,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
