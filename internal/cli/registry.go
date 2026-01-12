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

// registryCmd provides access to the central scmd registry
var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Search and browse the scmd command registry",
	Long: `Access the central scmd command registry with 100+ commands.

The registry provides a curated collection of slash commands from verified
publishers, organized by category with ratings and usage statistics.

Popular Categories:
  🔀 Git workflow       💻 Code analysis      🚀 DevOps
  📊 Data processing    📝 Documentation      🐛 Debugging
  🔒 Security           🐚 Shell utilities

Examples:
  scmd registry search git            Find git-related commands
  scmd registry featured              Show trending commands
  scmd registry search --category=devops
  scmd repo install official/<name>   Install a command`,
	Aliases: []string{"reg"},
}

// registrySearchCmd searches the registry
var registrySearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search the central registry for commands",
	Example: `  scmd registry search git
  scmd registry search --category=devops
  scmd registry search --verified --sort=downloads`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		registry := repos.NewRegistry("")

		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		category, _ := cmd.Flags().GetString("category")
		verified, _ := cmd.Flags().GetBool("verified")
		featured, _ := cmd.Flags().GetBool("featured")
		sortBy, _ := cmd.Flags().GetString("sort")
		limit, _ := cmd.Flags().GetInt("limit")

		opts := repos.SearchOptions{
			Query:    query,
			Category: category,
			Verified: verified,
			Featured: featured,
			SortBy:   sortBy,
			Limit:    limit,
		}

		results, err := registry.SearchCommands(ctx, opts)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Println("No commands found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "COMMAND\tREPO\tRATING\tDOWNLOADS\tDESCRIPTION")
		for _, r := range results {
			verified := ""
			if r.Verified {
				verified = "✓"
			}
			rating := fmt.Sprintf("%.1f", r.Rating)
			if r.RatingCount == 0 {
				rating = "-"
			}
			fmt.Fprintf(w, "%s%s\t%s\t%s\t%d\t%s\n",
				r.Name, verified, r.Repo, rating, r.Downloads, truncate(r.Description, 40))
		}
		w.Flush()

		return nil
	},
}

// registryFeaturedCmd shows featured/trending commands
var registryFeaturedCmd = &cobra.Command{
	Use:     "featured",
	Short:   "Show featured and trending commands",
	Aliases: []string{"trending", "popular"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		registry := repos.NewRegistry("")

		results, err := registry.GetFeatured(ctx)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Println("No featured commands available.")
			return nil
		}

		fmt.Println("🔥 Featured Commands:")
		fmt.Println()
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for i, r := range results {
			verified := ""
			if r.Verified {
				verified = " ✓"
			}
			fmt.Fprintf(w, "%d.\t%s%s\t(%s)\t%s\n",
				i+1, r.Name, verified, r.Repo, truncate(r.Description, 50))
		}
		w.Flush()

		return nil
	},
}

// registryCategoriesCmd lists available categories
var registryCategoriesCmd = &cobra.Command{
	Use:     "categories",
	Short:   "List available command categories",
	Aliases: []string{"cats"},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		registry := repos.NewRegistry("")

		categories, err := registry.GetCategories(ctx)
		if err != nil {
			return err
		}

		fmt.Println("Available Categories:")
		fmt.Println()
		for _, cat := range categories {
			count := ""
			if cat.Count > 0 {
				count = fmt.Sprintf(" (%d)", cat.Count)
			}
			fmt.Printf("  %s %s%s\n     %s\n\n",
				cat.Icon, cat.Name, count, cat.Description)
		}

		return nil
	},
}

// updateCmd checks for and applies updates
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for and install command updates",
	Long: `Check installed commands for available updates.

Use --check to only check without installing.
Use --all to update all commands at once.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		checkOnly, _ := cmd.Flags().GetBool("check")
		updateAll, _ := cmd.Flags().GetBool("all")

		dataDir := getDataDir()
		cache := repos.NewCache(dataDir)
		if err := cache.Load(); err != nil {
			return fmt.Errorf("load cache: %w", err)
		}

		mgr, err := getRepoManager()
		if err != nil {
			return err
		}

		// Check for updates
		updates, err := cache.CheckUpdates(func(repo, name string) (string, error) {
			r, ok := mgr.Get(repo)
			if !ok {
				return "", fmt.Errorf("repo not found")
			}
			manifest, err := mgr.FetchManifest(ctx, r)
			if err != nil {
				return "", err
			}
			for _, c := range manifest.Commands {
				if c.Name == name {
					// Fetch full spec to get version
					spec, err := mgr.FetchCommand(ctx, r, c.File)
					if err != nil {
						return "", err
					}
					return spec.Version, nil
				}
			}
			return "", fmt.Errorf("command not found")
		})

		if err != nil {
			return fmt.Errorf("check updates: %w", err)
		}

		if len(updates) == 0 {
			fmt.Println("All commands are up to date.")
			return nil
		}

		fmt.Printf("Found %d update(s):\n\n", len(updates))
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "COMMAND\tCURRENT\tLATEST")
		for _, u := range updates {
			fmt.Fprintf(w, "%s/%s\t%s\t%s\n", u.Repo, u.Command, u.Current, u.Latest)
		}
		w.Flush()

		if checkOnly {
			fmt.Println("\nRun 'scmd update' to install updates.")
			return nil
		}

		if !updateAll && len(updates) > 1 {
			fmt.Println("\nUse 'scmd update --all' to update all, or specify command names.")
			return nil
		}

		// Install updates
		installDir := filepath.Join(dataDir, "commands")
		for _, u := range updates {
			fmt.Printf("Updating %s/%s to %s...\n", u.Repo, u.Command, u.Latest)

			repo, _ := mgr.Get(u.Repo)
			manifest, _ := mgr.FetchManifest(ctx, repo)

			var cmdFile string
			for _, c := range manifest.Commands {
				if c.Name == u.Command {
					cmdFile = c.File
					break
				}
			}

			spec, err := mgr.FetchCommand(ctx, repo, cmdFile)
			if err != nil {
				fmt.Printf("  Error: %v\n", err)
				continue
			}

			if err := mgr.InstallCommand(spec, installDir); err != nil {
				fmt.Printf("  Error: %v\n", err)
				continue
			}

			cache.MarkInstalled(u.Repo, u.Command, spec.Version)
			fmt.Printf("  Updated to %s\n", spec.Version)
		}

		if err := cache.Save(); err != nil {
			return fmt.Errorf("save cache: %w", err)
		}

		return nil
	},
}

// lockCmd manages lockfiles for reproducible installations
var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Manage command lockfiles",
	Long: `Create and use lockfiles for reproducible command installations.

