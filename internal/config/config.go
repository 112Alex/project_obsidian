package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config представляет собой структуру конфигурации приложения
type Config struct {
	App      AppConfig
	Log      LogConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	Telegram TelegramConfig
	OpenAI   OpenAIConfig
	DeepSeek DeepSeekConfig
	Notion   NotionConfig
	FFmpeg   FFmpegConfig
}

// AppConfig содержит общие настройки приложения
type AppConfig struct {
	Name    string
	Version string
	Env     string
}

// LogConfig содержит настройки логирования
type LogConfig struct {
	Level string
}

// PostgresConfig содержит настройки подключения к PostgreSQL
type PostgresConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
	PoolMax  int
}

// DSN возвращает строку подключения к PostgreSQL
func (c PostgresConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.DBName, c.SSLMode)
}

// RedisConfig содержит настройки подключения к Redis
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// TelegramConfig содержит настройки для Telegram бота
type TelegramConfig struct {
	Token string
}

// OpenAIConfig содержит настройки для OpenAI API
type OpenAIConfig struct {
	APIKey      string
	WhisperModel string
	Timeout     time.Duration
}

// DeepSeekConfig содержит настройки для DeepSeek API
type DeepSeekConfig struct {
	APIKey  string
	Model   string
	Timeout time.Duration
}

// NotionConfig содержит настройки для Notion API
type NotionConfig struct {
	APIKey string
}

// FFmpegConfig содержит настройки для FFmpeg
type FFmpegConfig struct {
	BinaryPath string
}

// NewConfig создает и загружает конфигурацию из файла и переменных окружения
func NewConfig() (*Config, error) {
	// Установка значений по умолчанию
	setDefaults()

	// Чтение конфигурации из файла (необязательно)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		// Отсутствие файла не является критической ошибкой: берём значения из переменных окружения
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Чтение переменных окружения
	viper.AutomaticEnv()

	// Создание и заполнение структуры конфигурации
	var cfg Config

	cfg.App = AppConfig{
		Name:    viper.GetString("APP_NAME"),
		Version: viper.GetString("APP_VERSION"),
		Env:     viper.GetString("APP_ENV"),
	}

	cfg.Log = LogConfig{
		Level: viper.GetString("LOG_LEVEL"),
	}

	cfg.Postgres = PostgresConfig{
		Host:     viper.GetString("POSTGRES_HOST"),
		Port:     viper.GetString("POSTGRES_PORT"),
		Username: viper.GetString("POSTGRES_USER"),
		Password: viper.GetString("POSTGRES_PASSWORD"),
		DBName:   viper.GetString("POSTGRES_DB"),
		SSLMode:  viper.GetString("POSTGRES_SSLMODE"),
		PoolMax:  viper.GetInt("POSTGRES_POOL_MAX"),
	}

	cfg.Redis = RedisConfig{
		Addr:     viper.GetString("REDIS_ADDR"),
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       viper.GetInt("REDIS_DB"),
	}

	cfg.Telegram = TelegramConfig{
		Token: viper.GetString("TELEGRAM_TOKEN"),
	}

	cfg.OpenAI = OpenAIConfig{
		APIKey:      viper.GetString("OPENAI_API_KEY"),
		WhisperModel: viper.GetString("OPENAI_WHISPER_MODEL"),
		Timeout:     viper.GetDuration("OPENAI_TIMEOUT"),
	}

	cfg.DeepSeek = DeepSeekConfig{
		APIKey:  viper.GetString("DEEPSEEK_API_KEY"),
		Model:   viper.GetString("DEEPSEEK_MODEL"),
		Timeout: viper.GetDuration("DEEPSEEK_TIMEOUT"),
	}

	cfg.Notion = NotionConfig{
		APIKey: viper.GetString("NOTION_API_KEY"),
	}

	cfg.FFmpeg = FFmpegConfig{
		BinaryPath: viper.GetString("FFMPEG_BINARY_PATH"),
	}

	return &cfg, nil
}

// setDefaults устанавливает значения по умолчанию для конфигурации
func setDefaults() {
	// App
	viper.SetDefault("APP_NAME", "project_obsidian")
	viper.SetDefault("APP_VERSION", "0.1.0")
	viper.SetDefault("APP_ENV", "development")

	// Log
	viper.SetDefault("LOG_LEVEL", "info")

	// PostgreSQL
	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", "5432")
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "postgres")
	viper.SetDefault("POSTGRES_DB", "obsidian")
	viper.SetDefault("POSTGRES_SSLMODE", "disable")
	viper.SetDefault("POSTGRES_POOL_MAX", 10)

	// Redis
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)

	// OpenAI
	viper.SetDefault("OPENAI_WHISPER_MODEL", "whisper-1")
	viper.SetDefault("OPENAI_TIMEOUT", time.Second*30)

	// DeepSeek
	viper.SetDefault("DEEPSEEK_MODEL", "deepseek-chat")
	viper.SetDefault("DEEPSEEK_TIMEOUT", time.Second*30)

	// FFmpeg
	viper.SetDefault("FFMPEG_BINARY_PATH", "ffmpeg")
}