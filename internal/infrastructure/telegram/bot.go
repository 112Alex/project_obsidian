package telegram

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/112Alex/project_obsidian/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Bot представляет собой обертку над Telegram ботом
type Bot struct {
	api    *tgbotapi.BotAPI
	logger *logger.Logger

	// Обработчики команд и сообщений
	commandHandlers map[string]CommandHandler
	messageHandler  MessageHandler
	audioHandler    AudioHandler

	stop chan struct{}
}

// Stop останавливает бота
func (b *Bot) Stop() {
	select {
	case <-b.stop:
		// already closed
	default:
		close(b.stop)
	}
}

// CommandHandler представляет собой обработчик команды
type CommandHandler func(ctx context.Context, message *tgbotapi.Message) error

// MessageHandler представляет собой обработчик текстового сообщения
type MessageHandler func(ctx context.Context, message *tgbotapi.Message) error

// AudioHandler представляет собой обработчик аудио сообщения
type AudioHandler func(ctx context.Context, message *tgbotapi.Message, audio io.ReadCloser, fileName string) error

// NewBot создает нового Telegram бота
func NewBot(token string, logger *logger.Logger) (*Bot, error) {
	// Создание клиента Telegram Bot API
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	// Создание бота
	bot := &Bot{
		api:             api,
		logger:          logger,
		commandHandlers: make(map[string]CommandHandler),
		stop:            make(chan struct{}),
	}

	return bot, nil
}

// RegisterCommandHandler регистрирует обработчик команды
func (b *Bot) RegisterCommandHandler(command string, handler CommandHandler) {
	b.commandHandlers[command] = handler
}

// RegisterMessageHandler регистрирует обработчик текстовых сообщений
func (b *Bot) RegisterMessageHandler(handler MessageHandler) {
	b.messageHandler = handler
}

// RegisterAudioHandler регистрирует обработчик аудио сообщений
func (b *Bot) RegisterAudioHandler(handler AudioHandler) {
	b.audioHandler = handler
}

// Start запускает бота
func (b *Bot) Start() error {
	ctx := context.Background()
	b.logger.Info("Starting Telegram bot", "username", b.api.Self.UserName)

	// Настройка получения обновлений
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	// Получение канала обновлений
	updates := b.api.GetUpdatesChan(updateConfig)

	// Обработка обновлений
	for {
		select {
		case <-b.stop:
			b.logger.Info("Stopping Telegram bot")
			return nil
		case update := <-updates:
			go b.handleUpdate(ctx, update)
		}
	}
}

// handleUpdate обрабатывает обновление от Telegram
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	// Обработка сообщений
	if update.Message != nil {
		b.handleMessage(ctx, update.Message)
	}
}

// handleMessage обрабатывает сообщение
func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	// Логирование полученного сообщения
	b.logger.Debug("Received message",
		"chat_id", message.Chat.ID,
		"user_id", message.From.ID,
		"username", message.From.UserName,
		"text", message.Text,
	)

	// Обработка команд
	if message.IsCommand() {
		b.handleCommand(ctx, message)
		return
	}

	// Обработка аудио сообщений
	if message.Voice != nil && b.audioHandler != nil {
		b.handleVoice(ctx, message)
		return
	}

	if message.Audio != nil && b.audioHandler != nil {
		b.handleAudio(ctx, message)
		return
	}

	// Обработка текстовых сообщений
	if b.messageHandler != nil {
		err := b.messageHandler(ctx, message)
		if err != nil {
			b.logger.Error("Failed to handle message", "error", err)
			b.sendErrorMessage(message.Chat.ID, "Произошла ошибка при обработке сообщения")
		}
	}
}

// handleCommand обрабатывает команду
func (b *Bot) handleCommand(ctx context.Context, message *tgbotapi.Message) {
	// Получение имени команды
	command := message.Command()

	// Поиск обработчика команды
	handler, ok := b.commandHandlers[command]
	if !ok {
		b.logger.Warn("Unknown command", "command", command)
		b.sendErrorMessage(message.Chat.ID, "Неизвестная команда")
		return
	}

	// Вызов обработчика команды
	err := handler(ctx, message)
	if err != nil {
		b.logger.Error("Failed to handle command", "command", command, "error", err)
		b.sendErrorMessage(message.Chat.ID, "Произошла ошибка при обработке команды")
	}
}

