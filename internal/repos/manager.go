// Package repos provides repository management for scmd plugins
package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/scmd/scmd/internal/validation"
)

// Repository represents a plugin repository
type Repository struct {
	Name        string    `json:"name" yaml:"name"`
	URL         string    `json:"url" yaml:"url"`
	Description string    `json:"description,omitempty" yaml:"description,omitempty"`
	Enabled     bool      `json:"enabled" yaml:"enabled"`
	LastUpdated time.Time `json:"last_updated,omitempty" yaml:"last_updated,omitempty"`
}

// Manifest is the repo's scmd-repo.yaml file
type Manifest struct {
	Name        string         `yaml:"name"`
	Version     string         `yaml:"version"`
	Description string         `yaml:"description"`
	Author      string         `yaml:"author,omitempty"`
	Homepage    string         `yaml:"homepage,omitempty"`
	Commands    []Command      `yaml:"commands"`
	Templates   []TemplateRef  `yaml:"templates,omitempty"` // Template references in repository
}

// Command represents a slash command from a repo
type Command struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Usage       string   `yaml:"usage,omitempty"`
	Aliases     []string `yaml:"aliases,omitempty"`
	Category    string   `yaml:"category,omitempty"`
	File        string   `yaml:"file"` // Path to command YAML file in repo
	Path        string   `yaml:"path"` // Legacy support: alternative to File field
}

// CommandSpec is the full command specification from a YAML file
type CommandSpec struct {
	Name        string            `yaml:"name" json:"name"`
	Version     string            `yaml:"version" json:"version"`
	Description string            `yaml:"description" json:"description"`
	Usage       string            `yaml:"usage" json:"usage"`
	Aliases     []string          `yaml:"aliases,omitempty" json:"aliases,omitempty"`
	Category    string            `yaml:"category,omitempty" json:"category,omitempty"`
	Author      string            `yaml:"author,omitempty" json:"author,omitempty"`
	License     string            `yaml:"license,omitempty" json:"license,omitempty"`
	Homepage    string            `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	Repository  string            `yaml:"repository,omitempty" json:"repository,omitempty"`
	Tags        []string          `yaml:"tags,omitempty" json:"tags,omitempty"`
	Args        []ArgSpec         `yaml:"args,omitempty" json:"args,omitempty"`
	Flags       []FlagSpec        `yaml:"flags,omitempty" json:"flags,omitempty"`
	Prompt      PromptSpec        `yaml:"prompt" json:"prompt"`
	Model       ModelSpec         `yaml:"model,omitempty" json:"model,omitempty"`
	Examples    []string          `yaml:"examples,omitempty" json:"examples,omitempty"`
	Metadata    map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`

	// Template integration (NEW in v0.4.2)
	Template *TemplateRef `yaml:"template,omitempty" json:"template,omitempty"`

	// Enhanced features
	Dependencies []Dependency `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	Compose      *ComposeSpec `yaml:"compose,omitempty" json:"compose,omitempty"`
	Hooks        *HooksSpec   `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	Inputs       []InputSpec  `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	Outputs      *OutputSpec  `yaml:"outputs,omitempty" json:"outputs,omitempty"`
	Context      *ContextSpec `yaml:"context,omitempty" json:"context,omitempty"`
}

// TemplateRef references a template for command execution
type TemplateRef struct {
	// Name of the template to use (references ~/.scmd/templates/<name>.yaml)
	Name string `yaml:"name,omitempty" json:"name,omitempty"`

	// Variables to map from command context to template variables
	// Keys are template variable names, values are command context references
	// e.g., "Language": "{{.file_extension}}", "Code": "{{.file_content}}"
	Variables map[string]string `yaml:"variables,omitempty" json:"variables,omitempty"`

	// Inline template definition (alternative to referencing by name)
	Inline *InlineTemplate `yaml:"inline,omitempty" json:"inline,omitempty"`
}

// InlineTemplate defines a template inline within the command spec
type InlineTemplate struct {
	SystemPrompt       string            `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`
	UserPromptTemplate string            `yaml:"user_prompt_template" json:"user_prompt_template"`
	Variables          []TemplateVariable `yaml:"variables,omitempty" json:"variables,omitempty"`
}

// TemplateVariable defines a variable for inline templates
type TemplateVariable struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Default     string `yaml:"default,omitempty" json:"default,omitempty"`
	Required    bool   `yaml:"required,omitempty" json:"required,omitempty"`
}

// Dependency defines a command dependency
type Dependency struct {
	Command     string `yaml:"command" json:"command"`                     // e.g., "official/explain"
	Version     string `yaml:"version,omitempty" json:"version,omitempty"` // e.g., ">=1.0.0"
	Optional    bool   `yaml:"optional,omitempty" json:"optional,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// ComposeSpec defines command composition/chaining
type ComposeSpec struct {
	// Pipeline chains multiple commands together
	Pipeline []PipelineStep `yaml:"pipeline,omitempty" json:"pipeline,omitempty"`
	// Parallel runs commands in parallel and merges results
	Parallel []string `yaml:"parallel,omitempty" json:"parallel,omitempty"`
	// Fallback tries commands in order until one succeeds
	Fallback []string `yaml:"fallback,omitempty" json:"fallback,omitempty"`
}

// PipelineStep is a step in a command pipeline
type PipelineStep struct {
	Command   string            `yaml:"command" json:"command"`
	Args      map[string]string `yaml:"args,omitempty" json:"args,omitempty"`
	Transform string            `yaml:"transform,omitempty" json:"transform,omitempty"` // jq-like transform
	OnError   string            `yaml:"on_error,omitempty" json:"on_error,omitempty"`   // continue, stop, fallback
}

// HooksSpec defines pre/post execution hooks
type HooksSpec struct {
	Pre  []HookAction `yaml:"pre,omitempty" json:"pre,omitempty"`
	Post []HookAction `yaml:"post,omitempty" json:"post,omitempty"`
}

// HookAction is a hook action to run
type HookAction struct {
	Shell   string `yaml:"shell,omitempty" json:"shell,omitempty"`     // Shell command to run
	Command string `yaml:"command,omitempty" json:"command,omitempty"` // Another scmd command
	If      string `yaml:"if,omitempty" json:"if,omitempty"`           // Condition
}

// InputSpec defines structured input (like GitHub Actions inputs)
type InputSpec struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Required    bool     `yaml:"required,omitempty" json:"required,omitempty"`
	Default     string   `yaml:"default,omitempty" json:"default,omitempty"`
	Type        string   `yaml:"type,omitempty" json:"type,omitempty"` // string, file, choice, multiline
	Choices     []string `yaml:"choices,omitempty" json:"choices,omitempty"`
}

