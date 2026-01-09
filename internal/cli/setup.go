package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/backend/llamacpp"
	"github.com/scmd/scmd/internal/config"
	"github.com/scmd/scmd/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBold   = "\033[1m"
)

// ModelPreset represents a model configuration preset
type ModelPreset struct {
	Name        string
	Model       string
	Size        string
	Description string
	Speed       string
	Quality     string
}

var modelPresets = []ModelPreset{
	{
		Name:        "Fast (0.5B)",
		Model:       "qwen2.5-0.5b",
		Size:        "379 MB",
		Description: "Lightning fast responses, good for simple tasks",
		Speed:       "~60 tokens/sec",
		Quality:     "Good for basic queries",
	},
	{
		Name:        "Balanced (1.5B)",
		Model:       "qwen2.5-1.5b",
		Size:        "940 MB",
		Description: "Best balance of speed and quality (recommended)",
		Speed:       "~40 tokens/sec",
		Quality:     "Great for most tasks",
	},
	{
		Name:        "Best (3B)",
		Model:       "qwen2.5-3b",
		Size:        "1.9 GB",
		Description: "Higher quality responses, slower speed",
		Speed:       "~25 tokens/sec",
		Quality:     "Excellent comprehension",
	},
	{
		Name:        "Premium (7B)",
		Model:       "qwen2.5-7b",
		Size:        "3.8 GB",
		Description: "Highest quality, requires more resources",
		Speed:       "~12 tokens/sec",
		Quality:     "Best available quality",
	},
}

// SetupCommand returns the setup command
func SetupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure scmd with an interactive wizard",
		Long:  "Set up scmd with your preferred model and configuration through an interactive wizard",
		RunE:  runSetup,
	}

	cmd.Flags().Bool("force", false, "Force setup even if already completed")
	cmd.Flags().Bool("quiet", false, "Skip the welcome message")

	return cmd
}

// runSetup runs the interactive setup wizard
func runSetup(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	quiet, _ := cmd.Flags().GetBool("quiet")

	// Check if setup was already completed
	if !force && viper.GetBool("setup_completed") {
		fmt.Println("Setup has already been completed. Use --force to run again.")
		return nil
	}

	if !quiet {
		showWelcomeBanner()
	}

	fmt.Printf("\n%s%sSetup Progress:%s\n", colorBold, colorCyan, colorReset)

	// Stage 1: Model selection
	fmt.Printf("[1/3] Selecting AI model...\n")
	selectedModel, err := selectModel()
	if err != nil {
		return fmt.Errorf("model selection failed: %w", err)
	}
	fmt.Printf("%s✓ Model selected: %s%s\n\n", colorGreen, selectedModel, colorReset)

	// Stage 2: Download model
	fmt.Printf("[2/3] Downloading model (this may take a few minutes)...\n")
	if err := ensureModel(selectedModel); err != nil {
		return fmt.Errorf("model setup failed: %w", err)
	}
	fmt.Printf("%s✓ Model ready%s\n\n", colorGreen, colorReset)

	// Stage 3: Configuration
	fmt.Printf("[3/3] Saving configuration...\n")
	if err := updateConfiguration(selectedModel); err != nil {
		return fmt.Errorf("configuration update failed: %w", err)
	}
	fmt.Printf("%s✓ Configuration saved%s\n\n", colorGreen, colorReset)

	// Show success message
	showSuccessMessage(selectedModel)

	// Offer quick test
	if !quiet {
		offerQuickTest(selectedModel)
	}

	return nil
}

// showWelcomeBanner displays the welcome message
func showWelcomeBanner() {
	fmt.Println()
	fmt.Printf("%s%s╔══════════════════════════════════════════════════════════╗%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%s%s║                                                          ║%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%s%s║                    Welcome to scmd! 🚀                  ║%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%s%s║                                                          ║%s\n", colorBold, colorCyan, colorReset)
	fmt.Printf("%s%s╚══════════════════════════════════════════════════════════╝%s\n", colorBold, colorCyan, colorReset)
	fmt.Println()

	fmt.Println("scmd brings AI-powered slash commands to your terminal.")
	fmt.Printf("%s✨ 100%% offline and private - your data never leaves your machine%s\n", colorGreen, colorReset)
	fmt.Println()
	fmt.Println("Let's get you set up with a local AI model in under a minute!")
	fmt.Println()
}

