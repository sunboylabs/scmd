package chat

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/config"
)

// ConversationStore handles SQLite persistence for conversations
type ConversationStore struct {
	db *sql.DB
}

// ConversationSummary represents a conversation in list view
type ConversationSummary struct {
	ID           string
	Title        string
	Model        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	MessageCount int
}

// OpenConversationStore opens or creates the conversation database
func OpenConversationStore() (*ConversationStore, error) {
	// Get data directory from config
	dataDir := config.GetDataDir()
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "conversations.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Initialize schema
	if err := initSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return &ConversationStore{db: db}, nil
}

func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS conversations (
		id TEXT PRIMARY KEY,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		title TEXT,
		model TEXT,
		backend TEXT,
		metadata JSON
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		conversation_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		tokens INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_conv_updated ON conversations(updated_at DESC);
	CREATE INDEX IF NOT EXISTS idx_msg_conv ON messages(conversation_id);
	`

	_, err := db.Exec(schema)
	return err
}

// SaveConversation saves or updates a conversation
func (cs *ConversationStore) SaveConversation(session *Session) error {
	tx, err := cs.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()

	// Generate title from first message if not set
	title := ""
	messages := session.GetMessages()
	if len(messages) > 0 {
		title = generateTitle(messages[0].Content)
	}

	// Insert or update conversation
	_, err = tx.Exec(`
		INSERT INTO conversations (id, created_at, updated_at, title, model, backend, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			updated_at = ?,
			title = COALESCE(conversations.title, ?),
			model = ?,
			backend = ?
	`, session.GetConversationID(), now, now, title, session.GetModel(),
	   session.GetBackendName(), "{}",
	   now, title, session.GetModel(), session.GetBackendName())

	if err != nil {
		return fmt.Errorf("failed to save conversation: %w", err)
	}

	// Delete old messages and insert new ones
	_, err = tx.Exec("DELETE FROM messages WHERE conversation_id = ?", session.GetConversationID())
	if err != nil {
		return fmt.Errorf("failed to delete old messages: %w", err)
	}

	for _, msg := range messages {
		_, err = tx.Exec(`
			INSERT INTO messages (conversation_id, role, content, tokens, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, session.GetConversationID(), msg.Role, msg.Content, msg.Tokens, msg.CreatedAt)

		if err != nil {
			return fmt.Errorf("failed to save message: %w", err)
		}
	}

	return tx.Commit()
}

// LoadSession loads a conversation from the database
func (cs *ConversationStore) LoadSession(id string, backend backend.Backend) (*Session, error) {
	// Support partial ID matching
	fullID, err := cs.resolveConversationID(id)
	if err != nil {
		return nil, err
	}

	// Load conversation metadata
	var convID, title, model, backendName string
	var createdAt, updatedAt time.Time

	err = cs.db.QueryRow(`
		SELECT id, COALESCE(title, ''), model, backend, created_at, updated_at
		FROM conversations WHERE id = ?
	`, fullID).Scan(&convID, &title, &model, &backendName, &createdAt, &updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("conversation not found: %s", id)
		}
		return nil, fmt.Errorf("failed to load conversation: %w", err)
	}

	// Load messages
	rows, err := cs.db.Query(`
		SELECT role, content, tokens, created_at
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
	`, convID)

	if err != nil {
		return nil, fmt.Errorf("failed to load messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.Role, &msg.Content, &msg.Tokens, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate messages: %w", err)
	}

	// Create session
	session := &Session{
		conversationID: convID,
		messages:       messages,
		model:          model,
		backend:        backend,
		db:             cs,
		config: &Config{
			MaxContextMessages: 20,
			AutoSave:          true,
			AutoTitle:         true,
		},
	}

	return session, nil
}

