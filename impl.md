# scmd Implementation Guide: 3 Must-Borrow Features


**Version:** 1.0
**Date:** 2026-01-10
**Target Release:** Public Launch
**Estimated Timeline:** 4 weeks


---


## Executive Summary


This guide details the implementation of 3 critical features that will bring scmd to feature parity with competitors (sgpt, llm, mods) while maintaining its unique advantages (local-first, command generation, developer focus).


### The 3 Features:


1. **Interactive Conversation Mode** - Multi-turn conversations with context retention
2. **Beautiful Markdown Output** - Syntax highlighting and professional formatting
3. **Template/Pattern System** - User-customizable prompts for workflows


### Why These 3?


- **Conversation Mode**: Table stakes feature - every competitor has this
- **Beautiful Output**: First impression matters - makes scmd feel professional
- **Template System**: Extensibility without complexity - differentiates from basic tools


### Success Metrics:


✅ Users can have 5+ message conversations
✅ Code blocks have syntax highlighting
✅ Users can create and share custom templates
✅ scmd is ready for public launch


---


## Feature 1: Interactive Conversation Mode 🗣️


**Borrowed from:** sgpt (--repl), llm (conversations), mods (threads)


### Why This is Critical


- **Every competitor has this** - it's table stakes
- Users expect to ask follow-up questions
- Biggest gap in current scmd experience
- Transforms from "one-shot tool" to "AI pair"
- Increases user engagement and retention


### What to Build


```bash
# New commands
scmd chat                           # Start interactive session
scmd chat --continue <id>           # Resume previous conversation
scmd chat --model qwen2.5-7b        # Use specific model
scmd history                        # Show past conversations
scmd history show <id>              # Show specific conversation
scmd history search "docker"        # Search conversations
scmd history delete <id>            # Delete conversation
scmd history clear                  # Clear all history
```


### Database Schema (SQLite)


Create file: `~/.scmd/conversations.db`


```sql
-- Conversations table
CREATE TABLE conversations (
   id TEXT PRIMARY KEY,           -- UUID
   created_at TIMESTAMP,
   updated_at TIMESTAMP,
   title TEXT,                    -- Auto-generated from first message
   model TEXT,                    -- Model used (e.g., "qwen2.5-1.5b")
   backend TEXT,                  -- Backend used (llamacpp, ollama, etc.)
   metadata JSON                  -- Additional context (tags, folder, etc.)
);


-- Messages table
CREATE TABLE messages (
   id INTEGER PRIMARY KEY AUTOINCREMENT,
   conversation_id TEXT,
   role TEXT,                     -- 'user' or 'assistant'
   content TEXT,
   tokens INTEGER,                -- Token count (for analytics)
   created_at TIMESTAMP,
   FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);


-- Indexes for performance
CREATE INDEX idx_conv_updated ON conversations(updated_at DESC);
CREATE INDEX idx_msg_conv ON messages(conversation_id);


-- Full-text search
CREATE VIRTUAL TABLE messages_fts USING fts5(
   conversation_id UNINDEXED,
   content,
   content=messages,
   content_rowid=id
);


-- Triggers to keep FTS in sync
CREATE TRIGGER messages_ai AFTER INSERT ON messages BEGIN
   INSERT INTO messages_fts(rowid, conversation_id, content)
   VALUES (new.id, new.conversation_id, new.content);
END;


CREATE TRIGGER messages_ad AFTER DELETE ON messages BEGIN
   DELETE FROM messages_fts WHERE rowid = old.id;
END;


CREATE TRIGGER messages_au AFTER UPDATE ON messages BEGIN
   UPDATE messages_fts SET content = new.content WHERE rowid = old.id;
END;
```


### Go Implementation


#### File: `pkg/chat/session.go`


```go
package chat


import (
   "bufio"
   "fmt"
   "os"
   "strings"
   "time"


   "github.com/google/uuid"
)


type Message struct {
   Role      string    `json:"role"`      // "user" or "assistant"
   Content   string    `json:"content"`
   Tokens    int       `json:"tokens"`
   CreatedAt time.Time `json:"created_at"`
}


type Session struct {
   conversationID string
   messages       []Message
   model          string
   backend        Backend
   db             *ConversationStore
   config         *Config
}


func NewSession(model, backend string, config *Config) (*Session, error) {
   db, err := OpenConversationStore()
   if err != nil {
       return nil, err
   }


   convID := uuid.New().String()


   return &Session{
       conversationID: convID,
       messages:       []Message{},
       model:          model,
       backend:        backend,
       db:             db,
       config:         config,
   }, nil
}


func LoadSession(conversationID string) (*Session, error) {
   db, err := OpenConversationStore()
   if err != nil {
       return nil, err
   }


   session, err := db.LoadSession(conversationID)
   if err != nil {
       return nil, fmt.Errorf("conversation not found: %w", err)
   }


   return session, nil
}


func (s *Session) Run() error {
   reader := bufio.NewReader(os.Stdin)


   // Welcome message
   fmt.Println("🤖 scmd interactive mode")
   fmt.Printf("💬 Conversation: %s\n", s.conversationID[:8])
   fmt.Printf("🔧 Model: %s\n", s.model)
   fmt.Println("\nType your message and press Enter. Use /help for commands, Ctrl+D to exit.\n")


   // If resuming, show message count
   if len(s.messages) > 0 {
       fmt.Printf("📜 Resumed with %d previous messages\n\n", len(s.messages))
   }


   for {
       fmt.Print("You: ")
       input, err := reader.ReadString('\n')
       if err != nil {
           // EOF (Ctrl+D)
           fmt.Println("\n👋 Conversation saved. Use 'scmd chat --continue %s' to resume.",
               s.conversationID[:8])
           break
       }


       input = strings.TrimSpace(input)


       // Handle special commands
       if strings.HasPrefix(input, "/") {
           if err := s.handleCommand(input); err != nil {
               fmt.Printf("❌ Error: %v\n", err)
           }
           continue
       }


       if input == "" {
           continue
       }


       // Add user message
       s.addMessage("user", input)


       // Show thinking indicator
       fmt.Print("\n🤖 Assistant: ")


       // Generate response with full context
       response, tokens, err := s.generateResponse()
       if err != nil {
           fmt.Printf("\n❌ Error: %v\n\n", err)
           continue
       }


       // Add assistant message
       s.addMessageWithTokens("assistant", response, tokens)


       // Display response (already printed via streaming)
       fmt.Println("\n")


       // Auto-save after each exchange
       if err := s.db.SaveConversation(s); err != nil {
           fmt.Printf("⚠️  Warning: Failed to save conversation: %v\n", err)
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
       fmt.Println("✅ Context cleared (conversation history preserved)")


   case cmd == "/info":
       s.showInfo()


   case cmd == "/export":
       return s.exportToMarkdown()


   case strings.HasPrefix(cmd, "/model "):
       newModel := strings.TrimPrefix(cmd, "/model ")
       s.model = newModel
       fmt.Printf("✅ Switched to model: %s\n", newModel)


   case cmd == "/save":
       if err := s.db.SaveConversation(s); err != nil {
           return err
       }
       fmt.Println("✅ Conversation saved")


   case cmd == "/exit":
       return fmt.Errorf("exit requested")


   default:
       fmt.Printf("❌ Unknown command: %s (try /help)\n", cmd)
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
   fmt.Printf(`
