package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/backend/llamacpp"
	"github.com/scmd/scmd/internal/backend/mock"
	"github.com/scmd/scmd/internal/backend/ollama"
	"github.com/scmd/scmd/internal/backend/openai"
	"github.com/scmd/scmd/internal/command"
	"github.com/scmd/scmd/internal/command/builtin"
	"github.com/scmd/scmd/internal/config"
	"github.com/scmd/scmd/internal/repos"
	"github.com/scmd/scmd/pkg/version"
)

var (
	cfg          *config.Config
	verbose      bool
	promptFlag   string
	outputFlag   string
	formatFlag   string
	quietFlag    bool
	contextFlags []string
	backendFlag  string
	modelFlag    string

	// Global registries
	cmdRegistry     *command.Registry
	backendRegistry *backend.Registry
)

var rootCmd = &cobra.Command{
	Use:   "scmd",
	Short: "AI-powered slash commands in your terminal",
	Long: `scmd brings AI-powered slash commands to any terminal.

Backends (in order of preference):
  - Ollama (local): Runs free open-source models locally
  - OpenAI/Together/Groq: Set API key via environment variable

Examples:
  scmd                           Start interactive mode
  scmd explain file.go           Explain code
  cat foo.md | scmd -p "summarize this" > summary.md
  git diff | scmd review -o review.md

Environment Variables:
  OLLAMA_HOST          Ollama server URL (default: http://localhost:11434)
  OPENAI_API_KEY       OpenAI API key
  TOGETHER_API_KEY     Together.ai API key
  GROQ_API_KEY         Groq API key`,
	Version:           version.Short(),
	PersistentPreRunE: preRun,
	RunE:              runRoot,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Backend flags
	rootCmd.PersistentFlags().StringVarP(&backendFlag, "backend", "b", "", "backend to use: ollama, openai, together, groq")
	rootCmd.PersistentFlags().StringVarP(&modelFlag, "model", "m", "", "model to use (overrides default)")

	// Pipe/prompt flags
	rootCmd.PersistentFlags().StringVarP(&promptFlag, "prompt", "p", "", "inline prompt")
	rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "output file")
	rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "text", "output format: text, json, markdown")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "suppress progress")
	rootCmd.PersistentFlags().StringArrayVarP(&contextFlags, "context", "c", nil, "context files")

	// Add built-in commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(explainCmd)
	rootCmd.AddCommand(reviewCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(backendsCmd)
	rootCmd.AddCommand(repoCmd)
	rootCmd.AddCommand(registryCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(lockCmd)
	rootCmd.AddCommand(cacheCmd)
	rootCmd.AddCommand(slashCmd)
	rootCmd.AddCommand(modelsCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// backendsCmd lists available backends
var backendsCmd = &cobra.Command{
	Use:   "backends",
	Short: "List available LLM backends",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		fmt.Println("Available backends:")
		fmt.Println()

		for _, b := range backendRegistry.List() {
			avail, _ := b.IsAvailable(ctx)
			status := "✗"
			if avail {
				status = "✓"
			}
			fmt.Printf("  %s %-12s %s\n", status, b.Name(), b.ModelInfo().Name)
		}

		fmt.Println()
		fmt.Println("Environment variables:")
		fmt.Println("  OLLAMA_HOST        - Ollama server (default: http://localhost:11434)")
		fmt.Println("  OPENAI_API_KEY     - OpenAI API key")
		fmt.Println("  TOGETHER_API_KEY   - Together.ai API key")
		fmt.Println("  GROQ_API_KEY       - Groq API key")

		return nil
	},
}

// explainCmd wraps the builtin explain command
var explainCmd = &cobra.Command{
	Use:     "explain [file|concept]",
	Short:   "Explain code or concepts",
	Aliases: []string{"e", "what"},
	Example: `  scmd explain main.go
  scmd explain "what is a goroutine"
  cat file.py | scmd explain`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuiltinCommand("explain", args)
	},
}

// reviewCmd wraps the builtin review command
var reviewCmd = &cobra.Command{
	Use:     "review [file]",
	Short:   "Review code for issues and improvements",
	Aliases: []string{"r"},
	Example: `  scmd review main.go
  git diff | scmd review`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuiltinCommand("review", args)
	},
}

// configCmd wraps the builtin config command
var configCmd = &cobra.Command{
	Use:     "config [key] [value]",
	Short:   "View or modify configuration",
	Aliases: []string{"cfg"},
	Example: `  scmd config
  scmd config backends.default
  scmd config ui.colors true`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuiltinCommand("config", args)
	},
}

