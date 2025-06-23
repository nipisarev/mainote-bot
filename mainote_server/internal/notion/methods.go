package notion

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// CreateNote creates a new note in the Notion database
func (c *NotionClient) CreateNote(title, content, noteType, source string) (*Page, error) {
	// Truncate title if it's too long
	if len(title) > 50 {
		title = title[:50] + "..."
	}

	// If no title provided, create one from content
	if title == "" {
		if len(content) > 50 {
			title = content[:50] + "..."
		} else {
			title = content
		}
	}

	// Validate note type
	validTypes := map[string]bool{
		"idea":     true,
		"task":     true,
		"personal": true,
	}
	if !validTypes[noteType] {
		noteType = "task" // default type
	}

	// Create page properties
	properties := map[string]interface{}{
		"Name": TitleProperty{
			Title: []TextContent{
				{
					Text: PlainText{
						Content: title,
					},
				},
			},
		},
		"Type": SelectProperty{
			Select: SelectOption{
				Name: noteType,
			},
		},
		"Status": SelectProperty{
			Select: SelectOption{
				Name: "active",
			},
		},
		"Source": RichTextProperty{
			RichText: []TextContent{
				{
					Text: PlainText{
						Content: source,
					},
				},
			},
		},
		"Content": RichTextProperty{
			RichText: []TextContent{
				{
					Text: PlainText{
						Content: content,
					},
				},
			},
		},
		"Created": DateProperty{
			Date: DateValue{
				Start: time.Now().Format("2006-01-02"),
			},
		},
	}

	request := CreatePageRequest{
		Parent: Parent{
			DatabaseID: c.databaseID,
		},
		Properties: properties,
	}

	resp, err := c.makeRequest("POST", "/pages", request)
	if err != nil {
		return nil, fmt.Errorf("failed to make create page request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("notion API returned status %d", resp.StatusCode)
	}

	var page Page
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &page, nil
}

// UpdateNoteType updates the type of a note in Notion
func (c *NotionClient) UpdateNoteType(pageID, noteType string) error {
	// Validate note type
	validTypes := map[string]bool{
		"idea":     true,
		"task":     true,
		"personal": true,
	}
	if !validTypes[noteType] {
		return fmt.Errorf("invalid note type: %s", noteType)
	}

	properties := map[string]interface{}{
		"Type": SelectProperty{
			Select: SelectOption{
				Name: noteType,
			},
		},
	}

	request := UpdatePageRequest{
		Properties: properties,
	}

	resp, err := c.makeRequest("PATCH", "/pages/"+pageID, request)
	if err != nil {
		return fmt.Errorf("failed to make update page request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notion API returned status %d", resp.StatusCode)
	}

	return nil
}

// GetActiveTasks queries the Notion database for active tasks
func (c *NotionClient) GetActiveTasks() ([]Page, error) {
	filter := &Filter{
		And: []FilterCondition{
			{
				Property: "Status",
				Select: &SelectFilter{
					Equals: "active",
				},
			},
		},
	}

	request := QueryDatabaseRequest{
		Filter: filter,
	}

	resp, err := c.makeRequest("POST", "/databases/"+c.databaseID+"/query", request)
	if err != nil {
		return nil, fmt.Errorf("failed to make query database request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("notion API returned status %d", resp.StatusCode)
	}

	var response QueryDatabaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Results, nil
}

// FormatMorningNotification formats active tasks into a morning notification message
func (c *NotionClient) FormatMorningNotification(tasks []Page) string {
	if len(tasks) == 0 {
		return "Доброе утро! У вас нет активных задач на сегодня. Хорошего дня! 🌞"
	}

	message := "🌅 Доброе утро! Вот ваш план на сегодня:\n\n"

	for i, task := range tasks {
		title := "Без названия"
		noteType := "task"

		// Extract title from properties
		if nameProperty, ok := task.Properties["Name"].(map[string]interface{}); ok {
			if titleArray, ok := nameProperty["title"].([]interface{}); ok && len(titleArray) > 0 {
				if titleObj, ok := titleArray[0].(map[string]interface{}); ok {
					if textObj, ok := titleObj["text"].(map[string]interface{}); ok {
						if content, ok := textObj["content"].(string); ok {
							title = content
						}
					}
				}
			}
		}

		// Extract type from properties
		if typeProperty, ok := task.Properties["Type"].(map[string]interface{}); ok {
			if selectObj, ok := typeProperty["select"].(map[string]interface{}); ok {
				if name, ok := selectObj["name"].(string); ok {
					noteType = name
				}
			}
		}

		// Get emoji for task type
		emoji := "📝"
		switch noteType {
		case "idea":
			emoji = "💡"
		case "task":
			emoji = "✅"
		case "personal":
			emoji = "🏖"
		}

		message += fmt.Sprintf("%d. %s %s\n", i+1, emoji, title)
	}

	message += "\nУдачного и продуктивного дня! 💪"
	return message
}

// GetPageByID retrieves a specific page by its ID
func (c *NotionClient) GetPageByID(pageID string) (*Page, error) {
	resp, err := c.makeRequest("GET", "/pages/"+pageID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make get page request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("page not found")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("notion API returned status %d", resp.StatusCode)
	}

	var page Page
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &page, nil
}

// UpdatePageStatus updates the status of a page
func (c *NotionClient) UpdatePageStatus(pageID, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"active":   true,
		"done":     true,
		"archived": true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	properties := map[string]interface{}{
		"Status": SelectProperty{
			Select: SelectOption{
				Name: status,
			},
		},
	}

	request := UpdatePageRequest{
		Properties: properties,
	}

	resp, err := c.makeRequest("PATCH", "/pages/"+pageID, request)
	if err != nil {
		return fmt.Errorf("failed to make update page request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notion API returned status %d", resp.StatusCode)
	}

	return nil
}

// ExtractTextFromRichText extracts plain text from Notion rich text property
func ExtractTextFromRichText(richTextProperty interface{}) string {
	if richTextArray, ok := richTextProperty.([]interface{}); ok {
		var texts []string
		for _, item := range richTextArray {
			if textObj, ok := item.(map[string]interface{}); ok {
				if text, ok := textObj["text"].(map[string]interface{}); ok {
					if content, ok := text["content"].(string); ok {
						texts = append(texts, content)
					}
				}
			}
		}
		return strings.Join(texts, "")
	}
	return ""
}

// ExtractTitleFromTitle extracts plain text from Notion title property
func ExtractTitleFromTitle(titleProperty interface{}) string {
	if titleArray, ok := titleProperty.([]interface{}); ok && len(titleArray) > 0 {
		if titleObj, ok := titleArray[0].(map[string]interface{}); ok {
			if textObj, ok := titleObj["text"].(map[string]interface{}); ok {
				if content, ok := textObj["content"].(string); ok {
					return content
				}
			}
		}
	}
	return ""
}

// ExtractSelectValue extracts the name from a Notion select property
func ExtractSelectValue(selectProperty interface{}) string {
	if selectObj, ok := selectProperty.(map[string]interface{}); ok {
		if selectValue, ok := selectObj["select"].(map[string]interface{}); ok {
			if name, ok := selectValue["name"].(string); ok {
				return name
			}
		}
	}
	return ""
}
