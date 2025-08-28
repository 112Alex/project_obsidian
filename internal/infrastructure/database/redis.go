package database

import (
	"context"
	"fmt"
	"time"

	"github.com/112Alex/project_obsidian/internal/config"
	"github.com/redis/go-redis/v9"
)

// RedisClient представляет собой обертку над клиентом Redis
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient создает новое подключение к Redis
func NewRedisClient(ctx context.Context, cfg config.RedisConfig) (*RedisClient, error) {
	// Создание клиента Redis
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Проверка соединения
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &RedisClient{client: client}, nil
}

// Close закрывает соединение с Redis
func (r *RedisClient) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// Client возвращает клиент Redis
func (r *RedisClient) Client() *redis.Client {
	return r.client
}

// Set устанавливает значение по ключу
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Get получает значение по ключу
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Del удаляет ключ
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// LPush добавляет элемент в начало списка
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPush добавляет элемент в конец списка
func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

// LPop извлекает элемент из начала списка
func (r *RedisClient) LPop(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

// RPop извлекает элемент из конца списка
func (r *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// BLPop блокирующе извлекает элемент из начала списка
func (r *RedisClient) BLPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return r.client.BLPop(ctx, timeout, keys...).Result()
}

// BRPop блокирующе извлекает элемент из конца списка
func (r *RedisClient) BRPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	return r.client.BRPop(ctx, timeout, keys...).Result()
}

// LLen возвращает длину списка
func (r *RedisClient) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}
