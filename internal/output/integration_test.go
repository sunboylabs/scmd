package output

import (
	"os"
	"strings"
	"testing"
)

// TestIntegration_NO_COLOR tests that NO_COLOR environment variable is respected
func TestIntegration_NO_COLOR(t *testing.T) {
	// Save original NO_COLOR
	origNoColor := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", origNoColor)

	markdown := "# Test Heading\n\nThis is **bold** and this is `code`.\n\n```go\nfunc main() {}\n```"

	tests := []struct {
		name    string
		noColor string
		format  string
	}{
		{
			name:    "NO_COLOR with auto format",
			noColor: "1",
			format:  "auto",
		},
		{
			name:    "NO_COLOR with markdown format",
			noColor: "1",
			format:  "markdown",
		},
		{
			name:    "without NO_COLOR",
			noColor: "",
			format:  "plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NO_COLOR", tt.noColor)

			formatter, err := NewFormatter(&FormatterOptions{
				Format:   tt.format,
				Theme:    "dark",
				WordWrap: 80,
			})
			if err != nil {
				t.Fatalf("NewFormatter() error = %v", err)
			}

			result, err := formatter.Render(markdown)
			if err != nil {
				t.Errorf("Render() error = %v", err)
			}

			// Should contain the original text regardless
			if !strings.Contains(result, "Test Heading") {
				t.Error("Result should contain original text")
			}

			// When NO_COLOR is set, should not be colorized
			if tt.noColor != "" && formatter.IsColorized() {
				t.Error("Formatter should not be colorized when NO_COLOR is set")
			}
		})
	}
}

// TestIntegration_FullWorkflow tests the complete formatting workflow
func TestIntegration_FullWorkflow(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		theme    string
		wordWrap int
		markdown string
	}{
		{
			name:     "auto format with dark theme",
			format:   "auto",
			theme:    "dark",
			wordWrap: 80,
			markdown: "# Documentation\n\n## Features\n\n- Feature 1\n- Feature 2\n\n```python\ndef hello():\n    print('world')\n```",
		},
		{
			name:     "markdown format with light theme",
			format:   "markdown",
			theme:    "light",
			wordWrap: 100,
			markdown: "## API Reference\n\n### Methods\n\n1. `GET /api/v1/users`\n2. `POST /api/v1/users`",
		},
		{
			name:     "plain format",
			format:   "plain",
			theme:    "dark",
			wordWrap: 80,
			markdown: "Plain text output\n\nNo formatting applied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create formatter
			formatter, err := NewFormatter(&FormatterOptions{
				Format:   tt.format,
				Theme:    tt.theme,
				WordWrap: tt.wordWrap,
			})
			if err != nil {
				t.Fatalf("NewFormatter() error = %v", err)
			}

			// Render
			result, err := formatter.Render(tt.markdown)
			if err != nil {
				t.Errorf("Render() error = %v", err)
			}

			// Verify output is not empty
			if result == "" {
				t.Error("Render() returned empty result")
			}

			// Verify mode is correct
			expectedMode := FormatMode(tt.format)
			if formatter.GetMode() != expectedMode {
				t.Errorf("Mode = %v, want %v", formatter.GetMode(), expectedMode)
			}
		})
	}
}

// TestIntegration_ConfigVariations tests various configuration combinations
func TestIntegration_ConfigVariations(t *testing.T) {
	markdown := "# Test\n\nContent with **emphasis** and `code`."

	configs := []struct {
		format   string
		theme    string
		wordWrap int
	}{
		{"auto", "auto", 0},
		{"auto", "dark", 80},
		{"auto", "light", 100},
		{"markdown", "dark", 80},
		{"markdown", "light", 120},
		{"plain", "auto", 80},
	}

	for _, cfg := range configs {
		t.Run(cfg.format+"-"+cfg.theme, func(t *testing.T) {
			formatter, err := GetFormatterFromConfig(cfg.format, cfg.theme, cfg.wordWrap)
			if err != nil {
				t.Fatalf("GetFormatterFromConfig() error = %v", err)
			}

			result, err := formatter.Render(markdown)
			if err != nil {
				t.Errorf("Render() error = %v", err)
			}

			if result == "" {
				t.Error("Render() returned empty result")
			}
		})
	}
}

