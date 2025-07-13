package notifications

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	"mainote-server/internal/domain"
	"mainote-server/internal/repository"
)

// MorningNotificationFormatter handles formatting of morning notifications
type MorningNotificationFormatter struct {
	noteRepo repository.NoteRepository
}

// NewMorningNotificationFormatter creates a new formatter instance
func NewMorningNotificationFormatter(noteRepo repository.NoteRepository) *MorningNotificationFormatter {
	return &MorningNotificationFormatter{
		noteRepo: noteRepo,
	}
}

// CategoryInfo holds category display information
type CategoryInfo struct {
	Name  string
	Emoji string
	Order int
}

// GetCategoryOrder defines the order and display info for note categories
func (f *MorningNotificationFormatter) GetCategoryOrder() []CategoryInfo {
	return []CategoryInfo{
		{Name: "task", Emoji: "✅", Order: 1},
		{Name: "idea", Emoji: "💡", Order: 2},
		{Name: "personal", Emoji: "🏖", Order: 3},
		{Name: "work", Emoji: "💼", Order: 4},
		{Name: "general", Emoji: "📄", Order: 5},
	}
}

// GetCategoryEmoji returns the emoji for a given category
func (f *MorningNotificationFormatter) GetCategoryEmoji(category string) string {
	for _, info := range f.GetCategoryOrder() {
		if info.Name == category {
			return info.Emoji
		}
	}
	return "📄" // Default emoji
}

// FormatNoteTitle formats a note title with appropriate truncation
func (f *MorningNotificationFormatter) FormatNoteTitle(note domain.Note) string {
	title := ""

	// Use explicit title if available, otherwise extract from content
	if note.Title != nil && *note.Title != "" {
		title = *note.Title
	} else {
		// Extract first line from content as title
		lines := strings.Split(note.Content, "\n")
		if len(lines) > 0 && len(lines[0]) > 0 {
			title = lines[0]
		}
	}

	// Truncate title to reasonable length
	if len(title) > 40 {
		title = title[:40] + "..."
	}

	return title
}

// FormatNoteContent formats note content for preview
func (f *MorningNotificationFormatter) FormatNoteContent(note domain.Note) string {
	content := note.Content

	// Get first two lines of content for preview
	lines := strings.Split(content, "\n")
	if len(lines) > 2 {
		content = strings.Join(lines[:2], "\n")
	}

	// Truncate content if too long
	if len(content) > 80 {
		content = content[:80] + "..."
	}

	return content
}

// ShouldShowContentPreview determines if content preview should be shown
func (f *MorningNotificationFormatter) ShouldShowContentPreview(title, content string) bool {
	return content != title && content != "" && !strings.HasPrefix(content, title)
}

// FormatNoteItem formats a single note item for display
func (f *MorningNotificationFormatter) FormatNoteItem(note domain.Note) string {
	title := f.FormatNoteTitle(note)
	content := f.FormatNoteContent(note)

	var result strings.Builder

	if title != "" {
		result.WriteString(fmt.Sprintf("   • %s\n", title))
		// Only show content preview if it provides additional info
		if f.ShouldShowContentPreview(title, content) {
			result.WriteString(fmt.Sprintf("     %s\n", content))
		}
	} else {
		result.WriteString(fmt.Sprintf("   • %s\n", content))
	}

	return result.String()
}

// FormatCategorySection formats a complete category section
func (f *MorningNotificationFormatter) FormatCategorySection(category string, notes []domain.Note) string {
	if len(notes) == 0 {
		return ""
	}

	var result strings.Builder

	// Category header
	emoji := f.GetCategoryEmoji(category)
	categoryTitle := strings.ToUpper(category[:1]) + category[1:]
	result.WriteString(fmt.Sprintf("%s %s (%d):\n", emoji, categoryTitle, len(notes)))

	// Format notes (limit to 5 per category)
	maxNotes := 5
	for i, note := range notes {
		if i >= maxNotes {
			result.WriteString(fmt.Sprintf("   ...and %d more\n", len(notes)-maxNotes))
			break
		}
		result.WriteString(f.FormatNoteItem(note))
	}

	result.WriteString("\n")
	return result.String()
}

// GroupNotesByCategory groups notes by their category
func (f *MorningNotificationFormatter) GroupNotesByCategory(notes []domain.Note) map[string][]domain.Note {
	notesByCategory := make(map[string][]domain.Note)

	for _, note := range notes {
		category := note.Category
		if category == "" {
			category = "general"
		}
		notesByCategory[category] = append(notesByCategory[category], note)
	}

	return notesByCategory
}

// GenerateMessage generates the complete morning notification message
func (f *MorningNotificationFormatter) GenerateMessage(ctx context.Context, userWithSettings domain.UserWithSettings) string {
	// Start with greeting
	var message strings.Builder
	message.WriteString("🌅 Good morning! Hope you have a great day ahead!\n\n")

	// Fetch active notes for the user
	activeStatus := "active"
	notes, err := f.noteRepo.GetNotesForUser(ctx, userWithSettings.Settings.ChatID, nil, &activeStatus, 50, 0)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userWithSettings.User.ID.String()).
			Str("chat_id", userWithSettings.Settings.ChatID).
			Msg("Failed to fetch notes for morning notification")
		message.WriteString("This is your daily morning notification.")
		return message.String()
	}

	if len(notes.Notes) == 0 {
		message.WriteString("You have no active notes. Have a great day!")
		return message.String()
	}

	// Group notes by category
	notesByCategory := f.GroupNotesByCategory(notes.Notes)

	// Add notes section header
	message.WriteString("📝 Your active notes:\n\n")

	// Process categories in defined order
	categoryOrder := f.GetCategoryOrder()
	for _, categoryInfo := range categoryOrder {
		if categoryNotes, exists := notesByCategory[categoryInfo.Name]; exists {
			message.WriteString(f.FormatCategorySection(categoryInfo.Name, categoryNotes))
		}
	}

	// Add any remaining categories not in the predefined order
	for category, categoryNotes := range notesByCategory {
		found := false
		for _, categoryInfo := range categoryOrder {
			if category == categoryInfo.Name {
				found = true
				break
			}
		}

		if !found {
			message.WriteString(f.FormatCategorySection(category, categoryNotes))
		}
	}

	// Add closing message
	message.WriteString("Have a productive day! 🚀")

	return message.String()
}
