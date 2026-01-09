package llamacpp

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/scmd/scmd/internal/backend"
)

// Model represents a downloadable model
type Model struct {
	Name        string `json:"name"`
	Variant     string `json:"variant"` // e.g., "Q4_K_M", "Q8_0"
	URL         string `json:"url"`
	Size        int64  `json:"size"` // bytes
	SHA256      string `json:"sha256"`
	Description string `json:"description"`
	ContextSize int    `json:"context_size"`
	ToolCalling bool   `json:"tool_calling"`
}

// DefaultModels contains pre-configured models
// Using official Qwen GGUF releases from HuggingFace
var DefaultModels = []Model{
	// Qwen2.5 official models (recommended)
	// Note: Size is 0 (unknown) - will be determined from HTTP Content-Length during download
	{
		Name:        "qwen2.5-3b",
		Variant:     "q4_k_m",
		URL:         "https://huggingface.co/Qwen/Qwen2.5-3B-Instruct-GGUF/resolve/main/qwen2.5-3b-instruct-q4_k_m.gguf",
		Size:        0, // Unknown - determined during download
		Description: "Qwen2.5 3B (~2.1GB) - Good balance of speed and quality",
		ContextSize: 32768,
		ToolCalling: true,
	},
	{
		Name:        "qwen2.5-1.5b",
		Variant:     "q4_k_m",
		URL:         "https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct-GGUF/resolve/main/qwen2.5-1.5b-instruct-q4_k_m.gguf",
		Size:        0, // Unknown - determined during download
		Description: "Qwen2.5 1.5B (~1GB) - Fast and lightweight",
		ContextSize: 32768,
		ToolCalling: true,
	},
	{
		Name:        "qwen2.5-0.5b",
		Variant:     "q4_k_m",
		URL:         "https://huggingface.co/Qwen/Qwen2.5-0.5B-Instruct-GGUF/resolve/main/qwen2.5-0.5b-instruct-q4_k_m.gguf",
		Size:        0, // Unknown - determined during download
		Description: "Qwen2.5 0.5B (~500MB) - Smallest, fastest option",
		ContextSize: 32768,
		ToolCalling: true,
	},
	{
		Name:        "qwen2.5-7b",
		Variant:     "q3_k_m",
		URL:         "https://huggingface.co/Qwen/Qwen2.5-7B-Instruct-GGUF/resolve/main/qwen2.5-7b-instruct-q3_k_m.gguf",
		Size:        0, // Unknown - determined during download
		Description: "Qwen2.5 7B (~3.8GB) - Best quality, needs more RAM",
		ContextSize: 32768,
		ToolCalling: true,
	},
	// Qwen3 models from unsloth (alternative)
	{
		Name:        "qwen3-4b",
		Variant:     "Q4_K_M",
		URL:         "https://huggingface.co/unsloth/Qwen3-4B-Instruct-2507-GGUF/resolve/main/Qwen3-4B-Instruct-2507-Q4_K_M.gguf",
		Size:        2644000000, // ~2.6GB
		Description: "Qwen3 4B - Fast, efficient, tool calling support",
		ContextSize: 32768,
		ToolCalling: true,
	},
}

// ModelManager handles model downloading and management
type ModelManager struct {
	modelsDir  string
	httpClient *http.Client
	mu         sync.Mutex
}

// NewModelManager creates a new model manager
func NewModelManager(dataDir string) *ModelManager {
	return &ModelManager{
		modelsDir:  filepath.Join(dataDir, "models"),
		httpClient: &http.Client{},
	}
}