func runBuiltinCommand(name string, args []string) error {
	ctx := context.Background()
	mode := DetectIOMode()

	// Read stdin if piped
	var stdinContent string
	if mode.PipeIn {
		reader := NewStdinReader()
		content, err := reader.Read(ctx)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		stdinContent = content
	}

	// Setup output
	output, err := NewOutputWriter(&OutputConfig{FilePath: outputFlag, Mode: mode})
	if err != nil {
		return err
	}
	defer output.Close()

	// Get the best available backend
	activeBackend, err := getActiveBackend(ctx)
	if err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	// Create execution context
	execCtx := &command.ExecContext{
		Config:  cfg,
		Backend: activeBackend,
		UI:      NewConsoleUI(mode),
	}

	// Get the command
	c, ok := cmdRegistry.Get(name)
	if !ok {
		return fmt.Errorf("unknown command: %s", name)
	}

	// Build args
	cmdArgs := command.NewArgs()
	cmdArgs.Positional = args
	if stdinContent != "" {
		cmdArgs.Options["stdin"] = stdinContent
	}

	// Execute
	result, err := c.Execute(ctx, cmdArgs, execCtx)
	if err != nil {
		return err
	}

	if result.Output != "" {
		output.WriteLine(result.Output)
	}

	if !result.Success {
		if len(result.Suggestions) > 0 {
			fmt.Fprintln(os.Stderr, "Suggestions:")
			for _, s := range result.Suggestions {
				fmt.Fprintf(os.Stderr, "  - %s\n", s)
			}
		}
		return fmt.Errorf("%s", result.Error)
	}

	return nil
}

func preRun(_ *cobra.Command, _ []string) error {
	var err error

	// Load configuration
	cfg, err = config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Initialize registries
	cmdRegistry = command.NewRegistry()
	backendRegistry = backend.NewRegistry()

	// Register backends in order of preference
	dataDir := getDataDir()

	// 1. llama.cpp (local, built-in, no setup required)
	llamaBackend := llamacpp.New(dataDir)
	_ = backendRegistry.Register(llamaBackend)

	// 2. Ollama (local, if installed)
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}
	ollamaBackend := ollama.New(&ollama.Config{
		BaseURL: ollamaHost,
		Model:   cfg.Backends.Local.Model,
	})
	_ = backendRegistry.Register(ollamaBackend)

	// 3. Groq (fast, free tier available)
	if os.Getenv("GROQ_API_KEY") != "" {
		groqBackend := openai.NewGroq(os.Getenv("GROQ_API_KEY"))
		_ = backendRegistry.Register(groqBackend)
	}

	// 3. Together.ai (many models, competitive pricing)
	if os.Getenv("TOGETHER_API_KEY") != "" {
		togetherBackend := openai.NewTogether(os.Getenv("TOGETHER_API_KEY"))
		_ = backendRegistry.Register(togetherBackend)
	}

	// 4. OpenAI (proprietary, high quality)
	if os.Getenv("OPENAI_API_KEY") != "" {
		openaiBackend := openai.NewOpenAI(os.Getenv("OPENAI_API_KEY"))
		_ = backendRegistry.Register(openaiBackend)
	}

	// 5. Mock backend (fallback for testing)
	mockBackend := mock.New()
	_ = backendRegistry.Register(mockBackend)

	// Register built-in commands
	if err := builtin.RegisterAll(cmdRegistry); err != nil {
		return fmt.Errorf("register commands: %w", err)
	}

	// Load plugin commands from installed repos
	mgr := repos.NewManager(dataDir)
	_ = mgr.Load() // Ignore error, repos may not exist yet

	loader := repos.NewLoader(mgr, filepath.Join(dataDir, "commands"))
	_ = loader.RegisterAll(cmdRegistry) // Ignore errors, commands may not exist yet

	return nil
}

// getActiveBackend returns the best available backend
func getActiveBackend(ctx context.Context) (backend.Backend, error) {
	// If user specified a backend, use it
	if backendFlag != "" {
		b, ok := backendRegistry.Get(backendFlag)
		if !ok {
			return nil, fmt.Errorf("unknown backend: %s", backendFlag)
		}
		if modelFlag != "" {
			// Set model if supported
			if setter, ok := b.(interface{ SetModel(string) }); ok {
				setter.SetModel(modelFlag)
			}
		}
		return b, nil
	}

	// Try to find an available backend
	b, err := backendRegistry.GetAvailable(ctx)
	if err != nil {
		// Fall back to mock
		if mock, ok := backendRegistry.Get("mock"); ok {
			return mock, nil
		}
		return nil, err
	}

	if modelFlag != "" {
		if setter, ok := b.(interface{ SetModel(string) }); ok {
			setter.SetModel(modelFlag)
		}
	}

	return b, nil
}

