package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/112Alex/project_obsidian/internal/domain/entity"
	"github.com/112Alex/project_obsidian/internal/domain/repository"
)

// QueueRepositoryRedis реализует интерфейс QueueRepository для Redis
type QueueRepositoryRedis struct {
	redis *RedisClient
}

// NewQueueRepository создает новый репозиторий для работы с очередью задач
func NewQueueRepository(redis *RedisClient) repository.QueueRepository {
	return &QueueRepositoryRedis{redis: redis}
}

// Push добавляет задачу в очередь
func (r *QueueRepositoryRedis) Push(ctx context.Context, queueName string, job *entity.QueueJob) error {
	// Устанавливаем время создания задачи
	job.CreatedAt = time.Now()

	// Сериализуем задачу в JSON
	jobJSON, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Добавляем задачу в конец очереди
	err = r.redis.RPush(ctx, queueName, jobJSON)
	if err != nil {
		return fmt.Errorf("failed to push job to queue: %w", err)
	}

	return nil
}

// Pop извлекает задачу из очереди
func (r *QueueRepositoryRedis) Pop(ctx context.Context, queueName string) (*entity.QueueJob, error) {
	// Извлекаем задачу из начала очереди с блокировкой
	result, err := r.redis.BLPop(ctx, time.Second*1, queueName)
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to pop job from queue: %w", err)
	}

	// Если очередь пуста, возвращаем nil
	if len(result) < 2 {
		return nil, nil
	}

	// Десериализуем задачу из JSON
	var job entity.QueueJob
	err = json.Unmarshal([]byte(result[1]), &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// Size возвращает размер очереди
func (r *QueueRepositoryRedis) Size(ctx context.Context, queueName string) (int64, error) {
	// Получаем длину списка
	size, err := r.redis.LLen(ctx, queueName)
	if err != nil {
		return 0, fmt.Errorf("failed to get queue size: %w", err)
	}

	return size, nil
}