// OutputSpec defines command output structure
type OutputSpec struct {
	Format   string            `yaml:"format,omitempty" json:"format,omitempty"`     // text, json, markdown
	Schema   map[string]string `yaml:"schema,omitempty" json:"schema,omitempty"`     // JSON schema hints
	Template string            `yaml:"template,omitempty" json:"template,omitempty"` // Output template
}

// ContextSpec defines context requirements
type ContextSpec struct {
	Files     []string `yaml:"files,omitempty" json:"files,omitempty"`           // File patterns to include
	Git       bool     `yaml:"git,omitempty" json:"git,omitempty"`               // Include git context
	Env       []string `yaml:"env,omitempty" json:"env,omitempty"`               // Environment variables
	MaxTokens int      `yaml:"max_tokens,omitempty" json:"max_tokens,omitempty"` // Max context tokens
}

// ArgSpec defines a command argument
type ArgSpec struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     string `yaml:"default,omitempty"`
}

// FlagSpec defines a command flag
type FlagSpec struct {
	Name        string `yaml:"name"`
	Short       string `yaml:"short,omitempty"`
	Description string `yaml:"description"`
	Default     string `yaml:"default,omitempty"`
}

// PromptSpec defines the prompt template
type PromptSpec struct {
	System   string `yaml:"system,omitempty"`
	Template string `yaml:"template"`
}

// ModelSpec defines model preferences
type ModelSpec struct {
	Preferred   string  `yaml:"preferred,omitempty"`
	MinContext  int     `yaml:"min_context,omitempty"`
	Temperature float64 `yaml:"temperature,omitempty"`
	MaxTokens   int     `yaml:"max_tokens,omitempty"`
}

// Manager manages repositories
type Manager struct {
	mu         sync.RWMutex
	repos      map[string]*Repository
	dataDir    string
	httpClient *http.Client
}

