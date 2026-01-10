package builtin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/command"
	"github.com/scmd/scmd/internal/templates"
)

// ExplainCommand implements /explain
type ExplainCommand struct{}

// NewExplainCommand creates a new explain command
func NewExplainCommand() *ExplainCommand {
	return &ExplainCommand{}
}

// Name returns the command name
func (c *ExplainCommand) Name() string { return "explain" }

// Aliases returns command aliases
func (c *ExplainCommand) Aliases() []string { return []string{"e", "what"} }

// Description returns the command description
func (c *ExplainCommand) Description() string { return "Explain code or concepts" }

// Usage returns usage information
func (c *ExplainCommand) Usage() string { return "/explain <file|concept> [options]" }

// Category returns the command category
func (c *ExplainCommand) Category() command.Category { return command.CategoryCode }

// RequiresBackend returns true
func (c *ExplainCommand) RequiresBackend() bool { return true }

// Examples returns example usages
func (c *ExplainCommand) Examples() []string {
	return []string{
		"/explain main.go",
		"/explain what is a goroutine",
		"cat file.py | scmd explain",
	}
}

// Validate validates arguments
func (c *ExplainCommand) Validate(args *command.Args) error {
	// Need either a file/concept or piped input
	stdin, hasStdin := args.Options["stdin"]
	stdinEmpty := !hasStdin || strings.TrimSpace(stdin) == ""

	if len(args.Positional) == 0 && stdinEmpty {
		return fmt.Errorf("no input provided\n\nUsage:\n  scmd explain <file|concept>\n  cat file.py | scmd explain\n  echo 'code' | scmd explain")
	}

	// Check for empty stdin that was provided
	if hasStdin && strings.TrimSpace(stdin) == "" && len(args.Positional) == 0 {
		return fmt.Errorf("empty input provided - please provide code to explain")
	}

	return nil
}

// Execute runs the explain command
func (c *ExplainCommand) Execute(
	ctx context.Context,
	args *command.Args,
	execCtx *command.ExecContext,
) (*command.Result, error) {
	// Validate arguments first
	if err := c.Validate(args); err != nil {
		return command.NewErrorResult(err.Error()), nil
	}

	var content string
	var subject string

	// Check for piped input
	if stdin, ok := args.Options["stdin"]; ok && stdin != "" {
		content = stdin
		subject = "piped input"
	} else if len(args.Positional) > 0 {
		target := args.Positional[0]

		// Check if it's a file
		if isFile(target) {
			data, err := os.ReadFile(target)
			if err != nil {
				return command.NewErrorResult(
					fmt.Sprintf("cannot read file: %v", err),
					"Check the file path",
					"Ensure you have read permissions",
				), nil
			}
			content = string(data)
			subject = filepath.Base(target)
		} else {
			// It's a concept
			content = strings.Join(args.Positional, " ")
			subject = "concept"
		}
	}

	// Check for template
	templateName := args.GetOption("template")
	var systemPrompt, prompt string

	if templateName != "" {
		// Load and execute template
		mgr, err := templates.NewManager()
		if err != nil {
			return command.NewErrorResult(
				fmt.Sprintf("failed to load template manager: %v", err),
			), nil
		}

		// Detect language from filename or content
		language := detectLanguage(subject, content)

		templateData := map[string]interface{}{
			"Code":     content,
			"Language": language,
			"FocusOn":  args.GetOption("focus"),
		}

		var userPrompt string
		systemPrompt, userPrompt, err = mgr.Execute(templateName, templateData)
		if err != nil {
			return command.NewErrorResult(
				fmt.Sprintf("failed to execute template: %v", err),
				fmt.Sprintf("Check template with: scmd template show %s", templateName),
			), nil
		}
		prompt = userPrompt
	} else {
		// Build default prompt
		prompt = buildExplainPrompt(content, subject)
		systemPrompt = `You are a helpful programming assistant. Explain code and concepts clearly and concisely.
Use examples where helpful. Format your response in markdown.`
	}

	// Check if backend is available
	if execCtx.Backend == nil {
		return command.NewErrorResult(
			"no backend available",
			"Configure a backend with 'scmd config'",
		), nil
	}

	// Show progress
	stop := execCtx.UI.Spinner("Analyzing")
	defer stop()

	// Call backend
	req := &backend.CompletionRequest{
		Prompt:      prompt,
		MaxTokens:   2048,
		Temperature: 0.3,
		SystemPrompt: systemPrompt,
	}

	resp, err := execCtx.Backend.Complete(ctx, req)
	if err != nil {
		return command.NewErrorResult(
			fmt.Sprintf("backend error: %v", err),
		), nil
	}

	return command.NewResult(resp.Content), nil
}

func buildExplainPrompt(content, subject string) string {
	if subject == "concept" {
		return fmt.Sprintf("Explain the following concept:\n\n%s", content)
	}
	return fmt.Sprintf("Explain the following code from %s:\n\n```\n%s\n```", subject, content)
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
