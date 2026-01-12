package output

import (
	"sync"

	"github.com/charmbracelet/glamour"
)

// MarkdownRenderer provides lazy-loaded markdown rendering with Glamour
type MarkdownRenderer struct {
	theme    string
	wordWrap int
	renderer *glamour.TermRenderer
	mu       sync.Mutex
	initOnce sync.Once
	initErr  error
}

// NewMarkdownRenderer creates a new markdown renderer with lazy initialization
// This is lightweight and fast - actual Glamour renderer is only created when needed
func NewMarkdownRenderer(theme string, wordWrap int) *MarkdownRenderer {
	return &MarkdownRenderer{
		theme:    theme,
		wordWrap: wordWrap,
	}
}

// Render renders markdown content to formatted terminal output
// Lazy initializes the Glamour renderer on first use
func (r *MarkdownRenderer) Render(markdown string) (string, error) {
	// Lazy initialize
	if err := r.ensureInitialized(); err != nil {
		// Fallback to plain text on initialization error
		return markdown, err
	}

	// Use the renderer
	rendered, err := r.renderer.Render(markdown)
	if err != nil {
		// Fallback to plain text on render error
		return markdown, err
	}

	return rendered, nil
}

// ensureInitialized ensures the Glamour renderer is initialized
// This is called lazily on first use
func (r *MarkdownRenderer) ensureInitialized() error {
	r.initOnce.Do(func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		// Skip if already initialized
		if r.renderer != nil {
			return
		}

		// Configure Glamour renderer
		opts := []glamour.TermRendererOption{
			glamour.WithPreservedNewLines(),
		}

		// Set word wrap
		if r.wordWrap > 0 {
			opts = append(opts, glamour.WithWordWrap(r.wordWrap))
		}

		// Set theme
		switch r.theme {
		case "dark":
			opts = append(opts, glamour.WithStylePath("dark"))
		case "light":
			opts = append(opts, glamour.WithStylePath("light"))
		case "notty":
			opts = append(opts, glamour.WithStylePath("notty"))
		default:
			// Auto-detect
			opts = append(opts, glamour.WithAutoStyle())
		}

		// Create renderer
		renderer, err := glamour.NewTermRenderer(opts...)
		if err != nil {
			r.initErr = err
			return
		}

		r.renderer = renderer
	})

	return r.initErr
}

// IsInitialized returns true if the renderer has been initialized
func (r *MarkdownRenderer) IsInitialized() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.renderer != nil
}

// RendererOptions for creating renderers with custom settings
type RendererOptions struct {
	Theme    string
	WordWrap int
	Style    string
}

// DefaultRendererOptions returns default renderer options
func DefaultRendererOptions() *RendererOptions {
	return &RendererOptions{
		Theme:    "auto",
		WordWrap: 80,
	}
}

// NewMarkdownRendererWithOptions creates a renderer with custom options
func NewMarkdownRendererWithOptions(opts *RendererOptions) *MarkdownRenderer {
	if opts == nil {
		opts = DefaultRendererOptions()
	}
	return NewMarkdownRenderer(opts.Theme, opts.WordWrap)
}
