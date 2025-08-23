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
	if len(title) > 30 {
		title = title[:30] + "..."
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

// FormatNoteItem formats a single note item for display with index
func (f *MorningNotificationFormatter) FormatNoteItem(note domain.Note, index int) string {
	title := f.FormatNoteTitle(note)
	content := f.FormatNoteContent(note)

	var result strings.Builder

	if title != "" {
		result.WriteString(fmt.Sprintf("   %d. %s", index+1, title))
	} else {
		result.WriteString(fmt.Sprintf("   %d. %s", index+1, content))
	}

	// Extras
	extras := []string{}
	if note.DueAt != nil {
		extras = append(extras, fmt.Sprintf("⏰ %s", note.DueAt.Format("2006-01-02")))
	}
	if note.EffortMin > 0 {
		extras = append(extras, fmt.Sprintf("⏳ %dm", note.EffortMin))
	}
	if note.Priority != 0 {
		extras = append(extras, fmt.Sprintf("⭐ %d", note.Priority))
	}
	if len(extras) > 0 {
		result.WriteString("  (" + strings.Join(extras, " • ") + ")")
	}

	result.WriteString("\n")

	return result.String()
}

// FormatNoteItemCompact formats a single note item with cleaner, more compact styling
func (f *MorningNotificationFormatter) FormatNoteItemCompact(note domain.Note, index int) string {
	title := f.FormatNoteTitle(note)
	content := f.FormatNoteContent(note)

	var result strings.Builder

	if title != "" {
		result.WriteString(fmt.Sprintf("▸ %d. %s", index+1, title))
	} else {
		result.WriteString(fmt.Sprintf("▸ %d. %s", index+1, content))
	}

	extras := []string{}
	if note.DueAt != nil {
		extras = append(extras, fmt.Sprintf("⏰ %s", note.DueAt.Format("2006-01-02")))
	}
	if note.EffortMin > 0 {
		extras = append(extras, fmt.Sprintf("⏳ %dm", note.EffortMin))
	}
	if note.Priority != 0 {
		extras = append(extras, fmt.Sprintf("⭐ %d", note.Priority))
	}
	if len(extras) > 0 {
		result.WriteString("  (" + strings.Join(extras, " • ") + ")")
	}

	result.WriteString("\n")

	return result.String()
}

// FormatCategorySection formats a complete category section
func (f *MorningNotificationFormatter) FormatCategorySection(category string, notes []domain.Note) string {
	if len(notes) == 0 {
		return ""
	}

	var result strings.Builder

	// Category header with improved formatting
	emoji := f.GetCategoryEmoji(category)
	categoryTitle := strings.ToUpper(category[:1]) + category[1:]
	result.WriteString(fmt.Sprintf("┌─ %s %s (%d)\n", emoji, categoryTitle, len(notes)))

	// Format notes (limit to 5 per category)
	maxNotes := 5
	for i, note := range notes {
		if i >= maxNotes {
			result.WriteString(fmt.Sprintf("   │ ...and %d more\n", len(notes)-maxNotes))
			break
		}
		formattedNote := f.FormatNoteItem(note, i)
		// Add vertical line prefix for better visual hierarchy
		lines := strings.Split(strings.TrimSuffix(formattedNote, "\n"), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				result.WriteString(fmt.Sprintf("   │ %s\n", strings.TrimPrefix(line, "   ")))
			}
		}
	}

	result.WriteString("   └─────────────────────────────────────────\n\n")
	return result.String()
}

