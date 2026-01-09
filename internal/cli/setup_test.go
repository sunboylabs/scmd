package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scmd/scmd/internal/backend/llamacpp"
	"github.com/scmd/scmd/internal/config"
	"github.com/spf13/viper"
)

// TestSetupCommand_Basic tests the setup command can be created
func TestSetupCommand_Basic(t *testing.T) {
	cmd := SetupCommand()
	if cmd == nil {
		t.Fatal("setup command should not be nil")
	}

	if cmd.Use != "setup" {
		t.Errorf("expected 'setup', got %s", cmd.Use)
	}

	// Check flags exist
	if cmd.Flags().Lookup("force") == nil {
		t.Error("setup should have --force flag")
	}

	if cmd.Flags().Lookup("quiet") == nil {
		t.Error("setup should have --quiet flag")
	}
}

// TestSelectModel_ValidInput tests model selection with valid input
func TestSelectModel_ValidInput(t *testing.T) {
	// This would require mocking stdin, which we'll handle in integration tests
	// For now, just verify modelPresets are defined correctly
	if len(modelPresets) == 0 {
		t.Fatal("should have model presets defined")
	}

	for i, preset := range modelPresets {
		if preset.Name == "" {
			t.Errorf("preset %d has empty name", i)
		}
		if preset.Model == "" {
			t.Errorf("preset %d has empty model", i)
		}
		if preset.Size == "" {
			t.Errorf("preset %d has empty size", i)
		}
		if preset.Description == "" {
			t.Errorf("preset %d has empty description", i)
		}
		if preset.Speed == "" {
			t.Errorf("preset %d has empty speed", i)
		}
		if preset.Quality == "" {
			t.Errorf("preset %d has empty quality", i)
		}
	}
}

// TestGetModelURL_ValidModels tests getting URLs for valid models
func TestGetModelURL_ValidModels(t *testing.T) {
	testCases := []string{
		"qwen2.5-0.5b",
		"qwen2.5-1.5b",
		"qwen2.5-3b",
		"qwen2.5-7b",
	}

	for _, modelName := range testCases {
		url := getModelURL(modelName)
		if url == "" {
			t.Errorf("model %s should have a URL", modelName)
		}

		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			t.Errorf("model %s has invalid URL: %s", modelName, url)
		}
	}
}

// TestGetModelURL_InvalidModel tests getting URL for invalid model
func TestGetModelURL_InvalidModel(t *testing.T) {
	url := getModelURL("nonexistent-model")
	if url != "" {
		t.Error("invalid model should return empty URL")
	}
}

// TestUpdateConfiguration tests config save functionality
func TestUpdateConfiguration(t *testing.T) {
	// Create temp dir for config
	tmpDir := t.TempDir()

	// Set up test viper instance
	viper.Reset()
	viper.Set("config_dir", tmpDir)

	// Set data dir for testing
	t.Setenv("SCMD_DATA_DIR", tmpDir)

	// Test updating configuration
	err := updateConfiguration("qwen2.5-1.5b")
	if err != nil {
		t.Fatalf("updateConfiguration failed: %v", err)
	}

	// Verify config was set correctly
	if viper.GetString("backends.default") != "llamacpp" {
		t.Errorf("expected backends.default=llamacpp, got %s", viper.GetString("backends.default"))
	}

	if viper.GetString("backends.local.model") != "qwen2.5-1.5b" {
		t.Errorf("expected model=qwen2.5-1.5b, got %s", viper.GetString("backends.local.model"))
	}

	if !viper.GetBool("setup_completed") {
		t.Error("setup_completed should be true")
	}

	if !viper.GetBool("models.auto_download") {
		t.Error("models.auto_download should be true")
	}

	// Verify context_length is set to 0 (unlimited)
	if viper.GetInt("backends.local.context_length") != 0 {
		t.Error("context_length should be 0 for unlimited")
	}
}

// TestIsFirstRun_FreshInstall tests first run detection on fresh install
func TestIsFirstRun_FreshInstall(t *testing.T) {
	// Create isolated temp dir
	tmpDir := t.TempDir()

	// Reset viper and set to use temp dir
	viper.Reset()
	t.Setenv("SCMD_DATA_DIR", tmpDir)

	// Should be first run (no config exists)
	if !IsFirstRun() {
		t.Error("fresh install should be detected as first run")
	}
}

