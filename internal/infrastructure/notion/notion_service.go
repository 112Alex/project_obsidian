package notion

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/112Alex/project_obsidian/pkg/logger"
	"github.com/jomei/notionapi"
)

// NotionService представляет собой сервис для работы с Notion API
type NotionService struct {
	client *notionapi.Client
	logger *logger.Logger
}

// NewNotionService создает новый сервис для работы с Notion API
func NewNotionService(apiKey string, logger *logger.Logger) *NotionService {
	// Создание клиента Notion API
	client := notionapi.NewClient(notionapi.Token(apiKey))

	return &NotionService{
		client: client,
		logger: logger,
	}
}

// CreateDatabase создает новую базу данных в Notion
func (s *NotionService) CreateDatabase(ctx context.Context, userID int64, title string) (string, error) {
	parentPageID := fmt.Sprintf("%d", userID)
	// Логирование начала создания базы данных
	s.logger.Info("Creating Notion database",
		"parent_page_id", parentPageID,
		"title", title,
	)

	// Создание запроса на создание базы данных
	req := &notionapi.DatabaseCreateRequest{
		Parent: notionapi.Parent{
			Type:   notionapi.ParentTypePageID,
			PageID: notionapi.PageID(parentPageID),
		},
		Title: []notionapi.RichText{
			{
				Type: "text",
				Text: &notionapi.Text{
					Content: title,
				},
			},
		},
		Properties: map[string]notionapi.PropertyConfig{
			"Name": notionapi.TitlePropertyConfig{
				Type:  "title",
				Title: struct{}{},
			},
			"Description": notionapi.RichTextPropertyConfig{
				Type:     "rich_text",
				RichText: struct{}{},
			},
			"Date": notionapi.DatePropertyConfig{
				Type: "date",
				Date: struct{}{},
			},
			"Status": notionapi.SelectPropertyConfig{
				Type: "select",
				Select: notionapi.Select{
					Options: []notionapi.Option{
						{
							Name:  "Created",
							Color: "blue",
						},
						{
							Name:  "Processing",
							Color: "yellow",
						},
						{
							Name:  "Transcribed",
							Color: "green",
						},
						{
							Name:  "Summarized",
							Color: "purple",
						},
						{
							Name:  "Completed",
							Color: "green",
						},
						{
							Name:  "Failed",
							Color: "red",
						},
					},
				},
			},
			"Tags": notionapi.MultiSelectPropertyConfig{
				Type: "multi_select",
				MultiSelect: notionapi.Select{
					Options: []notionapi.Option{
						{Name: "Audio", Color: "blue"},
						{Name: "Transcription", Color: "purple"},
						{Name: "Summary", Color: "orange"},
					},
				},
			},
			"Duration": notionapi.NumberPropertyConfig{
				Type:   "number",
				Number: notionapi.NumberFormat{Format: notionapi.FormatNumberWithCommas},
			},
		},
	}

	// Выполнение запроса
	database, err := s.client.Database.Create(ctx, req)
	if err != nil {
		s.logger.Error("Failed to create Notion database",
			"error", err,
		)
		return "", fmt.Errorf("failed to create Notion database: %w", err)
	}

	// Логирование успешного создания базы данных
	s.logger.Info("Notion database created successfully",
		"database_id", database.ID,
	)

	return string(database.ID), nil
}

// CreatePage создает новую страницу в базе данных Notion
func (s *NotionService) CreatePage(ctx context.Context, databaseID, title, content string) (string, error) {
	// Логирование начала создания страницы
	s.logger.Info("Creating Notion page",
		"database_id", databaseID,
		"title", title,
	)

	// Tag functionality removed to match service interface
	// removed
	// removed

	dateNow := notionapi.Date(time.Now())
	// Создание запроса на создание страницы
	req := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			Type:       notionapi.ParentTypeDatabaseID,
			DatabaseID: notionapi.DatabaseID(databaseID),
		},
		Properties: notionapi.Properties{
			"Name": notionapi.TitleProperty{
				Title: []notionapi.RichText{
					{
						Type: "text",
						Text: &notionapi.Text{
							Content: title,
						},
					},
				},
			},
			"Date": notionapi.DateProperty{
				Date: &notionapi.DateObject{
					Start: &dateNow,
					End:   nil,
				},
			},
			"Status": notionapi.SelectProperty{
				Select: notionapi.Option{
					Name: "Completed",
				},
			},
		},
		Children: s.convertMarkdownToBlocks(content),
	}

	// Выполнение запроса
	page, err := s.client.Page.Create(ctx, req)
	if err != nil {
		s.logger.Error("Failed to create Notion page",
			"error", err,
		)
		return "", fmt.Errorf("failed to create Notion page: %w", err)
	}

	// Логирование успешного создания страницы
	s.logger.Info("Notion page created successfully",
		"page_id", page.ID,
	)

	return string(page.ID), nil
}

