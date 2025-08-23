package usecase

import (
	"context"
	"fmt"

	"mainote-server/internal/domain"
)

// aiUsecase implements domain.AIUsecase
type aiUsecase struct {
	aiService domain.AIService
}

// NewAIUsecase creates a new AI usecase instance
func NewAIUsecase(aiService domain.AIService) domain.AIUsecase {
	return &aiUsecase{
		aiService: aiService,
	}
}

// GenerateNoteTitle generates an appropriate title for a note
func (uc *aiUsecase) GenerateNoteTitle(ctx context.Context, note *domain.Note) (string, error) {
	if note == nil {
		return "", fmt.Errorf("note cannot be nil")
	}

	// Prepare the request
	req := domain.TitleGenerationRequest{
		Content:  note.Content,
		Category: note.Category,
		Source:   note.Source,
	}

	// Include transcription if available
	if note.Transcription != nil {
		req.Transcription = note.Transcription
	}

	// Generate title using AI service
	response, err := uc.aiService.GenerateTitle(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate title: %w", err)
	}

	if response == nil || response.Title == "" {
		return "", fmt.Errorf("empty title generated")
	}

	return response.Title, nil
}

// ProcessWithAI processes content using AI with available functions
func (uc *aiUsecase) ProcessWithAI(ctx context.Context, content string, functions []domain.AIFunction) (*domain.AIMessage, error) {
	if content == "" {
		return nil, fmt.Errorf("content cannot be empty")
	}

	// Create conversation with the content
	conversation := domain.AIConversation{
		Messages: []domain.AIMessage{
			{
				Role:    "user",
				Content: content,
			},
		},
	}

	// Call AI service with functions
	result, err := uc.aiService.CallFunction(ctx, conversation, functions)
	if err != nil {
		return nil, fmt.Errorf("failed to process with AI: %w", err)
	}

	return result, nil
}
