// Package ollama provides an Ollama LLM backend
package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/scmd/scmd/internal/backend"
)

// Backend implements the Ollama backend
type Backend struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// Config for Ollama backend
type Config struct {
	BaseURL string // Default: http://localhost:11434
	Model   string // Default: llama3.2 or qwen2.5-coder
	Timeout time.Duration
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		BaseURL: "http://localhost:11434",
		Model:   "qwen2.5-coder:1.5b",
		Timeout: 5 * time.Minute,
	}
}

// New creates a new Ollama backend
func New(cfg *Config) *Backend {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:11434"
	}
	if cfg.Model == "" {
		cfg.Model = "qwen2.5-coder:1.5b"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Minute
	}

	return &Backend{
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// Name returns the backend name
func (b *Backend) Name() string {
	return "ollama"
}

// Type returns the backend type
func (b *Backend) Type() backend.Type {
	return backend.TypeOllama
}

// Initialize initializes the backend
func (b *Backend) Initialize(ctx context.Context) error {
	// Check if Ollama is running
	available, err := b.IsAvailable(ctx)
	if err != nil {
		return fmt.Errorf("ollama not available: %w", err)
	}
	if !available {
		return fmt.Errorf("ollama server not responding at %s", b.baseURL)
	}
	return nil
}

// IsAvailable checks if Ollama is running
func (b *Backend) IsAvailable(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", b.baseURL+"/api/tags", nil)
	if err != nil {
		return false, err
	}

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return false, nil // Not available, but not an error
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// Shutdown shuts down the backend
func (b *Backend) Shutdown(_ context.Context) error {
	return nil
}

// generateRequest is the Ollama generate API request
type generateRequest struct {
	Model    string         `json:"model"`
	Prompt   string         `json:"prompt"`
	System   string         `json:"system,omitempty"`
	Stream   bool           `json:"stream"`
	Options  map[string]any `json:"options,omitempty"`
}

// generateResponse is the Ollama generate API response
type generateResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	Context   []int  `json:"context,omitempty"`
	CreatedAt string `json:"created_at"`

	// Timing info (only in final response)
	TotalDuration      int64 `json:"total_duration,omitempty"`
	LoadDuration       int64 `json:"load_duration,omitempty"`
	PromptEvalCount    int   `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64 `json:"prompt_eval_duration,omitempty"`
	EvalCount          int   `json:"eval_count,omitempty"`
	EvalDuration       int64 `json:"eval_duration,omitempty"`
}

// Complete performs a non-streaming completion
func (b *Backend) Complete(ctx context.Context, req *backend.CompletionRequest) (*backend.CompletionResponse, error) {
	ollamaReq := generateRequest{
		Model:  b.model,
		Prompt: req.Prompt,
		System: req.SystemPrompt,
		Stream: false,
		Options: map[string]any{
			"temperature": req.Temperature,
		},
	}

	if req.MaxTokens > 0 {
		ollamaReq.Options["num_predict"] = req.MaxTokens
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &backend.CompletionResponse{
		Content:      ollamaResp.Response,
		TokensUsed:   ollamaResp.EvalCount + ollamaResp.PromptEvalCount,
		FinishReason: backend.FinishComplete,
		Timing: &backend.Timing{
			PromptMS:     ollamaResp.PromptEvalDuration / 1_000_000,
			CompletionMS: ollamaResp.EvalDuration / 1_000_000,
			TokensPerSec: float64(ollamaResp.EvalCount) / (float64(ollamaResp.EvalDuration) / 1e9),
		},
	}, nil
}

// Stream performs a streaming completion
func (b *Backend) Stream(ctx context.Context, req *backend.CompletionRequest) (<-chan backend.StreamChunk, error) {
	ollamaReq := generateRequest{
		Model:  b.model,
		Prompt: req.Prompt,
		System: req.SystemPrompt,
		Stream: true,
		Options: map[string]any{
			"temperature": req.Temperature,
		},
	}

	if req.MaxTokens > 0 {
		ollamaReq.Options["num_predict"] = req.MaxTokens
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", b.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ollama error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	ch := make(chan backend.StreamChunk)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			var chunk generateResponse
			if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
				ch <- backend.StreamChunk{Error: err}
				return
			}

			ch <- backend.StreamChunk{
				Content: chunk.Response,
				Done:    chunk.Done,
			}

			if chunk.Done {
				return
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- backend.StreamChunk{Error: err}
		}
	}()

	return ch, nil
}

// SupportsToolCalling returns false (Ollama basic API doesn't support tools)
func (b *Backend) SupportsToolCalling() bool {
	return false
}

// CompleteWithTools is not supported
func (b *Backend) CompleteWithTools(_ context.Context, _ *backend.ToolRequest) (*backend.ToolResponse, error) {
	return nil, fmt.Errorf("tool calling not supported by Ollama backend")
}

// ModelInfo returns model information
func (b *Backend) ModelInfo() *backend.ModelInfo {
	return &backend.ModelInfo{
		Name:          b.model,
		ContextLength: 8192, // Default, varies by model
		Capabilities:  []string{"text", "code"},
	}
}

// EstimateTokens estimates token count (rough approximation)
func (b *Backend) EstimateTokens(text string) int {
	// Rough estimate: ~4 characters per token
	return len(text) / 4
}

// SetModel changes the active model
func (b *Backend) SetModel(model string) {
	b.model = model
}

// ListModels returns available models from Ollama
func (b *Backend) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", b.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		models[i] = m.Name
	}

	return models, nil
}