// TestIntegration_ErrorMessages tests rendering of styled messages
func TestIntegration_ErrorMessages(t *testing.T) {
	// Save original NO_COLOR
	origNoColor := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", origNoColor)

	tests := []struct {
		name     string
		noColor  string
		renderer func(string) string
	}{
		{"error without NO_COLOR", "", RenderError},
		{"error with NO_COLOR", "1", RenderError},
		{"success without NO_COLOR", "", RenderSuccess},
		{"success with NO_COLOR", "1", RenderSuccess},
		{"info without NO_COLOR", "", RenderInfo},
		{"info with NO_COLOR", "1", RenderInfo},
		{"warning without NO_COLOR", "", RenderWarning},
		{"warning with NO_COLOR", "1", RenderWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NO_COLOR", tt.noColor)

			result := tt.renderer("test message")
			if !strings.Contains(result, "test message") {
				t.Error("Result should contain the message")
			}
		})
	}
}

// TestIntegration_BufferedRendering tests buffered output
func TestIntegration_BufferedRendering(t *testing.T) {
	formatter, err := NewFormatter(&FormatterOptions{
		Format:   "plain",
		Theme:    "dark",
		WordWrap: 80,
	})
	if err != nil {
		t.Fatalf("NewFormatter() error = %v", err)
	}

	buffer := NewBufferedRender(formatter)

	// Write multiple chunks
	chunks := []string{
		"# Title\n\n",
		"Paragraph 1\n\n",
		"Paragraph 2\n\n",
		"```go\ncode\n```\n",
	}

	for _, chunk := range chunks {
		_, err := buffer.Write([]byte(chunk))
		if err != nil {
			t.Errorf("Write() error = %v", err)
		}
	}

	// Flush and verify
	result, err := buffer.Flush()
	if err != nil {
		t.Errorf("Flush() error = %v", err)
	}

	for _, chunk := range chunks {
		if !strings.Contains(result, strings.TrimSpace(chunk)) {
			t.Errorf("Result doesn't contain chunk: %s", chunk)
		}
	}
}

// TestIntegration_TerminalDetection tests terminal detection with various env vars
func TestIntegration_TerminalDetection(t *testing.T) {
	// Save original env vars
	origTerm := os.Getenv("TERM")
	origTermProgram := os.Getenv("TERM_PROGRAM")
	origScmdTheme := os.Getenv("SCMD_THEME")
	defer func() {
		os.Setenv("TERM", origTerm)
		os.Setenv("TERM_PROGRAM", origTermProgram)
		os.Setenv("SCMD_THEME", origScmdTheme)
	}()

	tests := []struct {
		name        string
		term        string
		termProgram string
		scmdTheme   string
	}{
		{
			name:        "standard terminal",
			term:        "xterm-256color",
			termProgram: "",
			scmdTheme:   "",
		},
		{
			name:        "iTerm2",
			term:        "xterm-256color",
			termProgram: "iTerm.app",
			scmdTheme:   "dark",
		},
		{
			name:        "kitty",
			term:        "xterm-kitty",
			termProgram: "",
			scmdTheme:   "light",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TERM", tt.term)
			os.Setenv("TERM_PROGRAM", tt.termProgram)
			os.Setenv("SCMD_THEME", tt.scmdTheme)

			info := DetectTerminal()
			if info == nil {
				t.Fatal("DetectTerminal() returned nil")
			}

			// Basic validation
			if info.ColorProfile == 0 {
				t.Error("ColorProfile should be set")
			}

			// Create formatter and verify it works
			formatter, err := NewFormatter(&FormatterOptions{
				Format:   "auto",
				Theme:    "auto",
				WordWrap: 80,
			})
			if err != nil {
				t.Fatalf("NewFormatter() error = %v", err)
			}

			result, err := formatter.Render("# Test")
			if err != nil {
				t.Errorf("Render() error = %v", err)
			}
			if result == "" {
				t.Error("Render() returned empty result")
			}
		})
	}
}
