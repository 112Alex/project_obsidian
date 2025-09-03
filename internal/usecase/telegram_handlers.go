package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/112Alex/project_obsidian/internal/domain/entity"
	"github.com/112Alex/project_obsidian/internal/domain/repository"
	"github.com/112Alex/project_obsidian/pkg/logger"
)

// TelegramHandlersUseCase представляет собой сценарий обработки команд Telegram бота
type TelegramHandlersUseCase struct {
	userRepo                repository.UserRepository
	jobRepo                 repository.JobRepository
	audioProcessingUseCase  *AudioProcessingUseCase
	notionProcessingUseCase *NotionProcessingUseCase
	logger                  *logger.Logger
}

// NewTelegramHandlersUseCase создает новый сценарий обработки команд Telegram бота
func NewTelegramHandlersUseCase(
	userRepo repository.UserRepository,
	jobRepo repository.JobRepository,
	audioProcessingUseCase *AudioProcessingUseCase,
	notionProcessingUseCase *NotionProcessingUseCase,
	logger *logger.Logger,
) *TelegramHandlersUseCase {
	return &TelegramHandlersUseCase{
		userRepo:                userRepo,
		jobRepo:                 jobRepo,
		audioProcessingUseCase:  audioProcessingUseCase,
		notionProcessingUseCase: notionProcessingUseCase,
		logger:                  logger,
	}
}

