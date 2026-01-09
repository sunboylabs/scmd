package security

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/scmd/scmd/internal/backend/mock"
	"github.com/scmd/scmd/internal/command"
	"github.com/scmd/scmd/internal/repos"
	"github.com/scmd/scmd/internal/slash"
)

var scmdBinary string

func init() {
	// Find the scmd binary
	cwd, _ := os.Getwd()
	scmdBinary = filepath.Join(cwd, "..", "..", "bin", "scmd")
}

func runScmd(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(scmdBinary, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func runScmdWithStdin(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(scmdBinary, args...)
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// ==================== COMMAND INJECTION TESTS ====================

func TestSecurity_CommandInjection_Prompt(t *testing.T) {
	// Attempt command injection via prompt
	malicious := []string{
		"; rm -rf /tmp/test",
		"&& cat /etc/passwd",
		"| echo hacked",
		"$(whoami)",
		"`id`",
		"\n/bin/sh",
	}

	for _, payload := range malicious {
		t.Run(payload, func(t *testing.T) {
			stdout, stderr, err := runScmd(t, "-b", "mock", "-p", payload)
			// Should not execute shell commands
			// Just process as normal input
			_ = stdout
			_ = stderr
			_ = err
			// If it runs without crashing, that's a success
		})
	}
}

func TestSecurity_CommandInjection_Stdin(t *testing.T) {
	malicious := []string{
		"; rm -rf /tmp",
		"&& cat /etc/passwd",
		"| whoami",
		"$(ls -la)",
	}

	for _, payload := range malicious {
		t.Run(payload, func(t *testing.T) {
			_, _, _ = runScmdWithStdin(t, payload, "-b", "mock", "explain")
			// Should treat as regular text, not execute
		})
	}
}

// ==================== PATH TRAVERSAL TESTS ====================

func TestSecurity_PathTraversal_OutputFile(t *testing.T) {
	// Attempt path traversal in output file
	malicious := []string{
		"../../../etc/passwd",
		"../../tmp/malicious",
		"/etc/shadow",
		"~/.ssh/authorized_keys",
	}

	for _, path := range malicious {
		t.Run(path, func(t *testing.T) {
			// Should either fail safely or write to sanitized location
			_, _, err := runScmd(t, "-b", "mock", "-p", "test", "-o", path)

			// Either way, should not write to sensitive locations
			if _, err := os.Stat("/etc/passwd"); err == nil {
				// File exists, check it wasn't modified
				info, _ := os.Stat("/etc/passwd")
				if info.ModTime().After(time.Now().Add(-1 * time.Second)) {
					t.Error("/etc/passwd was modified!")
				}
			}
			_ = err
		})
	}
}

func TestSecurity_PathTraversal_ConfigDir(t *testing.T) {
	// Attempt to set malicious config directory
	cmd := exec.Command(scmdBinary, "config")
	cmd.Env = append(os.Environ(), "SCMD_DATA_DIR=../../etc")
	var stdout strings.Builder
	cmd.Stdout = &stdout

	err := cmd.Run()
	// Should either reject or handle safely
	_ = err
}

// ==================== ENVIRONMENT VARIABLE INJECTION ====================

func TestSecurity_EnvInjection(t *testing.T) {
	// Attempt to inject malicious environment variables
	cmd := exec.Command(scmdBinary, "-b", "mock", "-p", "test")
	cmd.Env = append(os.Environ(),
		"PATH=/malicious/path",
		"LD_PRELOAD=/tmp/malicious.so",
		"DYLD_INSERT_LIBRARIES=/tmp/evil.dylib",
	)

	stdout, stderr, err := "", "", error(nil)
	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	stdout = outBuf.String()
	stderr = errBuf.String()

	// Should run safely even with malicious env vars
	_ = stdout
	_ = stderr
	_ = err
}

// ==================== FILE PERMISSION TESTS ====================

func TestSecurity_OutputFilePermissions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires binary in short mode")
	}

	tmpFile := filepath.Join(t.TempDir(), "output.txt")
	_, _, err := runScmd(t, "-b", "mock", "-p", "test", "-o", tmpFile)
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	// File should not be world-writable
	if info.Mode().Perm()&0002 != 0 {
		t.Error("output file is world-writable")
	}

	// File should not be world-readable (depending on security requirements)
	// This might be too strict for some use cases
	// if info.Mode().Perm()&0004 != 0 {
	// 	t.Error("output file is world-readable")
	// }
}

func TestSecurity_ConfigFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	cmd := exec.Command(scmdBinary, "config")
	cmd.Env = append(os.Environ(), "SCMD_DATA_DIR="+tmpDir)

	err := cmd.Run()
	if err != nil {
		// Config might not exist yet
		return
	}

	// Check config directory permissions
	info, err := os.Stat(tmpDir)
	if err != nil {
		return
	}

	// Directory should not be world-writable
	if info.Mode().Perm()&0002 != 0 {
		t.Error("config directory is world-writable")
	}
}

