package output

import (
	"os"
	"strings"
	"testing"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name    string
		opts    *FormatterOptions
		wantErr bool
	}{
		{
			name: "default options",
			opts: nil,
			wantErr: false,
		},
		{
			name: "auto format",
			opts: &FormatterOptions{
				Format:   "auto",
				Theme:    "dark",
				WordWrap: 80,
			},
			wantErr: false,
		},
		{
			name: "markdown format",
			opts: &FormatterOptions{
				Format:   "markdown",
				Theme:    "light",
				WordWrap: 100,
			},
			wantErr: false,
		},
		{
			name: "plain format",
			opts: &FormatterOptions{
				Format:   "plain",
				Theme:    "auto",
				WordWrap: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := NewFormatter(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFormatter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if formatter == nil {
				t.Error("NewFormatter() returned nil formatter")
			}
		})
	}
}

func TestFormatter_Render(t *testing.T) {
	tests := []struct {
		name     string
		opts     *FormatterOptions
		markdown string
		wantContains string
	}{
		{
			name: "plain text passthrough",
			opts: &FormatterOptions{
				Format: "plain",
			},
			markdown: "# Hello World\n\nThis is **bold**",
			wantContains: "# Hello World",
		},
		{
			name: "markdown with code block",
			opts: &FormatterOptions{
				Format: "plain",
			},
			markdown: "```go\nfunc main() {}\n```",
			wantContains: "func main()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := NewFormatter(tt.opts)
			if err != nil {
				t.Fatalf("NewFormatter() error = %v", err)
			}

			result, err := formatter.Render(tt.markdown)
			if err != nil {
				t.Errorf("Render() error = %v", err)
			}

			if !strings.Contains(result, tt.wantContains) {
				t.Errorf("Render() result doesn't contain expected text.\nGot: %s\nWant to contain: %s", result, tt.wantContains)
			}
		})
	}
}

func TestFormatter_IsColorized(t *testing.T) {
	// Save original NO_COLOR
	origNoColor := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", origNoColor)

	tests := []struct {
		name    string
		opts    *FormatterOptions
		noColor string
		want    bool
	}{
		{
			name: "plain format - not colorized",
			opts: &FormatterOptions{
				Format: "plain",
			},
			noColor: "",
			want:    false,
		},
		{
			name: "with NO_COLOR env",
			opts: &FormatterOptions{
				Format: "markdown",
			},
			noColor: "1",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NO_COLOR", tt.noColor)

			formatter, err := NewFormatter(tt.opts)
			if err != nil {
				t.Fatalf("NewFormatter() error = %v", err)
			}

			got := formatter.IsColorized()
			if got != tt.want {
				t.Errorf("IsColorized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatter_GetMode(t *testing.T) {
	tests := []struct {
		name string
		opts *FormatterOptions
		want FormatMode
	}{
		{
			name: "auto mode",
			opts: &FormatterOptions{Format: "auto"},
			want: FormatAuto,
		},
		{
			name: "markdown mode",
			opts: &FormatterOptions{Format: "markdown"},
			want: FormatMarkdown,
		},
		{
			name: "plain mode",
			opts: &FormatterOptions{Format: "plain"},
			want: FormatPlain,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, _ := NewFormatter(tt.opts)
			if got := formatter.GetMode(); got != tt.want {
				t.Errorf("GetMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderError(t *testing.T) {
	// Save original NO_COLOR
	origNoColor := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", origNoColor)

	tests := []struct {
		name    string
		message string
		noColor string
		wantContains string
	}{
		{
			name:    "with colors",
			message: "test error",
			noColor: "",
			wantContains: "✖ test error",
		},
		{
			name:    "without colors",
			message: "test error",
			noColor: "1",
			wantContains: "✖ test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("NO_COLOR", tt.noColor)
			result := RenderError(tt.message)
			if !strings.Contains(result, tt.wantContains) {
				t.Errorf("RenderError() = %v, want to contain %v", result, tt.wantContains)
			}
		})
	}
}

func TestRenderSuccess(t *testing.T) {
	result := RenderSuccess("test success")
	if !strings.Contains(result, "✓") {
		t.Errorf("RenderSuccess() should contain checkmark")
	}
}

func TestRenderInfo(t *testing.T) {
	result := RenderInfo("test info")
	if !strings.Contains(result, "ℹ") {
		t.Errorf("RenderInfo() should contain info symbol")
	}
}

func TestRenderWarning(t *testing.T) {
	result := RenderWarning("test warning")
	if !strings.Contains(result, "⚠") {
		t.Errorf("RenderWarning() should contain warning symbol")
	}
}

func TestRenderCodeBlock(t *testing.T) {
	code := "func main() {}"
	language := "go"

	result := RenderCodeBlock(code, language)
	if !strings.Contains(result, code) {
		t.Errorf("RenderCodeBlock() should contain the code")
	}
}

func TestRenderHeading(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		level int
		want  string
	}{
		{
			name:  "level 1",
			text:  "Title",
			level: 1,
			want:  "# Title",
		},
		{
			name:  "level 2",
			text:  "Subtitle",
			level: 2,
			want:  "## Subtitle",
		},
		{
			name:  "level 3",
			text:  "Section",
			level: 3,
			want:  "### Section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderHeading(tt.text, tt.level)
			if !strings.Contains(result, tt.want) {
				t.Errorf("RenderHeading() = %v, want to contain %v", result, tt.want)
			}
		})
	}
}

func TestBufferedRender(t *testing.T) {
	formatter, err := NewFormatter(&FormatterOptions{Format: "plain"})
	if err != nil {
		t.Fatalf("NewFormatter() error = %v", err)
	}

	buffer := NewBufferedRender(formatter)

	testContent := "# Test\n\nContent"
	_, err = buffer.Write([]byte(testContent))
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}

	result, err := buffer.Flush()
	if err != nil {
		t.Errorf("Flush() error = %v", err)
	}

	if !strings.Contains(result, "Test") {
		t.Errorf("Flush() result doesn't contain expected content")
	}
}

func TestGetDefaultFormatter(t *testing.T) {
	formatter, err := GetDefaultFormatter()
	if err != nil {
		t.Errorf("GetDefaultFormatter() error = %v", err)
	}
	if formatter == nil {
		t.Error("GetDefaultFormatter() returned nil")
	}
}

func TestGetFormatterFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		theme    string
		wordWrap int
		wantErr  bool
	}{
		{
			name:     "valid config",
			format:   "auto",
			theme:    "dark",
			wordWrap: 80,
			wantErr:  false,
		},
		{
			name:     "markdown format",
			format:   "markdown",
			theme:    "light",
			wordWrap: 100,
			wantErr:  false,
		},
		{
			name:     "plain format",
			format:   "plain",
			theme:    "auto",
			wordWrap: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := GetFormatterFromConfig(tt.format, tt.theme, tt.wordWrap)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFormatterFromConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if formatter == nil {
				t.Error("GetFormatterFromConfig() returned nil")
			}
		})
	}
}