Lockfiles record exact versions of installed commands, allowing
teams to share consistent command configurations.`,
}

// lockGenerateCmd generates a lockfile
var lockGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a lockfile from installed commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = "scmd.lock"
		}

		dataDir := getDataDir()
		cache := repos.NewCache(dataDir)
		if err := cache.Load(); err != nil {
			return fmt.Errorf("load cache: %w", err)
		}

		lf := cache.GenerateLockfile()

		if err := repos.SaveLockfile(lf, output); err != nil {
			return fmt.Errorf("save lockfile: %w", err)
		}

		fmt.Printf("Generated lockfile: %s (%d commands)\n", output, len(lf.Commands))
		return nil
	},
}

// lockInstallCmd installs from a lockfile
var lockInstallCmd = &cobra.Command{
	Use:   "install [lockfile]",
	Short: "Install commands from a lockfile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		input := "scmd.lock"
		if len(args) > 0 {
			input = args[0]
		}

		lf, err := repos.LoadLockfile(input)
		if err != nil {
			return fmt.Errorf("load lockfile: %w", err)
		}

		fmt.Printf("Installing %d commands from %s...\n", len(lf.Commands), input)

		mgr, err := getRepoManager()
		if err != nil {
			return err
		}

		dataDir := getDataDir()
		installDir := filepath.Join(dataDir, "commands")

		if err := mgr.InstallFromLockfile(ctx, lf, installDir); err != nil {
			return fmt.Errorf("install: %w", err)
		}

		fmt.Println("Done.")
		return nil
	},
}

// cacheCmd manages the local cache
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage the local command cache",
}

// cacheStatsCmd shows cache statistics
var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show cache statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := getDataDir()
		cache := repos.NewCache(dataDir)
		if err := cache.Load(); err != nil {
			return fmt.Errorf("load cache: %w", err)
		}

		stats := cache.Stats()

		fmt.Printf("Cache Statistics:\n")
		fmt.Printf("  Cached repositories:  %d\n", stats.CachedRepos)
		fmt.Printf("  Cached manifests:     %d\n", stats.CachedManifests)
		fmt.Printf("  Cached commands:      %d\n", stats.CachedCommands)
		fmt.Printf("  Installed commands:   %d\n", stats.InstalledCommands)
		if !stats.LastUpdated.IsZero() {
			fmt.Printf("  Last updated:         %s\n", stats.LastUpdated.Format("2006-01-02 15:04:05"))
		}

		return nil
	},
}

// cacheClearCmd clears the cache
var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the local cache",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := getDataDir()
		cache := repos.NewCache(dataDir)

		if err := cache.Clear(); err != nil {
			return fmt.Errorf("clear cache: %w", err)
		}

		fmt.Println("Cache cleared.")
		return nil
	},
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func init() {
	// Registry search flags
	registrySearchCmd.Flags().String("category", "", "filter by category")
	registrySearchCmd.Flags().Bool("verified", false, "show only verified commands")
	registrySearchCmd.Flags().Bool("featured", false, "show only featured commands")
	registrySearchCmd.Flags().String("sort", "downloads", "sort by: downloads, rating, name, updated")
	registrySearchCmd.Flags().Int("limit", 20, "max results to show")

	// Registry subcommands
	registryCmd.AddCommand(registrySearchCmd)
	registryCmd.AddCommand(registryFeaturedCmd)
	registryCmd.AddCommand(registryCategoriesCmd)

	// Update flags
	updateCmd.Flags().Bool("check", false, "check only, don't install")
	updateCmd.Flags().Bool("all", false, "update all commands")

	// Lock subcommands
	lockGenerateCmd.Flags().StringP("output", "o", "scmd.lock", "output file")
	lockCmd.AddCommand(lockGenerateCmd)
	lockCmd.AddCommand(lockInstallCmd)

	// Cache subcommands
	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cacheClearCmd)
}