// selectModel prompts the user to select a model preset
func selectModel() (string, error) {
	fmt.Printf("%s%sChoose your AI model:%s\n", colorBold, colorCyan, colorReset)
	fmt.Println()

	for i, preset := range modelPresets {
		if i == 1 { // Balanced - recommended
			fmt.Printf("%s[%d] %s - %s%s\n", colorGreen, i+1, preset.Name, preset.Size, colorReset)
		} else {
			fmt.Printf("[%d] %s - %s\n", i+1, preset.Name, preset.Size)
		}
		fmt.Printf("    %s\n", preset.Description)
		fmt.Printf("    Speed: %s | Quality: %s\n", preset.Speed, preset.Quality)
		fmt.Println()
	}

	fmt.Printf("%sRecommendation: Option 2 (Balanced) for most users%s\n", colorGreen, colorReset)
	fmt.Println()
	fmt.Print("Enter your choice [1-4] (default 2): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		input = "2" // Default to Balanced
	}

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(modelPresets) {
		fmt.Println("Invalid choice, using default (Balanced)")
		choice = 2
	}

	return modelPresets[choice-1].Model, nil
}

// ensureModel downloads the model if not already present
func ensureModel(modelName string) error {
	dataDir := config.DataDir()
	modelsDir := filepath.Join(dataDir, "models")
	modelPath := filepath.Join(modelsDir, modelName+".gguf")

	// Check if model exists
	if _, err := os.Stat(modelPath); err == nil {
		fmt.Printf("✓ Model %s is already downloaded\n", modelName)
		return nil
	}

	// Create models directory
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return fmt.Errorf("create models directory: %w", err)
	}

	fmt.Printf("Downloading %s model...\n", modelName)

	// Get model URL
	modelURL := getModelURL(modelName)
	if modelURL == "" {
		return fmt.Errorf("unknown model: %s", modelName)
	}

	// Download with progress bar
	downloader := llamacpp.NewDownloader()

	// Use our clean progress bar
	progress := ui.NewProgressBar(0, "Downloading", os.Stdout)

	if err := downloader.DownloadWithProgress(modelURL, modelPath, func(current, total int64) {
		if progress.Total == 0 && total > 0 {
			progress.Total = total
		}
		progress.Update(current)
	}); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	progress.Finish()
	fmt.Printf("✓ Model downloaded successfully to %s\n", modelPath)

	return nil
}

// getModelURL returns the download URL for a model
func getModelURL(modelName string) string {
	// Use the existing DefaultModels from llamacpp package
	// These URLs are verified and working
	for _, model := range llamacpp.DefaultModels {
		if model.Name == modelName {
			return model.URL
		}
	}
	return ""
}

// updateConfiguration saves the selected model to config
func updateConfiguration(modelName string) error {
	// Set the model
	viper.Set("backends.default", "llamacpp")
	viper.Set("backends.local.model", modelName)
	viper.Set("backends.local.context_length", 0) // 0 = use model's native context size (no limits)
	viper.Set("setup_completed", true)
	viper.Set("models.auto_download", true)

	// Save configuration
	configPath := config.ConfigPath()
	configDir := filepath.Dir(configPath)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("save configuration: %w", err)
	}

	return nil
}

// showSuccessMessage displays the success message
func showSuccessMessage(modelName string) {
	fmt.Println()
	fmt.Printf("%s%s🎉 Setup Complete!%s\n", colorBold, colorGreen, colorReset)
	fmt.Println()

	fmt.Println("You're all set! Your AI assistant is ready to help.")
	fmt.Printf("Model: %s (100%% offline and private)\n", modelName)
	fmt.Println()
}