// GetModelPath returns the path to a model, downloading if necessary
func (m *ModelManager) GetModelPath(ctx context.Context, modelName string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find model spec
	var model *Model
	for i := range DefaultModels {
		if DefaultModels[i].Name == modelName {
			model = &DefaultModels[i]
			break
		}
	}

	if model == nil {
		// Check if it's a local path
		if _, err := os.Stat(modelName); err == nil {
			return modelName, nil
		}
		return "", fmt.Errorf("unknown model: %s", modelName)
	}

	// Create models directory
	if err := os.MkdirAll(m.modelsDir, 0755); err != nil {
		return "", err
	}

	// Check if already downloaded
	filename := fmt.Sprintf("%s-%s.gguf", model.Name, model.Variant)
	modelPath := filepath.Join(m.modelsDir, filename)

	if _, err := os.Stat(modelPath); err == nil {
		return modelPath, nil
	}

	// In test mode, don't auto-download - return error instead
	if os.Getenv("SCMD_TEST_MODE") == "1" {
		return "", fmt.Errorf("model %s not found (auto-download disabled in test mode)", modelName)
	}

	// Download the model
	if model.Size > 0 {
		fmt.Printf("Downloading %s (%s)...\n", model.Name, formatBytes(model.Size))
	} else {
		fmt.Printf("Downloading %s...\n", model.Name)
	}
	if err := m.downloadModel(ctx, model, modelPath); err != nil {
		return "", err
	}

	return modelPath, nil
}

// downloadModel downloads a model from URL with enhanced retry and resume support
func (m *ModelManager) downloadModel(ctx context.Context, model *Model, destPath string) error {
	// Create enhanced downloader
	downloader := NewEnhancedDownloader(DefaultDownloadConfig())

	// Simple progress display
	var lastPercent int
	progressCallback := func(current, total int64) {
		if total <= 0 {
			return
		}
		percent := int(float64(current) * 100 / float64(total))
		if percent != lastPercent {
			lastPercent = percent
			fmt.Printf("\r  Progress: %d%% (%s / %s)", percent,
				formatBytes(current), formatBytes(total))
		}
	}

	// Download with all enhancements
	if err := downloader.DownloadWithProgress(ctx, model.URL, destPath, model.Size, progressCallback); err != nil {
		return err
	}

	fmt.Println() // New line after progress

	// Verify checksum if provided
	if model.SHA256 != "" {
		fmt.Printf("  Verifying download...\n")
		if err := verifyChecksum(destPath, model.SHA256); err != nil {
			os.Remove(destPath)
			return &DownloadError{
				Stage:   "verification",
				Err:     err,
				Message: "Downloaded file checksum doesn't match expected value.",
				Help: []string{
					"The file may be corrupted during download",
					"Try downloading again",
					"Check your internet connection",
				},
			}
		}
	}

	fmt.Printf("  ✓ Downloaded: %s\n", destPath)
	return nil
}

// verifyChecksum verifies the SHA256 checksum of a file
func verifyChecksum(path string, expectedHash string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return err
	}

	gotHash := hex.EncodeToString(hash.Sum(nil))
	if gotHash != expectedHash {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, gotHash)
	}

	return nil
}

// ListModels returns available models
func (m *ModelManager) ListModels() []Model {
	return DefaultModels
}

// ListDownloaded returns downloaded models
func (m *ModelManager) ListDownloaded() ([]string, error) {
	entries, err := os.ReadDir(m.modelsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var models []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".gguf") {
			models = append(models, e.Name())
		}
	}
	return models, nil
}

// DeleteModel deletes a downloaded model
func (m *ModelManager) DeleteModel(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Find model
	for _, model := range DefaultModels {
		if model.Name == name {
			filename := fmt.Sprintf("%s-%s.gguf", model.Name, model.Variant)
			modelPath := filepath.Join(m.modelsDir, filename)
			return os.Remove(modelPath)
		}
	}

	// Try as filename
	modelPath := filepath.Join(m.modelsDir, name)
	if _, err := os.Stat(modelPath); err == nil {
		return os.Remove(modelPath)
	}

	return fmt.Errorf("model not found: %s", name)
}