Conversation Info:
 ID:           %s
 Model:        %s
 Backend:      %s
 Messages:     %d
 Started:      %s
`, s.conversationID[:8], s.model, s.backend, len(s.messages),
       s.db.GetCreatedAt(s.conversationID).Format("2006-01-02 15:04"))
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


func (s *Session) generateResponse() (string, int, error) {
   // Build context from message history
   context := s.buildContext()


   // Call LLM backend with full context
   response, tokens, err := s.backend.GenerateWithContext(context, s.model)
   if err != nil {
       return "", 0, err
   }


   return response, tokens, nil
}


func (s *Session) buildContext() []Message {
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


   fmt.Printf("✅ Exported to %s\n", filename)
   return nil
}


func generateTitle(firstMessage string) string {
   // Extract first 50 characters for title
   if len(firstMessage) > 50 {
       return firstMessage[:47] + "..."
   }
   return firstMessage
}
```


#### File: `pkg/chat/store.go`


```go
package chat


import (
   "database/sql"
   "encoding/json"
   "fmt"
   "path/filepath"
   "time"


   _ "github.com/mattn/go-sqlite3"
)


type ConversationStore struct {
   db *sql.DB
}


type ConversationSummary struct {
   ID           string
   Title        string
   Model        string
   CreatedAt    time.Time
   UpdatedAt    time.Time
   MessageCount int
}


func OpenConversationStore() (*ConversationStore, error) {
   dbPath := filepath.Join(getConfigDir(), "conversations.db")


   db, err := sql.Open("sqlite3", dbPath)
   if err != nil {
       return nil, err
   }


   // Initialize schema
   if err := initSchema(db); err != nil {
       return nil, err
   }


   return &ConversationStore{db: db}, nil
}


func initSchema(db *sql.DB) error {
   schema := `
   CREATE TABLE IF NOT EXISTS conversations (
       id TEXT PRIMARY KEY,
       created_at TIMESTAMP,
       updated_at TIMESTAMP,
       title TEXT,
       model TEXT,
       backend TEXT,
       metadata JSON
   );


   CREATE TABLE IF NOT EXISTS messages (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       conversation_id TEXT,
       role TEXT,
       content TEXT,
       tokens INTEGER,
       created_at TIMESTAMP,
       FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
   );


   CREATE INDEX IF NOT EXISTS idx_conv_updated ON conversations(updated_at DESC);
   CREATE INDEX IF NOT EXISTS idx_msg_conv ON messages(conversation_id);
   `


   _, err := db.Exec(schema)
   return err
}


func (cs *ConversationStore) SaveConversation(session *Session) error {
   tx, err := cs.db.Begin()
   if err != nil {
       return err
   }
   defer tx.Rollback()


   now := time.Now()


   // Generate title from first message if not set
   title := ""
   if len(session.messages) > 0 {
       title = generateTitle(session.messages[0].Content)
   }


   // Insert or update conversation
   _, err = tx.Exec(`
       INSERT INTO conversations (id, created_at, updated_at, title, model, backend, metadata)
       VALUES (?, ?, ?, ?, ?, ?, ?)
       ON CONFLICT(id) DO UPDATE SET
           updated_at = ?,
           title = COALESCE(title, ?),
           model = ?,
           backend = ?
   `, session.conversationID, now, now, title, session.model, session.backend, "{}",
      now, title, session.model, session.backend)


   if err != nil {
       return err
   }


   // Delete old messages and insert new ones
   _, err = tx.Exec("DELETE FROM messages WHERE conversation_id = ?", session.conversationID)
   if err != nil {
       return err
   }


   for _, msg := range session.messages {
       _, err = tx.Exec(`
           INSERT INTO messages (conversation_id, role, content, tokens, created_at)
           VALUES (?, ?, ?, ?, ?)
       `, session.conversationID, msg.Role, msg.Content, msg.Tokens, msg.CreatedAt)


       if err != nil {
           return err
       }
   }


   return tx.Commit()
}


func (cs *ConversationStore) LoadSession(id string) (*Session, error) {
   // Load conversation metadata
   var convID, title, model, backend string
   var createdAt, updatedAt time.Time


   err := cs.db.QueryRow(`
       SELECT id, title, model, backend, created_at, updated_at
       FROM conversations WHERE id LIKE ?
   `, id+"%").Scan(&convID, &title, &model, &backend, &createdAt, &updatedAt)


   if err != nil {
       return nil, err
   }


   // Load messages
   rows, err := cs.db.Query(`
       SELECT role, content, tokens, created_at
       FROM messages
       WHERE conversation_id = ?
       ORDER BY created_at ASC
   `, convID)


   if err != nil {
       return nil, err
   }
   defer rows.Close()


   var messages []Message
   for rows.Next() {
       var msg Message
       if err := rows.Scan(&msg.Role, &msg.Content, &msg.Tokens, &msg.CreatedAt); err != nil {
           return nil, err
       }
       messages = append(messages, msg)
   }


   session := &Session{
       conversationID: convID,
       messages:       messages,
       model:          model,
       backend:        backend,
       db:             cs,
   }


   return session, nil
}


func (cs *ConversationStore) List(limit int) ([]ConversationSummary, error) {
   query := `
       SELECT c.id, c.title, c.model, c.created_at, c.updated_at,
              COUNT(m.id) as message_count
       FROM conversations c
       LEFT JOIN messages m ON c.id = m.conversation_id
       GROUP BY c.id
       ORDER BY c.updated_at DESC
       LIMIT ?
   `


   rows, err := cs.db.Query(query, limit)
   if err != nil {
       return nil, err
   }
   defer rows.Close()


   var conversations []ConversationSummary
   for rows.Next() {
       var conv ConversationSummary
       if err := rows.Scan(&conv.ID, &conv.Title, &conv.Model,
           &conv.CreatedAt, &conv.UpdatedAt, &conv.MessageCount); err != nil {
           return nil, err
       }
       conversations = append(conversations, conv)
   }


   return conversations, nil
}


func (cs *ConversationStore) Search(query string) ([]ConversationSummary, error) {
   // Use FTS if available, otherwise basic LIKE search
   sqlQuery := `
       SELECT DISTINCT c.id, c.title, c.model, c.created_at, c.updated_at,
              COUNT(m.id) as message_count
       FROM conversations c
       JOIN messages m ON c.id = m.conversation_id
       WHERE m.content LIKE ?
       GROUP BY c.id
       ORDER BY c.updated_at DESC
       LIMIT 20
   `


   rows, err := cs.db.Query(sqlQuery, "%"+query+"%")
   if err != nil {
       return nil, err
   }
   defer rows.Close()


   var conversations []ConversationSummary
   for rows.Next() {
       var conv ConversationSummary
       if err := rows.Scan(&conv.ID, &conv.Title, &conv.Model,
           &conv.CreatedAt, &conv.UpdatedAt, &conv.MessageCount); err != nil {
           return nil, err
       }
       conversations = append(conversations, conv)
   }


   return conversations, nil
}


func (cs *ConversationStore) Delete(id string) error {
   _, err := cs.db.Exec("DELETE FROM conversations WHERE id LIKE ?", id+"%")
   return err
}


func (cs *ConversationStore) GetCreatedAt(id string) time.Time {
   var createdAt time.Time
   cs.db.QueryRow("SELECT created_at FROM conversations WHERE id = ?", id).Scan(&createdAt)
   return createdAt
}


func (cs *ConversationStore) Close() error {
   return cs.db.Close()
}


func getConfigDir() string {
   home, _ := os.UserHomeDir()
   return filepath.Join(home, ".scmd")
}
```


