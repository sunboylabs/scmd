package repos

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/command"
	contextpkg "github.com/scmd/scmd/internal/context"
	"github.com/scmd/scmd/internal/tools"
)

// PluginCommand wraps a CommandSpec to implement the command.Command interface
type PluginCommand struct {
	spec *CommandSpec
}

// NewPluginCommand creates a new plugin command from a spec
func NewPluginCommand(spec *CommandSpec) *PluginCommand {
	return &PluginCommand{spec: spec}
}

// Name returns the command name
func (c *PluginCommand) Name() string {
	return c.spec.Name
}

// Aliases returns command aliases
func (c *PluginCommand) Aliases() []string {
	return c.spec.Aliases
}

// Description returns the command description
func (c *PluginCommand) Description() string {
	return c.spec.Description
}

// Usage returns usage information
func (c *PluginCommand) Usage() string {
	return c.spec.Usage
}

// Category returns the command category
func (c *PluginCommand) Category() command.Category {
	if c.spec.Category != "" {
		return command.Category(c.spec.Category)
	}
	return command.CategoryPlugin
}

// Examples returns example usages
func (c *PluginCommand) Examples() []string {
	return c.spec.Examples
}

// RequiresBackend returns true since plugin commands use LLM
func (c *PluginCommand) RequiresBackend() bool {
	return true
}

// Validate validates the command arguments
func (c *PluginCommand) Validate(args *command.Args) error {
	// Check required args
	for i, argSpec := range c.spec.Args {
		if argSpec.Required && i >= len(args.Positional) {
			// Check if default is provided
			if argSpec.Default == "" {
				return fmt.Errorf("missing required argument: %s", argSpec.Name)
			}
		}
	}
	return nil
}

