package notion

import (
	"fmt"
	"mainote-server/internal/domain"
)

// Service implements the domain.NotionService interface
type Service struct {
	client *NotionClient
}

// NewService creates a new Notion service
func NewService(client *NotionClient) *Service {
	return &Service{
		client: client,
	}
}

// CreateNote creates a new note in Notion
func (s *Service) CreateNote(title, content, category, source string) (*domain.NotionPage, error) {
	page, err := s.client.CreateNote(title, content, category, source)
	if err != nil {
		return nil, fmt.Errorf("failed to create note in Notion: %w", err)
	}

	return s.convertPageToDomain(page), nil
}

// UpdateNoteType updates the type of a note in Notion
func (s *Service) UpdateNoteType(pageID, noteType string) error {
	return s.client.UpdateNoteType(pageID, noteType)
}

// UpdatePageStatus updates the status of a page
func (s *Service) UpdatePageStatus(pageID, status string) error {
	return s.client.UpdatePageStatus(pageID, status)
}

// GetActiveTasks retrieves all active tasks from Notion
func (s *Service) GetActiveTasks() ([]*domain.NotionPage, error) {
	pages, err := s.client.GetActiveTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to get active tasks: %w", err)
	}

	result := make([]*domain.NotionPage, len(pages))
	for i, page := range pages {
		result[i] = s.convertPageToDomain(&page)
	}

	return result, nil
}

// GetPageByID retrieves a specific page by its ID
func (s *Service) GetPageByID(pageID string) (*domain.NotionPage, error) {
	page, err := s.client.GetPageByID(pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get page by ID: %w", err)
	}

	return s.convertPageToDomain(page), nil
}

// FormatMorningNotification formats active tasks into a notification message
func (s *Service) FormatMorningNotification(tasks []*domain.NotionPage) string {
	if len(tasks) == 0 {
		return "Доброе утро! У вас нет активных задач на сегодня. Хорошего дня! 🌞"
	}

	message := "🌅 Доброе утро! Вот ваш план на сегодня:\n\n"

	for i, task := range tasks {
		// Get emoji for task type
		emoji := "📝"
		switch task.Type {
		case "idea":
			emoji = "💡"
		case "task":
			emoji = "✅"
		case "personal":
			emoji = "🏖"
		}

		title := task.Title
		if title == "" {
			title = "Без названия"
		}

		message += fmt.Sprintf("%d. %s %s\n", i+1, emoji, title)
	}

	message += "\nУдачного и продуктивного дня! 💪"
	return message
}

// convertPageToDomain converts a Notion API page to domain model
func (s *Service) convertPageToDomain(page *Page) *domain.NotionPage {
	result := &domain.NotionPage{
		ID:        page.ID,
		CreatedAt: page.CreatedAt,
		UpdatedAt: page.UpdatedAt,
		URL:       page.URL,
	}

	// Extract properties
	if nameProperty, exists := page.Properties["Name"]; exists {
		result.Title = ExtractTitleFromTitle(nameProperty)
	}

	if contentProperty, exists := page.Properties["Content"]; exists {
		result.Content = ExtractTextFromRichText(contentProperty)
	}

	if typeProperty, exists := page.Properties["Type"]; exists {
		result.Type = ExtractSelectValue(typeProperty)
	}

	if statusProperty, exists := page.Properties["Status"]; exists {
		result.Status = ExtractSelectValue(statusProperty)
	}

	if sourceProperty, exists := page.Properties["Source"]; exists {
		result.Source = ExtractTextFromRichText(sourceProperty)
	}

	return result
}
