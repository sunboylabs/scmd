package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/scmd/scmd/internal/backend"
)

func TestNew(t *testing.T) {
	b := New(nil)
	assert.NotNil(t, b)
	assert.Equal(t, backend.TypeOpenAI, b.Type())
}

func TestNewOpenAI(t *testing.T) {
	b := NewOpenAI("test-key")
	assert.Equal(t, "openai", b.Name())
	assert.Equal(t, "gpt-4o-mini", b.model)
}

func TestNewTogether(t *testing.T) {
	b := NewTogether("test-key")
	assert.Equal(t, "together", b.Name())
}

func TestNewGroq(t *testing.T) {
	b := NewGroq("test-key")
	assert.Equal(t, "groq", b.Name())
}

func TestBackend_Name(t *testing.T) {
	tests := []struct {
		baseURL  string
		expected string
	}{
		{"https://api.openai.com/v1", "openai"},
		{"https://api.together.xyz/v1", "together"},
		{"https://api.groq.com/openai/v1", "groq"},
		{"https://custom.api.com/v1", "openai-compatible"},
	}

	for _, tt := range tests {
		b := New(&Config{BaseURL: tt.baseURL, APIKey: "test"})
		assert.Equal(t, tt.expected, b.Name())
	}
}

func TestBackend_ModelInfo(t *testing.T) {
	b := New(nil)
	info := b.ModelInfo()
	assert.NotNil(t, info)
	assert.NotEmpty(t, info.Name)
}

func TestBackend_EstimateTokens(t *testing.T) {
	b := New(nil)
	tokens := b.EstimateTokens("hello world test")
	assert.Greater(t, tokens, 0)
}

func TestBackend_SetModel(t *testing.T) {
	b := New(nil)
	b.SetModel("new-model")
	assert.Equal(t, "new-model", b.model)
}

func TestBackend_SetAPIKey(t *testing.T) {
	b := New(nil)
	b.SetAPIKey("new-key")
	assert.Equal(t, "new-key", b.apiKey)
}

func TestBackend_SupportsToolCalling(t *testing.T) {
	openai := NewOpenAI("test")
	assert.True(t, openai.SupportsToolCalling())

	groq := NewGroq("test")
	assert.False(t, groq.SupportsToolCalling())
}
