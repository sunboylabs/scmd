package chat

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/scmd/scmd/internal/backend"
)

// Message represents a single message in the conversation
type Message struct {
	Role      string    `json:"role"`      // "user" or "assistant"
	Content   string    `json:"content"`
	Tokens    int       `json:"tokens"`
	CreatedAt time.Time `json:"created_at"`
}

// Session represents an interactive chat session
type Session struct {
	conversationID string
	messages       []Message
	model          string
	backend        backend.Backend
	db             *ConversationStore
	config         *Config
}

// Config holds chat configuration
type Config struct {
	MaxContextMessages int
	AutoSave          bool
	AutoTitle         bool
}

// NewSession creates a new chat session
func NewSession(model string, backend backend.Backend, config *Config) (*Session, error) {
	db, err := OpenConversationStore()
	if err != nil {
		return nil, err
	}

	convID := uuid.New().String()

	if config == nil {
		config = &Config{
			MaxContextMessages: 20,
			AutoSave:          true,
			AutoTitle:         true,
		}
	}

	return &Session{
		conversationID: convID,
		messages:       []Message{},
		model:          model,
		backend:        backend,
		db:             db,
		config:         config,
	}, nil
}

// LoadSession loads an existing conversation session
func LoadSession(conversationID string, backend backend.Backend) (*Session, error) {
	db, err := OpenConversationStore()
	if err != nil {
		return nil, err
	}

	session, err := db.LoadSession(conversationID, backend)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	return session, nil
}

// Run starts the interactive REPL loop
func (s *Session) Run(ctx context.Context) error {
	reader := bufio.NewReader(os.Stdin)

	// Welcome message
	fmt.Println("scmd interactive mode")
	fmt.Printf("Conversation: %s\n", s.conversationID[:8])
	fmt.Printf("Model: %s\n", s.model)
	fmt.Println("\nType your message and press Enter. Use /help for commands, Ctrl+D to exit.\n")

	// If resuming, show message count
	if len(s.messages) > 0 {
		fmt.Printf("Resumed with %d previous messages\n\n", len(s.messages))
	}

	for {
		fmt.Print("You: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			// EOF (Ctrl+D)
			fmt.Printf("\nConversation saved. Use 'scmd chat --continue %s' to resume.\n",
				s.conversationID[:8])
			break
		}

		input = strings.TrimSpace(input)

		// Handle special commands
		if strings.HasPrefix(input, "/") {
			if err := s.handleCommand(input); err != nil {
				if err.Error() == "exit requested" {
					fmt.Printf("\nConversation saved. Use 'scmd chat --continue %s' to resume.\n",
						s.conversationID[:8])
					break
				}
				fmt.Printf("Error: %v\n", err)
			}
			continue
		}

		if input == "" {
			continue
		}

		// Add user message
		s.addMessage("user", input)

		// Show thinking indicator
		fmt.Print("\nAssistant: ")

		// Generate response with full context
		response, tokens, err := s.generateResponse(ctx)
		if err != nil {
			fmt.Printf("\nError: %v\n\n", err)
			continue
		}

		// Add assistant message
		s.addMessageWithTokens("assistant", response, tokens)

		// Display response
		fmt.Printf("%s\n\n", response)

		// Auto-save after each exchange
		if s.config.AutoSave {
			if err := s.db.SaveConversation(s); err != nil {
				fmt.Printf("Warning: Failed to save conversation: %v\n", err)
			}
		}
	}

	return nil
}

func (s *Session) handleCommand(cmd string) error {
	switch {
	case cmd == "/help":
		s.showHelp()

	case cmd == "/clear":
		s.messages = []Message{}
		fmt.Println("Context cleared (conversation history preserved)")

	case cmd == "/info":
		s.showInfo()

	case cmd == "/export":
		return s.exportToMarkdown()

	case strings.HasPrefix(cmd, "/model "):
		newModel := strings.TrimPrefix(cmd, "/model ")
		s.model = newModel
		fmt.Printf("Switched to model: %s\n", newModel)

	case cmd == "/save":
		if err := s.db.SaveConversation(s); err != nil {
			return err
		}
		fmt.Println("Conversation saved")

	case cmd == "/exit":
		return fmt.Errorf("exit requested")

	default:
		fmt.Printf("Unknown command: %s (try /help)\n", cmd)
	}

	return nil
}

