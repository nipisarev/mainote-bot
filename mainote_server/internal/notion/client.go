package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// NotionClient represents a client for interacting with Notion API
type NotionClient struct {
	apiKey     string
	databaseID string
	httpClient *http.Client
	baseURL    string
}

// NotionConfig holds configuration for Notion client
type NotionConfig struct {
	APIKey     string
	DatabaseID string
}

// NewNotionClient creates a new Notion client instance
func NewNotionClient(config NotionConfig) *NotionClient {
	return &NotionClient{
		apiKey:     config.APIKey,
		databaseID: config.DatabaseID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.notion.com/v1",
	}
}

// Page represents a Notion page structure
type Page struct {
	ID         string                 `json:"id"`
	CreatedAt  time.Time              `json:"created_time"`
	UpdatedAt  time.Time              `json:"last_edited_time"`
	Properties map[string]interface{} `json:"properties"`
	URL        string                 `json:"url"`
}

// Property types for Notion pages
type TitleProperty struct {
	Title []TextContent `json:"title"`
}

type TextContent struct {
	Text PlainText `json:"text"`
}

type PlainText struct {
	Content string `json:"content"`
}

type SelectProperty struct {
	Select SelectOption `json:"select"`
}

type SelectOption struct {
	Name string `json:"name"`
}

type RichTextProperty struct {
	RichText []TextContent `json:"rich_text"`
}

type DateProperty struct {
	Date DateValue `json:"date"`
}

type DateValue struct {
	Start string `json:"start"`
}

// CreatePageRequest represents a request to create a new page
type CreatePageRequest struct {
	Parent     Parent                 `json:"parent"`
	Properties map[string]interface{} `json:"properties"`
}

type Parent struct {
	DatabaseID string `json:"database_id"`
}

// UpdatePageRequest represents a request to update a page
type UpdatePageRequest struct {
	Properties map[string]interface{} `json:"properties"`
}

// QueryDatabaseRequest represents a request to query a database
type QueryDatabaseRequest struct {
	Filter *Filter `json:"filter,omitempty"`
}

type Filter struct {
	And []FilterCondition `json:"and,omitempty"`
	Or  []FilterCondition `json:"or,omitempty"`
}

type FilterCondition struct {
	Property string      `json:"property"`
	Select   *SelectFilter `json:"select,omitempty"`
}

type SelectFilter struct {
	Equals string `json:"equals"`
}

// QueryDatabaseResponse represents a response from querying a database
type QueryDatabaseResponse struct {
	Results []Page `json:"results"`
	HasMore bool   `json:"has_more"`
}

// makeRequest makes an HTTP request to Notion API
func (c *NotionClient) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var requestBody []byte
	var err error

	if body != nil {
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	url := c.baseURL + endpoint
	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	return c.httpClient.Do(req)
}