func runRoot(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	mode := DetectIOMode()

	// Read stdin if piped
	var stdinContent string
	if mode.PipeIn {
		reader := NewStdinReader()
		content, err := reader.Read(ctx)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		stdinContent = content
	}

	// Setup output
	output, err := NewOutputWriter(&OutputConfig{FilePath: outputFlag, Mode: mode})
	if err != nil {
		return err
	}
	defer output.Close()

	// Get the best available backend
	activeBackend, err := getActiveBackend(ctx)
	if err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	// Create execution context
	execCtx := &command.ExecContext{
		Config:  cfg,
		Backend: activeBackend,
		UI:      NewConsoleUI(mode),
	}

	// Handle -p flag
	if promptFlag != "" {
		return runPrompt(ctx, promptFlag, stdinContent, mode, output, execCtx)
	}

	// Handle command by name from internal registry (for slash commands in REPL)
	if len(args) > 0 {
		cmdName := args[0]
		if c, ok := cmdRegistry.Get(cmdName); ok {
			cmdArgs := command.NewArgs()
			cmdArgs.Positional = args[1:]
			if stdinContent != "" {
				cmdArgs.Options["stdin"] = stdinContent
			}
			result, err := c.Execute(ctx, cmdArgs, execCtx)
			if err != nil {
				return err
			}
			if result.Output != "" {
				output.WriteLine(result.Output)
			}
			if !result.Success {
				return fmt.Errorf("%s", result.Error)
			}
			return nil
		}
	}

	// Pipe to command
	if mode.PipeIn && len(args) > 0 {
		return runCommandWithStdin(ctx, args[0], args[1:], stdinContent, mode, output, execCtx)
	}

	// Interactive mode or help
	if mode.Interactive {
		return runREPL(execCtx)
	}

	return cmd.Help()
}

func runPrompt(ctx context.Context, prompt, stdin string, mode *IOMode, output *OutputWriter, execCtx *command.ExecContext) error {
	if execCtx.Backend == nil {
		return fmt.Errorf("no backend available. Install Ollama or set an API key")
	}

	if !quietFlag && mode.StderrIsTTY {
		fmt.Fprintf(mode.ProgressWriter(), "⏳ Using %s...\n", execCtx.Backend.Name())
	}

	fullPrompt := prompt
	if stdin != "" {
		fullPrompt = fmt.Sprintf("%s\n\nInput:\n%s", prompt, stdin)
	}

	req := &backend.CompletionRequest{
		Prompt:      fullPrompt,
		MaxTokens:   2048,
		Temperature: 0.7,
	}

	// Use streaming if TTY
	if mode.StdoutIsTTY && !mode.PipeOut {
		ch, err := execCtx.Backend.Stream(ctx, req)
		if err != nil {
			// Fall back to non-streaming
			resp, err := execCtx.Backend.Complete(ctx, req)
			if err != nil {
				return fmt.Errorf("completion failed: %w", err)
			}
			output.WriteLine(resp.Content)
			return nil
		}

		for chunk := range ch {
			if chunk.Error != nil {
				return fmt.Errorf("stream error: %w", chunk.Error)
			}
			fmt.Print(chunk.Content)
			if chunk.Done {
				fmt.Println()
				break
			}
		}
		return nil
	}

	// Non-streaming for pipes
	resp, err := execCtx.Backend.Complete(ctx, req)
	if err != nil {
		return fmt.Errorf("completion failed: %w", err)
	}

	output.WriteLine(resp.Content)
	return nil
}

func runCommandWithStdin(ctx context.Context, cmdName string, args []string, stdin string, mode *IOMode, output *OutputWriter, execCtx *command.ExecContext) error {
	c, ok := cmdRegistry.Get(cmdName)
	if !ok {
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	cmdArgs := command.NewArgs()
	cmdArgs.Positional = args
	cmdArgs.Options["stdin"] = stdin

	result, err := c.Execute(ctx, cmdArgs, execCtx)
	if err != nil {
		return err
	}

	if result.Output != "" {
		output.WriteLine(result.Output)
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Error)
	}

	return nil
}

func runREPL(execCtx *command.ExecContext) error {
	fmt.Println("scmd - AI-powered slash commands")

	// Show which backend is active
	if execCtx.Backend != nil {
		info := execCtx.Backend.ModelInfo()
		fmt.Printf("Using: %s (%s)\n", execCtx.Backend.Name(), info.Name)
	} else {
		fmt.Println("Warning: No backend available")
	}

	fmt.Println("Type /help for available commands")
	fmt.Println()

	// Simple REPL - for now just show help
	helpCmd, _ := cmdRegistry.Get("help")
	if helpCmd != nil {
		_, _ = helpCmd.Execute(context.Background(), command.NewArgs(), execCtx)
	}

	return nil
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// ConsoleUI implements command.UI for terminal output
type ConsoleUI struct {
	mode *IOMode
}

// NewConsoleUI creates a new console UI
func NewConsoleUI(mode *IOMode) *ConsoleUI {
	return &ConsoleUI{mode: mode}
}

// Write writes to stdout
func (u *ConsoleUI) Write(s string) {
	fmt.Print(s)
}

// WriteLine writes a line to stdout
func (u *ConsoleUI) WriteLine(s string) {
	fmt.Println(s)
}

// WriteError writes to stderr
func (u *ConsoleUI) WriteError(s string) {
	fmt.Fprintln(os.Stderr, s)
}

// Confirm prompts for confirmation
func (u *ConsoleUI) Confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	var response string
	fmt.Scanln(&response)
	return response == "y" || response == "Y"
}

// Spinner shows a loading spinner (simplified)
func (u *ConsoleUI) Spinner(message string) func() {
	if u.mode.StdoutIsTTY {
		fmt.Printf("⏳ %s...", message)
	}
	return func() {
		if u.mode.StdoutIsTTY {
			fmt.Println(" ✓")
		}
	}
}
