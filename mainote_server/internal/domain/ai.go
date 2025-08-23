package domain

import (
	"context"
)

// AIFunction represents a function that can be called by AI
type AIFunction struct {
	Name        string
	Description string
	Parameters  interface{}
}

// AIFunctionCall represents a function call made by AI
type AIFunctionCall struct {
	Name      string
	Arguments map[string]interface{}
}

// AIMessage represents a message in AI conversation
type AIMessage struct {
	Role         string          `json:"role"`
	Content      string          `json:"content"`
	FunctionCall *AIFunctionCall `json:"function_call,omitempty"`
	Functions    []AIFunction    `json:"functions,omitempty"`
}

// AIConversation represents a conversation context for AI
type AIConversation struct {
	Messages []AIMessage `json:"messages"`
}

// TitleGenerationRequest represents request for generating note title
type TitleGenerationRequest struct {
	Content       string  `json:"content"`
	Transcription *string `json:"transcription,omitempty"`
	Category      string  `json:"category"`
	Source        string  `json:"source"`
}

// TitleGenerationResponse represents response from title generation
type TitleGenerationResponse struct {
	Title      string  `json:"title"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

// AIService defines the interface for AI operations
type AIService interface {
	// GenerateTitle generates a title for a note based on its content
	GenerateTitle(ctx context.Context, req TitleGenerationRequest) (*TitleGenerationResponse, error)

	// CallFunction executes a function call with AI assistance
	CallFunction(ctx context.Context, conversation AIConversation, availableFunctions []AIFunction) (*AIMessage, error)
}

// AIUsecase defines the interface for AI business logic
type AIUsecase interface {
	// GenerateNoteTitle generates an appropriate title for a note
	GenerateNoteTitle(ctx context.Context, note *Note) (string, error)

	// ProcessWithAI processes content using AI with available functions
	ProcessWithAI(ctx context.Context, content string, functions []AIFunction) (*AIMessage, error)
}
