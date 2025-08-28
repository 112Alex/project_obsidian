package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/112Alex/project_obsidian/pkg/logger"
)

// SummarizationService представляет собой сервис для суммаризации текста с использованием DeepSeek API
type SummarizationService struct {
	apiKey     string
	apiBaseURL string
	model      string
	logger     *logger.Logger
}

// NewSummarizationService создает новый сервис для суммаризации текста
func NewSummarizationService(apiKey string, apiBaseURL string, model string, logger *logger.Logger) *SummarizationService {
	// Если базовый URL не указан, используем стандартный
	if apiBaseURL == "" {
		apiBaseURL = "https://api.deepseek.com"
	}

	// Если модель не указана, используем deepseek-chat
	if model == "" {
		model = "deepseek-chat"
	}

	return &SummarizationService{
		apiKey:     apiKey,
		apiBaseURL: apiBaseURL,
		model:      model,
		logger:     logger,
	}
}

// CompletionRequest представляет собой запрос на суммаризацию текста
type CompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Message представляет собой сообщение в запросе
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CompletionResponse представляет собой ответ от DeepSeek API
type CompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// SummarizeText суммаризирует текст
func (s *SummarizationService) SummarizeText(ctx context.Context, text string) (string, error) {
	// Логирование начала суммаризации
	s.logger.Info("Summarizing text",
		"text_length", len(text),
		"model", s.model,
	)

	// Создание запроса на суммаризацию
	prompt := fmt.Sprintf(
		"Пожалуйста, создай краткое и информативное резюме следующего текста. "+
			"Сохрани ключевые идеи, факты и выводы. "+
			"Текст: %s", text)

	req := CompletionRequest{
		Model: s.model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   1000,
		Temperature: 0.3,
	}

	// Выполнение запроса
	summary, err := s.createCompletion(ctx, req)
	if err != nil {
		s.logger.Error("Failed to summarize text",
			"error", err,
		)
		return "", fmt.Errorf("failed to summarize text: %w", err)
	}

	// Логирование успешной суммаризации
	s.logger.Info("Text summarized successfully",
		"summary_length", len(summary),
	)

	return summary, nil
}

// Summarize выполняет суммаризацию текста по умолчанию
func (s *SummarizationService) Summarize(ctx context.Context, text string) (string, error) {
	return s.SummarizeText(ctx, text)
}

// SummarizeTextWithBulletPoints суммаризирует текст с маркированным списком
func (s *SummarizationService) SummarizeTextWithBulletPoints(ctx context.Context, text string) (string, error) {
	// Логирование начала суммаризации
	s.logger.Info("Summarizing text with bullet points",
		"text_length", len(text),
		"model", s.model,
	)

	// Создание запроса на суммаризацию
	prompt := fmt.Sprintf(
		"Пожалуйста, создай краткое и информативное резюме следующего текста в виде маркированного списка. "+
			"Сохрани ключевые идеи, факты и выводы. "+
			"Текст: %s", text)

	req := CompletionRequest{
		Model: s.model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   1000,
		Temperature: 0.3,
	}

	// Выполнение запроса
	summary, err := s.createCompletion(ctx, req)
	if err != nil {
		s.logger.Error("Failed to summarize text with bullet points",
			"error", err,
		)
		return "", fmt.Errorf("failed to summarize text with bullet points: %w", err)
	}

	// Логирование успешной суммаризации
	s.logger.Info("Text summarized with bullet points successfully",
		"summary_length", len(summary),
	)

	return summary, nil
}

// SummarizeTextWithMarkdown суммаризирует текст с разметкой Markdown
func (s *SummarizationService) SummarizeTextWithMarkdown(ctx context.Context, text string) (string, error) {
	// Логирование начала суммаризации
	s.logger.Info("Summarizing text with Markdown",
		"text_length", len(text),
		"model", s.model,
	)

	// Создание запроса на суммаризацию
	prompt := fmt.Sprintf(
		"Пожалуйста, создай краткое и информативное резюме следующего текста с использованием разметки Markdown. "+
			"Используй заголовки, подзаголовки, маркированные списки и другие элементы Markdown для структурирования резюме. "+
			"Сохрани ключевые идеи, факты и выводы. "+
			"Текст: %s", text)

	req := CompletionRequest{
		Model: s.model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   1500,
		Temperature: 0.3,
	}

	// Выполнение запроса
	summary, err := s.createCompletion(ctx, req)
	if err != nil {
		s.logger.Error("Failed to summarize text with Markdown",
			"error", err,
		)
		return "", fmt.Errorf("failed to summarize text with Markdown: %w", err)
	}

	// Логирование успешной суммаризации
	s.logger.Info("Text summarized with Markdown successfully",
		"summary_length", len(summary),
	)

	return summary, nil
}

// createCompletion отправляет запрос на создание завершения
func (s *SummarizationService) createCompletion(ctx context.Context, req CompletionRequest) (string, error) {
	// Сериализация запроса
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Создание HTTP запроса
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v1/chat/completions", s.apiBaseURL),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Установка заголовков
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	// Выполнение запроса
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Чтение ответа
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Проверка статуса ответа
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned error: %s, status code: %d", string(respBody), resp.StatusCode)
	}

	// Десериализация ответа
	var completionResp CompletionResponse
	if err := json.Unmarshal(respBody, &completionResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w, response: %s", err, string(respBody))
	}

	// Проверка наличия выбора
	if len(completionResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response: %s", string(respBody))
	}

	return completionResp.Choices[0].Message.Content, nil
}