// handleVoice обрабатывает голосовое сообщение
func (b *Bot) handleVoice(ctx context.Context, message *tgbotapi.Message) {
	// Получение информации о голосовом сообщении
	voiceFileID := message.Voice.FileID
	voiceFileName := fmt.Sprintf("%s.ogg", voiceFileID)

	// Получение файла
	voiceFile, err := b.api.GetFile(tgbotapi.FileConfig{FileID: voiceFileID})
	if err != nil {
		b.logger.Error("Failed to get voice file", "error", err)
		b.sendErrorMessage(message.Chat.ID, "Не удалось получить голосовое сообщение")
		return
	}

	// Загрузка файла
	voiceURL := voiceFile.Link(b.api.Token)
	voiceReader, err := b.downloadFile(voiceURL)
	if err != nil {
		b.logger.Error("Failed to download voice file", "error", err)
		b.sendErrorMessage(message.Chat.ID, "Не удалось загрузить голосовое сообщение")
		return
	}
	defer voiceReader.Close()

	// Вызов обработчика аудио
	err = b.audioHandler(ctx, message, voiceReader, voiceFileName)
	if err != nil {
		b.logger.Error("Failed to handle voice message", "error", err)
		b.sendErrorMessage(message.Chat.ID, "Произошла ошибка при обработке голосового сообщения")
	}
}

// handleAudio обрабатывает аудио сообщение
func (b *Bot) handleAudio(ctx context.Context, message *tgbotapi.Message) {
	// Получение информации об аудио сообщении
	audioFileID := message.Audio.FileID
	audioFileName := message.Audio.FileName
	if audioFileName == "" {
		audioFileName = fmt.Sprintf("%s.mp3", audioFileID)
	}

	// Получение файла
	audioFile, err := b.api.GetFile(tgbotapi.FileConfig{FileID: audioFileID})
	if err != nil {
		b.logger.Error("Failed to get audio file", "error", err)
		b.sendErrorMessage(message.Chat.ID, "Не удалось получить аудио файл")
		return
	}

	// Загрузка файла
	audioURL := audioFile.Link(b.api.Token)
	audioReader, err := b.downloadFile(audioURL)
	if err != nil {
		b.logger.Error("Failed to download audio file", "error", err)
		b.sendErrorMessage(message.Chat.ID, "Не удалось загрузить аудио файл")
		return
	}
	defer audioReader.Close()

	// Вызов обработчика аудио
	err = b.audioHandler(ctx, message, audioReader, audioFileName)
	if err != nil {
		b.logger.Error("Failed to handle audio message", "error", err)
		b.sendErrorMessage(message.Chat.ID, "Произошла ошибка при обработке аудио файла")
	}
}

// downloadFile загружает файл по URL
func (b *Bot) downloadFile(url string) (io.ReadCloser, error) {
	// Создание временного файла
	tmpFile, err := os.CreateTemp("", "tg-audio-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	// Загрузка файла
	resp, err := http.Get(url)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// Копирование содержимого в файл
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Перемещение указателя в начало файла
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, fmt.Errorf("failed to seek file: %w", err)
	}

	file := tmpFile

	// Перемещение указателя в начало файла
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, fmt.Errorf("failed to seek file: %w", err)
	}

	// Создание ReadCloser, который удаляет файл при закрытии
	return &fileReadCloser{file: file}, nil
}

// fileReadCloser представляет собой обертку над файлом, которая удаляет файл при закрытии
type fileReadCloser struct {
	file *os.File
}

// Read реализует интерфейс io.Reader
func (f *fileReadCloser) Read(p []byte) (n int, err error) {
	return f.file.Read(p)
}

// Close реализует интерфейс io.Closer
func (f *fileReadCloser) Close() error {
	// Получение имени файла
	fileName := f.file.Name()

	// Закрытие файла
	err := f.file.Close()

	// Удаление файла
	os.Remove(fileName)

	return err
}

// SendMessage отправляет текстовое сообщение
func (b *Bot) SendMessage(chatID int64, text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	return b.api.Send(msg)
}

// SendMarkdownMessage отправляет сообщение с разметкой Markdown
func (b *Bot) SendMarkdownMessage(chatID int64, text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	return b.api.Send(msg)
}

// sendErrorMessage отправляет сообщение об ошибке
func (b *Bot) sendErrorMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.api.Send(msg)
	if err != nil {
		b.logger.Error("Failed to send error message", "error", err)
	}
}

// SaveAudioFile сохраняет аудиофайл на диск
func (b *Bot) SaveAudioFile(reader io.Reader, userID int64, fileName string) (string, error) {
	// Создание директории для сохранения файлов пользователя
	userDir := filepath.Join("uploads", fmt.Sprintf("user_%d", userID))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create user directory: %w", err)
	}

	// Создание файла
	filePath := filepath.Join(userDir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Копирование данных из reader в файл
	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}
