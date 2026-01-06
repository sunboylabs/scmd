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
	cmd       *exec.Cmd
	port      int
	modelPath string
	ready     bool
	mu        sync.Mutex
}

var (
	globalServer *Server
	serverMu     sync.Mutex
)

// StartServer starts a llama-server instance
func StartServer(modelPath string, port int) (*Server, error) {
	serverMu.Lock()
	defer serverMu.Unlock()

	// Check if already running with same model
	if globalServer != nil && globalServer.modelPath == modelPath && globalServer.ready {
		return globalServer, nil
	}

	// Stop existing server
	if globalServer != nil {
		globalServer.Stop()
	}

	// Find llama-server binary
	serverPath, err := findLlamaServer()
	if err != nil {
		return nil, err
	}

	// Start server
	args := []string{
		"-m", modelPath,
		"--port", fmt.Sprintf("%d", port),
		"-c", "4096",
		"-ngl", "99", // Use GPU if available
		"--log-disable",
	}

	cmd := exec.Command(serverPath, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start server: %w", err)
	}

	server := &Server{
		cmd:       cmd,
		port:      port,
		modelPath: modelPath,
	}

	// Wait for server to be ready
	if err := server.waitReady(30 * time.Second); err != nil {
		cmd.Process.Kill()
		return nil, err
	}

	server.ready = true
	globalServer = server

	return server, nil
}

// waitReady waits for the server to be ready
func (s *Server) waitReady(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	url := fmt.Sprintf("http://127.0.0.1:%d/health", s.port)

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("server not ready after %v", timeout)
}

// Stop stops the server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
		s.cmd.Wait()
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

	resp, err := http.DefaultClient.Do(httpReq)
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

	// Check bundled binary
	execPath, _ := os.Executable()
	if execPath != "" {
		bundledPath := filepath.Join(filepath.Dir(execPath), "llama-server")
		if runtime.GOOS == "windows" {
			bundledPath += ".exe"
		}
		candidates = append([]string{bundledPath}, candidates...)
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
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse response: %w\nRaw: %s", err, string(respBody))
	}

	content := strings.TrimSpace(result.Content)
	if content == "" {
		return "", fmt.Errorf("empty response from server.\nResponse: %s", string(respBody))
	}

	return content, nil
}

// runCGOInference uses CGO bindings for direct inference
// This requires the go-llama.cpp library to be properly linked
func (b *Backend) runCGOInference(ctx context.Context, prompt string, req *backend.CompletionRequest) (string, error) {
	// Start a server if not running
	server, err := StartServer(b.modelPath, 8089)
	if err != nil {
		return "", fmt.Errorf("start inference server: %w\n\nInstall llama-server: make install-llamacpp", err)
	}

	return server.Complete(ctx, prompt, req)
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
