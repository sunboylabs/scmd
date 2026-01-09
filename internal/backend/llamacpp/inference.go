package llamacpp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/scmd/scmd/internal/backend"
)

// Server wraps llama-server for inference
type Server struct {
	cmd         *exec.Cmd
	port        int
	modelPath   string
	contextSize int
	gpuLayers   int
	ready       bool
	mu          sync.Mutex
	logFile     *os.File
}

var (
	globalServer *Server
	serverMu     sync.Mutex
)

// ServerConfig holds server configuration
type ServerConfig struct {
	ModelPath   string
	Port        int
	ContextSize int
	GPULayers   int
}

// DefaultServerConfig returns default configuration
// Note: ContextSize should be set by the caller (usually from model metadata)
func DefaultServerConfig(modelPath string) *ServerConfig {
	return &ServerConfig{
		ModelPath:   modelPath,
		Port:        8089,
		ContextSize: 0,  // 0 = will be set by backend from model metadata
		GPULayers:   99, // Auto-detect and use GPU
	}
}

// ServerHealth contains detailed health information about the server
type ServerHealth struct {
	Running        bool
	ContextSize    int
	MatchesExpected bool
	Error          error
}

// IsServerRunning checks if a server is already running on the given port
func IsServerRunning(port int) bool {
	url := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	client := &http.Client{Timeout: 500 * time.Millisecond} // Reduced from 1s
	resp, err := client.Get(url)
	if err == nil {
		resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}
	return false
}

// CheckServerHealth performs a comprehensive health check
// Returns detailed health information including context size validation
func CheckServerHealth(port int, expectedContextSize int) *ServerHealth {
	health := &ServerHealth{
		Running:        false,
		ContextSize:    0,
		MatchesExpected: false,
	}

	// Check if server is responding
	url := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(url)
	if err != nil {
		health.Error = fmt.Errorf("server not responding: %w", err)
		return health
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		health.Error = fmt.Errorf("server unhealthy (status %d)", resp.StatusCode)
		return health
	}

	health.Running = true

	// Try to get context size by sending a test request
	// The error message will tell us the actual n_ctx
	testURL := fmt.Sprintf("http://127.0.0.1:%d/completion", port)
	testReq := map[string]interface{}{
		"prompt":    "test",
		"n_predict": 1,
	}
	jsonBody, _ := json.Marshal(testReq)
	testResp, err := client.Post(testURL, "application/json", bytes.NewReader(jsonBody))
	if err == nil {
		defer testResp.Body.Close()
		// Server responded, it's healthy
		// We can't easily get n_ctx from successful requests, so we'll validate on errors
		health.ContextSize = expectedContextSize
		health.MatchesExpected = true
	}

	return health
}

// Port returns the server port
func (s *Server) Port() int {
	return s.port
}

// StartServer starts a llama-server instance
func StartServer(modelPath string, port int) (*Server, error) {
	return StartServerWithConfig(&ServerConfig{
		ModelPath:   modelPath,
		Port:        port,
		ContextSize: 32768, // Use full 32K context that Qwen models support
		GPULayers:   99,
	})
}

