package output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// FormatMode represents the output formatting mode
type FormatMode string

const (
	FormatAuto     FormatMode = "auto"
	FormatMarkdown FormatMode = "markdown"
	FormatPlain    FormatMode = "plain"
)

// Formatter handles markdown rendering and output formatting
type Formatter struct {
	mode         FormatMode
	renderer     *MarkdownRenderer
	termInfo     *TerminalInfo
	colorize     bool
	termProfile  termenv.Profile
}

// FormatterOptions for creating a new formatter
type FormatterOptions struct {
	Format   string // "auto", "markdown", "plain"
	Theme    string // "auto", "dark", "light"
	WordWrap int    // word wrap width, 0 for terminal width
}

// NewFormatter creates a new output formatter with the given options
// This is lightweight and fast - markdown renderer is only initialized when needed
func NewFormatter(opts *FormatterOptions) (*Formatter, error) {
	if opts == nil {
		opts = &FormatterOptions{
			Format:   "auto",
			Theme:    "auto",
			WordWrap: 80,
		}
	}

	// Detect terminal capabilities
	termInfo := DetectTerminal()
	termProfile := termenv.ColorProfile()

	// Resolve format mode
	mode := FormatMode(opts.Format)
	if mode != FormatAuto && mode != FormatMarkdown && mode != FormatPlain {
		mode = FormatAuto
	}

	// Resolve theme
	theme := GetTheme(opts.Theme, termInfo)

	// Resolve word wrap
	wordWrap := GetWordWrap(opts.WordWrap, termInfo)

	// Determine if we should colorize output
	colorize := ShouldUseMarkdown(string(mode), termInfo)

	// Create lazy-loaded markdown renderer
	var renderer *MarkdownRenderer
	if colorize {
		renderer = NewMarkdownRenderer(theme, wordWrap)
	}

	return &Formatter{
		mode:        mode,
		renderer:    renderer,
		termInfo:    termInfo,
		colorize:    colorize,
		termProfile: termProfile,
	}, nil
}

// Render renders markdown content based on the formatter's mode
func (f *Formatter) Render(markdown string) (string, error) {
	// If colorization is disabled, return plain text
	if !f.colorize || f.renderer == nil {
		return markdown, nil
	}

	// Use the lazy-loaded renderer
	rendered, err := f.renderer.Render(markdown)
	if err != nil {
		// Fallback to plain text on error
		return markdown, nil
	}

	return rendered, nil
}

// RenderToWriter renders markdown to a writer
func (f *Formatter) RenderToWriter(markdown string, w io.Writer) error {
	rendered, err := f.Render(markdown)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(rendered))
	return err
}

// StreamRender handles streaming output (for real-time display)
func (f *Formatter) StreamRender(chunks <-chan string, w io.Writer) error {
	var buffer strings.Builder

	for chunk := range chunks {
		buffer.WriteString(chunk)
		// For streaming, just write chunks as-is
		// Final render happens at the end
		w.Write([]byte(chunk))
	}

	return nil
}

// IsColorized returns true if the formatter will apply colors/formatting
func (f *Formatter) IsColorized() bool {
	return f.colorize
}

// GetMode returns the current format mode
func (f *Formatter) GetMode() FormatMode {
	return f.mode
}

// Custom styles using lipgloss
var (
	CodeBlockStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		MarginTop(1).
		MarginBottom(1)

	HeadingStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		MarginTop(1).
		MarginBottom(1)

	ErrorStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Background(lipgloss.Color("52"))

	SuccessStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("82"))

	InfoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("33"))

	WarningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))
)

// RenderError renders an error message
func RenderError(message string) string {
	if os.Getenv("NO_COLOR") != "" {
		return "✖ " + message
	}
	return ErrorStyle.Render("✖ " + message)
}

// RenderSuccess renders a success message
func RenderSuccess(message string) string {
	if os.Getenv("NO_COLOR") != "" {
		return "✓ " + message
	}
	return SuccessStyle.Render("✓ " + message)
}

// RenderInfo renders an informational message
func RenderInfo(message string) string {
	if os.Getenv("NO_COLOR") != "" {
		return "ℹ " + message
	}
	return InfoStyle.Render("ℹ " + message)
}

// RenderWarning renders a warning message
func RenderWarning(message string) string {
	if os.Getenv("NO_COLOR") != "" {
		return "⚠ " + message
	}
	return WarningStyle.Render("⚠ " + message)
}

// RenderCodeBlock renders a code block with optional syntax highlighting
func RenderCodeBlock(code, language string) string {
	if os.Getenv("NO_COLOR") != "" {
		return fmt.Sprintf("```%s\n%s\n```", language, code)
	}
	block := fmt.Sprintf("```%s\n%s\n```", language, code)
	return CodeBlockStyle.Render(block)
}

// RenderHeading renders a heading
func RenderHeading(text string, level int) string {
	prefix := strings.Repeat("#", level)
	heading := fmt.Sprintf("%s %s", prefix, text)
	if os.Getenv("NO_COLOR") != "" {
		return heading
	}
	return HeadingStyle.Render(heading)
}

// BufferedRender collects all output and renders it at once
type BufferedRender struct {
	buffer    bytes.Buffer
	formatter *Formatter
}

// NewBufferedRender creates a new buffered renderer
func NewBufferedRender(formatter *Formatter) *BufferedRender {
	return &BufferedRender{
		formatter: formatter,
	}
}

// Write implements io.Writer
func (b *BufferedRender) Write(p []byte) (n int, err error) {
	return b.buffer.Write(p)
}

// Flush renders and returns the buffered content
func (b *BufferedRender) Flush() (string, error) {
	content := b.buffer.String()
	b.buffer.Reset()
	return b.formatter.Render(content)
}

// GetDefaultFormatter returns a formatter with default settings
func GetDefaultFormatter() (*Formatter, error) {
	return NewFormatter(&FormatterOptions{
		Format:   "auto",
		Theme:    "auto",
		WordWrap: 0, // Use terminal width
	})
}

// GetFormatterFromConfig creates a formatter from config values
func GetFormatterFromConfig(format, theme string, wordWrap int) (*Formatter, error) {
	return NewFormatter(&FormatterOptions{
		Format:   format,
		Theme:    theme,
		WordWrap: wordWrap,
	})
}
