package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// LLMClient handles communication with external LLM APIs
type LLMClient struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	provider   string
}

// LLMRequest represents a request to the LLM API
type LLMRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Message represents a single message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMResponse represents the response from the LLM API
type LLMResponse struct {
	Choices []Choice `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice represents a single response choice
type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// APIError represents an error from the LLM API
type APIError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// NewLLMClient creates a new LLM client
func NewLLMClient() (*LLMClient, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		// Fallback to OpenAI if Anthropic key not found
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("no LLM API key found in environment variables (ANTHROPIC_API_KEY or OPENAI_API_KEY)")
		}
	}

	baseURL := os.Getenv("LLM_API_URL")
	provider := os.Getenv("LLM_PROVIDER")
	
	// Set defaults based on available API key
	if provider == "" {
		if os.Getenv("ANTHROPIC_API_KEY") != "" {
			provider = "anthropic"
			if baseURL == "" {
				baseURL = "https://api.anthropic.com/v1/messages"
			}
		} else {
			provider = "openai"
			if baseURL == "" {
				baseURL = "https://api.openai.com/v1/chat/completions"
			}
		}
	}

	return &LLMClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:   apiKey,
		baseURL:  baseURL,
		provider: provider,
	}, nil
}

// GenerateChat generates a chat response using the configured LLM
func (c *LLMClient) GenerateChat(ctx context.Context, prompt string) (string, error) {
	if c.provider == "anthropic" {
		return c.generateChatAnthropic(ctx, prompt)
	}
	return c.generateChatOpenAI(ctx, prompt)
}

// generateChatOpenAI handles OpenAI API calls
func (c *LLMClient) generateChatOpenAI(ctx context.Context, prompt string) (string, error) {
	request := LLMRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Model:       "gpt-3.5-turbo",
		MaxTokens:   150, // Keep responses concise for chat
		Temperature: 0.7,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response LLMResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return response.Choices[0].Message.Content, nil
}

// generateChatAnthropic handles Anthropic Claude API calls
func (c *LLMClient) generateChatAnthropic(ctx context.Context, prompt string) (string, error) {
	// Anthropic has a different API format
	anthropicRequest := map[string]interface{}{
		"model":      "claude-3-haiku-20240307",
		"max_tokens": 150,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	requestBody, err := json.Marshal(anthropicRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Anthropic response format
	var anthropicResponse map[string]interface{}
	if err := json.Unmarshal(body, &anthropicResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract content from Anthropic response
	if content, ok := anthropicResponse["content"].([]interface{}); ok && len(content) > 0 {
		if textContent, ok := content[0].(map[string]interface{}); ok {
			if text, ok := textContent["text"].(string); ok {
				return text, nil
			}
		}
	}

	return "", fmt.Errorf("unexpected response format from Anthropic API")
}

// MockLLMClient provides a mock implementation for testing
type MockLLMClient struct {
	responses []string
	callCount int
}

// NewMockLLMClient creates a new mock LLM client with predefined responses
func NewMockLLMClient(responses []string) *MockLLMClient {
	return &MockLLMClient{
		responses: responses,
		callCount: 0,
	}
}

// GenerateChat returns a predefined mock response
func (m *MockLLMClient) GenerateChat(ctx context.Context, prompt string) (string, error) {
	if m.callCount >= len(m.responses) {
		return "I'm analyzing the situation and thinking strategically about our next move.", nil
	}
	
	response := m.responses[m.callCount]
	m.callCount++
	return response, nil
}