// StartServerWithConfig starts llama-server with custom configuration
func StartServerWithConfig(config *ServerConfig) (*Server, error) {
	serverMu.Lock()
	defer serverMu.Unlock()

	debug := os.Getenv("SCMD_DEBUG") != ""

	// Check if already running with same model
	if globalServer != nil && globalServer.modelPath == config.ModelPath && globalServer.ready {
		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Using existing server with model %s\n", config.ModelPath)
		}
		return globalServer, nil
	}

	// If running with different model, restart
	if globalServer != nil && globalServer.modelPath != config.ModelPath && globalServer.ready {
		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Model changed from %s to %s, restarting server\n",
				globalServer.modelPath, config.ModelPath)
		}
		globalServer.Stop()
		globalServer = nil
	}

	// Check if a server is already running on this port (external instance)
	if IsServerRunning(config.Port) {
		// Use the existing external server
		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Using existing llama-server on port %d\n", config.Port)
		}
		server := &Server{
			port:      config.Port,
			modelPath: config.ModelPath,
			ready:     true,
		}
		globalServer = server
		return server, nil
	}

	// Auto-tune configuration based on system resources if not explicitly set
	if config.ContextSize == 0 || config.GPULayers == 0 {
		resources, err := DetectSystemResources()
		if err == nil {
			// Get model size
			modelInfo, err := os.Stat(config.ModelPath)
			var modelSize int64
			if err == nil {
				modelSize = modelInfo.Size()
			} else {
				// Estimate based on common model sizes
				modelSize = 2 * 1024 * 1024 * 1024 // Default 2GB estimate
			}

			// Calculate optimal config
			optimalConfig := CalculateOptimalConfig(resources, modelSize)

			// Use optimal values if not set
			if config.ContextSize == 0 {
				config.ContextSize = optimalConfig.ContextSize
			}
			if config.GPULayers == 0 {
				config.GPULayers = optimalConfig.GPULayers
			}

			if debug {
				fmt.Fprintf(os.Stderr, "[DEBUG] Auto-tuned config: context=%d, gpu_layers=%d\n",
					config.ContextSize, config.GPULayers)
			}
		} else if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Could not detect system resources: %v\n", err)
		}
	}

	// Ensure we have sensible defaults
	// Note: ContextSize should be set by backend from model metadata
	// If still 0 here, use a large default (will be limited by model's actual max)
	if config.ContextSize == 0 {
		config.ContextSize = 131072 // Default to 128K (llama-server will cap at model's max)
	}
	if config.GPULayers < 0 {
		config.GPULayers = 99 // Default to full GPU
	}

	// Stop existing server if any
	if globalServer != nil {
		globalServer.Stop()
	}

	// Find llama-server binary
	serverPath, err := findLlamaServer()
	if err != nil {
		return nil, err
	}

	// Create log file
	dataDir := getDataDir()
	logDir := filepath.Join(dataDir, "logs")
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, "llama-server.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not create log file: %v\n", err)
		logFile = nil
	}

	// Check if CPU-only mode is enabled (for more conservative memory settings)
	cpuOnly := os.Getenv("SCMD_CPU_ONLY") != ""

	// Build arguments - use conservative settings for CPU-only mode
	args := []string{
		"-m", config.ModelPath,
		"--port", fmt.Sprintf("%d", config.Port),
		"-c", fmt.Sprintf("%d", config.ContextSize),
		"-ngl", fmt.Sprintf("%d", config.GPULayers),
		"--log-disable", // Disable verbose logging to stdout
	}

	// Add performance optimizations only when NOT in CPU-only mode
	// CPU-only mode uses conservative settings to maximize available context
	if !cpuOnly {
		args = append(args,
			// Performance optimizations
			"--parallel", "8", // Increased parallel requests for better throughput
			"--batch-size", "2048", // Batch size - larger = faster but more memory
			"--ubatch-size", "512", // Physical batch size
			"--mlock", // Lock model in memory (prevent swapping)

			// Threading optimizations
			"-t", fmt.Sprintf("%d", runtime.NumCPU()), // Use all CPU threads for prompt processing
			"--threads-batch", fmt.Sprintf("%d", runtime.NumCPU()), // Use all threads for batch processing

			// Cache optimizations
			"--cache-type-k", "f16", // Use f16 for key cache (faster)
			"--cache-type-v", "f16", // Use f16 for value cache (faster)

			// Flash attention for faster attention computation (auto mode)
			"--flash-attn", "on",

			// Continuous batching for better latency
			"--cont-batching",

			// Memory optimizations - keep model in RAM for faster access
			"--no-mmap", // Don't use memory mapping, load directly to RAM
		)
	} else {
		// CPU-only mode: conservative settings to maximize context size
		args = append(args,
			"--parallel", "4", // Fewer parallel requests to save memory
			"--batch-size", "512", // Smaller batch size
			"--ubatch-size", "256", // Smaller physical batch
			// NO --mlock (allow swapping if needed)
			// NO --no-mmap (use mmap to save RAM)
			"-t", fmt.Sprintf("%d", runtime.NumCPU()),
			"--threads-batch", fmt.Sprintf("%d", runtime.NumCPU()),
			"--cont-batching",
		)
	}

	// Apply CPU-only mode settings if enabled
	if cpuOnly {
		// Override GPU layers to force CPU-only mode
		for i := 0; i < len(args); i++ {
			if args[i] == "-ngl" && i+1 < len(args) {
				args[i+1] = "0"
				break
			}
		}

		// Remove GPU-dependent optimizations that can cause Metal initialization
		// These flags may trigger Metal even with -ngl 0
		filteredArgs := make([]string, 0, len(args))
		skipNext := false
		for i := 0; i < len(args); i++ {
			if skipNext {
				skipNext = false
				continue
			}

			// Remove flash attention (uses GPU/Metal)
			if args[i] == "--flash-attn" {
				skipNext = true
				continue
			}

			// Remove f16 cache types (may trigger Metal memory allocation)
			if args[i] == "--cache-type-k" || args[i] == "--cache-type-v" {
				skipNext = true
				continue
			}

			// Keep the arg
			filteredArgs = append(filteredArgs, args[i])
		}
		args = filteredArgs

		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Running in CPU-only mode (SCMD_CPU_ONLY set)\n")
			fmt.Fprintf(os.Stderr, "[DEBUG] Removed GPU-dependent optimizations\n")
		}
	}

	cmd := exec.Command(serverPath, args...)

	// Set environment variable to completely disable Metal (Apple GPU framework)
	// This prevents llama.cpp from initializing Metal even for non-model operations
	if cpuOnly {
		cmd.Env = append(os.Environ(), "GGML_METAL_DISABLE=1")
		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Set GGML_METAL_DISABLE=1 to prevent Metal initialization\n")
		}
	}
	if logFile != nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Starting llama-server: %s %v\n", serverPath, args)
	}

	if err := cmd.Start(); err != nil {
		if logFile != nil {
			logFile.Close()
		}
		return nil, fmt.Errorf("start server: %w", err)
	}

	server := &Server{
		cmd:         cmd,
		port:        config.Port,
		modelPath:   config.ModelPath,
		contextSize: config.ContextSize,
		gpuLayers:   config.GPULayers,
		logFile:     logFile,
	}

	// Wait for server to be ready (reduced timeout for faster failure feedback)
	if err := server.waitReady(10 * time.Second); err != nil {
		cmd.Process.Kill()
		if logFile != nil {
			logFile.Close()
		}
		return nil, err
	}

	server.ready = true
	globalServer = server

	// Write PID file for management
	pidPath := filepath.Join(dataDir, "llama-server.pid")
	os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644)

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] llama-server started successfully (PID: %d)\n", cmd.Process.Pid)
	}

	return server, nil
}