// FormatCategorySectionCompact formats a complete category section with cleaner styling
func (f *MorningNotificationFormatter) FormatCategorySectionCompact(category string, notes []domain.Note) string {
	if len(notes) == 0 {
		return ""
	}

	var result strings.Builder

	// Category header with cleaner formatting
	emoji := f.GetCategoryEmoji(category)
	categoryTitle := strings.ToUpper(category[:1]) + category[1:]
	result.WriteString(fmt.Sprintf("━━ %s %s (%d) ━━\n", emoji, categoryTitle, len(notes)))

	// Format notes (limit to 5 per category)
	maxNotes := 5
	for i, note := range notes {
		if i >= maxNotes {
			result.WriteString(fmt.Sprintf("▸ ...and %d more\n", len(notes)-maxNotes))
			break
		}
		result.WriteString(f.FormatNoteItemCompact(note, i))
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

	// Process categories in defined order using the new compact formatting
	categoryOrder := f.GetCategoryOrder()
	for _, categoryInfo := range categoryOrder {
		if categoryNotes, exists := notesByCategory[categoryInfo.Name]; exists {
			message.WriteString(f.FormatCategorySectionCompact(categoryInfo.Name, categoryNotes))
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
			message.WriteString(f.FormatCategorySectionCompact(category, categoryNotes))
		}
	}

	// Add closing message
	message.WriteString("Have a productive day! 🚀")

	return message.String()
}

// GenerateNotesMap generates a map of note numbers to note UUIDs for interactive browsing
func (f *MorningNotificationFormatter) GenerateNotesMap(notes []domain.Note) map[int]string {
	notesMap := make(map[int]string)

	log.Info().
		Int("input_notes_count", len(notes)).
		Msg("Starting to generate notes map")

	// Group notes by category first
	notesByCategory := f.GroupNotesByCategory(notes)

	log.Info().
		Int("categories_count", len(notesByCategory)).
		Interface("categories", func() map[string]int {
			counts := make(map[string]int)
			for cat, catNotes := range notesByCategory {
				counts[cat] = len(catNotes)
			}
			return counts
		}()).
		Msg("Grouped notes by category")

	// Process categories in defined order
	categoryOrder := f.GetCategoryOrder()
	noteNumber := 1

	for _, categoryInfo := range categoryOrder {
		if categoryNotes, exists := notesByCategory[categoryInfo.Name]; exists {
			log.Info().
				Str("category", categoryInfo.Name).
				Int("category_notes_count", len(categoryNotes)).
				Int("starting_note_number", noteNumber).
				Msg("Processing category")

			// Limit to 5 notes per category (matching the display limit)
			maxNotes := 5
			for i, note := range categoryNotes {
				if i >= maxNotes {
					break
				}
				notesMap[noteNumber] = note.NoteID.String()
				log.Debug().
					Int("note_number", noteNumber).
					Str("note_id", note.NoteID.String()).
					Str("category", categoryInfo.Name).
					Msg("Added note to map")
				noteNumber++
			}
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
			log.Info().
				Str("category", category).
				Int("category_notes_count", len(categoryNotes)).
				Int("starting_note_number", noteNumber).
				Msg("Processing additional category")

			maxNotes := 5
			for i, note := range categoryNotes {
				if i >= maxNotes {
					break
				}
				notesMap[noteNumber] = note.NoteID.String()
				log.Debug().
					Int("note_number", noteNumber).
					Str("note_id", note.NoteID.String()).
					Str("category", category).
					Msg("Added note to map")
				noteNumber++
			}
		}
	}

	log.Info().
		Int("final_notes_map_size", len(notesMap)).
		Interface("final_notes_map", notesMap).
		Msg("Completed generating notes map")

	return notesMap
}

// GenerateMessageWithNotesMap generates the complete morning notification message with notes map
func (f *MorningNotificationFormatter) GenerateMessageWithNotesMap(ctx context.Context, userWithSettings domain.UserWithSettings) (string, map[int]string) {
	// Fetch active notes for the user
	activeStatus := "active"
	notes, err := f.noteRepo.GetNotesForUser(ctx, userWithSettings.Settings.ChatID, nil, &activeStatus, 50, 0)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userWithSettings.User.ID.String()).
			Str("chat_id", userWithSettings.Settings.ChatID).
			Msg("Failed to fetch notes for morning notification")
		return "This is your daily morning notification.", nil
	}

	if len(notes.Notes) == 0 {
		log.Info().
			Str("user_id", userWithSettings.User.ID.String()).
			Str("chat_id", userWithSettings.Settings.ChatID).
			Msg("No active notes found for user")
		return "🌅 Good morning! Hope you have a great day ahead!\n\nYou have no active notes. Have a great day!", nil
	}

	log.Info().
		Str("user_id", userWithSettings.User.ID.String()).
		Str("chat_id", userWithSettings.Settings.ChatID).
		Int("notes_count", len(notes.Notes)).
		Msg("Fetched notes for morning notification")

	// Generate the message
	message := f.GenerateMessage(ctx, userWithSettings)

	// Generate the notes map
	notesMap := f.GenerateNotesMap(notes.Notes)

	log.Info().
		Str("user_id", userWithSettings.User.ID.String()).
		Int("notes_map_size", len(notesMap)).
		Interface("notes_map", notesMap).
		Msg("Generated notes map")

	return message, notesMap
}
