package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/scmd/scmd/internal/repos"
)

var repoManager *repos.Manager

// repoCmd is the parent command for repository operations
var repoCmd = &cobra.Command{
	Use:     "repo",
	Short:   "Manage slash command repositories",
	Aliases: []string{"repos"},
	Long: `Manage repositories that provide slash commands.

Repositories are remote sources of slash commands that you can install and use.
Each repository contains a manifest (scmd-repo.yaml) listing available commands.

Quick Start:
  1. Update repositories:  scmd repo update
  2. Search for commands:  scmd repo search <topic>
  3. View command details:  scmd repo show official/<command>
  4. Install a command:    scmd repo install official/<command>
  5. Use the command:      scmd /<command>

The official repository provides 100+ commands across categories like git, docker,
devops, code-review, documentation, debugging, data processing, and security.`,
}

// repoAddCmd adds a new repository
var repoAddCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add a repository",
	Args:  cobra.ExactArgs(2),
	Example: `  scmd repo add community https://raw.githubusercontent.com/scmd-community/commands/main
  scmd repo add myrepo https://example.com/scmd-commands`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, url := args[0], args[1]

		mgr, err := getRepoManager()
		if err != nil {
			return err
		}

		if err := mgr.Add(name, url); err != nil {
			return err
		}

		if err := mgr.Save(); err != nil {
			return fmt.Errorf("save repos: %w", err)
		}

		fmt.Printf("Added repository '%s' (%s)\n", name, url)

		// Try to fetch manifest to validate
		ctx := context.Background()
		repo, _ := mgr.Get(name)
		manifest, err := mgr.FetchManifest(ctx, repo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not fetch manifest: %v\n", err)
			return nil
		}

		fmt.Printf("  %d commands available\n", len(manifest.Commands))
		return nil
	},
}

// repoRemoveCmd removes a repository
var repoRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Remove a repository",
	Aliases: []string{"rm", "delete"},
	Args:    cobra.ExactArgs(1),
	Example: `  scmd repo remove community`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		mgr, err := getRepoManager()
		if err != nil {
			return err
		}

		if err := mgr.Remove(name); err != nil {
			return err
		}

		if err := mgr.Save(); err != nil {
			return fmt.Errorf("save repos: %w", err)
		}

		fmt.Printf("Removed repository '%s'\n", name)
		return nil
	},
}

// repoListCmd lists all repositories
var repoListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List configured repositories",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := getRepoManager()
		if err != nil {
			return err
		}

		repoList := mgr.List()
		if len(repoList) == 0 {
			fmt.Println("No repositories configured.")
			fmt.Println("Use 'scmd repo add <name> <url>' to add one.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tURL\tENABLED")
		for _, r := range repoList {
			enabled := "yes"
			if !r.Enabled {
				enabled = "no"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", r.Name, r.URL, enabled)
		}
		w.Flush()

		return nil
	},
}

// repoUpdateCmd fetches latest manifests from all repos
var repoUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update repository manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		mgr, err := getRepoManager()
		if err != nil {
			return err
		}

		repoList := mgr.List()
		if len(repoList) == 0 {
			fmt.Println("No repositories configured.")
			return nil
		}

		fmt.Println("Updating repositories...")
		for _, r := range repoList {
			if !r.Enabled {
				continue
			}
			fmt.Printf("  %s: ", r.Name)
			manifest, err := mgr.FetchManifest(ctx, r)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				continue
			}
			fmt.Printf("%d commands\n", len(manifest.Commands))
		}

		return nil
	},
}

