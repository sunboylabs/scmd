package repos

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/command"
	"github.com/scmd/scmd/internal/templates"
)

// MockBackend implements backend.Backend for testing
type MockBackend struct {
	response string
	err      error
}

func (m *MockBackend) Name() string {
	return "mock"
}

func (m *MockBackend) Type() backend.Type {
	return backend.TypeMock
}

func (m *MockBackend) Initialize(ctx context.Context) error {
	return nil
}

func (m *MockBackend) IsAvailable(ctx context.Context) (bool, error) {
	return true, nil
}

func (m *MockBackend) Shutdown(ctx context.Context) error {
	return nil
}

func (m *MockBackend) Complete(ctx context.Context, req *backend.CompletionRequest) (*backend.CompletionResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &backend.CompletionResponse{
		Content:    m.response,
		TokensUsed: 100,
	}, nil
}

func (m *MockBackend) Stream(ctx context.Context, req *backend.CompletionRequest) (<-chan backend.StreamChunk, error) {
	ch := make(chan backend.StreamChunk, 1)
	ch <- backend.StreamChunk{Content: m.response, Done: true}
	close(ch)
	return ch, nil
}

func (m *MockBackend) SupportsToolCalling() bool {
	return false
}

func (m *MockBackend) CompleteWithTools(ctx context.Context, req *backend.ToolRequest) (*backend.ToolResponse, error) {
	return nil, nil
}

func (m *MockBackend) EstimateTokens(text string) int {
	return len(text) / 4
}

func (m *MockBackend) ModelInfo() *backend.ModelInfo {
	return &backend.ModelInfo{
		Name:          "mock-model",
		ContextLength: 8192,
	}
}