// NewManager creates a new repository manager
func NewManager(dataDir string) *Manager {
	return &Manager{
		repos:   make(map[string]*Repository),
		dataDir: dataDir,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// reposFile returns the path to repos.json
func (m *Manager) reposFile() string {
	return filepath.Join(m.dataDir, "repos.json")
}

// Load loads repositories from disk
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.reposFile())
	if err != nil {
		if os.IsNotExist(err) {
			// Add default repos
			m.repos["official"] = &Repository{
				Name:        "official",
				URL:         "https://raw.githubusercontent.com/sunboylabs/commands/main",
				Description: "Official scmd commands",
				Enabled:     true,
			}
			return nil
		}
		return err
	}

	var repos []*Repository
	if err := json.Unmarshal(data, &repos); err != nil {
		return err
	}

	// Validate and load repositories
	for _, r := range repos {
		// Validate URL (SECURITY: prevent SSRF and file:// access)
		if err := validation.ValidateRepoURL(r.URL); err != nil {
			return fmt.Errorf("invalid repository '%s': %w", r.Name, err)
		}
		m.repos[r.Name] = r
	}

	return nil
}

// Save saves repositories to disk
func (m *Manager) Save() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := os.MkdirAll(m.dataDir, 0755); err != nil {
		return err
	}

	repos := make([]*Repository, 0, len(m.repos))
	for _, r := range m.repos {
		repos = append(repos, r)
	}

	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.reposFile(), data, 0644)
}

// Add adds a new repository
func (m *Manager) Add(name, url string) error {
	// Validate URL (SECURITY: prevent SSRF, file:// access, and other attacks)
	if err := validation.ValidateRepoURL(url); err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.repos[name]; exists {
		return fmt.Errorf("repository '%s' already exists", name)
	}

	m.repos[name] = &Repository{
		Name:    name,
		URL:     url,
		Enabled: true,
	}

	return nil
}

// Remove removes a repository
func (m *Manager) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.repos[name]; !exists {
		return fmt.Errorf("repository '%s' not found", name)
	}

	delete(m.repos, name)
	return nil
}

// Get returns a repository by name
func (m *Manager) Get(name string) (*Repository, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, ok := m.repos[name]
	return r, ok
}

// List returns all repositories
func (m *Manager) List() []*Repository {
	m.mu.RLock()
	defer m.mu.RUnlock()

	repos := make([]*Repository, 0, len(m.repos))
	for _, r := range m.repos {
		repos = append(repos, r)
	}
	return repos
}

// FetchManifest fetches and parses a repo's manifest
func (m *Manager) FetchManifest(ctx context.Context, repo *Repository) (*Manifest, error) {
	url := repo.URL + "/scmd-repo.yaml"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch manifest: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	// Normalize legacy format: handle path-only entries
	if err := m.normalizeManifest(ctx, repo, &manifest); err != nil {
		return nil, fmt.Errorf("normalize manifest: %w", err)
	}

	return &manifest, nil
}

// normalizeManifest handles legacy manifest formats by fetching metadata from command files
func (m *Manager) normalizeManifest(ctx context.Context, repo *Repository, manifest *Manifest) error {
	// Collect commands that need metadata fetching
	type fetchJob struct {
		index int
		cmd   *Command
	}

	var needsFetch []fetchJob
	for i := range manifest.Commands {
		cmd := &manifest.Commands[i]

		// Handle legacy path field
		if cmd.File == "" && cmd.Path != "" {
			cmd.File = cmd.Path
		}

		// If name or description is missing, we need to fetch the command file
		if (cmd.Name == "" || cmd.Description == "") && cmd.File != "" {
			needsFetch = append(needsFetch, fetchJob{index: i, cmd: cmd})
		}
	}

	// Nothing to fetch
	if len(needsFetch) == 0 {
		return nil
	}

	// Fetch command metadata in parallel (up to 10 concurrent requests)
	concurrency := 10
	if len(needsFetch) < concurrency {
		concurrency = len(needsFetch)
	}

	semaphore := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, job := range needsFetch {
		wg.Add(1)
		go func(j fetchJob) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			spec, err := m.FetchCommand(ctx, repo, j.cmd.File)
			if err != nil {
				// Skip commands that fail to fetch
				return
			}

			// Update command metadata
			mu.Lock()
			if j.cmd.Name == "" {
				j.cmd.Name = spec.Name
			}
			if j.cmd.Description == "" {
				j.cmd.Description = spec.Description
			}
			if j.cmd.Category == "" {
				j.cmd.Category = spec.Category
			}
			if j.cmd.Usage == "" && spec.Usage != "" {
				j.cmd.Usage = spec.Usage
			}
			if len(j.cmd.Aliases) == 0 && len(spec.Aliases) > 0 {
				j.cmd.Aliases = spec.Aliases
			}
			mu.Unlock()
		}(job)
	}

	wg.Wait()
	return nil
}