// HandleStart обрабатывает команду /start
func (uc *TelegramHandlersUseCase) HandleStart(ctx context.Context, telegramID int64, username string) (string, error) {
	// Логирование начала обработки команды /start
	uc.logger.Info("Handling /start command",
		"telegram_id", telegramID,
		"username", username,
	)

	// Получение или создание пользователя
	user, err := uc.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		// Если пользователь не найден, создаем нового
		user = &entity.User{
			TelegramID:       telegramID,
			Username:         username,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			NotionToken:      "",
			NotionDatabaseID: "",
		}

		err = uc.userRepo.Create(ctx, user)
		if err != nil {
			uc.logger.Error("Failed to create user",
				"error", err,
			)
			return "", fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Формирование приветственного сообщения
	welcomeMessage := fmt.Sprintf(
		"Привет, %s! 👋\n\n"+
			"Я бот для транскрибации аудио в текст и создания заметок в Notion. 🎙️📝\n\n"+
			"Отправь мне голосовое сообщение или аудиофайл, и я:\n"+
			"1️⃣ Преобразую его в текст\n"+
			"2️⃣ Создам краткое содержание\n"+
			"3️⃣ Сохраню в твою базу Notion (если настроено)\n\n"+
			"Доступные команды:\n"+
			"/help - показать справку\n"+
			"/notion - настроить интеграцию с Notion\n"+
			"/jobs - показать список задач",
		username,
	)

	// Логирование успешной обработки команды /start
	uc.logger.Info("Successfully handled /start command",
		"telegram_id", telegramID,
		"user_id", user.ID,
	)

	return welcomeMessage, nil
}

// HandleHelp обрабатывает команду /help
func (uc *TelegramHandlersUseCase) HandleHelp(ctx context.Context, telegramID int64) (string, error) {
	// Логирование начала обработки команды /help
	uc.logger.Info("Handling /help command",
		"telegram_id", telegramID,
	)

	// Формирование сообщения справки
	helpMessage := "🤖 *Справка по использованию бота* 🤖\n\n" +
		"*Основные возможности:*\n" +
		"• Транскрибация голосовых сообщений и аудиофайлов в текст\n" +
		"• Создание краткого содержания транскрибации\n" +
		"• Сохранение результатов в Notion\n\n" +
		"*Команды:*\n" +
		"/start - начать работу с ботом\n" +
		"/help - показать эту справку\n" +
		"/notion - настроить интеграцию с Notion\n" +
		"/jobs - показать список ваших задач\n\n" +
		"*Как использовать:*\n" +
		"1. Отправьте боту голосовое сообщение или аудиофайл\n" +
		"2. Дождитесь обработки (это может занять некоторое время)\n" +
		"3. Получите транскрипцию и краткое содержание\n" +
		"4. Если настроена интеграция с Notion, результаты будут автоматически сохранены\n\n" +
		"*Поддерживаемые форматы аудио:*\n" +
		"• Голосовые сообщения Telegram\n" +
		"• Аудиофайлы (.mp3, .wav, .ogg, .m4a)\n\n" +
		"*Настройка Notion:*\n" +
		"Используйте команду /notion для настройки интеграции с Notion. Вам потребуется токен интеграции Notion."

	// Логирование успешной обработки команды /help
	uc.logger.Info("Successfully handled /help command",
		"telegram_id", telegramID,
	)

	return helpMessage, nil
}

// HandleNotion обрабатывает команду /notion
func (uc *TelegramHandlersUseCase) HandleNotion(ctx context.Context, telegramID int64, args string) (string, error) {
	// Логирование начала обработки команды /notion
	uc.logger.Info("Handling /notion command",
		"telegram_id", telegramID,
	)

	// Получение пользователя
	user, err := uc.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		uc.logger.Error("Failed to get user",
			"error", err,
		)
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Если аргументы не предоставлены, отправляем инструкцию
	if args == "" {
		notionInstructions := "🔗 *Настройка интеграции с Notion* 🔗\n\n" +
			"Для настройки интеграции с Notion, выполните следующие шаги:\n\n" +
			"1. Перейдите на страницу [notion.so/my-integrations](https://www.notion.so/my-integrations)\n" +
			"2. Создайте новую интеграцию\n" +
			"3. Скопируйте токен интеграции\n" +
			"4. Отправьте команду `/notion ваш_токен`\n\n" +
			"После настройки интеграции, бот автоматически создаст базу данных в вашем Notion для хранения транскрипций."

		// Логирование отправки инструкций по настройке Notion
		uc.logger.Info("Sent Notion setup instructions",
			"telegram_id", telegramID,
		)

		return notionInstructions, nil
	}

	// Настройка интеграции с Notion
	notionToken := strings.TrimSpace(args)
	err = uc.notionProcessingUseCase.SetupNotionIntegration(ctx, user.ID, notionToken)
	if err != nil {
		uc.logger.Error("Failed to setup Notion integration",
			"error", err,
		)
		return "", fmt.Errorf("failed to setup Notion integration: %w", err)
	}

	// Формирование сообщения об успешной настройке
	successMessage := "✅ *Интеграция с Notion успешно настроена!* ✅\n\n" +
		"Теперь все транскрипции будут автоматически сохраняться в вашу базу данных Notion.\n\n" +
		"Вы можете отправить мне голосовое сообщение или аудиофайл для обработки."

	// Логирование успешной настройки интеграции с Notion
	uc.logger.Info("Successfully set up Notion integration",
		"telegram_id", telegramID,
		"user_id", user.ID,
	)

	return successMessage, nil
}

// HandleJobs обрабатывает команду /jobs
func (uc *TelegramHandlersUseCase) HandleJobs(ctx context.Context, telegramID int64) (string, error) {
	// Логирование начала обработки команды /jobs
	uc.logger.Info("Handling /jobs command",
		"telegram_id", telegramID,
	)

	// Получение пользователя
	user, err := uc.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		uc.logger.Error("Failed to get user",
			"error", err,
		)
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Получение списка задач пользователя
	jobs, err := uc.audioProcessingUseCase.GetUserJobs(ctx, user.ID)
	if err != nil {
		uc.logger.Error("Failed to get user jobs",
			"error", err,
		)
		return "", fmt.Errorf("failed to get user jobs: %w", err)
	}

	// Если у пользователя нет задач
	if len(jobs) == 0 {
		return "У вас пока нет задач. Отправьте мне голосовое сообщение или аудиофайл для обработки.", nil
	}

	// Формирование сообщения со списком задач
	messageBuilder := strings.Builder{}
	messageBuilder.WriteString("📋 *Ваши задачи:* 📋\n\n")

	for i, job := range jobs {
		// Получение статуса задачи в текстовом виде
		statusText := "Неизвестно"
		statusEmoji := "❓"

		switch job.Status {
		case entity.JobStatusPending:
			statusText = "В очереди"
			statusEmoji = "⏳"
		case entity.JobStatusProcessing:
			statusText = "Обрабатывается"
			statusEmoji = "⚙️"
		case entity.JobStatusTranscribed:
			statusText = "Транскрибировано"
			statusEmoji = "📝"
		case entity.JobStatusSummarized:
			statusText = "Суммаризировано"
			statusEmoji = "📊"
		case entity.JobStatusCompleted:
			statusText = "Завершено"
			statusEmoji = "✅"
		case entity.JobStatusFailed:
			statusText = "Ошибка"
			statusEmoji = "❌"
		}

		// Получение имени файла из пути
		fileName := filepath.Base(job.AudioFilePath)

		// Форматирование времени создания
		createdAt := job.CreatedAt.Format("02.01.2006 15:04")

		// Добавление информации о задаче
		messageBuilder.WriteString(fmt.Sprintf(
			"%d. %s *%s* (%s)\n   Создано: %s\n",
			i+1,
			statusEmoji,
			fileName,
			statusText,
			createdAt,
		))

		// Если задача завершена и есть ID страницы Notion
		if job.Status == entity.JobStatusCompleted && job.NotionPageID != "" {
			messageBuilder.WriteString("   📎 Сохранено в Notion\n")
		}

		// Добавление разделителя между задачами
		if i < len(jobs)-1 {
			messageBuilder.WriteString("\n")
		}
	}

	// Логирование успешной обработки команды /jobs
	uc.logger.Info("Successfully handled /jobs command",
		"telegram_id", telegramID,
		"user_id", user.ID,
		"jobs_count", len(jobs),
	)

	return messageBuilder.String(), nil
}

// HandleVoiceMessage обрабатывает голосовое сообщение
func (uc *TelegramHandlersUseCase) HandleVoiceMessage(ctx context.Context, telegramID int64, username string, fileID string, filePath string) (string, error) {
	// Логирование начала обработки голосового сообщения
	uc.logger.Info("Handling voice message",
		"telegram_id", telegramID,
		"file_id", fileID,
	)

	// Получение или создание пользователя
	user, err := uc.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		// Если пользователь не найден, создаем нового
		user = &entity.User{
			TelegramID:       telegramID,
			Username:         username,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			NotionToken:      "",
			NotionDatabaseID: "",
		}

		err = uc.userRepo.Create(ctx, user)
		if err != nil {
			uc.logger.Error("Failed to create user",
				"error", err,
			)
			return "", fmt.Errorf("failed to create user: %w", err)
		}

		// ID пользователя уже установлен в методе Create
	}

	// Обработка аудио файла
	fileName := filepath.Base(filePath)
	jobID, err := uc.audioProcessingUseCase.ProcessAudio(ctx, telegramID, filePath, fileName)
	if err != nil {
		uc.logger.Error("Failed to process audio file",
			"error", err,
		)
		return "", fmt.Errorf("failed to process audio file: %w", err)
	}

	// Формирование сообщения об успешном начале обработки
	responseMessage := "🎙️ *Голосовое сообщение принято в обработку!* 🎙️\n\n" +
		"Я начал обработку вашего голосового сообщения. Это может занять некоторое время.\n\n" +
		"Вы получите уведомление, когда транскрипция и суммаризация будут готовы.\n\n" +
		"Идентификатор задачи: `" + fmt.Sprintf("%d", jobID) + "`\n\n" +
		"Вы можете проверить статус задачи с помощью команды /jobs"

	// Логирование успешного начала обработки голосового сообщения
	uc.logger.Info("Successfully started processing voice message",
		"telegram_id", telegramID,
		"user_id", user.ID,
		"job_id", jobID,
	)

	return responseMessage, nil
}