func (s *Session) showHelp() {
	fmt.Println(`
Available Commands:
  /help      Show this help message
  /clear     Clear current context (keeps history)
  /info      Show conversation info
  /save      Force save conversation
  /export    Export conversation to markdown
  /model X   Switch to model X
  /exit      Exit session (or press Ctrl+D)
`)
}

func (s *Session) showInfo() {
	createdAt := s.db.GetCreatedAt(s.conversationID)
	fmt.Printf(`
Conversation Info:
  ID:           %s
  Model:        %s
  Backend:      %s
  Messages:     %d
  Started:      %s
`, s.conversationID[:8], s.model, s.backend.Name(), len(s.messages),
		createdAt.Format("2006-01-02 15:04"))
}

func (s *Session) addMessage(role, content string) {
	s.addMessageWithTokens(role, content, 0)
}

func (s *Session) addMessageWithTokens(role, content string, tokens int) {
	msg := Message{
		Role:      role,
		Content:   content,
		Tokens:    tokens,
		CreatedAt: time.Now(),
	}
	s.messages = append(s.messages, msg)
}

func (s *Session) generateResponse(ctx context.Context) (string, int, error) {
	// Build prompt from message history
	prompt := s.buildPrompt()

	// Call LLM backend
	req := &backend.CompletionRequest{
		Prompt:      prompt,
		MaxTokens:   4096,
		Temperature: 0.7,
	}

	resp, err := s.backend.Complete(ctx, req)
	if err != nil {
		return "", 0, err
	}

	return resp.Content, resp.TokensUsed, nil
}

func (s *Session) buildPrompt() string {
	// Build context from message history
	var sb strings.Builder

	messages := s.getContextMessages()
	for _, msg := range messages {
		role := "User"
		if msg.Role == "assistant" {
			role = "Assistant"
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n\n", role, msg.Content))
	}

	// Add instruction for the next response
	sb.WriteString("Assistant: ")

	return sb.String()
}

func (s *Session) getContextMessages() []Message {
	// Return last N messages to avoid context overflow
	maxMessages := s.config.MaxContextMessages
	if maxMessages <= 0 {
		maxMessages = 20 // Default
	}

	if len(s.messages) <= maxMessages {
		return s.messages
	}

	// Keep first message (often contains important context)
	// and last N-1 messages
	result := []Message{s.messages[0]}
	result = append(result, s.messages[len(s.messages)-(maxMessages-1):]...)
	return result
}

func (s *Session) exportToMarkdown() error {
	filename := fmt.Sprintf("conversation_%s.md", s.conversationID[:8])

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header
	fmt.Fprintf(f, "# Conversation %s\n\n", s.conversationID[:8])
	fmt.Fprintf(f, "**Model:** %s  \n", s.model)
	fmt.Fprintf(f, "**Messages:** %d  \n\n", len(s.messages))
	fmt.Fprintf(f, "---\n\n")

	// Write messages
	for _, msg := range s.messages {
		role := "**You:**"
		if msg.Role == "assistant" {
			role = "**Assistant:**"
		}
		fmt.Fprintf(f, "%s\n\n%s\n\n---\n\n", role, msg.Content)
	}

	fmt.Printf("Exported to %s\n", filename)
	return nil
}

// GetConversationID returns the conversation ID
func (s *Session) GetConversationID() string {
	return s.conversationID
}

// GetMessages returns all messages in the session
func (s *Session) GetMessages() []Message {
	return s.messages
}

// GetModel returns the model name
func (s *Session) GetModel() string {
	return s.model
}

// GetBackendName returns the backend name
func (s *Session) GetBackendName() string {
	return s.backend.Name()
}