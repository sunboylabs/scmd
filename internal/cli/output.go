package cli

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/scmd/scmd/internal/output"
)

// OutputWriter handles output to various destinations
type OutputWriter struct {
	mu        sync.Mutex
	writer    io.Writer
	buffered  *bufio.Writer
	file      *os.File
	mode      *IOMode
	formatter *output.Formatter
}

// OutputConfig configures the output writer
type OutputConfig struct {
	FilePath string
	Mode     *IOMode
	Format   string // "text", "json", or "markdown"
}

// NewOutputWriter creates a new output writer
func NewOutputWriter(cfg *OutputConfig) (*OutputWriter, error) {
	var writer io.Writer = os.Stdout
	var file *os.File

	if cfg.FilePath != "" {
		f, err := os.Create(cfg.FilePath)
		if err != nil {
			return nil, err
		}
		writer = f
		file = f
	}

	// Initialize formatter if appropriate
	var formatter *output.Formatter
	if cfg.Mode != nil && cfg.Mode.StdoutIsTTY && !cfg.Mode.PipeOut {
		// Only use formatter for TTY output
		f, err := output.GetDefaultFormatter()
		if err == nil {
			formatter = f
		}
	}

	ow := &OutputWriter{
		writer:    writer,
		mode:      cfg.Mode,
		file:      file,
		formatter: formatter,
	}

	// Use buffered writer for piped output or file output
	if cfg.Mode.PipeOut || cfg.FilePath != "" {
		ow.buffered = bufio.NewWriter(writer)
		ow.writer = ow.buffered
	}

	return ow, nil
}

// Write writes a string to the output
func (w *OutputWriter) Write(s string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, err := w.writer.Write([]byte(s))
	return err
}

// WriteLine writes a string followed by a newline
func (w *OutputWriter) WriteLine(s string) error {
	return w.Write(s + "\n")
}

// WriteBytes writes raw bytes to the output
func (w *OutputWriter) WriteBytes(b []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, err := w.writer.Write(b)
	return err
}

// WriteJSON writes a value as JSON
func (w *OutputWriter) WriteJSON(v interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	enc := json.NewEncoder(w.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// Flush flushes any buffered data
func (w *OutputWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buffered != nil {
		return w.buffered.Flush()
	}
	return nil
}

// Close flushes and closes the output
func (w *OutputWriter) Close() error {
	if err := w.Flush(); err != nil {
		return err
	}
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// IsTTY returns true if output is a terminal
func (w *OutputWriter) IsTTY() bool {
	return w.mode != nil && w.mode.StdoutIsTTY
}

// WriteMarkdown writes markdown content with formatting if TTY
func (w *OutputWriter) WriteMarkdown(markdown string) error {
	if w.formatter != nil {
		// Apply markdown formatting for TTY
		rendered, err := w.formatter.Render(markdown)
		if err != nil {
			// Fall back to plain markdown on error
			return w.WriteLine(markdown)
		}
		return w.Write(rendered)
	}
	// Plain text for non-TTY
	return w.WriteLine(markdown)
}