// getDataDir returns the scmd data directory
func getDataDir() string {
	if dir := os.Getenv("SCMD_DATA_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".scmd")
}

// waitReady waits for the server to be ready
func (s *Server) waitReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	url := fmt.Sprintf("http://127.0.0.1:%d/health", s.port)
	client := &http.Client{Timeout: 500 * time.Millisecond}

	// Exponential backoff for faster initial checks
	sleepDuration := 50 * time.Millisecond
	maxSleep := 500 * time.Millisecond

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(sleepDuration)

		// Increase sleep duration exponentially
		sleepDuration = sleepDuration * 2
		if sleepDuration > maxSleep {
			sleepDuration = maxSleep
		}
	}

	return fmt.Errorf("server not ready after %v", timeout)
}

// Stop stops the server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd != nil && s.cmd.Process != nil {
		// Try graceful shutdown first
		s.cmd.Process.Signal(os.Interrupt)

		// Wait up to 5 seconds for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- s.cmd.Wait()
		}()

		select {
		case <-done:
			// Graceful shutdown succeeded
		case <-time.After(5 * time.Second):
			// Force kill if not stopped gracefully
			s.cmd.Process.Kill()
			s.cmd.Wait()
		}

		// Clean up PID file
		dataDir := getDataDir()
		pidPath := filepath.Join(dataDir, "llama-server.pid")
		os.Remove(pidPath)
	}

	if s.logFile != nil {
		s.logFile.Close()
		s.logFile = nil
	}

	s.ready = false
	return nil
}