// HandleAudioFile обрабатывает аудио файл
func (uc *TelegramHandlersUseCase) HandleAudioFile(ctx context.Context, telegramID int64, username string, fileID string, filePath string) (string, error) {
	// Логирование начала обработки аудио файла
	uc.logger.Info("Handling audio file",
		"telegram_id", telegramID,
		"file_id", fileID,
	)

	// Получение или создание пользователя
	user, err := uc.userRepo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		// Если пользователь не найден, создаем нового
		user = &entity.User{
			TelegramID:       telegramID,
			Username:         username,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			NotionToken:      "",
			NotionDatabaseID: "",
		}

		err = uc.userRepo.Create(ctx, user)
		if err != nil {
			uc.logger.Error("Failed to create user",
				"error", err,
			)
			return "", fmt.Errorf("failed to create user: %w", err)
		}

		// ID пользователя устанавливается внутри метода Create
	}

	// Обработка аудио файла
	fileName := filepath.Base(filePath)
	jobID, err := uc.audioProcessingUseCase.ProcessAudio(ctx, telegramID, filePath, fileName)
	if err != nil {
		uc.logger.Error("Failed to process audio file",
			"error", err,
		)
		return "", fmt.Errorf("failed to process audio file: %w", err)
	}

	// Формирование сообщения об успешном начале обработки
	responseMessage := "🎵 *Аудиофайл принят в обработку!* 🎵\n\n" +
		"Я начал обработку вашего аудиофайла. Это может занять некоторое время.\n\n" +
		"Вы получите уведомление, когда транскрипция и суммаризация будут готовы.\n\n" +
		"Идентификатор задачи: `" + fmt.Sprintf("%d", jobID) + "`\n\n" +
		"Вы можете проверить статус задачи с помощью команды /jobs"

	// Логирование успешного начала обработки аудио файла
	uc.logger.Info("Successfully started processing audio file",
		"telegram_id", telegramID,
		"user_id", user.ID,
		"job_id", jobID,
	)

	return responseMessage, nil
}