// repoSearchCmd searches for commands across repos
var repoSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for commands in repositories",
	Example: `  scmd repo search git
  scmd repo search docker
  scmd repo search  # list all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		mgr, err := getRepoManager()
		if err != nil {
			return err
		}

		results, err := mgr.SearchCommands(ctx, query)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Println("No commands found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "COMMAND\tREPO\tDESCRIPTION")
		for _, r := range results {
			fmt.Fprintf(w, "%s\t%s\t%s\n", r.Command.Name, r.Repo, r.Command.Description)
		}
		w.Flush()

		return nil
	},
}

// repoShowCmd shows details about a specific command
var repoShowCmd = &cobra.Command{
	Use:   "show <repo>/<command>",
	Short: "Show details about a command from a repository",
	Args:  cobra.ExactArgs(1),
	Example: `  scmd repo show official/git-commit
  scmd repo show community/docker-compose`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Parse repo/command format
		repoCmd := args[0]
		var repoName, cmdName string
		for i, c := range repoCmd {
			if c == '/' {
				repoName = repoCmd[:i]
				cmdName = repoCmd[i+1:]
				break
			}
		}

		if repoName == "" || cmdName == "" {
			return fmt.Errorf("invalid format, use <repo>/<command>")
		}

		mgr, err := getRepoManager()
		if err != nil {
			return err
		}

		repo, ok := mgr.Get(repoName)
		if !ok {
			return fmt.Errorf("repository '%s' not found", repoName)
		}

		// Fetch manifest to find command
		manifest, err := mgr.FetchManifest(ctx, repo)
		if err != nil {
			return fmt.Errorf("fetch manifest: %w", err)
		}

		var cmdEntry *repos.Command
		for i := range manifest.Commands {
			if manifest.Commands[i].Name == cmdName {
				cmdEntry = &manifest.Commands[i]
				break
			}
		}

		if cmdEntry == nil {
			return fmt.Errorf("command '%s' not found in repository '%s'", cmdName, repoName)
		}

		// Fetch full command spec
		spec, err := mgr.FetchCommand(ctx, repo, cmdEntry.File)
		if err != nil {
			return fmt.Errorf("fetch command: %w", err)
		}

		// Display command info
		fmt.Printf("Name:        %s\n", spec.Name)
		fmt.Printf("Version:     %s\n", spec.Version)
		fmt.Printf("Description: %s\n", spec.Description)
		if spec.Author != "" {
			fmt.Printf("Author:      %s\n", spec.Author)
		}
		if spec.Category != "" {
			fmt.Printf("Category:    %s\n", spec.Category)
		}
		if spec.Usage != "" {
			fmt.Printf("Usage:       %s\n", spec.Usage)
		}
		if len(spec.Aliases) > 0 {
			fmt.Printf("Aliases:     %v\n", spec.Aliases)
		}

		if len(spec.Args) > 0 {
			fmt.Println("\nArguments:")
			for _, arg := range spec.Args {
				req := ""
				if arg.Required {
					req = " (required)"
				}
				fmt.Printf("  %s: %s%s\n", arg.Name, arg.Description, req)
			}
		}

		if len(spec.Flags) > 0 {
			fmt.Println("\nFlags:")
			for _, flag := range spec.Flags {
				short := ""
				if flag.Short != "" {
					short = fmt.Sprintf(" (-%s)", flag.Short)
				}
				fmt.Printf("  --%s%s: %s\n", flag.Name, short, flag.Description)
			}
		}

		if len(spec.Examples) > 0 {
			fmt.Println("\nExamples:")
			for _, ex := range spec.Examples {
				fmt.Printf("  %s\n", ex)
			}
		}

		return nil
	},
}

// repoInstallCmd installs a command from a repository
var repoInstallCmd = &cobra.Command{
	Use:   "install <repo>/<command>",
	Short: "Install a command from a repository",
	Args:  cobra.ExactArgs(1),
	Example: `  scmd repo install official/git-commit
  scmd repo install community/docker-compose`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Parse repo/command format
		repoCmd := args[0]
		var repoName, cmdName string
		for i, c := range repoCmd {
			if c == '/' {
				repoName = repoCmd[:i]
				cmdName = repoCmd[i+1:]
				break
			}
		}

		if repoName == "" || cmdName == "" {
			return fmt.Errorf("invalid format, use <repo>/<command>")
		}

		mgr, err := getRepoManager()
		if err != nil {
			return err
		}

		repo, ok := mgr.Get(repoName)
		if !ok {
			return fmt.Errorf("repository '%s' not found", repoName)
		}

		// Fetch manifest to find command
		manifest, err := mgr.FetchManifest(ctx, repo)
		if err != nil {
			return fmt.Errorf("fetch manifest: %w", err)
		}

		var cmdEntry *repos.Command
		for i := range manifest.Commands {
			if manifest.Commands[i].Name == cmdName {
				cmdEntry = &manifest.Commands[i]
				break
			}
		}

		if cmdEntry == nil {
			return fmt.Errorf("command '%s' not found in repository '%s'", cmdName, repoName)
		}

		// Fetch full command spec
		spec, err := mgr.FetchCommand(ctx, repo, cmdEntry.File)
		if err != nil {
			return fmt.Errorf("fetch command: %w", err)
		}

		// Save command locally
		installDir := filepath.Join(getDataDir(), "commands")
		if err := os.MkdirAll(installDir, 0755); err != nil {
			return fmt.Errorf("create install dir: %w", err)
		}

		// Save as installed command
		if err := mgr.InstallCommand(spec, installDir); err != nil {
			return fmt.Errorf("install command: %w", err)
		}

		fmt.Printf("Installed '%s' from '%s'\n", cmdName, repoName)
		fmt.Printf("Run with: scmd %s\n", cmdName)

		return nil
	},
}

// getRepoManager returns the repo manager, initializing if needed
func getRepoManager() (*repos.Manager, error) {
	if repoManager != nil {
		return repoManager, nil
	}

	dataDir := getDataDir()
	repoManager = repos.NewManager(dataDir)
	if err := repoManager.Load(); err != nil {
		return nil, fmt.Errorf("load repos: %w", err)
	}

	return repoManager, nil
}

// getDataDir returns the scmd data directory
func getDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".scmd")
}

func init() {
	// Add subcommands to repo command
	repoCmd.AddCommand(repoAddCmd)
	repoCmd.AddCommand(repoRemoveCmd)
	repoCmd.AddCommand(repoListCmd)
	repoCmd.AddCommand(repoUpdateCmd)
	repoCmd.AddCommand(repoSearchCmd)
	repoCmd.AddCommand(repoShowCmd)
	repoCmd.AddCommand(repoInstallCmd)
}