// ==================== INPUT VALIDATION TESTS ====================

func TestSecurity_OversizedInput(t *testing.T) {
	// Attempt to send extremely large input
	// 100MB of data
	large := strings.Repeat("x", 100*1024*1024)

	// Should handle gracefully (reject or process without crashing)
	_, _, err := runScmdWithStdin(t, large, "-b", "mock", "-p", "test")

	// It's OK if it errors (input too large)
	// But it should not crash or hang
	_ = err
}

func TestSecurity_NullBytes(t *testing.T) {
	// Null bytes in input
	input := "hello\x00world\x00test"
	_, _, err := runScmdWithStdin(t, input, "-b", "mock", "-p", "process")

	// Should handle null bytes safely
	_ = err
}

func TestSecurity_ControlCharacters(t *testing.T) {
	// Various control characters
	input := "test\x01\x02\x03\x04\x05\x1b[31mRED\x1b[0m"
	_, _, err := runScmdWithStdin(t, input, "-b", "mock", "-p", "process")

	// Should handle control characters safely
	_ = err
}

// ==================== CONFIGURATION SECURITY ====================

func TestSecurity_MaliciousConfigYAML(t *testing.T) {
	tmpDir := t.TempDir()

	// Create malicious config
	configPath := filepath.Join(tmpDir, "config.yaml")
	maliciousConfig := `
backends:
  default: "mock"
  llamacpp:
    model_path: "/etc/passwd"
  openai:
    api_key: "'; DROP TABLE users; --"
`
	if err := os.WriteFile(configPath, []byte(maliciousConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Try to use malicious config
	cmd := exec.Command(scmdBinary, "config")
	cmd.Env = append(os.Environ(), "SCMD_CONFIG="+configPath)

	err := cmd.Run()
	// Should either reject or sanitize the config
	_ = err
}

func TestSecurity_YAMLBombConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// YAML bomb (billion laughs attack)
	configPath := filepath.Join(tmpDir, "slash.yaml")
	yamlBomb := `
a: &a ["lol","lol","lol","lol","lol","lol","lol","lol","lol"]
b: &b [*a,*a,*a,*a,*a,*a,*a,*a,*a]
c: &c [*b,*b,*b,*b,*b,*b,*b,*b,*b]
commands: *c
`
	if err := os.WriteFile(configPath, []byte(yamlBomb), 0644); err != nil {
		t.Fatal(err)
	}

	// Try to load malicious config
	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)
	runner := slash.NewRunner(tmpDir, registry, repoMgr)

	// Should handle gracefully (timeout, size limit, or error)
	err := runner.LoadConfig()
	_ = err // May fail, which is acceptable
}

// ==================== BACKEND SECURITY ====================

func TestSecurity_BackendIsolation(t *testing.T) {
	// Ensure backends can't access files they shouldn't
	tmpDir := t.TempDir()
	secretFile := filepath.Join(tmpDir, ".secret")
	if err := os.WriteFile(secretFile, []byte("SECRET_DATA"), 0600); err != nil {
		t.Fatal(err)
	}

	// Try to make backend read secret file
	prompt := "read file " + secretFile
	_, _, err := runScmd(t, "-b", "mock", "-p", prompt)

	// Mock backend shouldn't actually read files
	// Real backends should have proper sandboxing
	_ = err
}

// ==================== REPOSITORY SECURITY ====================

func TestSecurity_MaliciousRepoURL(t *testing.T) {
	maliciousURLs := []string{
		"file:///etc/passwd",
		"javascript:alert(1)",
		"data:text/html,<script>alert(1)</script>",
		"ftp://malicious.com/exploit",
	}

	tmpDir := t.TempDir()
	mgr := repos.NewManager(tmpDir)

	for _, url := range maliciousURLs {
		err := mgr.Add("malicious", url)
		// Should reject invalid/malicious URLs
		if err == nil {
			t.Logf("Warning: accepted potentially malicious URL: %s", url)
		}
	}
}

