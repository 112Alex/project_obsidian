package openai

import (
	"context"
	"fmt"
	"os"

	"github.com/112Alex/project_obsidian/pkg/logger"
	openai "github.com/sashabaranov/go-openai"
)

// TranscriptionService представляет собой сервис для транскрибации аудио с использованием OpenAI Whisper API
type TranscriptionService struct {
	client *openai.Client
	logger *logger.Logger
	model  string
}

// NewTranscriptionService создает новый сервис для транскрибации аудио
func NewTranscriptionService(apiKey string, model string, logger *logger.Logger) *TranscriptionService {
	// Если модель не указана, используем whisper-1
	if model == "" {
		model = openai.Whisper1
	}

	// Создание клиента OpenAI
	client := openai.NewClient(apiKey)

	return &TranscriptionService{
		client: client,
		logger: logger,
		model:  model,
	}
}

// TranscribeAudio транскрибирует аудио файл
func (s *TranscriptionService) TranscribeAudio(ctx context.Context, audioPath string, language string) (string, error) {
	// Логирование начала транскрибации
	s.logger.Info("Transcribing audio",
		"path", audioPath,
		"language", language,
		"model", s.model,
	)

	// Открытие файла
	audioFile, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer audioFile.Close()

	// Создание запроса на транскрибацию
	req := openai.AudioRequest{
		Model:    s.model,
		FilePath: audioPath,
		Language: language,
		Format:   openai.AudioResponseFormatText,
	}

	// Выполнение запроса
	resp, err := s.client.CreateTranscription(ctx, req)
	if err != nil {
		s.logger.Error("Failed to transcribe audio",
			"error", err,
		)
		return "", fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Логирование успешной транскрибации
	s.logger.Info("Audio transcribed successfully",
		"text_length", len(resp.Text),
	)

	return resp.Text, nil
}

// Transcribe performs default transcription using Whisper
func (s *TranscriptionService) Transcribe(ctx context.Context, audioFilePath string) (string, error) {
	return s.TranscribeAudio(ctx, audioFilePath, "")
}

// TranscribeAudioWithTimestamps транскрибирует аудио файл с временными метками
func (s *TranscriptionService) TranscribeAudioWithTimestamps(ctx context.Context, audioPath string, language string) (string, error) {
	// Логирование начала транскрибации
	s.logger.Info("Transcribing audio with timestamps",
		"path", audioPath,
		"language", language,
		"model", s.model,
	)

	// Открытие файла
	audioFile, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer audioFile.Close()

	// Создание запроса на транскрибацию
	req := openai.AudioRequest{
		Model:    s.model,
		FilePath: audioPath,
		Language: language,
		Format:   openai.AudioResponseFormatSRT,
	}

	// Выполнение запроса
	resp, err := s.client.CreateTranscription(ctx, req)
	if err != nil {
		s.logger.Error("Failed to transcribe audio with timestamps",
			"error", err,
		)
		return "", fmt.Errorf("failed to transcribe audio with timestamps: %w", err)
	}

	// Логирование успешной транскрибации
	s.logger.Info("Audio transcribed with timestamps successfully",
		"text_length", len(resp.Text),
	)

	return resp.Text, nil
}

// TranscribeAudioWithVTT транскрибирует аудио файл с форматом VTT
func (s *TranscriptionService) TranscribeAudioWithVTT(ctx context.Context, audioPath string, language string) (string, error) {
	// Логирование начала транскрибации
	s.logger.Info("Transcribing audio with VTT format",
		"path", audioPath,
		"language", language,
		"model", s.model,
	)

	// Открытие файла
	audioFile, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer audioFile.Close()

	// Создание запроса на транскрибацию
	req := openai.AudioRequest{
		Model:    s.model,
		FilePath: audioPath,
		Language: language,
		Format:   openai.AudioResponseFormatJSON,
	}

	// Выполнение запроса
	resp, err := s.client.CreateTranscription(ctx, req)
	if err != nil {
		s.logger.Error("Failed to transcribe audio with VTT format",
			"error", err,
		)
		return "", fmt.Errorf("failed to transcribe audio with VTT format: %w", err)
	}

	// Логирование успешной транскрибации
	s.logger.Info("Audio transcribed with VTT format successfully",
		"text_length", len(resp.Text),
	)

	return resp.Text, nil
}

// TranscribeAudioWithVerbose транскрибирует аудио файл с подробным выводом
func (s *TranscriptionService) TranscribeAudioWithVerbose(ctx context.Context, audioPath string, language string) (string, error) {
	// Логирование начала транскрибации
	s.logger.Info("Transcribing audio with verbose output",
		"path", audioPath,
		"language", language,
		"model", s.model,
	)

	// Открытие файла
	audioFile, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer audioFile.Close()

	// Создание запроса на транскрибацию
	req := openai.AudioRequest{
		Model:    s.model,
		FilePath: audioPath,
		Language: language,
		Format:   openai.AudioResponseFormatVerboseJSON,
	}

	// Выполнение запроса
	resp, err := s.client.CreateTranscription(ctx, req)
	if err != nil {
		s.logger.Error("Failed to transcribe audio with verbose output",
			"error", err,
		)
		return "", fmt.Errorf("failed to transcribe audio with verbose output: %w", err)
	}

	// Логирование успешной транскрибации
	s.logger.Info("Audio transcribed with verbose output successfully",
		"text_length", len(resp.Text),
	)

	return resp.Text, nil
}
