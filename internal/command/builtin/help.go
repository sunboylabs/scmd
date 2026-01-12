// Package builtin provides built-in commands for scmd
package builtin

import (
	"context"
	"fmt"
	"strings"

	"github.com/scmd/scmd/internal/command"
)

// HelpCommand implements /help
type HelpCommand struct {
	registry *command.Registry
}

// NewHelpCommand creates a new help command
func NewHelpCommand(registry *command.Registry) *HelpCommand {
	return &HelpCommand{registry: registry}
}

// Name returns the command name
func (c *HelpCommand) Name() string { return "help" }

// Aliases returns command aliases
func (c *HelpCommand) Aliases() []string { return []string{"h", "?"} }

// Description returns the command description
func (c *HelpCommand) Description() string { return "Show help for commands" }

// Usage returns usage information
func (c *HelpCommand) Usage() string { return "/help [command]" }

// Category returns the command category
func (c *HelpCommand) Category() command.Category { return command.CategoryCore }

// RequiresBackend returns false
func (c *HelpCommand) RequiresBackend() bool { return false }

// Examples returns example usages
func (c *HelpCommand) Examples() []string {
	return []string{
		"/help",
		"/help explain",
	}
}

// Validate validates arguments
func (c *HelpCommand) Validate(_ *command.Args) error {
	return nil
}

// Execute runs the help command
func (c *HelpCommand) Execute(
	_ context.Context,
	args *command.Args,
	execCtx *command.ExecContext,
) (*command.Result, error) {
	if len(args.Positional) > 0 {
		return c.showCommandHelp(args.Positional[0], execCtx)
	}
	return c.showAllHelp(execCtx)
}

func (c *HelpCommand) showAllHelp(execCtx *command.ExecContext) (*command.Result, error) {
	var sb strings.Builder

	sb.WriteString("scmd - AI-powered slash commands\n\n")
	sb.WriteString("Commands:\n")

	categories := []command.Category{
		command.CategoryCore,
		command.CategoryCode,
		command.CategoryGit,
		command.CategoryConfig,
	}

	for _, cat := range categories {
		cmds := c.registry.ListByCategory(cat)
		if len(cmds) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("\n  %s:\n", cat))
		for _, cmd := range cmds {
			sb.WriteString(fmt.Sprintf("    %-12s %s\n", "/"+cmd.Name(), cmd.Description()))
		}
	}

	sb.WriteString("\nUse '/help <command>' for more information.\n")

	// Discovery section - promoted to top
	sb.WriteString("\n💡 Discover 100+ Commands:\n")
	sb.WriteString("  Search registry:     scmd registry search <topic>\n")
	sb.WriteString("  Browse categories:   scmd registry categories\n")
	sb.WriteString("  Install commands:    scmd repo install official/<name>\n")
	sb.WriteString("  List installed:      scmd slash list\n")
	sb.WriteString("\n  Local commands:      ~/.scmd/commands/*.yaml\n")

	execCtx.UI.Write(sb.String())
	return &command.Result{Success: true}, nil
}

func (c *HelpCommand) showCommandHelp(name string, execCtx *command.ExecContext) (*command.Result, error) {
	cmd, ok := c.registry.Get(name)
	if !ok {
		return &command.Result{
			Success:     false,
			Error:       fmt.Sprintf("unknown command: %s", name),
			Suggestions: []string{"/help"},
		}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s - %s\n\n", cmd.Name(), cmd.Description()))
	sb.WriteString(fmt.Sprintf("Usage: %s\n", cmd.Usage()))

	if aliases := cmd.Aliases(); len(aliases) > 0 {
		sb.WriteString(fmt.Sprintf("Aliases: %s\n", strings.Join(aliases, ", ")))
	}

	if examples := cmd.Examples(); len(examples) > 0 {
		sb.WriteString("\nExamples:\n")
		for _, ex := range examples {
			sb.WriteString(fmt.Sprintf("  %s\n", ex))
		}
	}

	execCtx.UI.Write(sb.String())
	return &command.Result{Success: true}, nil
}