#### File: `cmd/chat.go`


```go
package cmd


import (
   "fmt"


   "github.com/spf13/cobra"
   "your-project/pkg/chat"
)


var chatCmd = &cobra.Command{
   Use:   "chat",
   Short: "Start an interactive conversation",
   Long: `Start an interactive chat session with context retention.


Messages are saved automatically and can be resumed later.
Use Ctrl+D to exit, or type /exit.`,
   Example: `  scmd chat
 scmd chat --continue abc123
 scmd chat --model qwen2.5-7b`,
   RunE: func(cmd *cobra.Command, args []string) error {
       continueID, _ := cmd.Flags().GetString("continue")
       model, _ := cmd.Flags().GetString("model")
       backend, _ := cmd.Flags().GetString("backend")


       var session *chat.Session
       var err error


       if continueID != "" {
           // Resume existing conversation
           session, err = chat.LoadSession(continueID)
           if err != nil {
               return fmt.Errorf("failed to load conversation: %w", err)
           }
       } else {
           // Start new conversation
           config := chat.LoadConfig()
           session, err = chat.NewSession(model, backend, config)
           if err != nil {
               return fmt.Errorf("failed to create session: %w", err)
           }
       }


       return session.Run()
   },
}


var historyCmd = &cobra.Command{
   Use:   "history",
   Short: "Manage conversation history",
   Long:  "View, search, and manage your conversation history.",
}


var historyListCmd = &cobra.Command{
   Use:   "list",
   Short: "List recent conversations",
   RunE: func(cmd *cobra.Command, args []string) error {
       limit, _ := cmd.Flags().GetInt("limit")


       store, err := chat.OpenConversationStore()
       if err != nil {
           return err
       }
       defer store.Close()


       conversations, err := store.List(limit)
       if err != nil {
           return err
       }


       if len(conversations) == 0 {
           fmt.Println("No conversations found. Start one with: scmd chat")
           return nil
       }


       fmt.Println("Recent Conversations:\n")
       for i, conv := range conversations {
           fmt.Printf("%2d. %s  %s\n", i+1, conv.ID[:8], conv.Title)
           fmt.Printf("    Model: %s | Messages: %d | Updated: %s\n\n",
               conv.Model,
               conv.MessageCount,
               conv.UpdatedAt.Format("Jan 02, 15:04"))
       }


       fmt.Printf("Resume with: scmd chat --continue <id>\n")


       return nil
   },
}


var historySearchCmd = &cobra.Command{
   Use:   "search <query>",
   Short: "Search conversations",
   Args:  cobra.ExactArgs(1),
   RunE: func(cmd *cobra.Command, args []string) error {
       query := args[0]


       store, err := chat.OpenConversationStore()
       if err != nil {
           return err
       }
       defer store.Close()


       conversations, err := store.Search(query)
       if err != nil {
           return err
       }


       if len(conversations) == 0 {
           fmt.Printf("No conversations found matching '%s'\n", query)
           return nil
       }


       fmt.Printf("Found %d conversation(s) matching '%s':\n\n", len(conversations), query)
       for i, conv := range conversations {
           fmt.Printf("%2d. %s  %s\n", i+1, conv.ID[:8], conv.Title)
           fmt.Printf("    Updated: %s\n\n", conv.UpdatedAt.Format("Jan 02, 15:04"))
       }


       return nil
   },
}


var historyDeleteCmd = &cobra.Command{
   Use:   "delete <id>",
   Short: "Delete a conversation",
   Args:  cobra.ExactArgs(1),
   RunE: func(cmd *cobra.Command, args []string) error {
       id := args[0]


       store, err := chat.OpenConversationStore()
       if err != nil {
           return err
       }
       defer store.Close()


       if err := store.Delete(id); err != nil {
           return err
       }


       fmt.Printf("✅ Deleted conversation %s\n", id)
       return nil
   },
}


func init() {
   rootCmd.AddCommand(chatCmd)
   rootCmd.AddCommand(historyCmd)


   chatCmd.Flags().String("continue", "", "Continue a previous conversation by ID")
   chatCmd.Flags().StringP("model", "m", "", "Model to use")
   chatCmd.Flags().StringP("backend", "b", "", "Backend to use")


   historyCmd.AddCommand(historyListCmd)
   historyCmd.AddCommand(historySearchCmd)
   historyCmd.AddCommand(historyDeleteCmd)


   historyListCmd.Flags().IntP("limit", "n", 20, "Number of conversations to show")
}
```


### Configuration


Add to `~/.scmd/config.yaml`:


```yaml
chat:
 max_context_messages: 20    # Max messages to keep in context
 auto_save: true            # Auto-save after each message
 auto_title: true           # Auto-generate titles from first message
```


---


## Feature 2: Beautiful Markdown Output with Syntax Highlighting 🎨


**Borrowed from:** mods (Bubble Tea + Glamour), glow


### Why This is Critical


- Makes scmd feel **professional and polished**
- Code snippets need syntax highlighting
- Markdown is what LLMs output naturally
- First impression matters for public release
- Increases readability and user satisfaction


### What to Build


Transform plain text output into beautifully formatted, syntax-highlighted markdown with proper styling.


**Before:**
```
The code defines a function...
```javascript
function add(a, b) { return a + b; }
```
```


**After:** (with colors, formatting, borders, syntax highlighting)


### Dependencies


Add to `go.mod`:


```go
require (
   github.com/charmbracelet/glamour v0.6.0
   github.com/charmbracelet/lipgloss v0.9.1
   github.com/alecthomas/chroma/v2 v2.12.0
   github.com/muesli/termenv v0.15.2
)
```