// SendJobCompletionNotification отправляет уведомление о завершении задачи
func (uc *TelegramHandlersUseCase) SendJobCompletionNotification(ctx context.Context, jobIDStr string) (int64, string, error) {
	// Преобразование строки jobID в int64
	jobID, err := strconv.ParseInt(jobIDStr, 10, 64)
	if err != nil {
		uc.logger.Error("Failed to parse job ID",
			"error", err,
		)
		return 0, "", fmt.Errorf("failed to parse job ID: %w", err)
	}
	// Логирование начала отправки уведомления о завершении задачи
	uc.logger.Info("Sending job completion notification",
		"job_id", jobID,
	)

	// Получение задачи из базы данных
	job, err := uc.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		uc.logger.Error("Failed to get job",
			"error", err,
		)
		return 0, "", fmt.Errorf("failed to get job: %w", err)
	}

	// Получение пользователя из базы данных
	user, err := uc.userRepo.GetByTelegramID(ctx, job.UserID)
	if err != nil {
		uc.logger.Error("Failed to get user",
			"error", err,
		)
		return 0, "", fmt.Errorf("failed to get user: %w", err)
	}

	// Формирование сообщения о завершении задачи
	messageBuilder := strings.Builder{}
	messageBuilder.WriteString("✅ *Задача успешно выполнена!* ✅\n\n")

	// Добавление информации о транскрипции
	if job.Transcription != "" {
		// Ограничение длины транскрипции для сообщения
		transcriptionPreview := job.Transcription
		if len(transcriptionPreview) > 500 {
			transcriptionPreview = transcriptionPreview[:500] + "..."
		}

		messageBuilder.WriteString("📝 *Транскрипция:*\n")
		messageBuilder.WriteString(transcriptionPreview)
		messageBuilder.WriteString("\n\n")
	}

	// Добавление информации о суммаризации
	if job.Summary != "" {
		messageBuilder.WriteString("📊 *Краткое содержание:*\n")
		messageBuilder.WriteString(job.Summary)
		messageBuilder.WriteString("\n\n")
	}

	// Добавление информации о сохранении в Notion
	if job.NotionPageID != "" {
		messageBuilder.WriteString("📎 *Сохранено в Notion*\n")
	}

	// Логирование успешной отправки уведомления о завершении задачи
	uc.logger.Info("Successfully prepared job completion notification",
		"job_id", jobID,
		"user_id", job.UserID,
		"telegram_id", user.TelegramID,
	)

	return user.TelegramID, messageBuilder.String(), nil
}
