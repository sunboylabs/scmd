package ollama

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/scmd/scmd/internal/backend"
)

func TestNew(t *testing.T) {
	b := New(nil)
	assert.NotNil(t, b)
	assert.Equal(t, "ollama", b.Name())
	assert.Equal(t, backend.TypeOllama, b.Type())
}

func TestNew_WithConfig(t *testing.T) {
	cfg := &Config{
		BaseURL: "http://custom:8080",
		Model:   "custom-model",
	}
	b := New(cfg)
	assert.Equal(t, "http://custom:8080", b.baseURL)
	assert.Equal(t, "custom-model", b.model)
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "http://localhost:11434", cfg.BaseURL)
	assert.NotEmpty(t, cfg.Model)
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

func TestBackend_SupportsToolCalling(t *testing.T) {
	b := New(nil)
	assert.False(t, b.SupportsToolCalling())
}