### Go Implementation


#### File: `pkg/output/formatter.go`


```go
package output


import (
   "bytes"
   "fmt"
   "io"
   "os"
   "strings"


   "github.com/charmbracelet/glamour"
   "github.com/charmbracelet/lipgloss"
   "github.com/muesli/termenv"
)


type OutputFormatter struct {
   style        string
   renderer     *glamour.TermRenderer
   colorize     bool
   termProfile  termenv.Profile
}


func NewFormatter(style string, colorize bool) (*OutputFormatter, error) {
   // Auto-detect terminal capabilities
   termProfile := termenv.ColorProfile()


   if style == "auto" {
       style = detectStyle()
   }


   // Configure glamour renderer
   opts := []glamour.TermRendererOption{
       glamour.WithWordWrap(100),
       glamour.WithPreservedNewLines(),
   }


   switch style {
   case "dark":
       opts = append(opts, glamour.WithStylePath("dark"))
   case "light":
       opts = append(opts, glamour.WithStylePath("light"))
   case "notty":
       opts = append(opts, glamour.WithStylePath("notty"))
   default:
       opts = append(opts, glamour.WithStylePath("dark"))
   }


   renderer, err := glamour.NewTermRenderer(opts...)
   if err != nil {
       return nil, err
   }


   return &OutputFormatter{
       style:       style,
       renderer:    renderer,
       colorize:    colorize,
       termProfile: termProfile,
   }, nil
}


func (f *OutputFormatter) Render(markdown string) (string, error) {
   if !f.colorize || f.style == "notty" {
       return markdown, nil // Plain text mode
   }


   rendered, err := f.renderer.Render(markdown)
   if err != nil {
       // Fallback to plain text on error
       return markdown, nil
   }


   return rendered, nil
}


func (f *OutputFormatter) RenderToWriter(markdown string, w io.Writer) error {
   rendered, err := f.Render(markdown)
   if err != nil {
       return err
   }


   _, err = w.Write([]byte(rendered))
   return err
}


// Stream rendering for real-time output
func (f *OutputFormatter) StreamRender(chunks <-chan string, w io.Writer) error {
   var buffer strings.Builder


   for chunk := range chunks {
       buffer.WriteString(chunk)


       // For streaming, just write chunks as-is
       // Final render happens at the end
       w.Write([]byte(chunk))
   }


   return nil
}


func detectStyle() string {
   // Try to detect if terminal has dark or light background
   bg := termenv.BackgroundColor()


   // Check if background is dark or light
   if bg != nil {
       // Simple heuristic: if sum of RGB > 384 (128*3), it's light
       r, g, b, _ := bg.RGBA()
       sum := (r + g + b) / 256 // Normalize to 0-255 range


       if sum > 384 {
           return "light"
       }
   }


   return "dark" // Default to dark
}


// Custom styles using lipgloss
var (
   CodeBlockStyle = lipgloss.NewStyle().
       Border(lipgloss.RoundedBorder()).
       BorderForeground(lipgloss.Color("63")).
       Padding(1, 2).
       MarginTop(1).
       MarginBottom(1)


   HeadingStyle = lipgloss.NewStyle().
       Bold(true).
       Foreground(lipgloss.Color("86")).
       MarginTop(1).
       MarginBottom(1)


   ErrorStyle = lipgloss.NewStyle().
       Bold(true).
       Foreground(lipgloss.Color("196")).
       Background(lipgloss.Color("52"))


   SuccessStyle = lipgloss.NewStyle().
       Bold(true).
       Foreground(lipgloss.Color("82"))
)


func RenderError(message string) string {
   return ErrorStyle.Render("❌ " + message)
}


func RenderSuccess(message string) string {
   return SuccessStyle.Render("✅ " + message)
}
```


#### File: `pkg/output/syntax.go`


```go
package output


import (
   "bytes"
   "strings"


   "github.com/alecthomas/chroma/v2/formatters"
   "github.com/alecthomas/chroma/v2/lexers"
   "github.com/alecthomas/chroma/v2/styles"
)


// HighlightCode applies syntax highlighting to code
func HighlightCode(code, language string) (string, error) {
   // Get lexer for the language
   lexer := lexers.Get(language)
   if lexer == nil {
       lexer = lexers.Fallback
   }
   lexer = chroma.Coalesce(lexer)


   // Get style
   style := styles.Get("monokai")
   if style == nil {
       style = styles.Fallback
   }


   // Get formatter for terminal
   formatter := formatters.Get("terminal256")
   if formatter == nil {
       formatter = formatters.Fallback
   }


   // Tokenize
   iterator, err := lexer.Tokenise(nil, code)
   if err != nil {
       return code, err
   }


   // Format
   var buf bytes.Buffer
   err = formatter.Format(&buf, style, iterator)
   if err != nil {
       return code, err
   }


   return buf.String(), nil
}


// DetectLanguage tries to detect the programming language
func DetectLanguage(filename string) string {
   if filename == "" {
       return ""
   }


   lexer := lexers.Match(filename)
   if lexer != nil {
       return lexer.Config().Name
   }


   return ""
}
```


#### File: `pkg/output/spinner.go`


```go
package output


import (
   "time"


   "github.com/briandowns/spinner"
)


// ShowProgress displays a spinner with a message
func ShowProgress(message string) *spinner.Spinner {
   s := spinner.New(
       spinner.CharSets[14], // "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
       100*time.Millisecond,
   )
   s.Suffix = "  " + message
   s.Color("cyan")
   s.Start()
   return s
}


// Example usage:
// progress := ShowProgress("Analyzing code...")
// response := generateResponse()
// progress.Stop()
```


### Integration with Commands


Update existing commands to use the formatter:


```go
// File: cmd/explain.go


func runExplain(input string) error {
   // Show progress
   progress := output.ShowProgress("Thinking...")


   // Get response from LLM
   response, err := generateExplanation(input)
   progress.Stop()


   if err != nil {
       fmt.Println(output.RenderError(err.Error()))
       return err
   }


   // Format output based on flags
   outputFormat := viper.GetString("format")


   switch outputFormat {
   case "json":
       return outputJSON(response)


   case "markdown":
       fmt.Println(response) // Raw markdown
       return nil


   default: // "text" - render beautifully
       formatter, err := output.NewFormatter(
           viper.GetString("ui.style"),
           viper.GetBool("ui.colors"),
       )
       if err != nil {
           return err
       }


       return formatter.RenderToWriter(response, os.Stdout)
   }
}
```


### Configuration


Add to `~/.scmd/config.yaml`:


