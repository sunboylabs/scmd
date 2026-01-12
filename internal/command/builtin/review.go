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

// ReviewCommand implements /review
type ReviewCommand struct{}

// NewReviewCommand creates a new review command
func NewReviewCommand() *ReviewCommand {
	return &ReviewCommand{}
}

// Name returns the command name
func (c *ReviewCommand) Name() string { return "review" }

// Aliases returns command aliases
func (c *ReviewCommand) Aliases() []string { return []string{"r"} }

// Description returns the command description
func (c *ReviewCommand) Description() string { return "Review code for issues and improvements" }

// Usage returns usage information
func (c *ReviewCommand) Usage() string { return "/review <file> [options]" }

// Category returns the command category
func (c *ReviewCommand) Category() command.Category { return command.CategoryCode }

// RequiresBackend returns true
func (c *ReviewCommand) RequiresBackend() bool { return true }

// Examples returns example usages
func (c *ReviewCommand) Examples() []string {
	return []string{
		"/review main.go",
		"git diff | scmd review",
		"/review src/ --focus security",
	}
}

// Validate validates arguments
func (c *ReviewCommand) Validate(args *command.Args) error {
	stdin, hasStdin := args.Options["stdin"]
	stdinEmpty := !hasStdin || strings.TrimSpace(stdin) == ""

	if len(args.Positional) == 0 && stdinEmpty {
		return fmt.Errorf("no input provided\n\nUsage:\n  scmd review <file>\n  git diff | scmd review\n  cat file.py | scmd review")
	}

	// Check for empty stdin that was provided
	if hasStdin && strings.TrimSpace(stdin) == "" && len(args.Positional) == 0 {
		return fmt.Errorf("empty input provided - please provide code to review")
	}

	return nil
}

// Execute runs the review command
func (c *ReviewCommand) Execute(
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

	// Check for piped input (e.g., git diff)
	if stdin, ok := args.Options["stdin"]; ok && stdin != "" {
		content = stdin
		subject = "diff"
	} else if len(args.Positional) > 0 {
		target := args.Positional[0]

		if isFile(target) {
			data, err := os.ReadFile(target)
			if err != nil {
				return command.NewErrorResult(
					fmt.Sprintf("cannot read file: %v", err),
					"Check the file path",
				), nil
			}
			content = string(data)
			subject = filepath.Base(target)
		} else {
			return command.NewErrorResult(
				fmt.Sprintf("file not found: %s", target),
				"Check the file path",
			), nil
		}
	}

	// Get focus area if specified
	focus := args.GetOption("focus")

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
			"Context":  focus,
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
		prompt = buildReviewPrompt(content, subject, focus)
		systemPrompt = `You are an expert code reviewer. Analyze code for:
- Bugs and potential issues
- Security vulnerabilities
- Performance concerns
- Code quality and readability
- Best practices

Provide constructive feedback with specific suggestions.
Format your response in markdown with sections for each category.`
	}

	if execCtx.Backend == nil {
		return command.NewErrorResult(
			"no backend available",
			"Configure a backend with 'scmd config'",
		), nil
	}

	stop := execCtx.UI.Spinner("Reviewing")
	defer stop()

	req := &backend.CompletionRequest{
		Prompt:      prompt,
		MaxTokens:   4096,
		Temperature: 0.3,
		SystemPrompt: systemPrompt,
	}

	resp, err := execCtx.Backend.Complete(ctx, req)
	if err != nil {
		return command.NewErrorResult(
			fmt.Sprintf("backend error: %v", err),
		), nil
	}

	// The response is already in markdown format from the LLM
	// It will be formatted by the CLI output handler
	return command.NewResult(resp.Content), nil
}

func buildReviewPrompt(content, subject, focus string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Please review the following code from %s:\n\n", subject))

	if focus != "" {
		sb.WriteString(fmt.Sprintf("Focus especially on: %s\n\n", focus))
	}

	sb.WriteString("```\n")
	sb.WriteString(content)
	sb.WriteString("\n```")

	return sb.String()
}
