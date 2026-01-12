package output

import (
	"strings"
	"testing"
)

func TestNewMarkdownRenderer(t *testing.T) {
	tests := []struct {
		name     string
		theme    string
		wordWrap int
	}{
		{
			name:     "dark theme",
			theme:    "dark",
			wordWrap: 80,
		},
		{
			name:     "light theme",
			theme:    "light",
			wordWrap: 100,
		},
		{
			name:     "auto theme",
			theme:    "auto",
			wordWrap: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewMarkdownRenderer(tt.theme, tt.wordWrap)
			if renderer == nil {
				t.Error("NewMarkdownRenderer() returned nil")
			}
			if renderer.theme != tt.theme {
				t.Errorf("theme = %v, want %v", renderer.theme, tt.theme)
			}
			if renderer.wordWrap != tt.wordWrap {
				t.Errorf("wordWrap = %v, want %v", renderer.wordWrap, tt.wordWrap)
			}
		})
	}
}

func TestMarkdownRenderer_LazyInitialization(t *testing.T) {
	renderer := NewMarkdownRenderer("dark", 80)

	// Should not be initialized yet
	if renderer.IsInitialized() {
		t.Error("Renderer should not be initialized on creation")
	}

	// First render should initialize
	markdown := "# Test\n\nThis is a test."
	_, err := renderer.Render(markdown)
	if err != nil {
		t.Errorf("Render() error = %v", err)
	}

	// Should now be initialized
	if !renderer.IsInitialized() {
		t.Error("Renderer should be initialized after first render")
	}
}

func TestMarkdownRenderer_Render(t *testing.T) {
	tests := []struct {
		name     string
		theme    string
		markdown string
		verify   func(string) bool
	}{
		{
			name:     "simple heading",
			theme:    "dark",
			markdown: "# Hello World",
			verify:   func(s string) bool { return len(s) > 0 },
		},
		{
			name:     "code block",
			theme:    "dark",
			markdown: "```go\nfunc main() {}\n```",
			verify:   func(s string) bool { return len(s) > 0 },
		},
		{
			name:     "bold text",
			theme:    "light",
			markdown: "This is **bold**",
			verify:   func(s string) bool { return strings.Contains(s, "bold") },
		},
		{
			name:     "list",
			theme:    "dark",
			markdown: "- Item 1\n- Item 2",
			verify:   func(s string) bool { return len(s) > 0 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewMarkdownRenderer(tt.theme, 80)
			result, err := renderer.Render(tt.markdown)
			if err != nil {
				t.Errorf("Render() error = %v", err)
			}

			if result == "" {
				t.Error("Render() returned empty result")
			}

			if tt.verify != nil && !tt.verify(result) {
				t.Errorf("Render() result failed verification.\nGot: %s", result)
			}
		})
	}
}

func TestMarkdownRenderer_MultipleRenders(t *testing.T) {
	renderer := NewMarkdownRenderer("dark", 80)

	// First render
	result1, err := renderer.Render("# First")
	if err != nil {
		t.Errorf("First Render() error = %v", err)
	}

	// Second render should reuse initialized renderer
	result2, err := renderer.Render("# Second")
	if err != nil {
		t.Errorf("Second Render() error = %v", err)
	}

	if !strings.Contains(result1, "First") {
		t.Error("First render result doesn't contain expected text")
	}
	if !strings.Contains(result2, "Second") {
		t.Error("Second render result doesn't contain expected text")
	}
}

func TestDefaultRendererOptions(t *testing.T) {
	opts := DefaultRendererOptions()
	if opts == nil {
		t.Fatal("DefaultRendererOptions() returned nil")
	}

	if opts.Theme != "auto" {
		t.Errorf("Default theme = %v, want 'auto'", opts.Theme)
	}
	if opts.WordWrap != 80 {
		t.Errorf("Default wordWrap = %v, want 80", opts.WordWrap)
	}
}

func TestNewMarkdownRendererWithOptions(t *testing.T) {
	tests := []struct {
		name string
		opts *RendererOptions
	}{
		{
			name: "with custom options",
			opts: &RendererOptions{
				Theme:    "light",
				WordWrap: 100,
			},
		},
		{
			name: "with nil options",
			opts: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewMarkdownRendererWithOptions(tt.opts)
			if renderer == nil {
				t.Error("NewMarkdownRendererWithOptions() returned nil")
			}

			// Verify it can render
			_, err := renderer.Render("# Test")
			if err != nil {
				t.Errorf("Render() error = %v", err)
			}
		})
	}
}

func TestMarkdownRenderer_ErrorHandling(t *testing.T) {
	renderer := NewMarkdownRenderer("dark", 80)

	// Even with problematic markdown, should not crash
	markdown := "# Incomplete [link"
	result, err := renderer.Render(markdown)

	// Should return something even if there's an error
	if result == "" {
		t.Error("Render() returned empty string")
	}

	// Error is acceptable but should be handled gracefully
	if err != nil {
		t.Logf("Render() returned error (acceptable): %v", err)
	}
}

func TestMarkdownRenderer_WordWrap(t *testing.T) {
	longText := strings.Repeat("word ", 50) // Very long line

	tests := []struct {
		name     string
		wordWrap int
		markdown string
	}{
		{
			name:     "with word wrap",
			wordWrap: 40,
			markdown: longText,
		},
		{
			name:     "without word wrap",
			wordWrap: 0,
			markdown: longText,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewMarkdownRenderer("dark", tt.wordWrap)
			result, err := renderer.Render(tt.markdown)
			if err != nil {
				t.Errorf("Render() error = %v", err)
			}

			if result == "" {
				t.Error("Render() returned empty result")
			}
		})
	}
}