```yaml
ui:
 colors: true                # Enable colored output
 style: auto                 # auto, dark, light, notty
 streaming: true             # Enable streaming output
 pager: false               # Use pager for long output (less)


 markdown:
   render: true              # Render markdown (false = plain text)
   code_theme: monokai       # Syntax highlighting theme
   wrap_width: 100           # Character wrap width
   show_line_numbers: false  # Show line numbers in code blocks
```


### Custom Themes


Create file: `~/.scmd/themes/custom.json`


```json
{
 "document": {
   "color": "#f8f8f2"
 },
 "code_block": {
   "background_color": "#282a36",
   "border_style": "rounded",
   "chroma": {
     "name": "dracula"
   }
 },
 "heading": {
   "color": "#50fa7b",
   "bold": true
 },
 "link": {
   "color": "#8be9fd",
   "underline": true
 }
}
```


---


## Feature 3: Template/Pattern System 📋


**Borrowed from:** fabric (patterns), llm (templates)


### Why This is Critical


- **Extensibility without plugin complexity**
- Users can customize prompts for their workflows
- Teams can share templates
- Differentiates scmd from basic tools
- Easy to implement, high impact


### What to Build


```bash
# Create custom templates
scmd template create security-review \
 --prompt "Review this code for security issues..."


# List templates
scmd template list


# Use template
cat auth.js | scmd review --template security-review


# Share templates
scmd template export security-review > template.yaml
scmd template import template.yaml


# Template repositories
scmd template repo add company https://github.com/company/scmd-templates
scmd template search "kubernetes"
scmd template install company/k8s-review
```


### Template Structure


#### File: `~/.scmd/templates/security-review.yaml`


```yaml
name: security-review
version: "1.0"
author: "security-team"
description: "Security-focused code review with OWASP Top 10 emphasis"
tags:
 - security
 - owasp
 - review


# Can be used with different commands
compatible_commands:
 - review
 - explain


# System prompt (sets AI behavior)
system_prompt: |
 You are a security expert specializing in application security.
 Focus on OWASP Top 10 vulnerabilities and provide actionable remediation.
 Use clear severity ratings: Critical, High, Medium, Low.


# User prompt template (uses Go text/template syntax)
user_prompt_template: |
 Review the following {{.Language}} code for security issues:


 Focus areas:
 1. Injection attacks (SQL, Command, XSS)
 2. Broken authentication
 3. Sensitive data exposure
 4. XML external entities (XXE)
 5. Broken access control
 6. Security misconfiguration
 7. Cross-site scripting (XSS)
 8. Insecure deserialization
 9. Using components with known vulnerabilities
 10. Insufficient logging & monitoring


 Code:
 ```{{.Language}}
 {{.Code}}
 ```


 {{if .Context}}
 Additional context:
 {{.Context}}
 {{end}}


 Provide:
 - Severity rating for each issue
 - Specific vulnerabilities found
 - Remediation steps with code examples
 - Risk assessment


# Variables that can be used in template
variables:
 - name: Language
   description: "Programming language of the code"
   default: "auto-detect"


 - name: Code
   description: "The code to review"
   required: true


 - name: Context
   description: "Additional context about the codebase"
   required: false


# Output formatting
output:
 format: markdown
 sections:
   - title: "Security Assessment"
     required: true
   - title: "Vulnerabilities Found"
     required: true
   - title: "Recommendations"
     required: true


# Model recommendations
recommended_models:
 - qwen2.5-7b
 - gpt-4


# Examples
examples:
 - description: "Review authentication code"
   command: "cat login.js | scmd review --template security-review"


 - description: "Review with context"
   command: "scmd review auth.py --template security-review --context 'Flask app with JWT'"
```


### Go Implementation


#### File: `pkg/templates/template.go`


```go
package templates


import (
   "bytes"
   "fmt"
   "text/template"


   "gopkg.in/yaml.v3"
)


type Template struct {
   Name                string     `yaml:"name"`
   Version            string     `yaml:"version"`
   Author             string     `yaml:"author"`
   Description        string     `yaml:"description"`
   Tags               []string   `yaml:"tags"`
   CompatibleCommands []string   `yaml:"compatible_commands"`
   SystemPrompt       string     `yaml:"system_prompt"`
   UserPromptTemplate string     `yaml:"user_prompt_template"`
   Variables          []Variable `yaml:"variables"`
   Output             OutputConfig `yaml:"output"`
   RecommendedModels  []string   `yaml:"recommended_models"`
   Examples           []Example  `yaml:"examples"`
}


type Variable struct {
   Name        string `yaml:"name"`
   Description string `yaml:"description"`
   Default     string `yaml:"default"`
   Required    bool   `yaml:"required"`
}


type OutputConfig struct {
   Format   string    `yaml:"format"`
   Sections []Section `yaml:"sections"`
}


type Section struct {
   Title    string `yaml:"title"`
   Required bool   `yaml:"required"`
}


type Example struct {
   Description string `yaml:"description"`
   Command     string `yaml:"command"`
}


func (t *Template) Validate() error {
   if t.Name == "" {
       return fmt.Errorf("template name is required")
   }
   if t.UserPromptTemplate == "" {
       return fmt.Errorf("user_prompt_template is required")
   }
   if len(t.CompatibleCommands) == 0 {
       return fmt.Errorf("at least one compatible command is required")
   }
   return nil
}


func (t *Template) Execute(data map[string]interface{}) (string, string, error) {
   // Validate required variables
   for _, v := range t.Variables {
       if v.Required {
           if _, ok := data[v.Name]; !ok {
               return "", "", fmt.Errorf("required variable %s not provided", v.Name)
           }
       }
   }


   // Apply defaults
   for _, v := range t.Variables {
       if _, ok := data[v.Name]; !ok && v.Default != "" {
           data[v.Name] = v.Default
       }
   }


   // Execute user prompt template
   tmpl, err := template.New("user_prompt").Parse(t.UserPromptTemplate)
   if err != nil {
       return "", "", fmt.Errorf("failed to parse template: %w", err)
   }


   var userPromptBuf bytes.Buffer
   if err := tmpl.Execute(&userPromptBuf, data); err != nil {
       return "", "", fmt.Errorf("failed to execute template: %w", err)
   }


   return t.SystemPrompt, userPromptBuf.String(), nil
}


func LoadTemplate(path string) (*Template, error) {
   data, err := os.ReadFile(path)
   if err != nil {
       return nil, err
   }


   var tpl Template
   if err := yaml.Unmarshal(data, &tpl); err != nil {
       return nil, err
   }


   if err := tpl.Validate(); err != nil {
       return nil, err
   }


   return &tpl, nil
}


func (t *Template) Save(path string) error {
   data, err := yaml.Marshal(t)
   if err != nil {
       return err
   }


   return os.WriteFile(path, data, 0644)
}
```


#### File: `pkg/templates/manager.go`


