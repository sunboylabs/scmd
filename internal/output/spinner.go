package output

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/muesli/termenv"
)

// ProgressSpinner shows a loading spinner
type ProgressSpinner struct {
	spinner *spinner.Spinner
	message string
}

// ShowProgress displays a spinner with a message
func ShowProgress(message string) *ProgressSpinner {
	// Check if we're in a TTY
	if termenv.ColorProfile() == termenv.Ascii {
		// No spinner in non-TTY environments
		return &ProgressSpinner{
			message: message,
		}
	}

	s := spinner.New(
		spinner.CharSets[14], // "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
		100*time.Millisecond,
		spinner.WithWriter(os.Stderr), // Write to stderr to not interfere with output
	)
	s.Suffix = "  " + message
	s.Color("cyan")
	s.Start()

	return &ProgressSpinner{
		spinner: s,
		message: message,
	}
}

// Stop stops the spinner
func (p *ProgressSpinner) Stop() {
	if p.spinner != nil {
		p.spinner.Stop()
	}
}

// Success stops the spinner with a success message
func (p *ProgressSpinner) Success(message string) {
	if p.spinner != nil {
		p.spinner.FinalMSG = RenderSuccess(message) + "\n"
		p.spinner.Stop()
	}
}

// Error stops the spinner with an error message
func (p *ProgressSpinner) Error(message string) {
	if p.spinner != nil {
		p.spinner.FinalMSG = RenderError(message) + "\n"
		p.spinner.Stop()
	}
}

// Update changes the spinner message
func (p *ProgressSpinner) Update(message string) {
	if p.spinner != nil {
		p.spinner.Suffix = "  " + message
	}
	p.message = message
}

// SimpleProgress shows a simple progress message without spinner
func SimpleProgress(message string) func() {
	// For non-TTY environments or when spinner is not appropriate
	if termenv.ColorProfile() != termenv.Ascii {
		os.Stderr.WriteString(RenderInfo(message + "..."))
	}

	return func() {
		if termenv.ColorProfile() != termenv.Ascii {
			os.Stderr.WriteString(" done\n")
		}
	}
}