// Complete sends a completion request to the server
func (s *Server) Complete(ctx context.Context, prompt string, req *backend.CompletionRequest) (string, error) {
	debug := os.Getenv("SCMD_DEBUG") != ""
	url := fmt.Sprintf("http://127.0.0.1:%d/completion", s.port)

	// Build request body
	reqBody := map[string]interface{}{
		"prompt":      prompt,
		"n_predict":   req.MaxTokens,
		"temperature": req.Temperature,
		"stop":        []string{"<|im_end|>", "<|endoftext|>"},
		"stream":      false,
	}

	if req.MaxTokens == 0 {
		reqBody["n_predict"] = 2048
	}
	if req.Temperature == 0 {
		reqBody["temperature"] = 0.7
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Sending request to %s\n", url)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Use client with reasonable timeout (model inference can take time)
	// CPU-only mode is much slower, so use longer timeout
	timeout := 2 * time.Minute
	if os.Getenv("SCMD_CPU_ONLY") != "" {
		timeout = 10 * time.Minute // CPU inference can be 10-30x slower
		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Using extended timeout (%v) for CPU-only mode\n", timeout)
		}
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Response status: %d\n", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] Response body length: %d bytes\n", len(respBody))
		if len(respBody) < 1000 {
			fmt.Fprintf(os.Stderr, "[DEBUG] Response body: %s\n", string(respBody))
		}
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response - llama-server returns {"content": "...", ...}
	var result struct {
		Content string `json:"content"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w\nRaw: %s", err, string(respBody))
	}

	content := strings.TrimSpace(result.Content)
	if content == "" {
		// Check if there was an error in the response
		var errResult struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResult) == nil && errResult.Error != "" {
			return "", fmt.Errorf("llama-server: %s", errResult.Error)
		}
		return "", fmt.Errorf("empty response from model.\nPrompt was: %s...\nResponse: %s", truncate(prompt, 100), string(respBody))
	}

	return content, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// findLlamaServer finds the llama-server binary
func findLlamaServer() (string, error) {
	// Check common locations
	candidates := []string{
		"llama-server",
		"llama.cpp/build/bin/llama-server",
		"/usr/local/bin/llama-server",
		"/usr/lib/scmd/llama-server", // Linux package installation
		"/opt/llama.cpp/llama-server",
	}

	// Add platform-specific paths
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		candidates = append(candidates,
			filepath.Join(homeDir, ".local", "bin", "llama-server"),
			filepath.Join(homeDir, "llama.cpp", "build", "bin", "llama-server"),
		)
	}

	// Check bundled binary (highest priority)
	execPath, _ := os.Executable()
	if execPath != "" {
		execDir := filepath.Dir(execPath)

		// 1. Same directory as executable (direct bundle)
		bundledPath := filepath.Join(execDir, "llama-server")
		if runtime.GOOS == "windows" {
			bundledPath += ".exe"
		}
		candidates = append([]string{bundledPath}, candidates...)

		// 2. Homebrew libexec location (../libexec/llama-server)
		homebrewPath := filepath.Join(execDir, "..", "libexec", "llama-server")
		candidates = append([]string{homebrewPath}, candidates...)

		// 3. Archive bin/ subdirectory (bin/llama-server when extracted)
		binPath := filepath.Join(execDir, "bin", "llama-server")
		if runtime.GOOS == "windows" {
			binPath += ".exe"
		}
		candidates = append([]string{binPath}, candidates...)
	}

	for _, path := range candidates {
		if _, err := exec.LookPath(path); err == nil {
			return path, nil
		}
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("llama-server not found. Install with: make install-llamacpp")
}

// runServerInference uses llama-server for inference
func (b *Backend) runServerInference(ctx context.Context, prompt string, req *backend.CompletionRequest) (string, error) {
	debug := os.Getenv("SCMD_DEBUG") != ""

	// Use existing server URL
	url := b.serverURL + "/completion"

	reqBody := map[string]interface{}{
		"prompt":      prompt,
		"n_predict":   req.MaxTokens,
		"temperature": req.Temperature,
		"stop":        []string{"<|im_end|>", "<|endoftext|>"},
		"stream":      false,
	}

	if req.MaxTokens == 0 {
		reqBody["n_predict"] = 2048
	}
	if req.Temperature == 0 {
		reqBody["temperature"] = 0.7
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", ParseError(err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", ParseError(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Use client with reasonable timeout (model inference can take time)
	// CPU-only mode is much slower, so use longer timeout
	timeout := 2 * time.Minute
	if os.Getenv("SCMD_CPU_ONLY") != "" {
		timeout = 10 * time.Minute // CPU inference can be 10-30x slower
		if debug {
			fmt.Fprintf(os.Stderr, "[DEBUG] Using extended timeout (%v) for CPU-only mode\n", timeout)
		}
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", ParseError(err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", ParseError(fmt.Errorf("read response: %w", err))
	}

	if resp.StatusCode != http.StatusOK {
		return "", ParseError(fmt.Errorf("server error (HTTP %d): %s", resp.StatusCode, string(respBody)))
	}

	var result struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", ParseError(fmt.Errorf("parse response: %w", err))
	}

	content := strings.TrimSpace(result.Content)
	if content == "" {
		return "", ParseError(fmt.Errorf("empty response from server"))
	}

	return content, nil
}

// runCGOInference uses CGO bindings for direct inference
// This requires the go-llama.cpp library to be properly linked
func (b *Backend) runCGOInference(ctx context.Context, prompt string, req *backend.CompletionRequest) (string, error) {
	// Start a server if not running
	server, err := StartServer(b.modelPath, 8089)
	if err != nil {
		return "", ParseError(err)
	}

	result, err := server.Complete(ctx, prompt, req)
	if err != nil {
		return "", ParseError(err)
	}

	return result, nil
}

// SetServerURL sets the URL of an external llama-server
func (b *Backend) SetServerURL(url string) {
	b.serverURL = url
}

// StopServer stops the global inference server
func StopServer() {
	serverMu.Lock()
	defer serverMu.Unlock()
	if globalServer != nil {
		globalServer.Stop()
		globalServer = nil
	}
}
