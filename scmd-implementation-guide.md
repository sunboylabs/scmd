# scmd - Implementation Guide for Claude Code Agents

## Quick Reference

**Project:** scmd (slash command)  
**Language:** Go 1.22+  
**Repository:** github.com/scmd/scmd

---

## Agent Assignments

When working on scmd, identify which agent role you're acting as:

| Agent | Primary Focus | Quality Gates |
|-------|---------------|---------------|
| **PM** | Task coordination, priorities | Tasks have clear acceptance criteria |
| **ProdM** | Requirements, UX specs | User stories complete, errors helpful |
| **Arch** | Design, interfaces | No circular deps, ADRs documented |
| **Dev** | Implementation | Passes lint, documented, formatted |
| **Sec** | Security review | Checklist passed, no vulnerabilities |
| **QA** | Overall quality | Metrics met, cross-platform tested |
| **UnitTest** | Unit tests | >80% coverage, fast, no flaky tests |
| **IntTest** | Integration tests | All pass, isolated, reproducible |

---

## Phase 1: Project Bootstrap

### Step 1.1: Initialize Go Module

```bash
mkdir scmd && cd scmd
go mod init github.com/scmd/scmd
```

### Step 1.2: Install Dependencies

```bash
# CLI framework
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest

# TUI components
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbles@latest

# Terminal detection (for pipe support)
go get golang.org/x/term@latest

# Utilities
go get github.com/go-git/go-git/v5@latest
go get gopkg.in/yaml.v3@latest

# Testing
go get github.com/stretchr/testify@latest
```

### Step 1.3: Create Directory Structure

```bash
mkdir -p cmd/scmd
mkdir -p internal/{cli,command/builtin,backend/{local,ollama,claude,openai,mock},context,tools,plugins,repos,permissions,config,ui,models}
mkdir -p pkg/{version,errors,utils}
mkdir -p configs
mkdir -p scripts
mkdir -p tests/{unit,integration,e2e,fixtures,testutil}
mkdir -p docs/ADR
mkdir -p npm/platforms/{darwin-arm64,darwin-x64,linux-x64,linux-arm64,win32-x64}
```

### Step 1.4: File Inventory (Core I/O)

```
internal/cli/
├── root.go           # Main CLI with pipe support
├── mode.go           # TTY/pipe detection
├── stdin.go          # Stdin reader
├── output.go         # Output writer (stdout/file)
├── version.go        # Version command
├── repl.go           # Interactive REPL
└── completion.go     # Shell completions
```

---

## Phase 2: Implementation Order

Build files in this order to minimize dependency issues:

### 2.1: Foundation (No internal dependencies)

```
1. pkg/version/version.go
2. pkg/errors/errors.go
3. pkg/utils/strings.go
4. pkg/utils/files.go
5. pkg/utils/platform.go
```

### 2.2: Configuration

```
6. internal/config/config.go      # Config struct
7. internal/config/defaults.go    # Default values
8. internal/config/loader.go      # Load/save config
```

### 2.3: I/O Mode Detection & Piping

```
9. internal/cli/mode.go           # TTY/pipe detection
10. internal/cli/stdin.go         # Stdin reader
11. internal/cli/output.go        # Output writer
```

### 2.4: UI Components

```
12. internal/ui/colors.go         # Color utilities
13. internal/ui/spinner.go        # Loading spinner
14. internal/ui/progress.go       # Progress bar
15. internal/ui/stream.go         # Streaming output
16. internal/ui/prompt.go         # User prompts
```

### 2.5: Context Gathering

```
17. internal/context/interface.go # Context interface
18. internal/context/files.go     # File operations
19. internal/context/git.go       # Git operations
20. internal/context/project.go   # Project detection
21. internal/context/env.go       # Environment
22. internal/context/gatherer.go  # Main gatherer
```

### 2.6: Backend System

```
23. internal/backend/interface.go # Backend interface
24. internal/backend/registry.go  # Backend registry
25. internal/backend/mock/backend.go # Mock for testing
```

### 2.7: Command System

```
26. internal/command/interface.go # Command interface
27. internal/command/parser.go    # Argument parser
28. internal/command/registry.go  # Command registry
29. internal/command/builtin/help.go    # /help command
30. internal/command/builtin/prompt.go  # -p prompt command
```

