// Package mock provides a mock backend for testing
package mock

import (
	"context"

	"github.com/scmd/scmd/internal/backend"
)

// Backend is a mock LLM backend for testing
type Backend struct {
	response string
	err      error
	name     string
}

// New creates a new mock backend
func New() *Backend {
	return &Backend{
		name: "mock",
		response: "# Code Explanation\n\nThis is a **mock response** with markdown formatting.\n\n## Key Points\n\n- First important point\n- Second important point with `inline code`\n- Third point\n\n## Code Example\n\n```python\ndef example():\n    print(\"Hello, world!\")\n```\n\n> Note: This is a blockquote for emphasis.",
	}
}

// SetResponse sets the response to return
func (b *Backend) SetResponse(response string) {
	b.response = response
}

// SetError sets the error to return
func (b *Backend) SetError(err error) {
	b.err = err
}

// Name returns the backend name
func (b *Backend) Name() string {
	return b.name
}

// Type returns the backend type
func (b *Backend) Type() backend.Type {
	return backend.TypeMock
}

// Initialize initializes the backend
func (b *Backend) Initialize(_ context.Context) error {
	return nil
}

// IsAvailable returns true if the backend is available
func (b *Backend) IsAvailable(_ context.Context) (bool, error) {
	return true, nil
}

// Shutdown shuts down the backend
func (b *Backend) Shutdown(_ context.Context) error {
	return nil
}

// Complete performs a completion request
func (b *Backend) Complete(_ context.Context, _ *backend.CompletionRequest) (*backend.CompletionResponse, error) {
	if b.err != nil {
		return nil, b.err
	}
	return &backend.CompletionResponse{
		Content:      b.response,
		TokensUsed:   len(b.response) / 4,
		FinishReason: backend.FinishComplete,
	}, nil
}

// Stream returns a streaming response
func (b *Backend) Stream(_ context.Context, _ *backend.CompletionRequest) (<-chan backend.StreamChunk, error) {
	if b.err != nil {
		return nil, b.err
	}

	ch := make(chan backend.StreamChunk)
	go func() {
		defer close(ch)
		// Send response in chunks
		for i := 0; i < len(b.response); i += 10 {
			end := i + 10
			if end > len(b.response) {
				end = len(b.response)
			}
			ch <- backend.StreamChunk{Content: b.response[i:end]}
		}
		ch <- backend.StreamChunk{Done: true}
	}()
	return ch, nil
}

// SupportsToolCalling returns false for mock
func (b *Backend) SupportsToolCalling() bool {
	return false
}

// CompleteWithTools is not supported
func (b *Backend) CompleteWithTools(_ context.Context, _ *backend.ToolRequest) (*backend.ToolResponse, error) {
	return nil, nil
}

// ModelInfo returns mock model info
func (b *Backend) ModelInfo() *backend.ModelInfo {
	return &backend.ModelInfo{
		Name:          "mock-model",
		Size:          "0B",
		ContextLength: 8192,
		Capabilities:  []string{"text"},
	}
}

// EstimateTokens estimates tokens (rough approximation)
func (b *Backend) EstimateTokens(text string) int {
	return len(text) / 4
}