// TestIsFirstRun_WithSetupCompleted tests first run when setup is already done
func TestIsFirstRun_WithSetupCompleted(t *testing.T) {
	tmpDir := t.TempDir()

	viper.Reset()
	t.Setenv("SCMD_DATA_DIR", tmpDir)
	viper.Set("setup_completed", true)

	// Should NOT be first run
	if IsFirstRun() {
		t.Error("should not be first run when setup_completed=true")
	}
}

// TestIsFirstRun_WithModelsDirectory tests first run when models exist
func TestIsFirstRun_WithModelsDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	// Create models directory with a model
	modelsDir := filepath.Join(tmpDir, "models")
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create fake model file
	modelFile := filepath.Join(modelsDir, "test-model.gguf")
	if err := os.WriteFile(modelFile, []byte("fake model"), 0644); err != nil {
		t.Fatal(err)
	}

	viper.Reset()
	t.Setenv("SCMD_DATA_DIR", tmpDir)

	// Should NOT be first run (models exist)
	if IsFirstRun() {
		t.Error("should not be first run when models directory has models")
	}
}

// TestEnsureModel_AlreadyExists tests ensuring model when it already exists
func TestEnsureModel_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("SCMD_DATA_DIR", tmpDir)

	// Create models directory
	modelsDir := filepath.Join(tmpDir, "models")
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create fake model file
	modelName := "qwen2.5-1.5b"
	modelFile := filepath.Join(modelsDir, modelName+".gguf")
	if err := os.WriteFile(modelFile, []byte("fake model"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should succeed without downloading
	err := ensureModel(modelName)
	if err != nil {
		t.Errorf("ensureModel should succeed when model exists: %v", err)
	}
}

// TestEnsureModel_UnknownModel tests ensuring an unknown model
func TestEnsureModel_UnknownModel(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("SCMD_DATA_DIR", tmpDir)

	// Try to ensure unknown model
	err := ensureModel("totally-unknown-model-xyz")
	if err == nil {
		t.Error("ensureModel should fail with unknown model")
	}

	if !strings.Contains(err.Error(), "unknown model") {
		t.Errorf("error should mention unknown model, got: %v", err)
	}
}

// TestTestBackend_MockAvailable tests the backend test function
// This is marked as an integration test - skip in short mode
func TestTestBackend_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	t.Setenv("SCMD_DATA_DIR", tmpDir)

	// Test will fail if llama-server is not available
	// That's expected - we're testing the REAL flow
	err := testBackend("qwen2.5-1.5b")

	// We accept either success OR a clear error message about llama-server
	if err != nil {
		if !strings.Contains(err.Error(), "llama-server") &&
		   !strings.Contains(err.Error(), "backend check failed") &&
		   !strings.Contains(err.Error(), "model not found") {
			t.Errorf("unexpected error from testBackend: %v", err)
		}
		t.Logf("Backend test failed as expected (llama-server may not be installed): %v", err)
	}
}

