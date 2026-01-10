package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/backend/llamacpp"
	"github.com/scmd/scmd/internal/backend/ollama"
	"github.com/scmd/scmd/internal/backend/openai"
	"github.com/scmd/scmd/internal/chat"
	"github.com/scmd/scmd/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive conversation",
	Long: `Start an interactive chat session with context retention.

Messages are saved automatically and can be resumed later.
Use Ctrl+D to exit, or type /exit.`,
	Example: `  scmd chat
  scmd chat --continue abc123
  scmd chat --model qwen2.5-7b`,
	RunE: runChat,
}

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Manage conversation history",
	Long:  "View, search, and manage your conversation history.",
}

// historyListCmd lists recent conversations
var historyListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List recent conversations",
	RunE:    runHistoryList,
}

// historyShowCmd shows a specific conversation
var historyShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a conversation",
	Args:  cobra.ExactArgs(1),
	RunE:  runHistoryShow,
}

// historySearchCmd searches conversations
var historySearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search conversations",
	Args:  cobra.ExactArgs(1),
	RunE:  runHistorySearch,
}

// historyDeleteCmd deletes a conversation
var historyDeleteCmd = &cobra.Command{
	Use:     "delete <id>",
	Aliases: []string{"rm"},
	Short:   "Delete a conversation",
	Args:    cobra.ExactArgs(1),
	RunE:    runHistoryDelete,
}

// historyClearCmd clears all conversations
var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all conversations",
	Long:  "Delete all conversation history. This action cannot be undone.",
	RunE:  runHistoryClear,
}

func init() {
	// Chat command flags (avoiding -c which is used for context globally)
	chatCmd.Flags().String("continue", "", "Continue a previous conversation by ID")
	chatCmd.Flags().StringP("model", "m", "", "Model to use")
	chatCmd.Flags().String("backend", "", "Backend to use (llamacpp, ollama, openai)") // -b is already used globally

	// History list flags
	historyListCmd.Flags().IntP("limit", "n", 20, "Number of conversations to show")

	// History subcommands
	historyCmd.AddCommand(historyListCmd)
	historyCmd.AddCommand(historyShowCmd)
	historyCmd.AddCommand(historySearchCmd)
	historyCmd.AddCommand(historyDeleteCmd)
	historyCmd.AddCommand(historyClearCmd)
}

func runChat(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get flags
	continueID, _ := cmd.Flags().GetString("continue")
	modelName, _ := cmd.Flags().GetString("model")
	backendName, _ := cmd.Flags().GetString("backend")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine backend and model
	if backendName == "" {
		backendName = cfg.GetString("backends.default")
		if backendName == "" {
			backendName = "llamacpp" // Default to local
		}
	}

	if modelName == "" {
		modelName = cfg.GetString("backends.local.model")
		if modelName == "" {
			modelName = "qwen2.5-1.5b" // Default model
		}
	}

	// Initialize backend
	backendInstance, err := initializeBackend(backendName, modelName, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize backend: %w", err)
	}

	// Initialize backend
	if err := backendInstance.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize backend: %w", err)
	}
	defer backendInstance.Shutdown(ctx)

	var session *chat.Session

	if continueID != "" {
		// Resume existing conversation
		session, err = chat.LoadSession(continueID, backendInstance)
		if err != nil {
			return fmt.Errorf("failed to load conversation: %w", err)
		}
		fmt.Printf("Resumed conversation %s\n\n", continueID)
	} else {
		// Start new conversation
		chatConfig := &chat.Config{
			MaxContextMessages: viper.GetInt("chat.max_context_messages"),
			AutoSave:          viper.GetBool("chat.auto_save"),
			AutoTitle:         viper.GetBool("chat.auto_title"),
		}

		// Set defaults if not configured
		if chatConfig.MaxContextMessages == 0 {
			chatConfig.MaxContextMessages = 20
		}
		if !viper.IsSet("chat.auto_save") {
			chatConfig.AutoSave = true
		}
		if !viper.IsSet("chat.auto_title") {
			chatConfig.AutoTitle = true
		}

		session, err = chat.NewSession(modelName, backendInstance, chatConfig)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
	}

	// Run the interactive session
	return session.Run(ctx)
}