// offerQuickTest offers to run a quick test after setup
func offerQuickTest(modelName string) {
	fmt.Printf("%sTry these commands to get started:%s\n", colorCyan, colorReset)
	fmt.Println("  scmd /explain \"What is Docker?\"")
	fmt.Println("  echo 'print(\"Hello\")' | scmd explain")
	fmt.Println("  scmd /explain main.go")
	fmt.Println()

	fmt.Print("Would you like to run a quick test now? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		// Skip test on error
		fmt.Println("\nFor help, run: scmd --help")
		return
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input != "" && input != "y" && input != "yes" {
		fmt.Println("\nFor help, run: scmd --help")
		return
	}

	// Run a quick test
	fmt.Println()
	fmt.Printf("%sRunning quick test...%s\n", colorCyan, colorReset)
	fmt.Printf("%s> scmd /explain \"what is docker\"%s\n", colorYellow, colorReset)
	fmt.Println()

	// Actually test the backend
	if err := testBackend(modelName); err != nil {
		fmt.Printf("%s⚠️  Test failed: %v%s\n", colorYellow, err, colorReset)
		fmt.Println()
		fmt.Println("Don't worry! You can still use scmd.")
		fmt.Println("Try running: scmd backends")
		fmt.Println("To see available backends and troubleshoot.")
	} else {
		fmt.Printf("%s✓ Everything works perfectly!%s\n", colorGreen, colorReset)
	}
	fmt.Println()
	fmt.Println("For more commands and options:")
	fmt.Println("  scmd --help")
	fmt.Println()
}

// testBackend tests the configured backend with a simple query
func testBackend(modelName string) error {
	// Initialize llamacpp backend
	dataDir := config.DataDir()
	llamaBackend := llamacpp.New(dataDir)

	// Test if it's available
	ctx := context.Background()

	available, err := llamaBackend.IsAvailable(ctx)
	if err != nil {
		return fmt.Errorf("backend check failed: %w", err)
	}

	if !available {
		return fmt.Errorf("llama-server not found - please install llama.cpp: brew install llama.cpp")
	}

	// Try a simple completion
	req := &backend.CompletionRequest{
		Prompt:      "What is Docker? Answer in one sentence.",
		MaxTokens:   50,
		Temperature: 0.7,
	}

	resp, err := llamaBackend.Complete(ctx, req)
	if err != nil {
		return fmt.Errorf("completion failed: %w", err)
	}

	// Print the response
	if resp.Content != "" {
		fmt.Println(strings.TrimSpace(resp.Content))
		fmt.Println()
	}

	return nil
}

// IsFirstRun checks if this is the first run
func IsFirstRun() bool {
	// Check if setup_completed flag exists
	if viper.IsSet("setup_completed") {
		return !viper.GetBool("setup_completed")
	}

	// Check if config file exists
	configPath := config.ConfigPath()
	if _, err := os.Stat(configPath); err != nil {
		return true // Config doesn't exist, first run
	}

	// Check if models directory exists and has models
	modelsDir := filepath.Join(config.DataDir(), "models")
	if entries, err := os.ReadDir(modelsDir); err == nil && len(entries) > 0 {
		// Has models, probably not first run
		return false
	}

	return true
}

// RunSetupIfNeeded runs the setup wizard if it's the first run
func RunSetupIfNeeded() error {
	if !IsFirstRun() {
		return nil
	}

	fmt.Println()
	fmt.Printf("%s%s👋 It looks like this is your first time using scmd!%s\n", colorBold, colorYellow, colorReset)
	fmt.Println()
	fmt.Println("scmd needs to download an AI model (~940MB) for offline use.")
	fmt.Print("Would you like to run the setup wizard now? [Y/n]: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		// If they can't answer (non-interactive), skip setup
		fmt.Println("\nSkipping setup. Run 'scmd setup' later to configure.")
		return nil
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input != "" && input != "y" && input != "yes" {
		fmt.Println("You can run 'scmd setup' anytime to configure your AI model.")
		return nil
	}

	// Run the setup
	cmd := SetupCommand()
	return runSetup(cmd, []string{})
}