// ConvertMarkdownToBlocks satisfies the service.NotionService interface
func (s *NotionService) ConvertMarkdownToBlocks(ctx context.Context, markdown string) (interface{}, error) {
	return s.convertMarkdownToBlocks(markdown), nil
}

// convertMarkdownToBlocks конвертирует Markdown в блоки Notion
func (s *NotionService) convertMarkdownToBlocks(markdown string) []notionapi.Block {
	// Разделение Markdown на строки
	lines := strings.Split(markdown, "\n")

	// Создание блоков
	blocks := make([]notionapi.Block, 0)
	currentBlock := make([]string, 0)
	currentBlockType := ""

	// Функция для добавления текущего блока в список блоков
	addCurrentBlock := func() {
		if len(currentBlock) == 0 {
			return
		}

		text := strings.Join(currentBlock, "\n")

		switch currentBlockType {
		case "heading_1":
			blocks = append(blocks, notionapi.Heading1Block{
				Heading1: notionapi.Heading{
					RichText: []notionapi.RichText{
						{
							Type: "text",
							Text: &notionapi.Text{
								Content: text,
							},
						},
					},
				},
			})
		case "heading_2":
			blocks = append(blocks, notionapi.Heading2Block{
				Heading2: notionapi.Heading{
					RichText: []notionapi.RichText{
						{
							Type: "text",
							Text: &notionapi.Text{
								Content: text,
							},
						},
					},
				},
			})
		case "heading_3":
			blocks = append(blocks, notionapi.Heading3Block{
				Heading3: notionapi.Heading{
					RichText: []notionapi.RichText{
						{
							Type: "text",
							Text: &notionapi.Text{
								Content: text,
							},
						},
					},
				},
			})
		case "bulleted_list_item":
			blocks = append(blocks, notionapi.BulletedListItemBlock{
				BulletedListItem: notionapi.ListItem{
					RichText: []notionapi.RichText{
						{
							Type: "text",
							Text: &notionapi.Text{
								Content: text,
							},
						},
					},
				},
			})
		case "numbered_list_item":
			blocks = append(blocks, notionapi.NumberedListItemBlock{
				NumberedListItem: notionapi.ListItem{
					RichText: []notionapi.RichText{
						{
							Type: "text",
							Text: &notionapi.Text{
								Content: text,
							},
						},
					},
				},
			})
		case "code":
			blocks = append(blocks, notionapi.CodeBlock{
				Code: notionapi.Code{
					RichText: []notionapi.RichText{
						{
							Type: "text",
							Text: &notionapi.Text{
								Content: text,
							},
						},
					},
					Language: "plain text",
				},
			})
		case "quote":
			blocks = append(blocks, notionapi.QuoteBlock{
				Quote: notionapi.Quote{
					RichText: []notionapi.RichText{
						{
							Type: "text",
							Text: &notionapi.Text{Content: text},
						},
					},
				},
			})
		default:
			blocks = append(blocks, notionapi.ParagraphBlock{

				Paragraph: notionapi.Paragraph{
					RichText: []notionapi.RichText{
						{
							Type: "text",
							Text: &notionapi.Text{
								Content: text,
							},
						},
					},
				},
			})
		}

		currentBlock = make([]string, 0)
		currentBlockType = ""
	}

	// Обработка строк
	for _, line := range lines {
		// Пропуск пустых строк
		if strings.TrimSpace(line) == "" {
			addCurrentBlock()
			continue
		}

		// Определение типа блока
		blockType := ""
		if strings.HasPrefix(line, "# ") {
			blockType = "heading_1"
			line = strings.TrimPrefix(line, "# ")
		} else if strings.HasPrefix(line, "## ") {
			blockType = "heading_2"
			line = strings.TrimPrefix(line, "## ")
		} else if strings.HasPrefix(line, "### ") {
			blockType = "heading_3"
			line = strings.TrimPrefix(line, "### ")
		} else if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			blockType = "bulleted_list_item"
			line = strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")
		} else if strings.HasPrefix(line, "> ") {
			blockType = "quote"
			line = strings.TrimPrefix(line, "> ")
		} else if strings.HasPrefix(line, "```") {
			blockType = "code"
			line = strings.TrimPrefix(line, "```")
		} else if strings.HasPrefix(line, "1. ") || strings.HasPrefix(line, "1) ") {
			blockType = "numbered_list_item"
			line = strings.TrimPrefix(strings.TrimPrefix(line, "1. "), "1) ")
		} else {
			blockType = "paragraph"
		}

		// Если тип блока изменился, добавляем текущий блок в список блоков
		if currentBlockType != "" && currentBlockType != blockType {
			addCurrentBlock()
		}

		// Устанавливаем текущий тип блока
		currentBlockType = blockType

		// Добавляем строку в текущий блок
		currentBlock = append(currentBlock, line)
	}

	// Добавляем последний блок
	addCurrentBlock()

	return blocks
}
