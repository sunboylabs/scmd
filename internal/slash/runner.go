package slash

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/command"
	"github.com/scmd/scmd/internal/repos"
)

// SlashCommand represents a configured slash command
type SlashCommand struct {
	Name        string   `yaml:"name"`
	Command     string   `yaml:"command"`     // The actual command to run (e.g., "git-commit")
	Aliases     []string `yaml:"aliases"`     // Short aliases (e.g., ["gc", "gitc"])
	Description string   `yaml:"description"`
	Args        string   `yaml:"args"`        // Default args to pass
	Stdin       bool     `yaml:"stdin"`       // Whether to read stdin
}

// Config holds all slash command configurations
type Config struct {
	Commands []SlashCommand `yaml:"commands"`
}

// Runner executes slash commands
type Runner struct {
	config      *Config
	dataDir     string
	registry    *command.Registry
	repoManager *repos.Manager
	loader      *repos.Loader
}

// NewRunner creates a new slash command runner
func NewRunner(dataDir string, registry *command.Registry, repoManager *repos.Manager) *Runner {
	installDir := filepath.Join(dataDir, "commands")
	loader := repos.NewLoader(repoManager, installDir)

	return &Runner{
		config:      &Config{},
		dataDir:     dataDir,
		registry:    registry,
		repoManager: repoManager,
		loader:      loader,
	}
}

// LoadConfig loads slash command configuration
func (r *Runner) LoadConfig() error {
	// Ensure data directory exists
	if err := os.MkdirAll(r.dataDir, 0755); err != nil {
		return err
	}

	// Load from ~/.scmd/slash.yaml
	configPath := filepath.Join(r.dataDir, "slash.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config
			r.config = defaultConfig()
			return r.SaveConfig()
		}
		return err
	}

	return yaml.Unmarshal(data, r.config)
}

// SaveConfig saves the slash command configuration
func (r *Runner) SaveConfig() error {
	// Ensure data directory exists
	if err := os.MkdirAll(r.dataDir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(r.config)
	if err != nil {
		return err
	}

	configPath := filepath.Join(r.dataDir, "slash.yaml")
	return os.WriteFile(configPath, data, 0644)
}

// defaultConfig creates default slash commands
func defaultConfig() *Config {
	return &Config{
		Commands: []SlashCommand{
			{
				Name:        "explain",
				Command:     "explain",
				Aliases:     []string{"e", "exp"},
				Description: "Explain code or text",
				Stdin:       true,
			},
			{
				Name:        "review",
				Command:     "review",
				Aliases:     []string{"r", "rev"},
				Description: "Review code",
				Stdin:       true,
			},
			{
				Name:        "commit",
				Command:     "git-commit",
				Aliases:     []string{"gc", "gitc"},
				Description: "Generate git commit message",
				Stdin:       true,
			},
			{
				Name:        "summarize",
				Command:     "summarize",
				Aliases:     []string{"s", "sum", "tldr"},
				Description: "Summarize text",
				Stdin:       true,
			},
			{
				Name:        "fix",
				Command:     "explain-error",
				Aliases:     []string{"f", "err"},
				Description: "Explain and fix errors",
				Stdin:       true,
			},
		},
	}
}

// Parse parses a slash command string (e.g., "/gc some args")
func (r *Runner) Parse(input string) (cmd *SlashCommand, args []string, err error) {
	input = strings.TrimSpace(input)

	// Must start with /
	if !strings.HasPrefix(input, "/") {
		return nil, nil, fmt.Errorf("not a slash command (must start with /)")
	}

	// Remove leading /
	input = input[1:]

	// Split into command and args
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, nil, fmt.Errorf("empty slash command")
	}

	cmdName := parts[0]
	args = parts[1:]

	// Find matching command
	cmd = r.FindCommand(cmdName)
	if cmd == nil {
		return nil, nil, fmt.Errorf("unknown slash command: /%s", cmdName)
	}

	return cmd, args, nil
}

// FindCommand finds a slash command by name or alias
func (r *Runner) FindCommand(name string) *SlashCommand {
	name = strings.ToLower(name)

	for i := range r.config.Commands {
		cmd := &r.config.Commands[i]

		// Check name
		if strings.ToLower(cmd.Name) == name {
			return cmd
		}

		// Check aliases
		for _, alias := range cmd.Aliases {
			if strings.ToLower(alias) == name {
				return cmd
			}
		}
	}

	return nil
}

// Run executes a slash command
func (r *Runner) Run(ctx context.Context, slashCmd *SlashCommand, args []string, stdin string, be backend.Backend) (*command.Result, error) {
	// Load plugin commands
	if err := r.loader.RegisterAll(r.registry); err != nil {
		// Ignore - commands may not be installed yet
	}

	// Find the actual command
	cmd, ok := r.registry.Get(slashCmd.Command)
	if !ok {
		return nil, fmt.Errorf("command '%s' not found. Install with: scmd repo install <repo>/%s",
			slashCmd.Command, slashCmd.Command)
	}

	// Build args
	cmdArgs := command.NewArgs()
	cmdArgs.Positional = args

	if stdin != "" {
		cmdArgs.Options["stdin"] = stdin
	}

	// Add default args from slash config
	if slashCmd.Args != "" {
		for _, part := range strings.Fields(slashCmd.Args) {
			if strings.Contains(part, "=") {
				kv := strings.SplitN(part, "=", 2)
				cmdArgs.Options[kv[0]] = kv[1]
			}
		}
	}

	// Build execution context
	execCtx := &command.ExecContext{
		Backend: be,
		UI:      &simpleUI{},
	}

	return cmd.Execute(ctx, cmdArgs, execCtx)
}

