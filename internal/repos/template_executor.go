package repos

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/scmd/scmd/internal/backend"
	"github.com/scmd/scmd/internal/command"
	"github.com/scmd/scmd/internal/templates"
)

// TemplateExecutor handles execution of template-based commands
type TemplateExecutor struct {
	templateManager *templates.Manager
	dataDir         string
}

// NewTemplateExecutor creates a new template executor
func NewTemplateExecutor(dataDir string) (*TemplateExecutor, error) {
	templateManager, err := templates.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create template manager: %w", err)
	}

	return &TemplateExecutor{
		templateManager: templateManager,
		dataDir:         dataDir,
	}, nil
}

// NewTemplateExecutorWithManager creates a template executor with a custom template manager (for testing)
func NewTemplateExecutorWithManager(dataDir string, templateManager *templates.Manager) *TemplateExecutor {
	return &TemplateExecutor{
		templateManager: templateManager,
		dataDir:         dataDir,
	}
}

// ExecuteTemplateCommand executes a command that uses a template
func (te *TemplateExecutor) ExecuteTemplateCommand(
	ctx context.Context,
	spec *CommandSpec,
	args *command.Args,
	execCtx *command.ExecContext,
) (*command.Result, error) {
	if spec.Template == nil {
		return nil, fmt.Errorf("command spec has no template reference")
	}

	// Build data context from command args/flags
	data := te.buildDataContext(spec, args)

	// Get system and user prompts from template
	systemPrompt, userPrompt, err := te.executeTemplate(spec.Template, data)
	if err != nil {
		return &command.Result{
			Success: false,
			Error:   fmt.Sprintf("template execution failed: %v", err),
		}, nil
	}

	// Use existing backend for LLM inference
	if execCtx.Backend == nil {
		return &command.Result{
			Success: false,
			Error:   "no backend available",
		}, nil
	}

	// Create completion request
	req := &backend.CompletionRequest{
		Prompt:       userPrompt,
		SystemPrompt: systemPrompt,
		MaxTokens:    2048,
		Temperature:  0.7,
	}

	// Apply model preferences from command spec
	if spec.Model.MaxTokens > 0 {
		req.MaxTokens = spec.Model.MaxTokens
	}
	if spec.Model.Temperature > 0 {
		req.Temperature = spec.Model.Temperature
	}

	// Execute completion
	resp, err := execCtx.Backend.Complete(ctx, req)
	if err != nil {
		return &command.Result{
			Success: false,
			Error:   fmt.Sprintf("completion failed: %v", err),
		}, nil
	}

	return &command.Result{
		Success: true,
		Output:  resp.Content,
	}, nil
}

// executeTemplate resolves and executes a template reference
func (te *TemplateExecutor) executeTemplate(
	templateRef *TemplateRef,
	data map[string]interface{},
) (string, string, error) {
	// Case 1: Inline template
	if templateRef.Inline != nil {
		return te.executeInlineTemplate(templateRef.Inline, data)
	}

	// Case 2: Template reference by name
	if templateRef.Name != "" {
		return te.executeNamedTemplate(templateRef.Name, templateRef.Variables, data)
	}

	return "", "", fmt.Errorf("template reference must specify either 'name' or 'inline'")
}

// executeInlineTemplate executes an inline template definition
func (te *TemplateExecutor) executeInlineTemplate(
	inlineTemplate *InlineTemplate,
	data map[string]interface{},
) (string, string, error) {
	// Validate required variables
	for _, v := range inlineTemplate.Variables {
		if v.Required {
			if _, ok := data[v.Name]; !ok {
				return "", "", fmt.Errorf("required variable %s not provided", v.Name)
			}
		}
	}

	// Apply defaults
	for _, v := range inlineTemplate.Variables {
		if _, ok := data[v.Name]; !ok && v.Default != "" {
			data[v.Name] = v.Default
		}
	}

	// Execute user prompt template
	userPrompt, err := te.executeTemplateString(inlineTemplate.UserPromptTemplate, data)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute user prompt template: %w", err)
	}

	// System prompt is optional
	systemPrompt := inlineTemplate.SystemPrompt

	return systemPrompt, userPrompt, nil
}

