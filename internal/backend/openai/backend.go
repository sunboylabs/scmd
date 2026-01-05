// Package openai provides an OpenAI-compatible API backend
// Works with OpenAI, Together.ai, Groq, Anthropic, and other compatible APIs
package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/scmd/scmd/internal/backend"
)

// Backend implements an OpenAI-compatible backend
type Backend struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

// Config for OpenAI-compatible backend
type Config struct {
	BaseURL string // API base URL
	APIKey  string // API key (or from env)
	Model   string // Model name
	Timeout time.Duration
}

// Presets for popular providers
var (
	OpenAIConfig = &Config{
		BaseURL: "https://api.openai.com/v1",
		Model:   "gpt-4o-mini",
	}
	TogetherConfig = &Config{
		BaseURL: "https://api.together.xyz/v1",
		Model:   "meta-llama/Llama-3.2-3B-Instruct-Turbo",
	}
	GroqConfig = &Config{
		BaseURL: "https://api.groq.com/openai/v1",
		Model:   "llama-3.1-8b-instant",
	}
)

// New creates a new OpenAI-compatible backend
func New(cfg *Config) *Backend {
	if cfg == nil {
		cfg = OpenAIConfig
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Minute
	}

	// Try to get API key from environment if not provided
	apiKey := cfg.APIKey
	if apiKey == "" {
		// Check various environment variables
		for _, env := range []string{"OPENAI_API_KEY", "TOGETHER_API_KEY", "GROQ_API_KEY", "LLM_API_KEY"} {
			if key := os.Getenv(env); key != "" {
				apiKey = key
				break
			}
		}
	}

	return &Backend{
		baseURL: cfg.BaseURL,
		apiKey:  apiKey,
		model:   cfg.Model,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// NewOpenAI creates an OpenAI backend
func NewOpenAI(apiKey string) *Backend {
	cfg := *OpenAIConfig
	cfg.APIKey = apiKey
	return New(&cfg)
}

// NewTogether creates a Together.ai backend
func NewTogether(apiKey string) *Backend {
	cfg := *TogetherConfig
	cfg.APIKey = apiKey
	return New(&cfg)
}

// NewGroq creates a Groq backend
func NewGroq(apiKey string) *Backend {
	cfg := *GroqConfig
	cfg.APIKey = apiKey
	return New(&cfg)
}

// Name returns the backend name
func (b *Backend) Name() string {
	// Derive name from baseURL
	if strings.Contains(b.baseURL, "openai.com") {
		return "openai"
	}
	if strings.Contains(b.baseURL, "together") {
		return "together"
	}
	if strings.Contains(b.baseURL, "groq") {
		return "groq"
	}
	return "openai-compatible"
}

// Type returns the backend type
func (b *Backend) Type() backend.Type {
	return backend.TypeOpenAI
}

// Initialize initializes the backend
func (b *Backend) Initialize(_ context.Context) error {
	if b.apiKey == "" {
		return fmt.Errorf("API key required (set OPENAI_API_KEY, TOGETHER_API_KEY, or GROQ_API_KEY)")
	}
	return nil
}

// IsAvailable checks if the backend is configured
func (b *Backend) IsAvailable(_ context.Context) (bool, error) {
	return b.apiKey != "", nil
}

// Shutdown shuts down the backend
func (b *Backend) Shutdown(_ context.Context) error {
	return nil
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest is the OpenAI chat completion request
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream"`
	Stop        []string      `json:"stop,omitempty"`
}

// chatResponse is the OpenAI chat completion response
type chatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// streamChunk is a streaming response chunk
type streamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
}

// Complete performs a non-streaming completion
func (b *Backend) Complete(ctx context.Context, req *backend.CompletionRequest) (*backend.CompletionResponse, error) {
	messages := []ChatMessage{}

	if req.SystemPrompt != "" {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: req.Prompt,
	})

	chatReq := chatRequest{
		Model:       b.model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      false,
		Stop:        req.StopSequences,
	}

	if chatReq.MaxTokens == 0 {
		chatReq.MaxTokens = 2048
	}

	body, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+b.apiKey)

	resp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	finishReason := backend.FinishComplete
	switch chatResp.Choices[0].FinishReason {
	case "length":
		finishReason = backend.FinishLength
	case "stop":
		finishReason = backend.FinishStop
	}

	return &backend.CompletionResponse{
		Content:      chatResp.Choices[0].Message.Content,
		TokensUsed:   chatResp.Usage.TotalTokens,
		FinishReason: finishReason,
	}, nil
}

// Stream performs a streaming completion
func (b *Backend) Stream(ctx context.Context, req *backend.CompletionRequest) (<-chan backend.StreamChunk, error) {
	messages := []ChatMessage{}

	if req.SystemPrompt != "" {
		messages = append(messages, ChatMessage{
			Role:    "system",
			Content: req.SystemPrompt,
		})
	}

	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: req.Prompt,
	})

	chatReq := chatRequest{
		Model:       b.model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      true,
		Stop:        req.StopSequences,
	}

	if chatReq.MaxTokens == 0 {
		chatReq.MaxTokens = 2048
	}

	body, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+b.apiKey)

	resp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	ch := make(chan backend.StreamChunk)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			// Remove "data: " prefix
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")

			// Check for stream end
			if data == "[DONE]" {
				ch <- backend.StreamChunk{Done: true}
				return
			}

			var chunk streamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				ch <- backend.StreamChunk{Error: err}
				return
			}

			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				done := chunk.Choices[0].FinishReason != ""

				ch <- backend.StreamChunk{
					Content: content,
					Done:    done,
				}

				if done {
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- backend.StreamChunk{Error: err}
		}
	}()

	return ch, nil
}

// SupportsToolCalling returns true for OpenAI
func (b *Backend) SupportsToolCalling() bool {
	return strings.Contains(b.baseURL, "openai.com")
}

// CompleteWithTools performs completion with tool calling (OpenAI only)
func (b *Backend) CompleteWithTools(_ context.Context, _ *backend.ToolRequest) (*backend.ToolResponse, error) {
	return nil, fmt.Errorf("tool calling not yet implemented")
}

// ModelInfo returns model information
func (b *Backend) ModelInfo() *backend.ModelInfo {
	return &backend.ModelInfo{
		Name:          b.model,
		ContextLength: 128000, // Varies by model
		Capabilities:  []string{"text", "code", "chat"},
	}
}

// EstimateTokens estimates token count
func (b *Backend) EstimateTokens(text string) int {
	// Rough estimate: ~4 characters per token
	return len(text) / 4
}

// SetModel changes the active model
func (b *Backend) SetModel(model string) {
	b.model = model
}

// SetAPIKey sets the API key
func (b *Backend) SetAPIKey(key string) {
	b.apiKey = key
}