// GetDefaultModel returns the recommended default model
func GetDefaultModel() string {
	return "qwen2.5-1.5b" // Use qwen2.5-1.5b as default - fast, lightweight, and efficient
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// Backend implements the LLM backend using llama.cpp
type Backend struct {
	modelManager    *ModelManager
	modelName       string
	modelPath       string
	contextSize     int // 0 = use model's native context size
	contextSizeSet  bool
	initialized     bool
	mu              sync.Mutex

	// These will be set when CGO binding is available
	// For now, we use HTTP API to llama-server as fallback
	serverURL string
}

// New creates a new llama.cpp backend
func New(dataDir string) *Backend {
	return &Backend{
		modelManager:   NewModelManager(dataDir),
		modelName:      GetDefaultModel(),
		contextSize:    0, // 0 = use model's native context size (no limits)
		contextSizeSet: false,
	}
}

// Name returns the backend name
func (b *Backend) Name() string {
	return "llamacpp"
}

// Type returns the backend type
func (b *Backend) Type() backend.Type {
	return backend.TypeLocal
}

// SetContextSize sets a custom context size limit (0 = use model's native size)
func (b *Backend) SetContextSize(size int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.contextSize = size
	b.contextSizeSet = true
	b.initialized = false // Force re-initialization with new context size
}

// GetContextSize returns the effective context size
func (b *Backend) GetContextSize() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	// If explicitly set, use that
	if b.contextSizeSet && b.contextSize > 0 {
		return b.contextSize
	}

	// Otherwise, use model's native context size
	for _, m := range DefaultModels {
		if m.Name == b.modelName {
			return m.ContextSize
		}
	}

	// Fallback: use 32K as default for unknown models
	return 32768
}

// Initialize initializes the backend
func (b *Backend) Initialize(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.initialized {
		return nil
	}

	debug := os.Getenv("SCMD_DEBUG") != ""

	// Download model if needed
	modelPath, err := b.modelManager.GetModelPath(ctx, b.modelName)
	if err != nil {
		return fmt.Errorf("get model: %w", err)
	}
	b.modelPath = modelPath

	// Set context size from model metadata if not explicitly set
	if !b.contextSizeSet || b.contextSize == 0 {
		for _, m := range DefaultModels {
			if m.Name == b.modelName {
				b.contextSize = m.ContextSize
				if debug {
					fmt.Fprintf(os.Stderr, "[DEBUG] Using model's native context size: %d\n", b.contextSize)
				}
				break
			}
		}
		// Fallback
		if b.contextSize == 0 {
			b.contextSize = 32768
		}
	} else if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Using custom context size: %d\n", b.contextSize)
	}

	// Auto-start llama-server if not already running
	// This is the key fix from the evaluation feedback
	if !IsServerRunning(8089) {
		if os.Getenv("SCMD_NO_AUTOSTART") == "" {
			// Show helpful startup message (unless in quiet mode)
			quiet := os.Getenv("SCMD_QUIET") != ""
			if !quiet {
				fmt.Fprintln(os.Stderr, "⏳ Starting llama-server...")
			}

			if debug {
				fmt.Fprintf(os.Stderr, "[DEBUG] Auto-starting llama-server...\n")
			}

			config := DefaultServerConfig(modelPath)

			// Use backend's context size (respects user override or model's native size)
			config.ContextSize = b.contextSize

			// Detect resources and show performance info
			if !quiet {
				resources, err := DetectSystemResources()
				if err == nil {
					modelInfo, _ := os.Stat(modelPath)
					var modelSize int64
					if modelInfo != nil {
						modelSize = modelInfo.Size()
					}
					optimalConfig := CalculateOptimalConfig(resources, modelSize)
					// Don't override context size from auto-tuning - use backend's explicit value
					// config.ContextSize = optimalConfig.ContextSize  // REMOVED
					config.GPULayers = optimalConfig.GPULayers

					// Show performance mode
					if config.GPULayers == 0 {
						fmt.Fprintln(os.Stderr, "⚠️  Running in CPU mode (slower, but uses less memory)")
						fmt.Fprintln(os.Stderr, "   Expect 30-60 seconds per query")
						fmt.Fprintln(os.Stderr, "   Tip: Use cloud backend for faster results: scmd -b openai")
					} else if resources.HasGPU {
						fmt.Fprintf(os.Stderr, "✅ GPU acceleration enabled (%s)\n", resources.GPUType)
						fmt.Fprintln(os.Stderr, "   Expected response time: 5-10 seconds (optimized)")
						fmt.Fprintln(os.Stderr, "   First query may be slower due to model warmup")
					}
				}
			}

			_, err := StartServerWithConfig(config)
			if err != nil {
				return fmt.Errorf("auto-start llama-server: %w\n\nTip: Install llama-server with: brew install llama.cpp", err)
			}

			if !quiet {
				fmt.Fprintln(os.Stderr, "✅ llama-server ready")
				fmt.Fprintln(os.Stderr, "")
			}

			if debug {
				fmt.Fprintf(os.Stderr, "[DEBUG] llama-server is ready\n")
			}
		}
	}

	b.initialized = true
	return nil
}

