package repos

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	m := NewManager("/tmp/test")
	assert.NotNil(t, m)
	assert.Equal(t, "/tmp/test", m.dataDir)
}

func TestManager_AddRemove(t *testing.T) {
	m := NewManager("/tmp/test")

	// Add repo
	err := m.Add("test", "https://example.com/repo")
	assert.NoError(t, err)

	// Get repo
	repo, ok := m.Get("test")
	assert.True(t, ok)
	assert.Equal(t, "test", repo.Name)
	assert.Equal(t, "https://example.com/repo", repo.URL)
	assert.True(t, repo.Enabled)

	// Add duplicate
	err = m.Add("test", "https://other.com/repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Remove repo
	err = m.Remove("test")
	assert.NoError(t, err)

	// Get removed
	_, ok = m.Get("test")
	assert.False(t, ok)

	// Remove non-existent
	err = m.Remove("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_List(t *testing.T) {
	m := NewManager("/tmp/test")

	// Empty list
	list := m.List()
	assert.Empty(t, list)

	// Add repos
	_ = m.Add("repo1", "https://example.com/repo1")
	_ = m.Add("repo2", "https://example.com/repo2")

	list = m.List()
	assert.Len(t, list, 2)
}

func TestManager_LoadSave(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manager and add repo
	m := NewManager(tmpDir)
	_ = m.Add("test", "https://example.com/repo")

	// Save
	err := m.Save()
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filepath.Join(tmpDir, "repos.json"))
	assert.NoError(t, err)

	// Create new manager and load
	m2 := NewManager(tmpDir)
	err = m2.Load()
	require.NoError(t, err)

	// Verify repo loaded
	repo, ok := m2.Get("test")
	assert.True(t, ok)
	assert.Equal(t, "https://example.com/repo", repo.URL)
}

func TestManager_LoadDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	// Load without file creates defaults
	m := NewManager(tmpDir)
	err := m.Load()
	require.NoError(t, err)

	// Should have default repo
	repo, ok := m.Get("official")
	assert.True(t, ok)
	assert.Contains(t, repo.URL, "commands")
}

func TestManager_FetchManifest(t *testing.T) {
	// Create test server
	manifest := `
name: test-repo
version: "1.0.0"
description: Test repository
commands:
  - name: test-cmd
    description: Test command
    file: commands/test.yaml
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/scmd-repo.yaml" {
			w.Write([]byte(manifest))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	m := NewManager("/tmp/test")
	repo := &Repository{
		Name:    "test",
		URL:     server.URL,
		Enabled: true,
	}

	ctx := context.Background()
	result, err := m.FetchManifest(ctx, repo)
	require.NoError(t, err)

	assert.Equal(t, "test-repo", result.Name)
	assert.Equal(t, "1.0.0", result.Version)
	assert.Len(t, result.Commands, 1)
	assert.Equal(t, "test-cmd", result.Commands[0].Name)
}

func TestManager_FetchCommand(t *testing.T) {
	// Create test server
	cmdSpec := `name: git-commit
version: "1.0.0"
description: Generate commit messages
usage: "git-commit [files]"
args:
  - name: files
    description: Files to commit
    required: false
prompt:
  system: "You are a git expert."
  template: "Generate a commit message for these changes"
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/commands/git-commit.yaml" {
			w.Write([]byte(cmdSpec))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	m := NewManager("/tmp/test")
	repo := &Repository{
		Name:    "test",
		URL:     server.URL,
		Enabled: true,
	}

	ctx := context.Background()
	result, err := m.FetchCommand(ctx, repo, "commands/git-commit.yaml")
	require.NoError(t, err)

	assert.Equal(t, "git-commit", result.Name)
	assert.Equal(t, "1.0.0", result.Version)
	assert.Contains(t, result.Prompt.Template, "Generate")
}