```go
package templates


import (
   "fmt"
   "os"
   "path/filepath"
   "strings"
)


type TemplateManager struct {
   templatesDir string
   cache        map[string]*Template
}


func NewTemplateManager() (*TemplateManager, error) {
   configDir, err := os.UserHomeDir()
   if err != nil {
       return nil, err
   }


   templatesDir := filepath.Join(configDir, ".scmd", "templates")
   if err := os.MkdirAll(templatesDir, 0755); err != nil {
       return nil, err
   }


   return &TemplateManager{
       templatesDir: templatesDir,
       cache:        make(map[string]*Template),
   }, nil
}


func (tm *TemplateManager) Load(name string) (*Template, error) {
   // Check cache
   if t, ok := tm.cache[name]; ok {
       return t, nil
   }


   // Load from file
   path := filepath.Join(tm.templatesDir, name+".yaml")
   tpl, err := LoadTemplate(path)
   if err != nil {
       return nil, fmt.Errorf("template %s not found: %w", name, err)
   }


   // Cache it
   tm.cache[name] = tpl
   return tpl, nil
}


func (tm *TemplateManager) List() ([]*Template, error) {
   pattern := filepath.Join(tm.templatesDir, "*.yaml")
   files, err := filepath.Glob(pattern)
   if err != nil {
       return nil, err
   }


   var templates []*Template
   for _, file := range files {
       tpl, err := LoadTemplate(file)
       if err != nil {
           // Skip invalid templates
           continue
       }
       templates = append(templates, tpl)
   }


   return templates, nil
}


func (tm *TemplateManager) Create(tpl *Template) error {
   if err := tpl.Validate(); err != nil {
       return err
   }


   path := filepath.Join(tm.templatesDir, tpl.Name+".yaml")


   // Check if exists
   if _, err := os.Stat(path); err == nil {
       return fmt.Errorf("template %s already exists", tpl.Name)
   }


   return tpl.Save(path)
}


func (tm *TemplateManager) Delete(name string) error {
   path := filepath.Join(tm.templatesDir, name+".yaml")


   if err := os.Remove(path); err != nil {
       return fmt.Errorf("failed to delete template %s: %w", name, err)
   }


   delete(tm.cache, name)
   return nil
}


func (tm *TemplateManager) Search(query string) ([]*Template, error) {
   templates, err := tm.List()
   if err != nil {
       return nil, err
   }


   query = strings.ToLower(query)
   var matches []*Template


   for _, tpl := range templates {
       // Search in name, description, and tags
       if strings.Contains(strings.ToLower(tpl.Name), query) ||
          strings.Contains(strings.ToLower(tpl.Description), query) {
           matches = append(matches, tpl)
           continue
       }


       for _, tag := range tpl.Tags {
           if strings.Contains(strings.ToLower(tag), query) {
               matches = append(matches, tpl)
               break
           }
       }
   }


   return matches, nil
}


func (tm *TemplateManager) Execute(name string, data map[string]interface{}) (string, string, error) {
   tpl, err := tm.Load(name)
   if err != nil {
       return "", "", err
   }


   return tpl.Execute(data)
}
```


#### File: `cmd/template.go`


