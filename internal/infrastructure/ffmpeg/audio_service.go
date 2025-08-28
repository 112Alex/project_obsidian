package ffmpeg

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/112Alex/project_obsidian/pkg/logger"
)

// AudioService представляет собой сервис для работы с аудио файлами
type AudioService struct {
	ffmpegPath string
	logger     *logger.Logger
}

// NewAudioService создает новый сервис для работы с аудио файлами
func NewAudioService(ffmpegPath string, logger *logger.Logger) *AudioService {
	return &AudioService{
		ffmpegPath: ffmpegPath,
		logger:     logger,
	}
}

// SaveAudio сохраняет аудиофайл
func (s *AudioService) SaveAudio(ctx context.Context, userID int64, audioData io.Reader, filename string) (string, error) {
	// Создание директории для сохранения файлов пользователя
	userDir := filepath.Join("uploads", fmt.Sprintf("user_%d", userID))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create user directory: %w", err)
	}

	// Создание файла
	filePath := filepath.Join(userDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Копирование данных из reader в файл
	if _, err := io.Copy(file, audioData); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

// ConvertToWAV конвертирует аудио файл в формат WAV
func (s *AudioService) ConvertToWAV(ctx context.Context, inputPath string) (string, error) {
	// Создание выходного пути
	outputPath := changeExt(inputPath, ".wav")

	// Логирование начала конвертации
	s.logger.Info("Converting audio to WAV",
		"input", inputPath,
		"output", outputPath,
	)

	// Формирование команды FFmpeg
	cmd := exec.CommandContext(
		ctx,
		s.ffmpegPath,
		"-i", inputPath,
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		"-y",
		outputPath,
	)

	// Выполнение команды
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("Failed to convert audio",
			"error", err,
			"output", string(output),
		)
		return "", fmt.Errorf("failed to convert audio: %w\nOutput: %s", err, string(output))
	}

	// Проверка существования выходного файла
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("output file not created: %w", err)
	}

	return outputPath, nil
}

// NormalizeAudio нормализует громкость аудио файла
func (s *AudioService) NormalizeAudio(ctx context.Context, inputPath string) (string, error) {
	// Создание выходного пути
	outputPath := addSuffix(inputPath, "_normalized")

	// Логирование начала нормализации
	s.logger.Info("Normalizing audio",
		"input", inputPath,
		"output", outputPath,
	)

	// Формирование команды FFmpeg
	cmd := exec.CommandContext(
		ctx,
		s.ffmpegPath,
		"-i", inputPath,
		"-filter:a", "loudnorm=I=-16:TP=-1.5:LRA=11",
		"-y",
		outputPath,
	)

	// Выполнение команды
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("Failed to normalize audio",
			"error", err,
			"output", string(output),
		)
		return "", fmt.Errorf("failed to normalize audio: %w\nOutput: %s", err, string(output))
	}

	// Проверка существования выходного файла
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("output file not created: %w", err)
	}

	return outputPath, nil
}

// RemoveNoise удаляет шум из аудио файла
func (s *AudioService) RemoveNoise(ctx context.Context, inputPath string) (string, error) {
	// Создание выходного пути
	outputPath := addSuffix(inputPath, "_denoised")

	// Логирование начала удаления шума
	s.logger.Info("Removing noise from audio",
		"input", inputPath,
		"output", outputPath,
	)

	// Формирование команды FFmpeg
	cmd := exec.CommandContext(
		ctx,
		s.ffmpegPath,
		"-i", inputPath,
		"-af", "afftdn=nf=-25",
		"-y",
		outputPath,
	)

	// Выполнение команды
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("Failed to remove noise",
			"error", err,
			"output", string(output),
		)
		return "", fmt.Errorf("failed to remove noise: %w\nOutput: %s", err, string(output))
	}

	// Проверка существования выходного файла
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("output file not created: %w", err)
	}

	return outputPath, nil
}

// ProcessAudioForTranscription обрабатывает аудио файл для транскрибации
func (s *AudioService) ProcessAudioForTranscription(ctx context.Context, inputPath string) (string, error) {
	// Конвертация в WAV
	wavPath, err := s.ConvertToWAV(ctx, inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to convert to WAV: %w", err)
	}

	// Нормализация аудио
	normalizedPath, err := s.NormalizeAudio(ctx, wavPath)
	if err != nil {
		return "", fmt.Errorf("failed to normalize audio: %w", err)
	}

	// Удаление шума
	denoisedPath, err := s.RemoveNoise(ctx, normalizedPath)
	if err != nil {
		return "", fmt.Errorf("failed to remove noise: %w", err)
	}

	return denoisedPath, nil
}

// GetAudioDuration возвращает длительность аудио файла в секундах
func (s *AudioService) GetAudioDuration(ctx context.Context, inputPath string) (float64, error) {
	// Формирование команды FFprobe
	cmd := exec.CommandContext(
		ctx,
		s.ffmpegPath,
		"-i", inputPath,
		"-show_entries", "format=duration",
		"-v", "quiet",
		"-of", "csv=p=0",
	)

	// Выполнение команды
	output, err := cmd.Output()
	if err != nil {
		s.logger.Error("Failed to get audio duration",
			"error", err,
		)
		return 0, fmt.Errorf("failed to get audio duration: %w", err)
	}

	// Парсинг длительности
	var duration float64
	if _, err := fmt.Sscanf(string(output), "%f", &duration); err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// changeExt изменяет расширение файла
func changeExt(path string, newExt string) string {
	ext := filepath.Ext(path)
	return path[:len(path)-len(ext)] + newExt
}

// addSuffix добавляет суффикс к имени файла перед расширением
func addSuffix(path string, suffix string) string {
	ext := filepath.Ext(path)
	return path[:len(path)-len(ext)] + suffix + ext
}

func (s *AudioService) ProcessAudio(ctx context.Context, audioPath string, fileName string) (string, error) {
	// Currently we ignore fileName as processing depends only on path.
	return s.ProcessAudioForTranscription(ctx, audioPath)
}