// simpleUI implements command.UI with minimal output
type simpleUI struct{}

func (u *simpleUI) Write(s string)      { fmt.Print(s) }
func (u *simpleUI) WriteLine(s string)  { fmt.Println(s) }
func (u *simpleUI) WriteError(s string) { fmt.Fprintln(os.Stderr, s) }
func (u *simpleUI) Confirm(prompt string) bool { return true }
func (u *simpleUI) Spinner(message string) func() {
	// No-op spinner for non-interactive use
	return func() {}
}

// List returns all configured slash commands
func (r *Runner) List() []SlashCommand {
	return r.config.Commands
}

// Add adds a new slash command
func (r *Runner) Add(cmd SlashCommand) error {
	// Check for duplicates
	if r.FindCommand(cmd.Name) != nil {
		return fmt.Errorf("slash command '%s' already exists", cmd.Name)
	}

	for _, alias := range cmd.Aliases {
		if r.FindCommand(alias) != nil {
			return fmt.Errorf("alias '%s' already exists", alias)
		}
	}

	r.config.Commands = append(r.config.Commands, cmd)
	return r.SaveConfig()
}

// Remove removes a slash command
func (r *Runner) Remove(name string) error {
	name = strings.ToLower(name)

	for i, cmd := range r.config.Commands {
		if strings.ToLower(cmd.Name) == name {
			r.config.Commands = append(r.config.Commands[:i], r.config.Commands[i+1:]...)
			return r.SaveConfig()
		}
	}

	return fmt.Errorf("slash command '%s' not found", name)
}

// AddAlias adds an alias to an existing command
func (r *Runner) AddAlias(cmdName, alias string) error {
	// Check alias doesn't exist
	if r.FindCommand(alias) != nil {
		return fmt.Errorf("alias '%s' already exists", alias)
	}

	for i, cmd := range r.config.Commands {
		if strings.ToLower(cmd.Name) == strings.ToLower(cmdName) {
			r.config.Commands[i].Aliases = append(r.config.Commands[i].Aliases, alias)
			return r.SaveConfig()
		}
	}

	return fmt.Errorf("command '%s' not found", cmdName)
}

// GenerateShellIntegration generates shell functions for slash commands
func (r *Runner) GenerateShellIntegration(shell string) string {
	switch shell {
	case "bash", "zsh":
		return r.generateBashZsh()
	case "fish":
		return r.generateFish()
	default:
		return r.generateBashZsh()
	}
}

func (r *Runner) generateBashZsh() string {
	var sb strings.Builder

	sb.WriteString(`# scmd slash command integration
# Add to your ~/.bashrc or ~/.zshrc

# Main slash command function
/ () {
    if [ -z "$1" ]; then
        scmd slash list
        return
    fi

    # Check if there's piped input
    if [ ! -t 0 ]; then
        cat | scmd slash run "$@"
    else
        scmd slash run "$@"
    fi
}

# Individual command aliases
`)

	for _, cmd := range r.config.Commands {
		// Main command alias
		fmt.Fprintf(&sb, "alias /%s='/ %s'\n", cmd.Name, cmd.Name)

		// Short aliases
		for _, alias := range cmd.Aliases {
			fmt.Fprintf(&sb, "alias /%s='/ %s'\n", alias, cmd.Name)
		}
	}

	sb.WriteString(`
# Completion (bash)
_scmd_slash_completions() {
    local commands="`)

	var names []string
	for _, cmd := range r.config.Commands {
		names = append(names, cmd.Name)
		names = append(names, cmd.Aliases...)
	}
	sb.WriteString(strings.Join(names, " "))

	sb.WriteString(`"
    COMPREPLY=($(compgen -W "$commands" -- "${COMP_WORDS[1]}"))
}
complete -F _scmd_slash_completions /
`)

	return sb.String()
}

func (r *Runner) generateFish() string {
	var sb strings.Builder

	sb.WriteString(`# scmd slash command integration for fish
# Add to your ~/.config/fish/config.fish

function /
    if test (count $argv) -eq 0
        scmd slash list
        return
    end

    if not isatty stdin
        cat | scmd slash run $argv
    else
        scmd slash run $argv
    end
end

# Individual command aliases
`)

	for _, cmd := range r.config.Commands {
		fmt.Fprintf(&sb, "alias /%s '/ %s'\n", cmd.Name, cmd.Name)
		for _, alias := range cmd.Aliases {
			fmt.Fprintf(&sb, "alias /%s '/ %s'\n", alias, cmd.Name)
		}
	}

	sb.WriteString(`
# Completions
set -l slash_commands `)

	var names []string
	for _, cmd := range r.config.Commands {
		names = append(names, cmd.Name)
		names = append(names, cmd.Aliases...)
	}
	sb.WriteString(strings.Join(names, " "))

	sb.WriteString(`
complete -c / -f -a "$slash_commands"
`)

	return sb.String()
}