func runHistoryList(cmd *cobra.Command, args []string) error {
	limit, _ := cmd.Flags().GetInt("limit")

	store, err := chat.OpenConversationStore()
	if err != nil {
		return fmt.Errorf("failed to open conversation store: %w", err)
	}
	defer store.Close()

	conversations, err := store.List(limit)
	if err != nil {
		return fmt.Errorf("failed to list conversations: %w", err)
	}

	if len(conversations) == 0 {
		fmt.Println("No conversations found. Start one with: scmd chat")
		return nil
	}

	fmt.Println("Recent Conversations:\n")
	for i, conv := range conversations {
		// Display title or truncated first message
		title := conv.Title
		if title == "" {
			title = "(no title)"
		}

		fmt.Printf("%2d. [%s] %s\n", i+1, conv.ID[:8], title)
		fmt.Printf("    Model: %s | Messages: %d | Updated: %s\n",
			conv.Model,
			conv.MessageCount,
			conv.UpdatedAt.Format("Jan 02, 15:04"))
	}

	fmt.Printf("\nResume with: scmd chat --continue <id>\n")
	fmt.Printf("View details: scmd history show <id>\n")

	return nil
}

func runHistoryShow(cmd *cobra.Command, args []string) error {
	id := args[0]

	store, err := chat.OpenConversationStore()
	if err != nil {
		return fmt.Errorf("failed to open conversation store: %w", err)
	}
	defer store.Close()

	content, err := store.ShowConversation(id)
	if err != nil {
		return fmt.Errorf("failed to show conversation: %w", err)
	}

	fmt.Println(content)
	return nil
}

func runHistorySearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	store, err := chat.OpenConversationStore()
	if err != nil {
		return fmt.Errorf("failed to open conversation store: %w", err)
	}
	defer store.Close()

	conversations, err := store.Search(query)
	if err != nil {
		return fmt.Errorf("failed to search conversations: %w", err)
	}

	if len(conversations) == 0 {
		fmt.Printf("No conversations found matching '%s'\n", query)
		return nil
	}

	fmt.Printf("Found %d conversation(s) matching '%s':\n\n", len(conversations), query)
	for i, conv := range conversations {
		title := conv.Title
		if title == "" {
			title = "(no title)"
		}

		fmt.Printf("%2d. [%s] %s\n", i+1, conv.ID[:8], title)
		fmt.Printf("    Updated: %s\n", conv.UpdatedAt.Format("Jan 02, 15:04"))
	}

	return nil
}

func runHistoryDelete(cmd *cobra.Command, args []string) error {
	id := args[0]

	// Confirm deletion
	fmt.Printf("Delete conversation %s? (y/N): ", id)
	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) != "y" {
		fmt.Println("Cancelled.")
		return nil
	}

	store, err := chat.OpenConversationStore()
	if err != nil {
		return fmt.Errorf("failed to open conversation store: %w", err)
	}
	defer store.Close()

	if err := store.Delete(id); err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	fmt.Printf("Deleted conversation %s\n", id)
	return nil
}

func runHistoryClear(cmd *cobra.Command, args []string) error {
	// Confirm clearing all
	fmt.Print("This will delete ALL conversations. Are you sure? (yes/N): ")
	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	store, err := chat.OpenConversationStore()
	if err != nil {
		return fmt.Errorf("failed to open conversation store: %w", err)
	}
	defer store.Close()

	if err := store.ClearAll(); err != nil {
		return fmt.Errorf("failed to clear conversations: %w", err)
	}

	fmt.Println("All conversations have been deleted.")
	return nil
}

// initializeBackend creates the appropriate backend instance
func initializeBackend(backendName, modelName string, cfg *config.Config) (backend.Backend, error) {
	dataDir := config.GetDataDir()

	// Create backend based on type
	switch backendName {
	case "llamacpp", "local":
		// Use llamacpp backend
		llamaBackend := llamacpp.New(dataDir)
		llamaBackend.SetModel(modelName)

		// Apply context size from config
		if contextSize := cfg.GetInt("backends.local.context_length"); contextSize > 0 {
			llamaBackend.SetContextSize(contextSize)
		}

		return llamaBackend, nil

	case "ollama":
		host := viper.GetString("backends.ollama.host")
		if host == "" {
			host = "http://localhost:11434"
		}
		ollamaConfig := &ollama.Config{
			BaseURL: host,
			Model:   modelName,
		}
		return ollama.New(ollamaConfig), nil

	case "openai":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
		}
		openaiConfig := &openai.Config{
			APIKey:  apiKey,
			Model:   modelName,
			BaseURL: "https://api.openai.com/v1",
		}
		return openai.New(openaiConfig), nil

	default:
		return nil, fmt.Errorf("unknown backend: %s", backendName)
	}
}