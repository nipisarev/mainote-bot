package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog/log"
	"github.com/sashabaranov/go-openai"

	"mainote-server/internal/domain"
)

// OpenAIService implements domain.AIService using OpenAI API
type OpenAIService struct {
	client *openai.Client
	model  string
}

// NewOpenAIService creates a new OpenAI service instance
func NewOpenAIService(apiKey string) domain.AIService {
	client := openai.NewClient(apiKey)
	return &OpenAIService{
		client: client,
		model:  "gpt-4o-mini", // Using GPT-4o mini as requested
	}
}

// GenerateTitle generates a title for a note based on its content
func (s *OpenAIService) GenerateTitle(ctx context.Context, req domain.TitleGenerationRequest) (*domain.TitleGenerationResponse, error) {
	// Prepare the content for title generation
	content := req.Content
	if req.Transcription != nil && *req.Transcription != "" {
		content = fmt.Sprintf("Transcription: %s\nNote: %s", *req.Transcription, req.Content)
	}

	// Define the function for title generation
	titleFunction := openai.FunctionDefinition{
		Name:        "generate_note_title",
		Description: "Generate an appropriate title for a note based on its content, transcription, category and source",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "A concise, descriptive title for the note (3-8 words)",
				},
				"confidence": map[string]interface{}{
					"type":        "number",
					"description": "Confidence level in the generated title (0.0 to 1.0)",
					"minimum":     0.0,
					"maximum":     1.0,
				},
				"reasoning": map[string]interface{}{
					"type":        "string",
					"description": "Brief explanation of why this title was chosen",
				},
			},
			"required": []string{"title", "confidence", "reasoning"},
		},
	}

	// Prepare system prompt
	systemPrompt := `You are an AI assistant specialized in generating concise, meaningful titles for notes. 
Your task is to analyze the provided content and generate an appropriate title that:
1. Is 3-8 words long
2. Captures the main topic or action
3. Is clear and descriptive
4. Uses Russian language for Russian content, English for English content
5. Considers the source and category of the note

Always use the generate_note_title function to provide your response with confidence and reasoning.`

	// Prepare user prompt
	userPrompt := fmt.Sprintf(`Please generate a title for this note:

Content: %s
Category: %s
Source: %s

Generate an appropriate title using the generate_note_title function.`, content, req.Category, req.Source)

	// Make the API call
	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Functions: []openai.FunctionDefinition{titleFunction},
		FunctionCall: &openai.FunctionCall{
			Name: "generate_note_title",
		},
		Temperature: 0.3, // Lower temperature for more consistent results
		MaxTokens:   200,
	})

	if err != nil {
		sentry.CaptureException(err)
		// Check for specific error types
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "quota") {
			log.Error().Err(err).Msg("OpenAI quota exceeded - please check billing")
			return nil, fmt.Errorf("OpenAI quota exceeded: %w", err)
		}
		log.Error().Err(err).Msg("Failed to generate title with OpenAI")
		return nil, fmt.Errorf("failed to generate title: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	choice := resp.Choices[0]
	if choice.Message.FunctionCall == nil {
		return nil, fmt.Errorf("no function call in response")
	}

	// Parse the function call response
	var result domain.TitleGenerationResponse
	if err := json.Unmarshal([]byte(choice.Message.FunctionCall.Arguments), &result); err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Str("arguments", choice.Message.FunctionCall.Arguments).Msg("Failed to parse function call arguments")
		return nil, fmt.Errorf("failed to parse function call response: %w", err)
	}

	// Validate the result
	if result.Title == "" {
		return nil, fmt.Errorf("empty title generated")
	}

	// Clean and validate title
	result.Title = strings.TrimSpace(result.Title)
	if len(strings.Fields(result.Title)) > 8 {
		// Truncate if too long
		words := strings.Fields(result.Title)
		result.Title = strings.Join(words[:8], " ")
	}

	log.Info().
		Str("title", result.Title).
		Float64("confidence", result.Confidence).
		Str("reasoning", result.Reasoning).
		Msg("Generated note title")

	return &result, nil
}

// CallFunction executes a function call with AI assistance
func (s *OpenAIService) CallFunction(ctx context.Context, conversation domain.AIConversation, availableFunctions []domain.AIFunction) (*domain.AIMessage, error) {
	// Convert domain messages to OpenAI format
	messages := make([]openai.ChatCompletionMessage, len(conversation.Messages))
	for i, msg := range conversation.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Convert domain functions to OpenAI format
	functions := make([]openai.FunctionDefinition, len(availableFunctions))
	for i, fn := range availableFunctions {
		functions[i] = openai.FunctionDefinition{
			Name:        fn.Name,
			Description: fn.Description,
			Parameters:  fn.Parameters,
		}
	}

	// Make the API call
	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    messages,
		Functions:   functions,
		Temperature: 0.7,
		MaxTokens:   1000,
	})

	if err != nil {
		sentry.CaptureException(err)
		log.Error().Err(err).Msg("Failed to call OpenAI function")
		return nil, fmt.Errorf("failed to call function: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	choice := resp.Choices[0]
	result := &domain.AIMessage{
		Role:    choice.Message.Role,
		Content: choice.Message.Content,
	}

	// Handle function call if present
	if choice.Message.FunctionCall != nil {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(choice.Message.FunctionCall.Arguments), &args); err != nil {
			sentry.CaptureException(err)
			log.Error().Err(err).Msg("Failed to parse function call arguments")
			return nil, fmt.Errorf("failed to parse function call arguments: %w", err)
		}

		result.FunctionCall = &domain.AIFunctionCall{
			Name:      choice.Message.FunctionCall.Name,
			Arguments: args,
		}
	}

	return result, nil
}
