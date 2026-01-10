package output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Formatter handles markdown rendering and output formatting
type Formatter struct {
	style        string
	renderer     *glamour.TermRenderer
	colorize     bool
	termProfile  termenv.Profile
}

// NewFormatter creates a new output formatter
func NewFormatter(style string, colorize bool) (*Formatter, error) {
	// Auto-detect terminal capabilities
	termProfile := termenv.ColorProfile()

	if style == "auto" {
		style = detectStyle()
	}

	// Configure glamour renderer
	opts := []glamour.TermRendererOption{
		glamour.WithWordWrap(100),
		glamour.WithPreservedNewLines(),
	}

	switch style {
	case "dark":
		opts = append(opts, glamour.WithStylePath("dark"))
	case "light":
		opts = append(opts, glamour.WithStylePath("light"))
	case "notty":
		opts = append(opts, glamour.WithStylePath("notty"))
	default:
		opts = append(opts, glamour.WithAutoStyle())
	}

	renderer, err := glamour.NewTermRenderer(opts...)
	if err != nil {
		return nil, err
	}

	return &Formatter{
		style:       style,
		renderer:    renderer,
		colorize:    colorize,
		termProfile: termProfile,
	}, nil
}

// Render renders markdown content
func (f *Formatter) Render(markdown string) (string, error) {
	if !f.colorize || f.style == "notty" {
		return markdown, nil // Plain text mode
	}

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

// detectStyle detects whether to use dark or light theme
func detectStyle() string {
	// For now, default to dark theme
	// Terminal theme detection is complex and varies by terminal
	// Users can override with environment variables or config
	if os.Getenv("SCMD_THEME") == "light" {
		return "light"
	}
	return "dark" // Default to dark
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
	return ErrorStyle.Render("✖ " + message)
}

// RenderSuccess renders a success message
func RenderSuccess(message string) string {
	return SuccessStyle.Render("✓ " + message)
}

// RenderInfo renders an informational message
func RenderInfo(message string) string {
	return InfoStyle.Render("ℹ " + message)
}

// RenderWarning renders a warning message
func RenderWarning(message string) string {
	return WarningStyle.Render("⚠ " + message)
}

// RenderCodeBlock renders a code block with optional syntax highlighting
func RenderCodeBlock(code, language string) string {
	// For now, just use the lipgloss style
	// Syntax highlighting will be added in syntax.go
	block := fmt.Sprintf("```%s\n%s\n```", language, code)
	return CodeBlockStyle.Render(block)
}

// RenderHeading renders a heading
func RenderHeading(text string, level int) string {
	prefix := strings.Repeat("#", level)
	return HeadingStyle.Render(fmt.Sprintf("%s %s", prefix, text))
}

// BufferedRender collects all output and renders it at once
type BufferedRender struct {
	buffer bytes.Buffer
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
	// Check if colors are enabled
	colorize := os.Getenv("NO_COLOR") == "" && termenv.ColorProfile() != termenv.Ascii

	return NewFormatter("auto", colorize)
}