func TestTemplateRefValidation(t *testing.T) {
	tests := []struct {
		name    string
		ref     *TemplateRef
		wantErr bool
	}{
		{
			name: "valid named template",
			ref: &TemplateRef{
				Name: "security-review",
				Variables: map[string]string{
					"Language": "{{.file_extension}}",
					"Code":     "{{.file_content}}",
				},
			},
			wantErr: false,
		},
		{
			name: "valid inline template",
			ref: &TemplateRef{
				Inline: &InlineTemplate{
					SystemPrompt:       "You are a code reviewer",
					UserPromptTemplate: "Review this code: {{.Code}}",
					Variables: []TemplateVariable{
						{Name: "Code", Required: true},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty template ref",
			ref:     &TemplateRef{},
			wantErr: true,
		},
		{
			name: "both name and inline",
			ref: &TemplateRef{
				Name: "test",
				Inline: &InlineTemplate{
					UserPromptTemplate: "test",
				},
			},
			wantErr: true,
		},
		{
			name: "inline without user prompt",
			ref: &TemplateRef{
				Inline: &InlineTemplate{
					SystemPrompt: "test",
				},
			},
			wantErr: true,
		},
	}

	mgr := NewManager("testdata")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mgr.validateTemplateRef(tt.ref)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTemplateRef() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInlineTemplateExecution(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "scmd-template-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create template executor
	executor, err := NewTemplateExecutor(tmpDir)
	if err != nil {
		t.Fatalf("failed to create template executor: %v", err)
	}

	// Create command spec with inline template
	spec := &CommandSpec{
		Name:        "test-review",
		Version:     "1.0.0",
		Description: "Test command with inline template",
		Args: []ArgSpec{
			{Name: "file", Description: "File to review", Required: true},
		},
		Template: &TemplateRef{
			Inline: &InlineTemplate{
				SystemPrompt: "You are a code reviewer",
				UserPromptTemplate: `Review this {{.Language}} code:

{{.Code}}

Focus on:
1. Code quality
2. Best practices`,
				Variables: []TemplateVariable{
					{Name: "Language", Default: "unknown"},
					{Name: "Code", Required: true},
				},
			},
		},
	}

	// Create args with stdin content
	args := &command.Args{
		Positional: []string{"test.go"},
		Options: map[string]string{
			"stdin": "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}",
		},
	}

	// Create execution context
	mockBackend := &MockBackend{
		response: "The code looks good!",
	}

	execCtx := &command.ExecContext{
		Backend: mockBackend,
		DataDir: tmpDir,
	}

	// Execute command
	result, err := executor.ExecuteTemplateCommand(context.Background(), spec, args, execCtx)
	if err != nil {
		t.Fatalf("ExecuteTemplateCommand() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success=true, got success=false with error: %s", result.Error)
	}

	if result.Output != "The code looks good!" {
		t.Errorf("Expected output 'The code looks good!', got '%s'", result.Output)
	}
}

func TestNamedTemplateExecution(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "scmd-template-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create templates directory
	templatesDir := filepath.Join(tmpDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("failed to create templates dir: %v", err)
	}

	// Create a test template
	testTemplate := &templates.Template{
		Name:               "test-template",
		Version:            "1.0",
		Description:        "Test template",
		CompatibleCommands: []string{"analyze", "review"},
		SystemPrompt:       "You are a helpful assistant",
		UserPromptTemplate: `Analyze this {{.Language}} code:

{{.Code}}

{{if .Context}}
Context: {{.Context}}
{{end}}`,
		Variables: []templates.Variable{
			{Name: "Language", Required: true},
			{Name: "Code", Required: true},
			{Name: "Context", Required: false},
		},
	}

	templatePath := filepath.Join(templatesDir, "test-template.yaml")
	if err := testTemplate.Save(templatePath); err != nil {
		t.Fatalf("failed to save template: %v", err)
	}

	// Verify the file exists
	if _, err := os.Stat(templatePath); err != nil {
		t.Fatalf("template file does not exist after save: %v", err)
	}

	// Create template manager with custom directory
	templateManager, err := templates.NewManagerWithDir(templatesDir)
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	// Verify we can load the template directly
	tpl, err := templateManager.Load("test-template")
	if err != nil {
		t.Fatalf("failed to load template after manager creation: %v (looking in: %s)", err, templatesDir)
	}
	if tpl.Name != "test-template" {
		t.Fatalf("loaded template has wrong name: %s", tpl.Name)
	}

	// Create template executor with custom manager
	executor := NewTemplateExecutorWithManager(tmpDir, templateManager)

	// Create command spec with template reference
	spec := &CommandSpec{
		Name:        "test-analyze",
		Version:     "1.0.0",
		Description: "Test command with template reference",
		Args: []ArgSpec{
			{Name: "file", Description: "File to analyze", Required: true},
		},
		Flags: []FlagSpec{
			{Name: "context", Description: "Additional context"},
		},
		Template: &TemplateRef{
			Name: "test-template",
			Variables: map[string]string{
				"Language": "{{.file_extension}}",
				"Code":     "{{.file_content}}",
				"Context":  "{{.context}}",
			},
		},
	}

	// Create args with stdin content
	args := &command.Args{
		Positional: []string{"test.py"},
		Options: map[string]string{
			"stdin":   "def hello():\n    print('Hello, World!')",
			"context": "This is a greeting function",
		},
	}

	// Create execution context
	mockBackend := &MockBackend{
		response: "The Python code defines a simple greeting function.",
	}

	execCtx := &command.ExecContext{
		Backend: mockBackend,
		DataDir: tmpDir,
	}

	// Execute command
	result, err := executor.ExecuteTemplateCommand(context.Background(), spec, args, execCtx)
	if err != nil {
		t.Fatalf("ExecuteTemplateCommand() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success=true, got success=false with error: %s", result.Error)
	}

	if result.Output != "The Python code defines a simple greeting function." {
		t.Errorf("Unexpected output: %s", result.Output)
	}
}

func TestLanguageDetection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scmd-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	executor, err := NewTemplateExecutor(tmpDir)
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	tests := []struct {
		ext      string
		expected string
	}{
		{"go", "Go"},
		{"py", "Python"},
		{"js", "JavaScript"},
		{"ts", "TypeScript"},
		{"rs", "Rust"},
		{"java", "Java"},
		{"cpp", "C++"},
		{"unknown", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := executor.detectLanguage(tt.ext)
			if result != tt.expected {
				t.Errorf("detectLanguage(%s) = %s, want %s", tt.ext, result, tt.expected)
			}
		})
	}
}

func TestDataContextBuilding(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scmd-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	executor, err := NewTemplateExecutor(tmpDir)
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	spec := &CommandSpec{
		Args: []ArgSpec{
			{Name: "file", Description: "Input file"},
			{Name: "output", Description: "Output file", Default: "out.txt"},
		},
		Flags: []FlagSpec{
			{Name: "verbose", Description: "Verbose mode", Default: "false"},
		},
	}

	args := &command.Args{
		Positional: []string{"input.go"},
		Options: map[string]string{
			"stdin":   "package main",
			"verbose": "true",
		},
	}

	data := executor.buildDataContext(spec, args)

	// Check positional args
	if data["file"] != "input.go" {
		t.Errorf("Expected file='input.go', got '%v'", data["file"])
	}

	// Check default value
	if data["output"] != "out.txt" {
		t.Errorf("Expected output='out.txt', got '%v'", data["output"])
	}

	// Check flags
	if data["verbose"] != "true" {
		t.Errorf("Expected verbose='true', got '%v'", data["verbose"])
	}

	// Check stdin mapping
	if data["stdin"] != "package main" {
		t.Errorf("Expected stdin='package main', got '%v'", data["stdin"])
	}

	if data["file_content"] != "package main" {
		t.Errorf("Expected file_content='package main', got '%v'", data["file_content"])
	}

	// Check language detection
	if data["file_extension"] != "go" {
		t.Errorf("Expected file_extension='go', got '%v'", data["file_extension"])
	}

	if data["Language"] != "Go" {
		t.Errorf("Expected Language='Go', got '%v'", data["Language"])
	}
}

func TestCommandSpecWithTemplate(t *testing.T) {
	// Test YAML marshaling/unmarshaling
	spec := &CommandSpec{
		Name:        "security-audit",
		Version:     "1.0.0",
		Description: "Security audit using template",
		Args: []ArgSpec{
			{Name: "file", Description: "File to audit", Required: true},
		},
		Template: &TemplateRef{
			Name: "security-review",
			Variables: map[string]string{
				"Language": "{{.file_extension}}",
				"Code":     "{{.file_content}}",
			},
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(spec)
	if err != nil {
		t.Fatalf("failed to marshal spec: %v", err)
	}

	// Unmarshal back
	var spec2 CommandSpec
	if err := yaml.Unmarshal(data, &spec2); err != nil {
		t.Fatalf("failed to unmarshal spec: %v", err)
	}

	// Verify template reference
	if spec2.Template == nil {
		t.Fatal("Template reference was lost during marshaling")
	}

	if spec2.Template.Name != "security-review" {
		t.Errorf("Expected template name 'security-review', got '%s'", spec2.Template.Name)
	}

	if len(spec2.Template.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(spec2.Template.Variables))
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that commands without templates still work
	tmpDir, err := os.MkdirTemp("", "scmd-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create command spec WITHOUT template
	spec := &CommandSpec{
		Name:        "old-command",
		Version:     "1.0.0",
		Description: "Legacy command without template",
		Prompt: PromptSpec{
			System:   "You are a helpful assistant",
			Template: "Answer: {{.input}}",
		},
	}

	// Verify it can be marshaled
	data, err := yaml.Marshal(spec)
	if err != nil {
		t.Fatalf("failed to marshal spec: %v", err)
	}

	// Verify it can be unmarshaled
	var spec2 CommandSpec
	if err := yaml.Unmarshal(data, &spec2); err != nil {
		t.Fatalf("failed to unmarshal spec: %v", err)
	}

	// Verify template is nil (backward compatible)
	if spec2.Template != nil {
		t.Error("Template should be nil for legacy commands")
	}

	// Verify prompt still works
	if spec2.Prompt.Template != "Answer: {{.input}}" {
		t.Error("Prompt template was corrupted")
	}
}