### 2.8: CLI Entry Point

```
31. internal/cli/root.go          # Root command with pipe support
32. internal/cli/version.go       # Version command
33. cmd/scmd/main.go              # Entry point
```

### 2.8: First Built-in Command

```
30. internal/command/builtin/explain.go
31. internal/command/builtin/explain_test.go
```

---

## Phase 3: Core Implementation

### 3.1: Entry Point

```go
// cmd/scmd/main.go

package main

import (
    "os"

    "github.com/scmd/scmd/internal/cli"
)

func main() {
    if err := cli.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### 3.2: Version Package

```go
// pkg/version/version.go

package version

import (
    "fmt"
    "runtime"
)

// Set by ldflags
var (
    Version = "dev"
    Commit  = "none"
    Date    = "unknown"
)

// Info returns version information
func Info() string {
    return fmt.Sprintf("scmd %s (%s) built %s with %s",
        Version, Commit[:min(7, len(Commit))], Date, runtime.Version())
}

// Short returns just the version
func Short() string {
    return Version
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

### 3.3: Error Types

```go
// pkg/errors/errors.go

package errors

import (
    "errors"
    "fmt"
)

// Standard errors
var (
    ErrNotFound       = errors.New("not found")
    ErrInvalidInput   = errors.New("invalid input")
    ErrPermission     = errors.New("permission denied")
    ErrTimeout        = errors.New("operation timed out")
    ErrBackendFailed  = errors.New("backend operation failed")
    ErrConfigInvalid  = errors.New("invalid configuration")
)

// CommandError represents a command execution error
type CommandError struct {
    Command     string
    Message     string
    Suggestions []string
    Cause       error
}

func (e *CommandError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s: %v", e.Command, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Command, e.Message)
}

func (e *CommandError) Unwrap() error {
    return e.Cause
}

// NewCommandError creates a new command error
func NewCommandError(cmd, msg string, suggestions ...string) *CommandError {
    return &CommandError{
        Command:     cmd,
        Message:     msg,
        Suggestions: suggestions,
    }
}

// Wrap wraps an error with command context
func Wrap(cmd string, err error) *CommandError {
    return &CommandError{
        Command: cmd,
        Message: err.Error(),
        Cause:   err,
    }
}
```

### 3.4: Configuration

```go
// internal/config/config.go

package config

import (
    "os"
    "path/filepath"
)

// Config represents scmd configuration
type Config struct {
    Version  string         `mapstructure:"version"`
    Backends BackendsConfig `mapstructure:"backends"`
    UI       UIConfig       `mapstructure:"ui"`
    Models   ModelsConfig   `mapstructure:"models"`
}

// BackendsConfig for LLM backends
type BackendsConfig struct {
    Default string            `mapstructure:"default"`
    Local   LocalBackendConfig `mapstructure:"local"`
}

// LocalBackendConfig for local llama.cpp
type LocalBackendConfig struct {
    Model         string `mapstructure:"model"`
    ModelPath     string `mapstructure:"model_path"`
    ContextLength int    `mapstructure:"context_length"`
    GPULayers     int    `mapstructure:"gpu_layers"`
    Threads       int    `mapstructure:"threads"`
}

// UIConfig for UI preferences
type UIConfig struct {
    Streaming bool `mapstructure:"streaming"`
    Colors    bool `mapstructure:"colors"`
    Verbose   bool `mapstructure:"verbose"`
}

// ModelsConfig for model management
type ModelsConfig struct {
    Directory    string `mapstructure:"directory"`
    AutoDownload bool   `mapstructure:"auto_download"`
}

// DataDir returns the scmd data directory
func DataDir() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".scmd")
}

// ConfigPath returns the config file path
func ConfigPath() string {
    return filepath.Join(DataDir(), "config.yaml")
}
```

```go
// internal/config/defaults.go

package config

import (
    "path/filepath"
)

// Default returns default configuration
func Default() *Config {
    return &Config{
        Version: "1.0",
        Backends: BackendsConfig{
            Default: "local",
            Local: LocalBackendConfig{
                Model:         "qwen2.5-coder-1.5b",
                ContextLength: 8192,
                GPULayers:     0,
                Threads:       0,
            },
        },
        UI: UIConfig{
            Streaming: true,
            Colors:    true,
            Verbose:   false,
        },
        Models: ModelsConfig{
            Directory:    filepath.Join(DataDir(), "models"),
            AutoDownload: true,
        },
    }
}
```

```go
// internal/config/loader.go

package config

import (
    "os"
    "path/filepath"

    "github.com/spf13/viper"
)

// Load loads configuration from file and environment
func Load() (*Config, error) {
    v := viper.New()

    // Set defaults
    defaults := Default()
    v.SetDefault("version", defaults.Version)
    v.SetDefault("backends.default", defaults.Backends.Default)
    v.SetDefault("backends.local.model", defaults.Backends.Local.Model)
    v.SetDefault("backends.local.context_length", defaults.Backends.Local.ContextLength)
    v.SetDefault("ui.streaming", defaults.UI.Streaming)
    v.SetDefault("ui.colors", defaults.UI.Colors)
    v.SetDefault("models.directory", defaults.Models.Directory)
    v.SetDefault("models.auto_download", defaults.Models.AutoDownload)

    // Config file
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath(DataDir())

    // Environment variables
    v.SetEnvPrefix("SCMD")
    v.AutomaticEnv()

    // Read config
    if err := v.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
        // Config not found is OK, use defaults
    }

    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}

// Save saves configuration to file
func Save(cfg *Config) error {
    dir := DataDir()
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    v := viper.New()
    v.Set("version", cfg.Version)
    v.Set("backends", cfg.Backends)
    v.Set("ui", cfg.UI)
    v.Set("models", cfg.Models)

    return v.WriteConfigAs(filepath.Join(dir, "config.yaml"))
}

// EnsureDataDir creates the data directory if it doesn't exist
func EnsureDataDir() error {
    return os.MkdirAll(DataDir(), 0755)
}
```

### 3.5: I/O Mode Detection

```go
// internal/cli/mode.go

package cli

import (
    "os"

    "golang.org/x/term"
)

// IOMode represents the input/output mode
type IOMode struct {
    HasStdin    bool // Data is being piped in
    StdinIsTTY  bool // Stdin is a terminal
    StdoutIsTTY bool // Stdout is a terminal
    StderrIsTTY bool // Stderr is a terminal
    Interactive bool // Full interactive mode
    PipeIn      bool // Receiving piped input
    PipeOut     bool // Output is being piped
}

// DetectIOMode determines how scmd is being invoked
func DetectIOMode() *IOMode {
    stdinIsTTY := term.IsTerminal(int(os.Stdin.Fd()))
    stdoutIsTTY := term.IsTerminal(int(os.Stdout.Fd()))
    stderrIsTTY := term.IsTerminal(int(os.Stderr.Fd()))

    return &IOMode{
        HasStdin:    !stdinIsTTY,
        StdinIsTTY:  stdinIsTTY,
        StdoutIsTTY: stdoutIsTTY,
        StderrIsTTY: stderrIsTTY,
        Interactive: stdinIsTTY && stdoutIsTTY,
        PipeIn:      !stdinIsTTY,
        PipeOut:     !stdoutIsTTY,
    }
}

// ShouldStream returns true if output should stream
func (m *IOMode) ShouldStream() bool     { return m.StdoutIsTTY }
func (m *IOMode) ShouldShowProgress() bool { return m.StdoutIsTTY && m.StderrIsTTY }
func (m *IOMode) ShouldUseColors() bool   { return m.StdoutIsTTY }
func (m *IOMode) ProgressWriter() *os.File {
    if m.PipeOut { return os.Stderr }
    return os.Stdout
}
```

### 3.6: Stdin Reader

```go
// internal/cli/stdin.go

package cli

import (
    "context"
    "io"
    "os"
    "time"
)

// StdinReader handles piped input
type StdinReader struct {
    timeout time.Duration
    maxSize int64
}

func NewStdinReader() *StdinReader {
    return &StdinReader{timeout: 30 * time.Second, maxSize: 10 * 1024 * 1024}
}

// Read reads all stdin with timeout
func (r *StdinReader) Read(ctx context.Context) (string, error) {
    ctx, cancel := context.WithTimeout(ctx, r.timeout)
    defer cancel()

    type result struct { data []byte; err error }
    ch := make(chan result, 1)

    go func() {
        data, err := io.ReadAll(io.LimitReader(os.Stdin, r.maxSize))
        ch <- result{data, err}
    }()

    select {
    case res := <-ch:
        return string(res.data), res.err
    case <-ctx.Done():
        return "", ctx.Err()
    }
}
```

### 3.7: Output Writer

```go
// internal/cli/output.go

package cli

import (
    "bufio"
    "encoding/json"
    "io"
    "os"
    "sync"
)

type OutputWriter struct {
    mu       sync.Mutex
    writer   io.Writer
    buffered *bufio.Writer
    file     *os.File
    mode     *IOMode
}

type OutputConfig struct {
    FilePath string
    Mode     *IOMode
}

func NewOutputWriter(cfg *OutputConfig) (*OutputWriter, error) {
    var writer io.Writer = os.Stdout
    var file *os.File

    if cfg.FilePath != "" {
        f, err := os.Create(cfg.FilePath)
        if err != nil { return nil, err }
        writer, file = f, f
    }

    ow := &OutputWriter{writer: writer, mode: cfg.Mode, file: file}
    if cfg.Mode.PipeOut || cfg.FilePath != "" {
        ow.buffered = bufio.NewWriter(writer)
        ow.writer = ow.buffered
    }
    return ow, nil
}

func (w *OutputWriter) Write(s string) error {
    w.mu.Lock(); defer w.mu.Unlock()
    _, err := w.writer.Write([]byte(s)); return err
}

func (w *OutputWriter) WriteLine(s string) error { return w.Write(s + "\n") }

func (w *OutputWriter) WriteJSON(v interface{}) error {
    w.mu.Lock(); defer w.mu.Unlock()
    enc := json.NewEncoder(w.writer)
    enc.SetIndent("", "  ")
    return enc.Encode(v)
}

func (w *OutputWriter) Flush() error {
    w.mu.Lock(); defer w.mu.Unlock()
    if w.buffered != nil { return w.buffered.Flush() }
    return nil
}

func (w *OutputWriter) Close() error {
    w.Flush()
    if w.file != nil { return w.file.Close() }
    return nil
}
```

### 3.8: CLI Root with Pipe Support

```go
// internal/cli/root.go

package cli

import (
    "context"
    "fmt"

    "github.com/spf13/cobra"

    "github.com/scmd/scmd/internal/config"
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
)

var rootCmd = &cobra.Command{
    Use:   "scmd",
    Short: "AI-powered slash commands in your terminal",
    Long: `scmd brings AI-powered slash commands to any terminal.

Examples:
  scmd                           Start interactive mode
  scmd explain file.go           Explain code
  cat foo.md | scmd -p "summarize this" > summary.md
  git diff | scmd review -o review.md`,
    Version: version.Short(),
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        var err error
        cfg, err = config.Load()
        if err != nil {
            return fmt.Errorf("load config: %w", err)
        }
        return nil
    },
    RunE: runRoot,
}

func init() {
    // Global flags
    rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
    
    // Pipe/prompt flags
    rootCmd.PersistentFlags().StringVarP(&promptFlag, "prompt", "p", "", "inline prompt")
    rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "output file")
    rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "text", "output format: text, json, markdown")
    rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "suppress progress")
    rootCmd.PersistentFlags().StringArrayVarP(&contextFlags, "context", "c", nil, "context files")

    rootCmd.AddCommand(versionCmd)
    rootCmd.CompletionOptions.DisableDefaultCmd = true
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

    // Handle -p flag
    if promptFlag != "" {
        return runPrompt(ctx, promptFlag, stdinContent, mode, output)
    }

    // Pipe to command
    if mode.PipeIn && len(args) > 0 {
        return runCommandWithStdin(ctx, args[0], args[1:], stdinContent, mode, output)
    }

    // Interactive mode
    if mode.Interactive {
        return runREPL()
    }

    return cmd.Help()
}

func runPrompt(ctx context.Context, prompt, stdin string, mode *IOMode, output *OutputWriter) error {
    // TODO: Execute prompt with backend
    if !quietFlag && mode.StderrIsTTY {
        fmt.Fprintln(mode.ProgressWriter(), "⏳ Processing...")
    }
    
    // Placeholder
    result := fmt.Sprintf("Prompt: %s\nInput length: %d bytes", prompt, len(stdin))
    output.WriteLine(result)
    return nil
}

func runCommandWithStdin(ctx context.Context, cmd string, args []string, stdin string, mode *IOMode, output *OutputWriter) error {
    // TODO: Route to command with stdin
    return fmt.Errorf("command %s not implemented yet", cmd)
}

func runREPL() error {
    fmt.Println("Interactive mode coming soon...")
    fmt.Println("Use 'scmd --help' to see available commands")
    return nil
}

func Execute() error {
    return rootCmd.Execute()
}
```

```go
// internal/cli/version.go

package cli

import (
    "fmt"

    "github.com/spf13/cobra"

    "github.com/scmd/scmd/pkg/version"
)

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Print version information",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println(version.Info())
    },
}
```

### 3.9: Command Interface

```go
// internal/command/interface.go

package command

import (
    "context"
)

// Command defines the interface for all scmd commands
type Command interface {
    Name() string
    Aliases() []string
    Description() string
    Usage() string
    Examples() []string
    Category() Category
    Execute(ctx context.Context, args *Args, execCtx *ExecContext) (*Result, error)
    Validate(args *Args) error
    RequiresBackend() bool
}

// Category classifies commands
type Category string

const (
    CategoryCore   Category = "core"
    CategoryCode   Category = "code"
    CategoryGit    Category = "git"
    CategoryConfig Category = "config"
)

// Args represents parsed command arguments
type Args struct {
    Positional []string
    Flags      map[string]bool
    Options    map[string]string
    Raw        string
}

// NewArgs creates a new Args instance
func NewArgs() *Args {
    return &Args{
        Positional: []string{},
        Flags:      make(map[string]bool),
        Options:    make(map[string]string),
    }
}

// Result represents command execution result
type Result struct {
    Success     bool
    Output      string
    Error       string
    Suggestions []string
    ExitCode    int
}

// ExecContext provides execution dependencies
type ExecContext struct {
    Config *Config
    UI     UI
    // Backend will be added later
}

// UI interface for user interaction
type UI interface {
    Write(s string)
    WriteLine(s string)
    WriteError(s string)
    Confirm(prompt string) bool
    Spinner(message string) func()
}

// Config interface for command access to config
type Config interface {
    GetString(key string) string
    GetBool(key string) bool
    GetInt(key string) int
}
```

### 3.7: Help Command

```go
// internal/command/builtin/help.go

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

func (c *HelpCommand) Name() string        { return "help" }
func (c *HelpCommand) Aliases() []string   { return []string{"h", "?"} }
func (c *HelpCommand) Description() string { return "Show help for commands" }
func (c *HelpCommand) Usage() string       { return "/help [command]" }
func (c *HelpCommand) Category() command.Category { return command.CategoryCore }
func (c *HelpCommand) RequiresBackend() bool      { return false }

func (c *HelpCommand) Examples() []string {
    return []string{
        "/help",
        "/help explain",
    }
}

func (c *HelpCommand) Validate(args *command.Args) error {
    return nil
}

func (c *HelpCommand) Execute(
    ctx context.Context,
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

    for _, cat := range []command.Category{
        command.CategoryCore,
        command.CategoryCode,
        command.CategoryGit,
        command.CategoryConfig,
    } {
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
```

### 3.8: Command Registry

```go
// internal/command/registry.go

package command

import (
    "fmt"
    "sync"
)

// Registry manages all available commands
type Registry struct {
    mu       sync.RWMutex
    commands map[string]Command
    aliases  map[string]string
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
    return &Registry{
        commands: make(map[string]Command),
        aliases:  make(map[string]string),
    }
}

// Register adds a command to the registry
func (r *Registry) Register(cmd Command) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    name := cmd.Name()
    if _, exists := r.commands[name]; exists {
        return fmt.Errorf("command already registered: %s", name)
    }

    r.commands[name] = cmd

    for _, alias := range cmd.Aliases() {
        if existing, exists := r.aliases[alias]; exists {
            return fmt.Errorf("alias %s already used by %s", alias, existing)
        }
        r.aliases[alias] = name
    }

    return nil
}

// Get retrieves a command by name or alias
func (r *Registry) Get(nameOrAlias string) (Command, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    if cmd, ok := r.commands[nameOrAlias]; ok {
        return cmd, true
    }

    if name, ok := r.aliases[nameOrAlias]; ok {
        return r.commands[name], true
    }

    return nil, false
}

// List returns all registered commands
func (r *Registry) List() []Command {
    r.mu.RLock()
    defer r.mu.RUnlock()

    cmds := make([]Command, 0, len(r.commands))
    for _, cmd := range r.commands {
        cmds = append(cmds, cmd)
    }
    return cmds
}

// ListByCategory returns commands filtered by category
func (r *Registry) ListByCategory(cat Category) []Command {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var cmds []Command
    for _, cmd := range r.commands {
        if cmd.Category() == cat {
            cmds = append(cmds, cmd)
        }
    }
    return cmds
}
```

---

## Phase 4: Testing Setup

### 4.1: Test Utilities

```go
// tests/testutil/mock_ui.go

package testutil

import (
    "bytes"
    "fmt"
)

// MockUI implements command.UI for testing
type MockUI struct {
    Output  bytes.Buffer
    Errors  bytes.Buffer
    Prompts []string
    PromptResponses []bool
    promptIndex int
}

// NewMockUI creates a new mock UI
func NewMockUI() *MockUI {
    return &MockUI{
        PromptResponses: []bool{true},
    }
}

func (u *MockUI) Write(s string) {
    u.Output.WriteString(s)
}

func (u *MockUI) WriteLine(s string) {
    u.Output.WriteString(s + "\n")
}

func (u *MockUI) WriteError(s string) {
    u.Errors.WriteString(s + "\n")
}

func (u *MockUI) Confirm(prompt string) bool {
    u.Prompts = append(u.Prompts, prompt)
    if u.promptIndex < len(u.PromptResponses) {
        resp := u.PromptResponses[u.promptIndex]
        u.promptIndex++
        return resp
    }
    return true
}

func (u *MockUI) Spinner(message string) func() {
    u.Output.WriteString(fmt.Sprintf("[spinner] %s\n", message))
    return func() {
        u.Output.WriteString("[spinner done]\n")
    }
}

// GetOutput returns all output
func (u *MockUI) GetOutput() string {
    return u.Output.String()
}

// GetErrors returns all errors
func (u *MockUI) GetErrors() string {
    return u.Errors.String()
}
```

```go
// tests/testutil/mock_backend.go

package testutil

import (
    "context"

    "github.com/scmd/scmd/internal/backend"
)

// MockBackend implements backend.Backend for testing
type MockBackend struct {
    response string
    err      error
}

// NewMockBackend creates a new mock backend
func NewMockBackend() *MockBackend {
    return &MockBackend{}
}

// SetResponse sets the response to return
func (b *MockBackend) SetResponse(response string) {
    b.response = response
}

// SetError sets the error to return
func (b *MockBackend) SetError(err error) {
    b.err = err
}

func (b *MockBackend) Name() string       { return "mock" }
func (b *MockBackend) Type() backend.Type { return "mock" }
func (b *MockBackend) SupportsToolCalling() bool { return false }

func (b *MockBackend) Initialize(ctx context.Context) error {
    return nil
}

func (b *MockBackend) IsAvailable(ctx context.Context) (bool, error) {
    return true, nil
}

func (b *MockBackend) Shutdown(ctx context.Context) error {
    return nil
}

func (b *MockBackend) Stream(
    ctx context.Context,
    req *backend.CompletionRequest,
) (<-chan backend.StreamChunk, error) {
    if b.err != nil {
        return nil, b.err
    }

    ch := make(chan backend.StreamChunk)
    go func() {
        defer close(ch)
        ch <- backend.StreamChunk{Content: b.response}
        ch <- backend.StreamChunk{Done: true}
    }()
    return ch, nil
}

func (b *MockBackend) Complete(
    ctx context.Context,
    req *backend.CompletionRequest,
) (*backend.CompletionResponse, error) {
    if b.err != nil {
        return nil, b.err
    }
    return &backend.CompletionResponse{
        Content:      b.response,
        FinishReason: backend.FinishComplete,
    }, nil
}

func (b *MockBackend) CompleteWithTools(
    ctx context.Context,
    req *backend.ToolRequest,
) (*backend.ToolResponse, error) {
    return nil, nil
}

func (b *MockBackend) ModelInfo() *backend.ModelInfo {
    return &backend.ModelInfo{
        Name:          "mock-model",
        ContextLength: 8192,
    }
}

func (b *MockBackend) EstimateTokens(text string) int {
    return len(text) / 4
}
```

### 4.2: Sample Unit Test

```go
// internal/command/registry_test.go

package command

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

type mockCommand struct {
    name        string
    aliases     []string
    description string
    category    Category
}

func (c *mockCommand) Name() string        { return c.name }
func (c *mockCommand) Aliases() []string   { return c.aliases }
func (c *mockCommand) Description() string { return c.description }
func (c *mockCommand) Usage() string       { return "/" + c.name }
func (c *mockCommand) Examples() []string  { return nil }
func (c *mockCommand) Category() Category  { return c.category }
func (c *mockCommand) RequiresBackend() bool { return false }
func (c *mockCommand) Validate(args *Args) error { return nil }
func (c *mockCommand) Execute(ctx context.Context, args *Args, execCtx *ExecContext) (*Result, error) {
    return &Result{Success: true}, nil
}

func TestRegistry_Register(t *testing.T) {
    r := NewRegistry()

    cmd := &mockCommand{
        name:    "test",
        aliases: []string{"t"},
    }

    err := r.Register(cmd)
    require.NoError(t, err)

    // Should find by name
    found, ok := r.Get("test")
    assert.True(t, ok)
    assert.Equal(t, "test", found.Name())

    // Should find by alias
    found, ok = r.Get("t")
    assert.True(t, ok)
    assert.Equal(t, "test", found.Name())
}

func TestRegistry_Register_Duplicate(t *testing.T) {
    r := NewRegistry()

    cmd1 := &mockCommand{name: "test"}
    cmd2 := &mockCommand{name: "test"}

    err := r.Register(cmd1)
    require.NoError(t, err)

    err = r.Register(cmd2)
    assert.Error(t, err)
}

func TestRegistry_ListByCategory(t *testing.T) {
    r := NewRegistry()

    r.Register(&mockCommand{name: "cmd1", category: CategoryCode})
    r.Register(&mockCommand{name: "cmd2", category: CategoryCode})
    r.Register(&mockCommand{name: "cmd3", category: CategoryGit})

    codeCmds := r.ListByCategory(CategoryCode)
    assert.Len(t, codeCmds, 2)

    gitCmds := r.ListByCategory(CategoryGit)
    assert.Len(t, gitCmds, 1)
}
```

---

## Phase 5: Makefile

```makefile
# Makefile

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X github.com/scmd/scmd/pkg/version.Version=$(VERSION)

.PHONY: all build test lint clean install dev fmt

all: lint test build

# Build
build:
	go build -ldflags "$(LDFLAGS)" -o bin/scmd ./cmd/scmd

# Test
test:
	go test -race -coverprofile=coverage.out ./...

test-short:
	go test -short -race ./...

coverage:
	go tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

# Development
dev:
	go run ./cmd/scmd

install: build
	cp bin/scmd /usr/local/bin/

# Clean
clean:
	rm -rf bin/ dist/ coverage.out coverage.html

# Dependencies
deps:
	go mod tidy
	go mod verify

# Generate
generate:
	go generate ./...
```

---

## Phase 6: Linter Configuration

```yaml
# .golangci.yaml

run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - misspell

linters-settings:
  errcheck:
    check-type-assertions: true
  govet:
    check-shadowing: true

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
```

---

## Agent Handoff Checklist

When completing a task, ensure:

### Developer Agent
- [ ] Code compiles: `go build ./...`
- [ ] Code formatted: `gofmt -s -w .`
- [ ] No lint errors: `golangci-lint run`
- [ ] All exported symbols documented
- [ ] No hardcoded strings

### Unit Test Agent
- [ ] Tests pass: `go test ./...`
- [ ] Coverage adequate: `go test -cover ./...`
- [ ] Edge cases covered
- [ ] No flaky tests

### Security Agent
- [ ] No shell injection
- [ ] Input validated
- [ ] No secrets in code
- [ ] Permissions checked

### QA Agent
- [ ] Feature complete per spec
- [ ] Error messages helpful
- [ ] Documentation updated
- [ ] Cross-platform considered

---

## Quick Commands

```bash
# Build and run
make dev

# Run tests
make test

# Check lint
make lint

# Build binary
make build

# Install locally
make install
```

---

*This guide is designed to be fed to Claude Code agents for implementation.*