// executeNamedTemplate executes a template by name from the template manager
func (te *TemplateExecutor) executeNamedTemplate(
	name string,
	variableMapping map[string]string,
	data map[string]interface{},
) (string, string, error) {
	// Load template from manager
	tpl, err := te.templateManager.Load(name)
	if err != nil {
		return "", "", fmt.Errorf("failed to load template %s: %w", name, err)
	}

	// Map command context to template variables
	templateData := make(map[string]interface{})
	for templateVar, commandVar := range variableMapping {
		// Resolve the command variable mapping
		value, err := te.resolveVariableMapping(commandVar, data)
		if err != nil {
			return "", "", fmt.Errorf("failed to resolve variable %s: %w", commandVar, err)
		}
		templateData[templateVar] = value
	}

	// Fill in any missing variables from data
	for key, value := range data {
		if _, exists := templateData[key]; !exists {
			templateData[key] = value
		}
	}

	// Execute template with mapped data
	return tpl.Execute(templateData)
}

// resolveVariableMapping resolves a variable mapping like "{{.file_content}}"
func (te *TemplateExecutor) resolveVariableMapping(
	mapping string,
	data map[string]interface{},
) (interface{}, error) {
	// If it's a template expression, evaluate it
	if strings.HasPrefix(mapping, "{{") && strings.HasSuffix(mapping, "}}") {
		return te.executeTemplateString(mapping, data)
	}

	// Otherwise, treat it as a direct value
	return mapping, nil
}

// executeTemplateString executes a template string with the given data
func (te *TemplateExecutor) executeTemplateString(
	tmplStr string,
	data map[string]interface{},
) (string, error) {
	tmpl, err := template.New("template").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// buildDataContext builds the data context for template execution
func (te *TemplateExecutor) buildDataContext(
	spec *CommandSpec,
	args *command.Args,
) map[string]interface{} {
	data := make(map[string]interface{})

	// Add positional args by name
	for i, argSpec := range spec.Args {
		if i < len(args.Positional) {
			data[argSpec.Name] = args.Positional[i]

			// Auto-detect file extension if this is a file argument
			if strings.Contains(argSpec.Description, "file") || strings.Contains(argSpec.Name, "file") {
				ext := filepath.Ext(args.Positional[i])
				if ext != "" {
					data["file_extension"] = strings.TrimPrefix(ext, ".")
				}
			}
		} else if argSpec.Default != "" {
			data[argSpec.Name] = argSpec.Default
		}
	}

	// Add flags by name
	for _, flagSpec := range spec.Flags {
		if val, ok := args.Options[flagSpec.Name]; ok {
			data[flagSpec.Name] = val
		} else if flagSpec.Default != "" {
			data[flagSpec.Name] = flagSpec.Default
		}
	}

	// Add stdin if present
	if stdin, ok := args.Options["stdin"]; ok {
		data["stdin"] = stdin
		data["input"] = stdin        // alias
		data["file_content"] = stdin // commonly used for code content
	}

	// Add all positional args as array
	data["args"] = args.Positional

	// Auto-detect language from file extension if available
	if _, ok := data["file_extension"]; ok && data["file_extension"] != "" {
		data["Language"] = te.detectLanguage(data["file_extension"].(string))
	}

	// If we have stdin/file content, add it as "Code" for convenience
	if content, ok := data["file_content"]; ok {
		data["Code"] = content
	}

	return data
}

// detectLanguage maps file extensions to language names
func (te *TemplateExecutor) detectLanguage(ext string) string {
	languageMap := map[string]string{
		"go":   "Go",
		"py":   "Python",
		"js":   "JavaScript",
		"ts":   "TypeScript",
		"java": "Java",
		"c":    "C",
		"cpp":  "C++",
		"cc":   "C++",
		"cxx":  "C++",
		"rs":   "Rust",
		"rb":   "Ruby",
		"php":  "PHP",
		"sh":   "Bash",
		"bash": "Bash",
		"zsh":  "Zsh",
		"sql":  "SQL",
		"yaml": "YAML",
		"yml":  "YAML",
		"json": "JSON",
		"xml":  "XML",
		"html": "HTML",
		"css":  "CSS",
		"md":   "Markdown",
	}

	if lang, ok := languageMap[ext]; ok {
		return lang
	}

	return "text"
}