// IsAvailable checks if the backend is available
func (b *Backend) IsAvailable(ctx context.Context) (bool, error) {
	// Check if llama-server binary exists
	_, err := findLlamaServer()
	if err != nil {
		// llama-server not found, backend not available
		return false, nil
	}

	// Check if we can access the model (or download it)
	// This doesn't download yet, just checks if we have the ability to
	return true, nil
}

// SetModel sets the model to use
func (b *Backend) SetModel(model string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.modelName = model
	b.initialized = false // Force re-initialization
	return nil
}

// ListModels returns available models
func (b *Backend) ListModels(ctx context.Context) ([]string, error) {
	models := b.modelManager.ListModels()
	names := make([]string, len(models))
	for i, m := range models {
		names[i] = m.Name
	}
	return names, nil
}

// Complete generates a completion
func (b *Backend) Complete(ctx context.Context, req *backend.CompletionRequest) (*backend.CompletionResponse, error) {
	debug := os.Getenv("SCMD_DEBUG") != ""

	if err := b.Initialize(ctx); err != nil {
		return nil, err
	}

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Model path: %s\n", b.modelPath)
	}

	// Build prompt with system message
	prompt := b.buildPrompt(req)

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Prompt length: %d chars\n", len(prompt))
		fmt.Fprintf(os.Stderr, "[DEBUG] Prompt preview: %s...\n", truncateStr(prompt, 200))
	}

	// Use inference engine
	response, err := b.runInference(ctx, prompt, req)
	if err != nil {
		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Inference error: %v\n", err)
		}
		return nil, err
	}

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Response length: %d chars\n", len(response))
		fmt.Fprintf(os.Stderr, "[DEBUG] Response: %s\n", truncateStr(response, 500))
	}

	return &backend.CompletionResponse{
		Content:    response,
		TokensUsed: 0, // TODO: track tokens
	}, nil
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// Stream generates a streaming completion
func (b *Backend) Stream(ctx context.Context, req *backend.CompletionRequest) (<-chan backend.StreamChunk, error) {
	if err := b.Initialize(ctx); err != nil {
		return nil, err
	}

	ch := make(chan backend.StreamChunk, 100)

	go func() {
		defer close(ch)

		// Build prompt
		prompt := b.buildPrompt(req)

		// For now, non-streaming fallback
		response, err := b.runInference(ctx, prompt, req)
		if err != nil {
			ch <- backend.StreamChunk{Error: err}
			return
		}

		// Send as single chunk
		ch <- backend.StreamChunk{Content: response, Done: true}
	}()

	return ch, nil
}

// buildPrompt constructs the prompt from request
func (b *Backend) buildPrompt(req *backend.CompletionRequest) string {
	var sb strings.Builder

	// Qwen chat format
	if req.SystemPrompt != "" {
		sb.WriteString("<|im_start|>system\n")
		sb.WriteString(req.SystemPrompt)
		sb.WriteString("<|im_end|>\n")
	}

	sb.WriteString("<|im_start|>user\n")
	sb.WriteString(req.Prompt)
	sb.WriteString("<|im_end|>\n")
	sb.WriteString("<|im_start|>assistant\n")

	return sb.String()
}

// runInference runs the actual inference
// This is a placeholder - actual implementation depends on CGO bindings
func (b *Backend) runInference(ctx context.Context, prompt string, req *backend.CompletionRequest) (string, error) {
	// Check if we have a local llama-server running
	if b.serverURL != "" {
		return b.runServerInference(ctx, prompt, req)
	}

	// Try to use CGO bindings (when available)
	return b.runCGOInference(ctx, prompt, req)
}

// Shutdown cleans up resources
func (b *Backend) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.initialized = false
	return nil
}

// SupportsToolCalling returns true since Qwen3 models support tool calling
func (b *Backend) SupportsToolCalling() bool {
	// Qwen3 models have native tool calling support
	return true
}