// Execute runs the plugin command
func (c *PluginCommand) Execute(ctx context.Context, args *command.Args, execCtx *command.ExecContext) (*command.Result, error) {
	if err := c.Validate(args); err != nil {
		return &command.Result{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	if execCtx.Backend == nil {
		return &command.Result{
			Success: false,
			Error:   "no backend available",
		}, nil
	}

	// Execute pre-hooks
	if c.spec.Hooks != nil && len(c.spec.Hooks.Pre) > 0 {
		if err := c.executeHooks(ctx, c.spec.Hooks.Pre); err != nil {
			return &command.Result{
				Success: false,
				Error:   fmt.Sprintf("pre-hook failed: %v", err),
			}, nil
		}
	}

	// Check for template-based execution (NEW in v0.4.2)
	if c.spec.Template != nil {
		templateExecutor, err := NewTemplateExecutor(execCtx.DataDir)
		if err != nil {
			return &command.Result{
				Success: false,
				Error:   fmt.Sprintf("failed to initialize template executor: %v", err),
			}, nil
		}

		result, err := templateExecutor.ExecuteTemplateCommand(ctx, c.spec, args, execCtx)
		if err != nil {
			return &command.Result{
				Success: false,
				Error:   fmt.Sprintf("template execution failed: %v", err),
			}, nil
		}

		// Execute post-hooks
		if c.spec.Hooks != nil && len(c.spec.Hooks.Post) > 0 {
			if err := c.executeHooks(ctx, c.spec.Hooks.Post); err != nil {
				return &command.Result{
					Success: false,
					Error:   fmt.Sprintf("post-hook failed: %v", err),
				}, nil
			}
		}

		return result, nil
	}

	// Check for composition - delegate to composer if present
	if c.spec.Compose != nil && execCtx.Registry != nil {
		// Create loader and composer for composition execution
		manager := NewManager(execCtx.DataDir)
		installDir := execCtx.DataDir + "/commands"
		loader := NewLoader(manager, installDir)
		composer := NewComposer(execCtx.Registry, loader)
		result, err := composer.ExecuteComposed(ctx, c.spec, args, execCtx)
		if err != nil {
			return &command.Result{
				Success: false,
				Error:   fmt.Sprintf("composition failed: %v", err),
			}, nil
		}

		// Execute post-hooks after composition
		if c.spec.Hooks != nil && len(c.spec.Hooks.Post) > 0 {
			if err := c.executeHooks(ctx, c.spec.Hooks.Post); err != nil {
				return &command.Result{
					Success: false,
					Error:   fmt.Sprintf("post-hook failed: %v", err),
				}, nil
			}
		}

		return result, nil
	}

	// Build template context
	tmplCtx := c.buildTemplateContext(args)

	// Execute prompt template
	prompt, err := c.executeTemplate(c.spec.Prompt.Template, tmplCtx)
	if err != nil {
		return &command.Result{
			Success: false,
			Error:   fmt.Sprintf("template error: %v", err),
		}, nil
	}

	// Execute system template if present
	system := ""
	if c.spec.Prompt.System != "" {
		system, err = c.executeTemplate(c.spec.Prompt.System, tmplCtx)
		if err != nil {
			return &command.Result{
				Success: false,
				Error:   fmt.Sprintf("system template error: %v", err),
			}, nil
		}
	}

	// Gather automatic context if specified
	if c.spec.Context != nil {
		gatherer := contextpkg.NewGatherer("") // Use current working directory

		// Convert repos.ContextSpec to contextpkg.ContextSpec
		contextSpec := &contextpkg.ContextSpec{
			Files:     c.spec.Context.Files,
			Git:       c.spec.Context.Git,
			Env:       c.spec.Context.Env,
			MaxTokens: c.spec.Context.MaxTokens,
		}

		autoContext, err := gatherer.Gather(ctx, contextSpec)
		if err != nil {
			// Log warning but don't fail - context is supplementary
			if execCtx.UI != nil {
				execCtx.UI.WriteError(fmt.Sprintf("Warning: Failed to gather context: %v", err))
			}
		} else if autoContext != nil {
			// Prepend context to prompt
			contextStr := autoContext.Format()
			if contextStr != "" {
				prompt = contextStr + "\n---\n\n" + prompt
			}
		}
	}

	// Use tool calling if backend supports it
	var output string
	if execCtx.Backend.SupportsToolCalling() {
		// Create tool registry with confirmation UI
		var confirmUI tools.ConfirmUI
		if execCtx.UI != nil {
			confirmUI = execCtx.UI
		}

		toolRegistry := tools.DefaultRegistry(confirmUI)
		toolExecutor := tools.NewExecutor(toolRegistry, execCtx.Backend)

		output, err = toolExecutor.ExecuteWithTools(ctx, prompt, system)
		if err != nil {
			return &command.Result{
				Success: false,
				Error:   fmt.Sprintf("tool execution failed: %v", err),
			}, nil
		}
	} else {
		// Fall back to basic completion if no tool calling
		req := &backend.CompletionRequest{
			Prompt:       prompt,
			SystemPrompt: system,
			MaxTokens:    2048,
			Temperature:  0.7,
		}

		// Apply model preferences
		if c.spec.Model.MaxTokens > 0 {
			req.MaxTokens = c.spec.Model.MaxTokens
		}
		if c.spec.Model.Temperature > 0 {
			req.Temperature = c.spec.Model.Temperature
		}

		resp, err := execCtx.Backend.Complete(ctx, req)
		if err != nil {
			return &command.Result{
				Success: false,
				Error:   fmt.Sprintf("completion failed: %v", err),
			}, nil
		}
		output = resp.Content
	}

	// Execute post-hooks
	if c.spec.Hooks != nil && len(c.spec.Hooks.Post) > 0 {
		if err := c.executeHooks(ctx, c.spec.Hooks.Post); err != nil {
			return &command.Result{
				Success: false,
				Error:   fmt.Sprintf("post-hook failed: %v", err),
			}, nil
		}
	}

	return &command.Result{
		Success: true,
		Output:  output,
	}, nil
}

// buildTemplateContext creates the context for template execution
func (c *PluginCommand) buildTemplateContext(args *command.Args) map[string]interface{} {
	ctx := make(map[string]interface{})

	// Add positional args by name
	for i, argSpec := range c.spec.Args {
		if i < len(args.Positional) {
			ctx[argSpec.Name] = args.Positional[i]
		} else if argSpec.Default != "" {
			ctx[argSpec.Name] = argSpec.Default
		} else {
			ctx[argSpec.Name] = ""
		}
	}

	// Add flags by name
	for _, flagSpec := range c.spec.Flags {
		if val, ok := args.Options[flagSpec.Name]; ok {
			ctx[flagSpec.Name] = val
		} else if flagSpec.Default != "" {
			ctx[flagSpec.Name] = flagSpec.Default
		} else {
			ctx[flagSpec.Name] = ""
		}
	}

	// Add stdin if present
	if stdin, ok := args.Options["stdin"]; ok {
		ctx["stdin"] = stdin
		ctx["input"] = stdin // alias
	}

	// Add all positional args as array
	ctx["args"] = args.Positional

	// Join all positional args
	ctx["all_args"] = strings.Join(args.Positional, " ")

	return ctx
}

// executeTemplate executes a Go template with the given context
func (c *PluginCommand) executeTemplate(tmplStr string, ctx map[string]interface{}) (string, error) {
	tmpl, err := template.New("prompt").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// executeHooks executes a list of hook commands
func (c *PluginCommand) executeHooks(ctx context.Context, hooks []HookAction) error {
	for _, hook := range hooks {
		// Check condition if present
		if hook.If != "" {
			// Simple condition check - in production this would be more sophisticated
			// For now, skip conditional hooks
			continue
		}

		// Execute shell command
		if hook.Shell != "" {
			cmd := exec.CommandContext(ctx, "sh", "-c", hook.Shell)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("shell hook failed: %s (output: %s)", err, string(output))
			}
			continue
		}

		// Execute scmd command
		if hook.Command != "" {
			// Parse command and args
			parts := strings.Fields(hook.Command)
			if len(parts) == 0 {
				continue
			}
			// For now, execute as shell command
			// TODO: Use command registry to execute scmd commands
			cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("command hook failed: %s (output: %s)", err, string(output))
			}
		}
	}
	return nil
}

// Loader loads plugin commands from installed command specs
type Loader struct {
	manager    *Manager
	installDir string
}

// NewLoader creates a new plugin loader
func NewLoader(manager *Manager, installDir string) *Loader {
	return &Loader{
		manager:    manager,
		installDir: installDir,
	}
}

// LoadAll loads all installed plugin commands
func (l *Loader) LoadAll() ([]*PluginCommand, error) {
	specs, err := l.manager.LoadInstalledCommands(l.installDir)
	if err != nil {
		return nil, err
	}

	commands := make([]*PluginCommand, len(specs))
	for i, spec := range specs {
		commands[i] = NewPluginCommand(spec)
	}

	return commands, nil
}

// RegisterAll registers all plugin commands with the command registry
func (l *Loader) RegisterAll(registry *command.Registry) error {
	commands, err := l.LoadAll()
	if err != nil {
		return err
	}

	for _, cmd := range commands {
		if err := registry.Register(cmd); err != nil {
			// Skip if already registered (e.g., builtin with same name)
			continue
		}
	}

	return nil
}
