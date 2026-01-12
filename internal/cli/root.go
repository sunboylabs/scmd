package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	cfg             *config.Config
	verbose         bool
	promptFlag      string
	outputFlag      string
	formatFlag      string
	quietFlag       bool
	contextFlags    []string
	backendFlag     string
	modelFlag       string
	contextSizeFlag int

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
	rootCmd.PersistentFlags().IntVar(&contextSizeFlag, "context-size", 0, "max context size (0 = use model's native max)")

	// Pipe/prompt flags
	rootCmd.PersistentFlags().StringVarP(&promptFlag, "prompt", "p", "", "inline prompt")
	rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "output file")
	rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "auto", "output format: auto, markdown, plain")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "suppress progress")
	rootCmd.PersistentFlags().StringArrayVarP(&contextFlags, "context", "c", nil, "context files")

	// Add built-in commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(explainCmd)
	rootCmd.AddCommand(reviewCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(killProcessCmd)
	rootCmd.AddCommand(backendsCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(repoCmd)
	rootCmd.AddCommand(registryCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(lockCmd)
	rootCmd.AddCommand(cacheCmd)
	rootCmd.AddCommand(slashCmd)
	rootCmd.AddCommand(modelsCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(chatCmd)      // New chat command
	rootCmd.AddCommand(historyCmd)   // New history command
	rootCmd.AddCommand(templateCmd)  // New template command
	rootCmd.AddCommand(SetupCommand())

	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// completionCmd generates shell completion scripts
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for scmd.

To load completions:

Bash:
  $ source <(scmd completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ scmd completion bash > /etc/bash_completion.d/scmd
  # macOS:
  $ scmd completion bash > /usr/local/etc/bash_completion.d/scmd

Zsh:
  # If shell completion is not already enabled in your environment, you will need to enable it:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ scmd completion zsh > "${fpath[1]}/_scmd"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ scmd completion fish | source

  # To load completions for each session, execute once:
  $ scmd completion fish > ~/.config/fish/completions/scmd.fish

PowerShell:
  PS> scmd completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> scmd completion powershell > scmd.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	},
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
  cat file.py | scmd explain
  scmd explain loop.py --template beginner-explain`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuiltinCommandWithCmd(cmd, "explain", args)
	},
}

// reviewCmd wraps the builtin review command
var reviewCmd = &cobra.Command{
	Use:     "review [file]",
	Short:   "Review code for issues and improvements",
	Aliases: []string{"r"},
	Example: `  scmd review main.go
  git diff | scmd review
  scmd review auth.py --template security-review`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuiltinCommandWithCmd(cmd, "review", args)
	},
}

func init() {
	// Add --template flag to review and explain commands
	reviewCmd.Flags().String("template", "", "Use a prompt template")
	explainCmd.Flags().String("template", "", "Use a prompt template")
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

// killProcessCmd wraps the builtin kill-process command
var killProcessCmd = &cobra.Command{
	Use:     "kill-process <name>",
	Short:   "Find and kill processes by name",
	Aliases: []string{"kp", "killp"},
	Example: `  scmd kill-process cursor
  scmd /kp node
  scmd kill-process chrome`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuiltinCommand("kill-process", args)
	},
}

func runBuiltinCommand(name string, args []string) error {
	return runBuiltinCommandWithCmd(nil, name, args)
}

func runBuiltinCommandWithCmd(cmd *cobra.Command, name string, args []string) error {
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
	output, err := NewOutputWriter(&OutputConfig{
		FilePath: outputFlag,
		Mode:     mode,
		Format:   formatFlag,
		Config:   cfg,
	})
	if err != nil {
		return err
	}
	defer output.Close()

	// Get the best available backend
	activeBackend, err := getActiveBackend(ctx)
	if err != nil {
		// If user explicitly specified a backend, fail immediately
		if backendFlag != "" {
			return err
		}
		// Otherwise, just warn if verbose
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
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
		return NewCommandNotFoundError(name, cmdRegistry.Names())
	}

	// Build args
	cmdArgs := command.NewArgs()
	cmdArgs.Positional = args
	if stdinContent != "" {
		cmdArgs.Options["stdin"] = stdinContent
	}

	// Pass --template flag if present (for review and explain commands)
	if cmd != nil {
		if templateFlag, _ := cmd.Flags().GetString("template"); templateFlag != "" {
			cmdArgs.Options["template"] = templateFlag
		}
	}

	// Execute
	result, err := c.Execute(ctx, cmdArgs, execCtx)
	if err != nil {
		return err
	}

	if result.Output != "" {
		// Format output based on format flag
		if err := formatAndWriteOutput(output, result, formatFlag); err != nil {
			return err
		}
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

func preRun(cmd *cobra.Command, _ []string) error {
	var err error

	// Skip first-run check for certain commands
	skipCommands := map[string]bool{
		"help":       true,
		"version":    true,
		"setup":      true,
		"completion": true,
	}

	cmdName := cmd.Name()
	if !skipCommands[cmdName] && IsFirstRun() {
		// Check if running in interactive mode
		mode := DetectIOMode()
		if mode.Interactive {
			if err := RunSetupIfNeeded(); err != nil {
				return fmt.Errorf("setup failed: %w", err)
			}
		}
	}

	// Validate format flag if provided
	if formatFlag != "" {
		validFormats := []string{"auto", "markdown", "plain"}
		valid := false
		for _, f := range validFormats {
			if formatFlag == f {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid format '%s': must be one of: auto, markdown, plain", formatFlag)
		}
	}

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

	// Apply context size from config or CLI flag
	// Priority: CLI flag > config > model's native size
	contextSize := cfg.Backends.Local.ContextLength // From config
	if contextSizeFlag > 0 {
		contextSize = contextSizeFlag // CLI flag overrides config
	}
	if contextSize > 0 {
		llamaBackend.SetContextSize(contextSize)
	}

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

	// Apply configured default backend (if set)
	if cfg.Backends.Default != "" {
		if err := backendRegistry.SetDefault(cfg.Backends.Default); err != nil {
			// Log warning but don't fail - will fall back to availability check
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: configured backend '%s' not available, will auto-select\n", cfg.Backends.Default)
			}
		}
	}

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
			// Get list of available backends
			availableBackends := []string{}
			for _, backend := range backendRegistry.List() {
				availableBackends = append(availableBackends, backend.Name())
			}
			return nil, NewBackendNotFoundError(backendFlag, availableBackends)
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
	output, err := NewOutputWriter(&OutputConfig{
		FilePath: outputFlag,
		Mode:     mode,
		Format:   formatFlag,
		Config:   cfg,
	})
	if err != nil {
		return err
	}
	defer output.Close()

	// Get the best available backend
	activeBackend, err := getActiveBackend(ctx)
	if err != nil {
		// If user explicitly specified a backend, fail immediately
		if backendFlag != "" {
			return err
		}
		// Otherwise, just warn if verbose
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
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
				if err := formatAndWriteOutput(output, result, formatFlag); err != nil {
					return err
				}
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
		return NewNoBackendError()
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
		return NewCommandNotFoundError(cmdName, cmdRegistry.Names())
	}

	cmdArgs := command.NewArgs()
	cmdArgs.Positional = args
	cmdArgs.Options["stdin"] = stdin

	result, err := c.Execute(ctx, cmdArgs, execCtx)
	if err != nil {
		return err
	}

	if result.Output != "" {
		if err := formatAndWriteOutput(output, result, formatFlag); err != nil {
			return err
		}
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Error)
	}

	return nil
}

// formatAndWriteOutput formats the output based on the format flag and writes it
func formatAndWriteOutput(output *OutputWriter, result *command.Result, format string) error {
	switch format {
	case "json":
		// Create a JSON structure for the result
		jsonOutput := map[string]interface{}{
			"success": result.Success,
			"output":  result.Output,
		}
		if result.Error != "" {
			jsonOutput["error"] = result.Error
		}
		if len(result.Suggestions) > 0 {
			jsonOutput["suggestions"] = result.Suggestions
		}
		return output.WriteJSON(jsonOutput)
	case "markdown":
		// For markdown, wrap the output in a code block if it's not already markdown
		if !strings.Contains(result.Output, "```") {
			return output.WriteLine("```\n" + result.Output + "\n```")
		}
		return output.WriteLine(result.Output)
	default: // "text"
		// Check if the output looks like markdown (has headers, code blocks, etc.)
		if looksLikeMarkdown(result.Output) {
			return output.WriteMarkdown(result.Output)
		}
		return output.WriteLine(result.Output)
	}
}

// looksLikeMarkdown checks if a string appears to be markdown
func looksLikeMarkdown(s string) bool {
	// Check for common markdown patterns
	return strings.Contains(s, "```") ||
		strings.Contains(s, "##") ||
		strings.Contains(s, "**") ||
		strings.Contains(s, "- ") ||
		strings.Contains(s, "* ") ||
		strings.Contains(s, "`") ||
		strings.Contains(s, "[") && strings.Contains(s, "](")
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
	// Intercept slash commands before cobra processes them
	// Search all args for a slash command (not just position 1)
	// BUT: Exclude paths that look like files (e.g., /tmp/file.py, /home/user/script.sh)
	slashIndex := -1
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "/") {
			// Check if this looks like a file path vs a slash command
			// File paths typically have:
			// - Multiple slashes (e.g., /tmp/test.py)
			// - Known file extensions
			// - Exist on filesystem
			if isLikelyFilePath(arg) {
				continue
			}
			slashIndex = i
			break
		}
	}

	if slashIndex > 0 {
		// Found slash command - extract it and all args (including flags before it)
		slashCmd := os.Args[slashIndex]
		// Pass everything except the executable and the slash command itself
		// This includes flags before the slash command and args after it
		allArgs := append(os.Args[1:slashIndex], os.Args[slashIndex+1:]...)
		return runSlashCommand(slashCmd, allArgs)
	}

	return rootCmd.Execute()
}

