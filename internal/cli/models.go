package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/scmd/scmd/internal/backend/llamacpp"
	"github.com/scmd/scmd/internal/config"
)

// modelsCmd manages local models
var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Manage local LLM models",
	Long: `Manage local LLM models for offline inference.

scmd uses llama.cpp with small, efficient models that run locally.
No API keys or internet required after initial download.`,
	Aliases: []string{"model"},
}

// modelsListCmd lists available models
var modelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available models",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := getDataDir()
		mgr := llamacpp.NewModelManager(dataDir)

		// Get downloaded models
		downloaded, _ := mgr.ListDownloaded()
		downloadedSet := make(map[string]bool)
		for _, d := range downloaded {
			downloadedSet[d] = true
		}

		fmt.Println("Available Models:")
		fmt.Println()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSIZE\tSTATUS\tDESCRIPTION")

		for _, m := range mgr.ListModels() {
			status := "not downloaded"
			filename := fmt.Sprintf("%s-%s.gguf", m.Name, m.Variant)
			if downloadedSet[filename] {
				status = "✓ ready"
			}

			size := formatSize(m.Size)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Name, size, status, m.Description)
		}
		w.Flush()

		fmt.Println()
		fmt.Println("Default model:", llamacpp.GetDefaultModel())
		fmt.Println()
		fmt.Println("Download a model: scmd models pull <name>")

		return nil
	},
}

// modelsPullCmd downloads a model
var modelsPullCmd = &cobra.Command{
	Use:   "pull <model>",
	Short: "Download a model",
	Args:  cobra.ExactArgs(1),
	Example: `  scmd models pull qwen3-4b
  scmd models pull qwen3-1.7b`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dataDir := getDataDir()
		mgr := llamacpp.NewModelManager(dataDir)

		modelName := args[0]
		fmt.Printf("Pulling model: %s\n", modelName)

		path, err := mgr.GetModelPath(ctx, modelName)
		if err != nil {
			return err
		}

		fmt.Printf("Model ready: %s\n", path)
		return nil
	},
}

// modelsRemoveCmd removes a downloaded model
var modelsRemoveCmd = &cobra.Command{
	Use:   "remove <model>",
	Short: "Remove a downloaded model",
	Aliases: []string{"rm", "delete"},
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := getDataDir()
		mgr := llamacpp.NewModelManager(dataDir)

		if err := mgr.DeleteModel(args[0]); err != nil {
			return err
		}

		fmt.Printf("Removed model: %s\n", args[0])
		return nil
	},
}

// modelsInfoCmd shows model information
var modelsInfoCmd = &cobra.Command{
	Use:   "info <model>",
	Short: "Show model information",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := getDataDir()
		mgr := llamacpp.NewModelManager(dataDir)

		modelName := args[0]

		for _, m := range mgr.ListModels() {
			if m.Name == modelName {
				fmt.Printf("Name:         %s\n", m.Name)
				fmt.Printf("Variant:      %s\n", m.Variant)
				fmt.Printf("Size:         %s\n", formatSize(m.Size))
				fmt.Printf("Context:      %d tokens\n", m.ContextSize)
				fmt.Printf("Tool Calling: %v\n", m.ToolCalling)
				fmt.Printf("Description:  %s\n", m.Description)
				fmt.Printf("URL:          %s\n", m.URL)
				return nil
			}
		}

		return fmt.Errorf("model not found: %s", modelName)
	},
}

// modelsSetDefaultCmd sets the default model
var modelsSetDefaultCmd = &cobra.Command{
	Use:   "default <model>",
	Short: "Set the default model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Update config
		cfg.Backends.Local.Model = args[0]
		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("Default model set to: %s\n", args[0])
		return nil
	},
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func init() {
	modelsCmd.AddCommand(modelsListCmd)
	modelsCmd.AddCommand(modelsPullCmd)
	modelsCmd.AddCommand(modelsRemoveCmd)
	modelsCmd.AddCommand(modelsInfoCmd)
	modelsCmd.AddCommand(modelsSetDefaultCmd)
}