```go
package cmd


import (
   "fmt"
   "strings"


   "github.com/spf13/cobra"
   "your-project/pkg/templates"
)


var templateCmd = &cobra.Command{
   Use:   "template",
   Short: "Manage prompt templates",
   Long:  "Create, list, and manage prompt templates for customized AI interactions.",
}


var templateListCmd = &cobra.Command{
   Use:   "list",
   Short: "List available templates",
   RunE: func(cmd *cobra.Command, args []string) error {
       tm, err := templates.NewTemplateManager()
       if err != nil {
           return err
       }


       tpls, err := tm.List()
       if err != nil {
           return err
       }


       if len(tpls) == 0 {
           fmt.Println("No templates found.")
           fmt.Println("\nCreate one with: scmd template create <name>")
           return nil
       }


       fmt.Println("Available Templates:\n")
       for _, t := range tpls {
           fmt.Printf("📋 %s (v%s)\n", t.Name, t.Version)
           fmt.Printf("   %s\n", t.Description)
           if len(t.Tags) > 0 {
               fmt.Printf("   Tags: %s\n", strings.Join(t.Tags, ", "))
           }
           fmt.Printf("   Compatible: %s\n\n", strings.Join(t.CompatibleCommands, ", "))
       }


       return nil
   },
}


var templateShowCmd = &cobra.Command{
   Use:   "show <name>",
   Short: "Show template details",
   Args:  cobra.ExactArgs(1),
   RunE: func(cmd *cobra.Command, args []string) error {
       name := args[0]


       tm, err := templates.NewTemplateManager()
       if err != nil {
           return err
       }


       tpl, err := tm.Load(name)
       if err != nil {
           return err
       }


       fmt.Printf("📋 Template: %s (v%s)\n\n", tpl.Name, tpl.Version)
       fmt.Printf("Author: %s\n", tpl.Author)
       fmt.Printf("Description: %s\n\n", tpl.Description)


       if len(tpl.Tags) > 0 {
           fmt.Printf("Tags: %s\n\n", strings.Join(tpl.Tags, ", "))
       }


       fmt.Printf("Compatible Commands: %s\n\n", strings.Join(tpl.CompatibleCommands, ", "))


       if len(tpl.Variables) > 0 {
           fmt.Println("Variables:")
           for _, v := range tpl.Variables {
               req := ""
               if v.Required {
                   req = " (required)"
               }
               fmt.Printf("  - %s%s: %s\n", v.Name, req, v.Description)
               if v.Default != "" {
                   fmt.Printf("    Default: %s\n", v.Default)
               }
           }
           fmt.Println()
       }


       if len(tpl.Examples) > 0 {
           fmt.Println("Examples:")
           for _, ex := range tpl.Examples {
               fmt.Printf("  %s\n", ex.Description)
               fmt.Printf("  $ %s\n\n", ex.Command)
           }
       }


       return nil
   },
}


var templateCreateCmd = &cobra.Command{
   Use:   "create <name>",
   Short: "Create a new template interactively",
   Args:  cobra.ExactArgs(1),
   RunE: func(cmd *cobra.Command, args []string) error {
       name := args[0]


       fmt.Printf("Creating template: %s\n\n", name)


       // Interactive prompts
       description := promptString("Description: ")
       author := promptString("Author: ")
       tags := promptList("Tags (comma-separated): ")
       commands := promptList("Compatible commands (comma-separated): ")


       fmt.Println("\nEnter system prompt (defines AI behavior):")
       fmt.Println("(Press Ctrl+D when done)")
       systemPrompt := promptMultiline()


       fmt.Println("\nEnter user prompt template:")
       fmt.Println("(Use {{.VariableName}} for variables)")
       fmt.Println("(Press Ctrl+D when done)")
       userPrompt := promptMultiline()


       tpl := &templates.Template{
           Name:                name,
           Version:            "1.0",
           Author:             author,
           Description:        description,
           Tags:               tags,
           CompatibleCommands: commands,
           SystemPrompt:       systemPrompt,
           UserPromptTemplate: userPrompt,
       }


       tm, _ := templates.NewTemplateManager()
       if err := tm.Create(tpl); err != nil {
           return err
       }


       fmt.Printf("\n✅ Template '%s' created successfully\n", name)
       fmt.Printf("\nUse with: scmd review --template %s\n", name)
       fmt.Printf("View with: scmd template show %s\n", name)


       return nil
   },
}


var templateDeleteCmd = &cobra.Command{
   Use:   "delete <name>",
   Short: "Delete a template",
   Args:  cobra.ExactArgs(1),
   RunE: func(cmd *cobra.Command, args []string) error {
       name := args[0]


       tm, err := templates.NewTemplateManager()
       if err != nil {
           return err
       }


       // Confirm deletion
       fmt.Printf("Delete template '%s'? (y/N): ", name)
       var confirm string
       fmt.Scanln(&confirm)


       if strings.ToLower(confirm) != "y" {
           fmt.Println("Cancelled.")
           return nil
       }


       if err := tm.Delete(name); err != nil {
           return err
       }


       fmt.Printf("✅ Deleted template '%s'\n", name)
       return nil
   },
}


var templateExportCmd = &cobra.Command{
   Use:   "export <name>",
   Short: "Export template to stdout",
   Args:  cobra.ExactArgs(1),
   RunE: func(cmd *cobra.Command, args []string) error {
       name := args[0]


       tm, err := templates.NewTemplateManager()
       if err != nil {
           return err
       }


       tpl, err := tm.Load(name)
       if err != nil {
           return err
       }


       data, err := yaml.Marshal(tpl)
       if err != nil {
           return err
       }


       fmt.Println(string(data))
       return nil
   },
}


var templateImportCmd = &cobra.Command{
   Use:   "import <file>",
   Short: "Import template from file",
   Args:  cobra.ExactArgs(1),
   RunE: func(cmd *cobra.Command, args []string) error {
       file := args[0]


       tpl, err := templates.LoadTemplate(file)
       if err != nil {
           return err
       }


       tm, err := templates.NewTemplateManager()
       if err != nil {
           return err
       }


       if err := tm.Create(tpl); err != nil {
           return err
       }


       fmt.Printf("✅ Imported template '%s'\n", tpl.Name)
       return nil
   },
}


var templateSearchCmd = &cobra.Command{
   Use:   "search <query>",
   Short: "Search templates",
   Args:  cobra.ExactArgs(1),
   RunE: func(cmd *cobra.Command, args []string) error {
       query := args[0]


       tm, err := templates.NewTemplateManager()
       if err != nil {
           return err
       }


       matches, err := tm.Search(query)
       if err != nil {
           return err
       }


       if len(matches) == 0 {
           fmt.Printf("No templates found matching '%s'\n", query)
           return nil
       }


       fmt.Printf("Found %d template(s) matching '%s':\n\n", len(matches), query)
       for _, t := range matches {
           fmt.Printf("📋 %s - %s\n", t.Name, t.Description)
       }


       return nil
   },
}


func init() {
   rootCmd.AddCommand(templateCmd)


   templateCmd.AddCommand(templateListCmd)
   templateCmd.AddCommand(templateShowCmd)
   templateCmd.AddCommand(templateCreateCmd)
   templateCmd.AddCommand(templateDeleteCmd)
   templateCmd.AddCommand(templateExportCmd)
   templateCmd.AddCommand(templateImportCmd)
   templateCmd.AddCommand(templateSearchCmd)


   // Add --template flag to existing commands
   reviewCmd.Flags().String("template", "", "Use a prompt template")
   explainCmd.Flags().String("template", "", "Use a prompt template")
}


// Helper functions
func promptString(prompt string) string {
   fmt.Print(prompt)
   var input string
   fmt.Scanln(&input)
   return input
}


func promptList(prompt string) []string {
   fmt.Print(prompt)
   var input string
   fmt.Scanln(&input)


   parts := strings.Split(input, ",")
   var result []string
   for _, p := range parts {
       result = append(result, strings.TrimSpace(p))
   }
   return result
}


func promptMultiline() string {
   scanner := bufio.NewScanner(os.Stdin)
   var lines []string


   for scanner.Scan() {
       lines = append(lines, scanner.Text())
   }


   return strings.Join(lines, "\n")
}
```


### Integration with Existing Commands


Update `cmd/review.go` and `cmd/explain.go`:


```go
func runReview(file string) error {
   code, err := readFile(file)
   if err != nil {
       return err
   }


   var systemPrompt, userPrompt string


   templateName := viper.GetString("template")
   if templateName != "" {
       // Use template
       tm, _ := templates.NewTemplateManager()


       data := map[string]interface{}{
           "Code":     code,
           "Language": detectLanguage(file),
           "Context":  viper.GetString("context"),
       }


       systemPrompt, userPrompt, err = tm.Execute(templateName, data)
       if err != nil {
           return err
       }
   } else {
       // Use default prompt
       userPrompt = fmt.Sprintf("Review this code:\n\n%s", code)
   }


   // Generate response with system and user prompts
   response, err := generateResponseWithSystem(systemPrompt, userPrompt)
   if err != nil {
       return err
   }


   // Format and display
   return displayResponse(response)
}
```


### Built-in Templates to Ship


Include these templates by default in the release:


#### 1. `security-review.yaml`
- OWASP Top 10 focus
- Severity ratings
- Remediation steps


#### 2. `performance.yaml`
- Performance bottlenecks
- Big O analysis
- Optimization suggestions


#### 3. `api-design.yaml`
- REST best practices
- HTTP method correctness
- Error handling
- API documentation


#### 4. `testing.yaml`
- Test coverage analysis
- Edge case identification
- Test quality review


#### 5. `documentation.yaml`
- Documentation completeness
- Code comment quality
- README suggestions


#### 6. `accessibility.yaml`
- WCAG compliance
- Keyboard navigation
- Screen reader support


#### 7. `beginner-explain.yaml`
- Explain code to beginners
- Step-by-step breakdown
- Analogies and examples


#### 8. `error-handling.yaml`
- Error handling patterns
- Exception safety
- Logging recommendations


---


## Implementation Timeline


### Week 1: Interactive Conversation Mode
- **Day 1-2:** Database schema & storage implementation
 - Set up SQLite
 - Create migrations
 - Implement CRUD operations


- **Day 3-4:** Interactive session logic
 - Build chat loop
 - Command handling
 - Context management


- **Day 5-7:** History management & CLI
 - History commands
 - Search functionality
 - Testing & debugging


### Week 2: Beautiful Markdown Output
- **Day 1-3:** Glamour integration & theming
 - Add dependencies
 - Create formatter
 - Custom themes