// List returns recent conversations
func (cs *ConversationStore) List(limit int) ([]ConversationSummary, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		SELECT c.id, COALESCE(c.title, ''), c.model, c.created_at, c.updated_at,
		       COUNT(m.id) as message_count
		FROM conversations c
		LEFT JOIN messages m ON c.id = m.conversation_id
		GROUP BY c.id
		ORDER BY c.updated_at DESC
		LIMIT ?
	`

	rows, err := cs.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer rows.Close()

	var conversations []ConversationSummary
	for rows.Next() {
		var conv ConversationSummary
		if err := rows.Scan(&conv.ID, &conv.Title, &conv.Model,
			&conv.CreatedAt, &conv.UpdatedAt, &conv.MessageCount); err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, conv)
	}

	return conversations, rows.Err()
}

// Search finds conversations containing the query
func (cs *ConversationStore) Search(query string) ([]ConversationSummary, error) {
	sqlQuery := `
		SELECT DISTINCT c.id, COALESCE(c.title, ''), c.model, c.created_at, c.updated_at,
		       COUNT(m.id) as message_count
		FROM conversations c
		JOIN messages m ON c.id = m.conversation_id
		WHERE m.content LIKE ? OR c.title LIKE ?
		GROUP BY c.id
		ORDER BY c.updated_at DESC
		LIMIT 20
	`

	searchPattern := "%" + query + "%"
	rows, err := cs.db.Query(sqlQuery, searchPattern, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search conversations: %w", err)
	}
	defer rows.Close()

	var conversations []ConversationSummary
	for rows.Next() {
		var conv ConversationSummary
		if err := rows.Scan(&conv.ID, &conv.Title, &conv.Model,
			&conv.CreatedAt, &conv.UpdatedAt, &conv.MessageCount); err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, conv)
	}

	return conversations, rows.Err()
}

// Delete removes a conversation
func (cs *ConversationStore) Delete(id string) error {
	// Support partial ID matching
	fullID, err := cs.resolveConversationID(id)
	if err != nil {
		return err
	}

	result, err := cs.db.Exec("DELETE FROM conversations WHERE id = ?", fullID)
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check deletion: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("conversation not found: %s", id)
	}

	return nil
}

// GetCreatedAt returns when a conversation was created
func (cs *ConversationStore) GetCreatedAt(id string) time.Time {
	var createdAt time.Time
	cs.db.QueryRow("SELECT created_at FROM conversations WHERE id = ?", id).Scan(&createdAt)
	return createdAt
}

// ShowConversation displays a conversation's messages
func (cs *ConversationStore) ShowConversation(id string) (string, error) {
	// Support partial ID matching
	fullID, err := cs.resolveConversationID(id)
	if err != nil {
		return "", err
	}

	// Get conversation info
	var title, model string
	var createdAt time.Time
	err = cs.db.QueryRow(`
		SELECT COALESCE(title, ''), model, created_at
		FROM conversations WHERE id = ?
	`, fullID).Scan(&title, &model, &createdAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("conversation not found: %s", id)
		}
		return "", fmt.Errorf("failed to load conversation: %w", err)
	}

	// Load messages
	rows, err := cs.db.Query(`
		SELECT role, content, created_at
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
	`, fullID)

	if err != nil {
		return "", fmt.Errorf("failed to load messages: %w", err)
	}
	defer rows.Close()

	// Build output
	var output string
	output += fmt.Sprintf("Conversation: %s\n", fullID[:8])
	if title != "" {
		output += fmt.Sprintf("Title: %s\n", title)
	}
	output += fmt.Sprintf("Model: %s\n", model)
	output += fmt.Sprintf("Started: %s\n\n", createdAt.Format("2006-01-02 15:04"))
	output += "---\n\n"

	for rows.Next() {
		var role, content string
		var msgTime time.Time
		if err := rows.Scan(&role, &content, &msgTime); err != nil {
			return "", fmt.Errorf("failed to scan message: %w", err)
		}

		roleDisplay := "You"
		if role == "assistant" {
			roleDisplay = "Assistant"
		}
		output += fmt.Sprintf("**%s** (%s):\n%s\n\n",
			roleDisplay, msgTime.Format("15:04"), content)
	}

	return output, rows.Err()
}

// ClearAll deletes all conversations
func (cs *ConversationStore) ClearAll() error {
	_, err := cs.db.Exec("DELETE FROM conversations")
	if err != nil {
		return fmt.Errorf("failed to clear conversations: %w", err)
	}
	return nil
}

// Close closes the database connection
func (cs *ConversationStore) Close() error {
	return cs.db.Close()
}

// resolveConversationID finds full ID from partial match
func (cs *ConversationStore) resolveConversationID(partial string) (string, error) {
	// If it looks like a full UUID, use it directly
	if len(partial) >= 32 {
		return partial, nil
	}

	// Otherwise, try to find a matching conversation
	var fullID string
	err := cs.db.QueryRow(`
		SELECT id FROM conversations
		WHERE id LIKE ?
		ORDER BY updated_at DESC
		LIMIT 1
	`, partial+"%").Scan(&fullID)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no conversation found matching: %s", partial)
		}
		return "", fmt.Errorf("failed to resolve conversation ID: %w", err)
	}

	return fullID, nil
}

// generateTitle creates a title from the first message
func generateTitle(firstMessage string) string {
	// Extract first 50 characters for title
	if len(firstMessage) > 50 {
		// Try to break at word boundary
		title := firstMessage[:50]
		lastSpace := strings.LastIndex(title, " ")
		if lastSpace > 30 {
			title = title[:lastSpace]
		}
		return title + "..."
	}
	return firstMessage
}

// ExportConversationToJSON exports a conversation to JSON format
func (cs *ConversationStore) ExportConversationToJSON(id string) ([]byte, error) {
	fullID, err := cs.resolveConversationID(id)
	if err != nil {
		return nil, err
	}

	// Load conversation and messages
	var conv struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		Model     string    `json:"model"`
		Backend   string    `json:"backend"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Messages  []Message `json:"messages"`
	}

	err = cs.db.QueryRow(`
		SELECT id, COALESCE(title, ''), model, backend, created_at, updated_at
		FROM conversations WHERE id = ?
	`, fullID).Scan(&conv.ID, &conv.Title, &conv.Model, &conv.Backend,
		&conv.CreatedAt, &conv.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to load conversation: %w", err)
	}

	// Load messages
	rows, err := cs.db.Query(`
		SELECT role, content, tokens, created_at
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
	`, fullID)

	if err != nil {
		return nil, fmt.Errorf("failed to load messages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.Role, &msg.Content, &msg.Tokens, &msg.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		conv.Messages = append(conv.Messages, msg)
	}

	return json.MarshalIndent(conv, "", "  ")
}