// CompleteWithTools generates a completion that may include tool calls
func (b *Backend) CompleteWithTools(ctx context.Context, req *backend.ToolRequest) (*backend.ToolResponse, error) {
	if err := b.Initialize(ctx); err != nil {
		return nil, err
	}

	// Build prompt with tool definitions in Qwen format
	prompt := b.buildToolPrompt(req)

	response, err := b.runInference(ctx, prompt, &req.CompletionRequest)
	if err != nil {
		return nil, err
	}

	// Parse response for tool calls
	toolCalls := b.parseToolCalls(response)

	return &backend.ToolResponse{
		Content:   response,
		ToolCalls: toolCalls,
	}, nil
}

// buildToolPrompt constructs a prompt with tool definitions
func (b *Backend) buildToolPrompt(req *backend.ToolRequest) string {
	var sb strings.Builder

	// System message with tool definitions
	sb.WriteString("<|im_start|>system\n")
	if req.SystemPrompt != "" {
		sb.WriteString(req.SystemPrompt)
		sb.WriteString("\n\n")
	}

	// Add tool definitions if any
	if len(req.Tools) > 0 {
		sb.WriteString("You have access to the following tools:\n\n")
		for _, tool := range req.Tools {
			sb.WriteString(fmt.Sprintf("### %s\n", tool.Name))
			sb.WriteString(fmt.Sprintf("%s\n", tool.Description))
			if len(tool.Parameters) > 0 {
				sb.WriteString("Parameters:\n")
				for name, param := range tool.Parameters {
					req := ""
					if param.Required {
						req = " (required)"
					}
					sb.WriteString(fmt.Sprintf("- %s (%s)%s: %s\n", name, param.Type, req, param.Description))
				}
			}
			sb.WriteString("\n")
		}
		sb.WriteString("To use a tool, respond with:\n")
		sb.WriteString("<tool_call>{\"name\": \"tool_name\", \"parameters\": {...}}</tool_call>\n")
	}
	sb.WriteString("<|im_end|>\n")

	sb.WriteString("<|im_start|>user\n")
	sb.WriteString(req.Prompt)
	sb.WriteString("<|im_end|>\n")
	sb.WriteString("<|im_start|>assistant\n")

	return sb.String()
}

// parseToolCalls extracts tool calls from response
func (b *Backend) parseToolCalls(response string) []backend.ToolCall {
	var calls []backend.ToolCall

	// Look for <tool_call>...</tool_call> patterns
	start := 0
	for {
		openTag := strings.Index(response[start:], "<tool_call>")
		if openTag == -1 {
			break
		}
		openTag += start + len("<tool_call>")

		closeTag := strings.Index(response[openTag:], "</tool_call>")
		if closeTag == -1 {
			break
		}

		jsonStr := strings.TrimSpace(response[openTag : openTag+closeTag])

		var call struct {
			Name       string                 `json:"name"`
			Parameters map[string]interface{} `json:"parameters"`
		}

		if err := json.Unmarshal([]byte(jsonStr), &call); err == nil {
			calls = append(calls, backend.ToolCall{
				Name:       call.Name,
				Parameters: call.Parameters,
			})
		}

		start = openTag + closeTag + len("</tool_call>")
	}

	return calls
}

// ModelInfo returns information about the current model
func (b *Backend) ModelInfo() *backend.ModelInfo {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Find model in defaults
	for _, m := range DefaultModels {
		if m.Name == b.modelName {
			return &backend.ModelInfo{
				Name:          m.Name,
				Size:          formatBytes(m.Size),
				Quantization:  m.Variant,
				ContextLength: m.ContextSize,
				Capabilities:  []string{"chat", "tool_calling"},
			}
		}
	}

	// Unknown model
	return &backend.ModelInfo{
		Name:          b.modelName,
		ContextLength: b.contextSize,
		Capabilities:  []string{"chat"},
	}
}

// EstimateTokens provides a rough token count estimate
// Uses a simple heuristic: ~4 characters per token for English
func (b *Backend) EstimateTokens(text string) int {
	// Rough estimate: average of 4 chars per token
	return len(text) / 4
}