// isLikelyFilePath determines if a string starting with / is a file path vs a slash command
func isLikelyFilePath(arg string) bool {
	// If it contains more than one slash, it's likely a path
	if strings.Count(arg, "/") > 1 {
		return true
	}

	// If it has a file extension, it's likely a path
	if strings.Contains(arg, ".") {
		return true
	}

	// If the file exists on filesystem, it's definitely a path
	if _, err := os.Stat(arg); err == nil {
		return true
	}

	return false
}

// runSlashCommand handles /command style invocations
func runSlashCommand(cmd string, args []string) error {
	// Strip leading slash
	cmdName := strings.TrimPrefix(cmd, "/")

	// Parse flags from args (e.g., --backend, --model, etc.)
	// This sets the global flag variables like backendFlag, modelFlag
	if err := rootCmd.ParseFlags(args); err != nil {
		return err
	}

	// Get the non-flag arguments (the actual command arguments)
	cmdArgs := rootCmd.Flags().Args()

	// Initialize everything via preRun
	if err := preRun(rootCmd, nil); err != nil {
		return err
	}

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
	output, err := NewOutputWriter(&OutputConfig{
		FilePath: outputFlag,
		Mode:     mode,
		Format:   formatFlag,
		Config:   cfg,
	})
	if err != nil {
		return err
	}
	defer output.Close()

	// Get the best available backend
	activeBackend, err := getActiveBackend(ctx)
	if err != nil {
		// If user explicitly specified a backend, fail immediately
		if backendFlag != "" {
			return err
		}
		// Otherwise, just warn if verbose
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
		}
	}

	// Create execution context
	execCtx := &command.ExecContext{
		Config:  cfg,
		Backend: activeBackend,
		UI:      NewConsoleUI(mode),
	}

	// Look up command in registry
	c, ok := cmdRegistry.Get(cmdName)
	if !ok {
		// Try aliases
		aliasMap := map[string]string{
			"e":   "explain",
			"r":   "review",
			"gc":  "commit",
			"sum": "summarize",
			"cfg": "config",
		}
		if alias, found := aliasMap[cmdName]; found {
			c, ok = cmdRegistry.Get(alias)
		}
	}

	if !ok {
		return NewCommandNotFoundError(cmdName, cmdRegistry.Names())
	}

	// Build command args (use cmdArgs from flag parsing, which has non-flag arguments)
	commandArgs := command.NewArgs()
	commandArgs.Positional = cmdArgs
	if stdinContent != "" {
		commandArgs.Options["stdin"] = stdinContent
	}

	// Execute
	result, err := c.Execute(ctx, commandArgs, execCtx)
	if err != nil {
		return err
	}

	if result.Output != "" {
		if err := formatAndWriteOutput(output, result, formatFlag); err != nil {
			return err
		}
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