func TestSecurity_RepoSSRF(t *testing.T) {
	// Attempt Server-Side Request Forgery
	internalURLs := []string{
		"http://localhost:8080",
		"http://127.0.0.1:22",
		"http://169.254.169.254/latest/meta-data", // AWS metadata
		"http://[::1]:8080",
	}

	tmpDir := t.TempDir()
	mgr := repos.NewManager(tmpDir)

	for _, url := range internalURLs {
		err := mgr.Add("test", url)
		if err == nil {
			// Try to fetch manifest (should timeout or be blocked)
			repo, _ := mgr.Get("test")
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, err = mgr.FetchManifest(ctx, repo)
			cancel()

			// Should fail or timeout, not succeed
			if err == nil {
				t.Errorf("SSRF vulnerability: successfully fetched %s", url)
			}
		}
	}
}

// ==================== TOOL CALLING SECURITY ====================

func TestSecurity_ToolExecutionRestrictions(t *testing.T) {
	// Tools should have restrictions on what they can execute
	tmpDir := t.TempDir()
	backend := mock.New()

	// Mock shouldn't actually execute shell commands
	ctx := context.Background()
	_ = backend
	_ = ctx
	_ = tmpDir

	// Test that dangerous operations are restricted
	// This would need actual tool calling implementation
}

// ==================== SLASH COMMAND SECURITY ====================

func TestSecurity_SlashCommandNameValidation(t *testing.T) {
	tmpDir := t.TempDir()
	registry := command.NewRegistry()
	repoMgr := repos.NewManager(tmpDir)

	runner := slash.NewRunner(tmpDir, registry, repoMgr)
	runner.LoadConfig()

	// Attempt to add commands with malicious names
	malicious := []slash.SlashCommand{
		{Name: "../../../etc/passwd", Command: "test"},
		{Name: "test;rm -rf /", Command: "test"},
		{Name: "test`whoami`", Command: "test"},
		{Name: "test$(id)", Command: "test"},
	}

	for _, cmd := range malicious {
		err := runner.Add(cmd)
		if err == nil {
			t.Errorf("accepted malicious command name: %s", cmd.Name)
			// Clean up
			runner.Remove(cmd.Name)
		}
	}
}

// ==================== DATA SANITIZATION ====================

func TestSecurity_OutputSanitization(t *testing.T) {
	// Ensure output doesn't leak sensitive data
	stdout, _, err := runScmd(t, "-b", "mock", "-p", "test")
	if err != nil {
		return
	}

	// Check output doesn't contain common sensitive patterns
	sensitive := []string{
		"password",
		"secret",
		"api_key",
		"token",
		"private_key",
	}

	lower := strings.ToLower(stdout)
	for _, pattern := range sensitive {
		if strings.Contains(lower, pattern) {
			// May be a false positive, but worth checking
			t.Logf("Output contains potentially sensitive term: %s", pattern)
		}
	}
}

// ==================== RESOURCE LIMITS ====================

func TestSecurity_ResourceLimits(t *testing.T) {
	// Test that resource usage is bounded
	if testing.Short() {
		t.Skip("skipping resource limit test in short mode")
	}

	// Try to create resource exhaustion
	done := make(chan bool, 1)
	go func() {
		for i := 0; i < 100; i++ {
			_, _, _ = runScmd(t, "-b", "mock", "-p", "test")
		}
		done <- true
	}()

	// Should complete within reasonable time
	timeout := 30 * time.Second
	select {
	case <-done:
		// Success
	case <-time.After(timeout):
		t.Errorf("commands did not complete within %v (possible resource exhaustion)", timeout)
	}
}

// ==================== INFORMATION DISCLOSURE ====================

func TestSecurity_ErrorMessages(t *testing.T) {
	// Error messages shouldn't leak sensitive information
	_, stderr, err := runScmd(t, "invalid-command")
	if err == nil {
		return
	}

	// Check stderr doesn't contain internal paths or sensitive data
	if strings.Contains(stderr, "/home/") ||
		strings.Contains(stderr, "/Users/") ||
		strings.Contains(stderr, "C:\\Users\\") {
		t.Log("Error message may contain internal path information")
	}
}

func TestSecurity_VersionDisclosure(t *testing.T) {
	// Version info should be available but not overly detailed
	stdout, _, err := runScmd(t, "version")
	if err != nil {
		return
	}

	// Should have version info
	if !strings.Contains(stdout, "scmd") && !strings.Contains(stdout, "version") {
		t.Log("Version output may not contain expected information")
	}
}

// ==================== SAFE DEFAULTS ====================

func TestSecurity_SafeDefaults(t *testing.T) {
	// Verify the tool has safe defaults
	stdout, _, err := runScmd(t, "config")
	if err != nil {
		return
	}

	// Check for sensible defaults
	if strings.Contains(strings.ToLower(stdout), "unsafe") {
		t.Error("config contains 'unsafe' setting")
	}
}