- **Day 4-5:** Syntax highlighting & code blocks
 - Chroma integration
 - Language detection
 - Code block styling


- **Day 6-7:** Streaming output & polish
 - Stream renderer
 - Progress spinners
 - Final touches


### Week 3: Template System
- **Day 1-2:** Template structure & manager
 - YAML schema
 - Template loading
 - Validation


- **Day 3-4:** CLI commands & execution
 - Create commands
 - Template execution
 - Variable substitution


- **Day 5-7:** Built-in templates & documentation
 - Write 8 templates
 - Examples
 - Documentation


### Week 4: Testing & Polish
- **Day 1-2:** Integration testing
 - Test all features together
 - Edge cases
 - Error handling


- **Day 3-4:** Documentation
 - README updates
 - Command help text
 - Tutorial


- **Day 5-7:** Release prep
 - Final polish
 - Performance tuning
 - Release notes
 - Public launch! 🚀


---


## Testing Checklist


### Conversation Mode
- [ ] Start new conversation
- [ ] Resume conversation by ID
- [ ] List conversations
- [ ] Search conversations
- [ ] Delete conversations
- [ ] Export to markdown
- [ ] Handle interruptions (Ctrl+C, Ctrl+D)
- [ ] Long conversations (50+ messages)
- [ ] Special characters in messages
- [ ] Context window management


### Beautiful Output
- [ ] Render markdown correctly
- [ ] Syntax highlighting works for major languages
- [ ] Light/dark theme switching
- [ ] Streaming output displays properly
- [ ] Progress spinners work
- [ ] JSON output format intact
- [ ] Plain text mode (--no-colors)
- [ ] Long output handling


### Template System
- [ ] Create template
- [ ] List templates
- [ ] Show template details
- [ ] Use template in review
- [ ] Use template in explain
- [ ] Template variables work
- [ ] Required variables enforced
- [ ] Export/import templates
- [ ] Search templates
- [ ] Delete templates


---


## Success Metrics


After implementing these features, scmd should achieve:


### User Experience
✅ Professional, polished first impression
✅ Feature parity with competitors (sgpt, llm, mods)
✅ Unique value: local + templates + cmd generation
✅ Extensible without plugin complexity


### Technical
✅ Conversation history persists
✅ Beautiful, syntax-highlighted output
✅ Users can create and share custom templates
✅ Performance remains fast (<2s for analysis)


### Business
✅ Ready for public launch
✅ Competitive positioning established
✅ Community contribution enabled (templates)
✅ Word-of-mouth appeal ("wow" factor)


---


## Appendices


### A. Configuration File Reference


Complete `~/.scmd/config.yaml`:


```yaml
version: "1.0"


# Backend configuration
backends:
 default: llamacpp
 local:
   model: qwen2.5-1.5b
   context_length: 0  # 0 = use model's native max


# UI configuration
ui:
 colors: true
 style: auto  # auto, dark, light, notty
 streaming: true
 pager: false
 verbose: false


 markdown:
   render: true
   code_theme: monokai
   wrap_width: 100
   show_line_numbers: false


# Chat configuration
chat:
 max_context_messages: 20
 auto_save: true
 auto_title: true


# Model configuration
models:
 directory: ~/.scmd/models
 auto_download: true


# Template configuration
templates:
 directory: ~/.scmd/templates
 auto_update: false
```


### B. Dependencies to Add


```bash
go get github.com/charmbracelet/glamour@v0.6.0
go get github.com/charmbracelet/lipgloss@v0.9.1
go get github.com/alecthomas/chroma/v2@v2.12.0
go get github.com/muesli/termenv@v0.15.2
go get github.com/briandowns/spinner@v1.23.0
go get github.com/mattn/go-sqlite3@v1.14.18
go get github.com/google/uuid@v1.5.0
go get gopkg.in/yaml.v3@v3.0.1
```


### C. Directory Structure After Implementation


```
~/.scmd/
├── config.yaml
├── conversations.db
├── models/
│   ├── qwen2.5-1.5b.gguf
│   └── qwen2.5-7b.gguf
├── templates/
│   ├── security-review.yaml
│   ├── performance.yaml
│   ├── api-design.yaml
│   ├── testing.yaml
│   ├── documentation.yaml
│   ├── accessibility.yaml
│   ├── beginner-explain.yaml
│   └── error-handling.yaml
└── cache/
   └── repos/
```


### D. Example Template: Performance Review


```yaml
name: performance
version: "1.0"
author: "scmd"
description: "Performance optimization and bottleneck analysis"
tags:
 - performance
 - optimization
 - profiling


compatible_commands:
 - review
 - explain


system_prompt: |
 You are a performance optimization expert.
 Analyze code for performance bottlenecks, algorithmic complexity,
 and suggest concrete optimizations.
 Provide Big O analysis where relevant.


user_prompt_template: |
 Analyze the following {{.Language}} code for performance:


 ```{{.Language}}
 {{.Code}}
 ```


 Focus on:
 1. Time complexity (Big O)
 2. Space complexity
 3. Bottlenecks and hot paths
 4. Memory allocation patterns
 5. Loop optimizations
 6. Caching opportunities
 7. Algorithmic improvements


 {{if .Context}}
 Context: {{.Context}}
 {{end}}


 Provide:
 - Current complexity analysis
 - Identified bottlenecks
 - Optimization suggestions with code examples
 - Expected performance impact


variables:
 - name: Language
   description: "Programming language"
   default: "auto-detect"
 - name: Code
   description: "Code to analyze"
   required: true
 - name: Context
   description: "Performance requirements or constraints"
   required: false


output:
 format: markdown
 sections:
   - title: "Complexity Analysis"
     required: true
   - title: "Bottlenecks"
     required: true
   - title: "Optimizations"
     required: true


recommended_models:
 - qwen2.5-7b


examples:
 - description: "Review sorting algorithm"
   command: "cat sort.py | scmd review --template performance"
 - description: "With performance target"
   command: "scmd review api.js --template performance --context 'Must handle 10k req/sec'"
```


---


## Final Notes


### Priority Order


If you need to implement these incrementally:


1. **Conversation Mode** (Week 1) - Most critical, biggest gap
2. **Beautiful Output** (Week 2) - High impact, first impression
3. **Template System** (Week 3) - Differentiator, extensibility


### Quick Wins


Some features can be implemented faster:


- **Progress spinners** - 1 day
- **Basic markdown rendering** - 2 days
- **Simple templates** - 3 days


### Don't Forget


- [ ] Update README with new features
- [ ] Add examples to documentation
- [ ] Create demo GIFs/videos
- [ ] Write release blog post
- [ ] Prepare for Hacker News/Reddit launch


---


**Ready to ship!** These 3 features will make scmd competitive with established tools while maintaining its unique advantages. Good luck with the implementation! 🚀