func TestManager_SearchCommands(t *testing.T) {
	// Allow localhost URLs for test server
	t.Setenv("SCMD_ALLOW_LOCALHOST", "1")

	// Create test server
	manifest := `
name: test-repo
version: "1.0.0"
description: Test repository
commands:
  - name: git-commit
    description: Generate git commit messages
    category: git
    file: commands/git-commit.yaml
  - name: docker-compose
    description: Generate docker-compose files
    category: docker
    file: commands/docker-compose.yaml
  - name: explain-code
    description: Explain code snippets
    category: code
    file: commands/explain.yaml
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/scmd-repo.yaml" {
			w.Write([]byte(manifest))
		} else {
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	m := NewManager("/tmp/test")
	_ = m.Add("test", server.URL)

	ctx := context.Background()

	// Search for git
	results, err := m.SearchCommands(ctx, "git")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "git-commit", results[0].Command.Name)

	// Search for docker
	results, err = m.SearchCommands(ctx, "docker")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "docker-compose", results[0].Command.Name)

	// Search all
	results, err = m.SearchCommands(ctx, "")
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestManager_InstallCommand(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	spec := &CommandSpec{
		Name:        "test-cmd",
		Version:     "1.0.0",
		Description: "Test command",
		Prompt: PromptSpec{
			Template: "Test prompt: {{.input}}",
		},
	}

	err := m.InstallCommand(spec, tmpDir)
	require.NoError(t, err)

	// Verify file created
	_, err = os.Stat(filepath.Join(tmpDir, "test-cmd.yaml"))
	assert.NoError(t, err)
}

func TestManager_LoadInstalledCommands(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Install some commands
	for _, name := range []string{"cmd1", "cmd2", "cmd3"} {
		spec := &CommandSpec{
			Name:        name,
			Version:     "1.0.0",
			Description: "Test " + name,
			Prompt:      PromptSpec{Template: "Test"},
		}
		_ = m.InstallCommand(spec, tmpDir)
	}

	// Load commands
	commands, err := m.LoadInstalledCommands(tmpDir)
	require.NoError(t, err)
	assert.Len(t, commands, 3)
}

func TestManager_UninstallCommand(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	// Install command
	spec := &CommandSpec{
		Name:    "to-remove",
		Version: "1.0.0",
		Prompt:  PromptSpec{Template: "Test"},
	}
	_ = m.InstallCommand(spec, tmpDir)

	// Uninstall
	err := m.UninstallCommand("to-remove", tmpDir)
	assert.NoError(t, err)

	// Verify removed
	_, err = os.Stat(filepath.Join(tmpDir, "to-remove.yaml"))
	assert.True(t, os.IsNotExist(err))

	// Uninstall non-existent
	err = m.UninstallCommand("nonexistent", tmpDir)
	assert.Error(t, err)
}

func TestContainsLower(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"Hello World", "hello", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "xyz", false},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tt := range tests {
		got := contains(tt.s, tt.substr)
		assert.Equal(t, tt.want, got, "contains(%q, %q)", tt.s, tt.substr)
	}
}

func TestMatchesQuery(t *testing.T) {
	cmd := Command{
		Name:        "git-commit",
		Description: "Generate commit messages",
		Category:    "git",
	}

	assert.True(t, matchesQuery(cmd, ""))
	assert.True(t, matchesQuery(cmd, "git"))
	assert.True(t, matchesQuery(cmd, "commit"))
	assert.True(t, matchesQuery(cmd, "Generate"))
	assert.False(t, matchesQuery(cmd, "docker"))
}

// TestManager_FetchManifestLegacyFormat tests handling of legacy manifest format with path field
func TestManager_FetchManifestLegacyFormat(t *testing.T) {
	// Allow localhost URLs for test server
	t.Setenv("SCMD_ALLOW_LOCALHOST", "1")

	// Create test server with legacy manifest format (path instead of name/description)
	legacyManifest := `
name: legacy-repo
version: "1.0.0"
description: Legacy format repository
commands:
  - path: commands/git/commit.yaml
  - path: commands/docker/compose.yaml
`

	// Command specs for fetching
	commitSpec := `name: commit
version: "1.0.0"
description: Generate commit messages
category: git
prompt:
  template: "Generate commit"
`

	composeSpec := `name: compose
version: "1.0.0"
description: Generate docker-compose files
category: docker
prompt:
  template: "Generate compose"
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/scmd-repo.yaml":
			w.Write([]byte(legacyManifest))
		case "/commands/git/commit.yaml":
			w.Write([]byte(commitSpec))
		case "/commands/docker/compose.yaml":
			w.Write([]byte(composeSpec))
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	m := NewManager("/tmp/test")
	repo := &Repository{
		Name:    "legacy",
		URL:     server.URL,
		Enabled: true,
	}

	ctx := context.Background()
	manifest, err := m.FetchManifest(ctx, repo)
	require.NoError(t, err)

	// Verify manifest was fetched
	assert.Equal(t, "legacy-repo", manifest.Name)
	assert.Len(t, manifest.Commands, 2)

	// Verify commands were normalized with metadata from command files
	assert.Equal(t, "commit", manifest.Commands[0].Name)
	assert.Equal(t, "Generate commit messages", manifest.Commands[0].Description)
	assert.Equal(t, "git", manifest.Commands[0].Category)
	assert.Equal(t, "commands/git/commit.yaml", manifest.Commands[0].File)

	assert.Equal(t, "compose", manifest.Commands[1].Name)
	assert.Equal(t, "Generate docker-compose files", manifest.Commands[1].Description)
	assert.Equal(t, "docker", manifest.Commands[1].Category)
	assert.Equal(t, "commands/docker/compose.yaml", manifest.Commands[1].File)
}