// TestModelPresets_ValidConfiguration tests all presets point to valid models
func TestModelPresets_ValidConfiguration(t *testing.T) {
	for _, preset := range modelPresets {
		// Verify the model exists in llamacpp.DefaultModels
		url := getModelURL(preset.Model)
		if url == "" {
			t.Errorf("preset %s references unknown model %s", preset.Name, preset.Model)
		}

		// Verify it's a valid model in the llamacpp package
		found := false
		for _, model := range llamacpp.DefaultModels {
			if model.Name == preset.Model {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("preset %s model %s not found in llamacpp.DefaultModels", preset.Name, preset.Model)
		}
	}
}

// TestModelPresets_DefaultIsBalanced tests that default choice (2) is the balanced option
func TestModelPresets_DefaultIsBalanced(t *testing.T) {
	// Default is index 1 (choice 2)
	defaultPreset := modelPresets[1]

	if !strings.Contains(defaultPreset.Name, "Balanced") {
		t.Errorf("default preset should be Balanced, got: %s", defaultPreset.Name)
	}

	if defaultPreset.Model != "qwen2.5-1.5b" {
		t.Errorf("default preset should use qwen2.5-1.5b, got: %s", defaultPreset.Model)
	}
}

// TestModelPresets_SizeOrdering tests models are ordered by size
func TestModelPresets_SizeOrdering(t *testing.T) {
	expectedOrder := []string{"0.5B", "1.5B", "3B", "7B"}

	for i, expected := range expectedOrder {
		if !strings.Contains(modelPresets[i].Name, expected) {
			t.Errorf("preset %d should contain %s, got: %s", i, expected, modelPresets[i].Name)
		}
	}
}

// TestBackendAvailability tests that we can check if backends are available
func TestBackendAvailability(t *testing.T) {
	tmpDir := t.TempDir()
	backend := llamacpp.New(tmpDir)

	ctx := context.Background()
	available, err := backend.IsAvailable(ctx)

	// Either it's available (llama.cpp installed) or not
	// Both are valid outcomes - we're testing the check works
	if err != nil {
		t.Logf("Backend availability check returned error (expected if llama.cpp not installed): %v", err)
	}

	t.Logf("Backend available: %v", available)
}

// TestSetupConfigDefaults tests that setup sets correct default values
func TestSetupConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	viper.Reset()
	t.Setenv("SCMD_DATA_DIR", tmpDir)

	// Run updateConfiguration
	err := updateConfiguration("qwen2.5-3b")
	if err != nil {
		t.Fatalf("updateConfiguration failed: %v", err)
	}

	// Verify all expected config values
	tests := []struct {
		key      string
		expected interface{}
	}{
		{"backends.default", "llamacpp"},
		{"backends.local.model", "qwen2.5-3b"},
		{"backends.local.context_length", 0},
		{"setup_completed", true},
		{"models.auto_download", true},
	}

	for _, tt := range tests {
		actual := viper.Get(tt.key)
		if actual != tt.expected {
			t.Errorf("config %s: expected %v, got %v", tt.key, tt.expected, actual)
		}
	}
}

// TestSetupCommand_Flags tests setup command has correct flags
func TestSetupCommand_Flags(t *testing.T) {
	cmd := SetupCommand()

	// Test --force flag
	forceFlag := cmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Fatal("setup should have --force flag")
	}
	if forceFlag.DefValue != "false" {
		t.Error("--force should default to false")
	}

	// Test --quiet flag
	quietFlag := cmd.Flags().Lookup("quiet")
	if quietFlag == nil {
		t.Fatal("setup should have --quiet flag")
	}
	if quietFlag.DefValue != "false" {
		t.Error("--quiet should default to false")
	}
}

// TestBackendConfiguration tests that backend configuration is correct
func TestBackendConfiguration(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("SCMD_DATA_DIR", tmpDir)

	// Update config with a model
	err := updateConfiguration("qwen2.5-1.5b")
	if err != nil {
		t.Fatalf("updateConfiguration failed: %v", err)
	}

	// Verify the backend name is "llamacpp" not "local"
	// This was a bug - defaults.go had "local" which doesn't exist
	defaultBackend := viper.GetString("backends.default")
	if defaultBackend != "llamacpp" {
		t.Errorf("default backend should be 'llamacpp', got '%s'", defaultBackend)
	}

	// Verify this matches an actual backend
	backend := llamacpp.New(tmpDir)
	if backend.Name() != defaultBackend {
		t.Errorf("configured backend '%s' doesn't match llamacpp backend name '%s'",
			defaultBackend, backend.Name())
	}
}

// TestConfigPath tests that config path is correctly determined
func TestConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("SCMD_DATA_DIR", tmpDir)

	configPath := config.ConfigPath()

	// Should be in the data directory
	if !strings.Contains(configPath, tmpDir) {
		t.Errorf("config path should be in temp dir, got: %s", configPath)
	}

	// Should end with config.yaml
	if !strings.HasSuffix(configPath, "config.yaml") {
		t.Errorf("config path should end with config.yaml, got: %s", configPath)
	}
}

// TestDataDir tests that data directory is correctly determined
func TestDataDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("SCMD_DATA_DIR", tmpDir)

	dataDir := config.DataDir()

	if dataDir != tmpDir {
		t.Errorf("data dir should be %s, got: %s", tmpDir, dataDir)
	}
}