// FetchCommand fetches a command spec from a repo
func (m *Manager) FetchCommand(ctx context.Context, repo *Repository, cmdPath string) (*CommandSpec, error) {
	url := repo.URL + "/" + cmdPath

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch command: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch command: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var spec CommandSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parse command: %w", err)
	}

	return &spec, nil
}

// SearchCommands searches for commands across all repos
func (m *Manager) SearchCommands(ctx context.Context, query string) ([]SearchResult, error) {
	m.mu.RLock()
	repos := make([]*Repository, 0, len(m.repos))
	for _, r := range m.repos {
		if r.Enabled {
			repos = append(repos, r)
		}
	}
	m.mu.RUnlock()

	var results []SearchResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, repo := range repos {
		wg.Add(1)
		go func(r *Repository) {
			defer wg.Done()

			manifest, err := m.FetchManifest(ctx, r)
			if err != nil {
				return
			}

			for _, cmd := range manifest.Commands {
				if matchesQuery(cmd, query) {
					mu.Lock()
					results = append(results, SearchResult{
						Repo:    r.Name,
						Command: cmd,
					})
					mu.Unlock()
				}
			}
		}(repo)
	}

	wg.Wait()
	return results, nil
}

// SearchResult represents a search result
type SearchResult struct {
	Repo    string
	Command Command
}

func matchesQuery(cmd Command, query string) bool {
	if query == "" {
		return true
	}
	// Simple substring match
	return contains(cmd.Name, query) ||
		contains(cmd.Description, query) ||
		contains(cmd.Category, query)
}

func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && containsLower(s, substr))
}

func containsLower(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

// InstallCommand saves a command spec to local storage
func (m *Manager) InstallCommand(spec *CommandSpec, installDir string) error {
	// Validate template reference if present
	if spec.Template != nil {
		if err := m.validateTemplateRef(spec.Template); err != nil {
			return fmt.Errorf("invalid template reference: %w", err)
		}
	}

	data, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Errorf("marshal command: %w", err)
	}

	filename := spec.Name + ".yaml"
	filepath := filepath.Join(installDir, filename)

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("write command: %w", err)
	}

	return nil
}

// validateTemplateRef validates a template reference
func (m *Manager) validateTemplateRef(ref *TemplateRef) error {
	// Must have either name or inline, not both
	hasName := ref.Name != ""
	hasInline := ref.Inline != nil

	if !hasName && !hasInline {
		return fmt.Errorf("template must specify either 'name' or 'inline'")
	}

	if hasName && hasInline {
		return fmt.Errorf("template cannot specify both 'name' and 'inline'")
	}

	// Validate inline template if present
	if hasInline {
		if ref.Inline.UserPromptTemplate == "" {
			return fmt.Errorf("inline template must have user_prompt_template")
		}

		// Validate required variables
		for _, v := range ref.Inline.Variables {
			if v.Name == "" {
				return fmt.Errorf("inline template variable must have a name")
			}
		}
	}

	return nil
}

// LoadInstalledCommands loads all installed commands from local storage
func (m *Manager) LoadInstalledCommands(installDir string) ([]*CommandSpec, error) {
	entries, err := os.ReadDir(installDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var commands []*CommandSpec
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) < 5 || name[len(name)-5:] != ".yaml" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(installDir, name))
		if err != nil {
			continue
		}

		var spec CommandSpec
		if err := yaml.Unmarshal(data, &spec); err != nil {
			continue
		}

		commands = append(commands, &spec)
	}

	return commands, nil
}

// UninstallCommand removes an installed command
func (m *Manager) UninstallCommand(name, installDir string) error {
	filename := name + ".yaml"
	filepath := filepath.Join(installDir, filename)

	if err := os.Remove(filepath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("command '%s' is not installed", name)
		}
		return err
	}

	return nil
}
