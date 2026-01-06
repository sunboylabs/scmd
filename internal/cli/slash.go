package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/scmd/scmd/internal/slash"
)

var slashRunner *slash.Runner

// slashCmd is the parent command for slash command operations
var slashCmd = &cobra.Command{
	Use:   "slash",
	Short: "Manage and run slash commands",
	Long: `Manage slash commands - quick aliases for AI commands.

Slash commands provide a fast way to invoke AI commands:
  /gc           -> git-commit (generate commit message)
  /explain      -> explain code
  /review       -> review code
  /sum          -> summarize text

Use 'scmd slash init' to set up shell integration.`,
}

// slashRunCmd runs a slash command
var slashRunCmd = &cobra.Command{
	Use:   "run <command> [args...]",
	Short: "Run a slash command",
	Args:  cobra.MinimumNArgs(1),
	Example: `  scmd slash run gc
  cat main.go | scmd slash run explain
  scmd slash run review --focus=security`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		runner, err := getSlashRunner()
		if err != nil {
			return err
		}

		// Find command
		slashCmd := runner.FindCommand(args[0])
		if slashCmd == nil {
			return fmt.Errorf("unknown slash command: %s\nRun 'scmd slash list' to see available commands", args[0])
		}

		// Read stdin if available
		var stdin string
		if slashCmd.Stdin {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				stdin = string(data)
			}
		}

		// Get backend
		be, err := getActiveBackend(ctx)
		if err != nil {
			return err
		}

		// Run command
		result, err := runner.Run(ctx, slashCmd, args[1:], stdin, be)
		if err != nil {
			return err
		}

		if !result.Success {
			return fmt.Errorf("%s", result.Error)
		}

		fmt.Print(result.Output)
		if !strings.HasSuffix(result.Output, "\n") {
			fmt.Println()
		}

		return nil
	},
}

// slashListCmd lists all slash commands
var slashListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List configured slash commands",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		runner, err := getSlashRunner()
		if err != nil {
			return err
		}

		commands := runner.List()
		if len(commands) == 0 {
			fmt.Println("No slash commands configured.")
			fmt.Println("Run 'scmd slash add' to create one.")
			return nil
		}

		fmt.Println("Slash Commands:")
		fmt.Println()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "COMMAND\tALIASES\tRUNS\tDESCRIPTION")
		for _, c := range commands {
			aliases := strings.Join(c.Aliases, ", ")
			if aliases == "" {
				aliases = "-"
			}
			fmt.Fprintf(w, "/%s\t%s\t%s\t%s\n", c.Name, aliases, c.Command, c.Description)
		}
		w.Flush()

		fmt.Println()
		fmt.Println("Usage: /<command> [args]  or  scmd slash run <command> [args]")

		return nil
	},
}

// slashAddCmd adds a new slash command
var slashAddCmd = &cobra.Command{
	Use:   "add <name> <command>",
	Short: "Add a new slash command",
	Args:  cobra.ExactArgs(2),
	Example: `  scmd slash add commit git-commit
  scmd slash add doc generate-docs --alias=d,docs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		runner, err := getSlashRunner()
		if err != nil {
			return err
		}

		name, command := args[0], args[1]
		aliasFlag, _ := cmd.Flags().GetString("alias")
		desc, _ := cmd.Flags().GetString("description")
		stdin, _ := cmd.Flags().GetBool("stdin")

		var aliases []string
		if aliasFlag != "" {
			aliases = strings.Split(aliasFlag, ",")
			for i := range aliases {
				aliases[i] = strings.TrimSpace(aliases[i])
			}
		}

		slashCmd := slash.SlashCommand{
			Name:        name,
			Command:     command,
			Aliases:     aliases,
			Description: desc,
			Stdin:       stdin,
		}

		if err := runner.Add(slashCmd); err != nil {
			return err
		}

		fmt.Printf("Added slash command: /%s -> %s\n", name, command)
		if len(aliases) > 0 {
			fmt.Printf("Aliases: %s\n", strings.Join(aliases, ", "))
		}

		return nil
	},
}

// slashRemoveCmd removes a slash command
var slashRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Remove a slash command",
	Aliases: []string{"rm", "delete"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		runner, err := getSlashRunner()
		if err != nil {
			return err
		}

		if err := runner.Remove(args[0]); err != nil {
			return err
		}

		fmt.Printf("Removed slash command: /%s\n", args[0])
		return nil
	},
}

// slashAliasCmd adds an alias to a command
var slashAliasCmd = &cobra.Command{
	Use:   "alias <command> <alias>",
	Short: "Add an alias to a slash command",
	Args:  cobra.ExactArgs(2),
	Example: `  scmd slash alias commit gc
  scmd slash alias explain e`,
	RunE: func(cmd *cobra.Command, args []string) error {
		runner, err := getSlashRunner()
		if err != nil {
			return err
		}

		if err := runner.AddAlias(args[0], args[1]); err != nil {
			return err
		}

		fmt.Printf("Added alias: /%s -> /%s\n", args[1], args[0])
		return nil
	},
}

// slashInitCmd sets up shell integration
var slashInitCmd = &cobra.Command{
	Use:   "init [shell]",
	Short: "Generate shell integration for slash commands",
	Long: `Generate shell integration code for slash commands.

This creates a '/' function in your shell that enables:
  /gc              -> run git-commit
  cat file | /exp  -> run explain with file as input
  /sum article.md  -> run summarize

Supported shells: bash, zsh, fish`,
	Args: cobra.MaximumNArgs(1),
	Example: `  # Add to your ~/.bashrc or ~/.zshrc:
  eval "$(scmd slash init bash)"

  # For fish, add to ~/.config/fish/config.fish:
  scmd slash init fish | source`,
	RunE: func(cmd *cobra.Command, args []string) error {
		runner, err := getSlashRunner()
		if err != nil {
			return err
		}

		shell := "bash"
		if len(args) > 0 {
			shell = args[0]
		}

		integration := runner.GenerateShellIntegration(shell)
		fmt.Print(integration)

		return nil
	},
}

// slashInteractiveCmd starts interactive slash mode
var slashInteractiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start interactive slash command mode",
	Aliases: []string{"i", "repl"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		runner, err := getSlashRunner()
		if err != nil {
			return err
		}

		be, err := getActiveBackend(ctx)
		if err != nil {
			return err
		}

		fmt.Println("scmd interactive mode. Type /help for commands, /quit to exit.")
		fmt.Println()

		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("/ ")
			if !scanner.Scan() {
				break
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			// Handle special commands
			switch input {
			case "/quit", "/exit", "/q":
				fmt.Println("Goodbye!")
				return nil
			case "/help", "/?":
				fmt.Println("Commands:")
				for _, c := range runner.List() {
					aliases := ""
					if len(c.Aliases) > 0 {
						aliases = fmt.Sprintf(" (%s)", strings.Join(c.Aliases, ", "))
					}
					fmt.Printf("  /%s%s - %s\n", c.Name, aliases, c.Description)
				}
				fmt.Println("  /quit - exit")
				continue
			case "/list", "/ls":
				for _, c := range runner.List() {
					fmt.Printf("  /%s -> %s\n", c.Name, c.Command)
				}
				continue
			}

			// Parse and run slash command
			if !strings.HasPrefix(input, "/") {
				input = "/" + input
			}

			slashCmd, cmdArgs, err := runner.Parse(input)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			// For interactive mode, prompt for input if command expects stdin
			var stdin string
			if slashCmd.Stdin && len(cmdArgs) == 0 {
				fmt.Print("Input (end with empty line):\n")
				var lines []string
				for scanner.Scan() {
					line := scanner.Text()
					if line == "" {
						break
					}
					lines = append(lines, line)
				}
				stdin = strings.Join(lines, "\n")
			}

			result, err := runner.Run(ctx, slashCmd, cmdArgs, stdin, be)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			if !result.Success {
				fmt.Printf("Error: %s\n", result.Error)
				continue
			}

			fmt.Println()
			fmt.Println(result.Output)
			fmt.Println()
		}

		return nil
	},
}

func getSlashRunner() (*slash.Runner, error) {
	if slashRunner != nil {
		return slashRunner, nil
	}

	dataDir := getDataDir()
	mgr, err := getRepoManager()
	if err != nil {
		return nil, err
	}

	slashRunner = slash.NewRunner(dataDir, cmdRegistry, mgr)
	if err := slashRunner.LoadConfig(); err != nil {
		return nil, fmt.Errorf("load slash config: %w", err)
	}

	return slashRunner, nil
}

func init() {
	// Add flags
	slashAddCmd.Flags().String("alias", "", "comma-separated aliases (e.g., gc,gitc)")
	slashAddCmd.Flags().String("description", "", "command description")
	slashAddCmd.Flags().Bool("stdin", true, "command accepts stdin input")

	// Add subcommands
	slashCmd.AddCommand(slashRunCmd)
	slashCmd.AddCommand(slashListCmd)
	slashCmd.AddCommand(slashAddCmd)
	slashCmd.AddCommand(slashRemoveCmd)
	slashCmd.AddCommand(slashAliasCmd)
	slashCmd.AddCommand(slashInitCmd)
	slashCmd.AddCommand(slashInteractiveCmd)
}
