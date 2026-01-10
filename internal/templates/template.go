package templates

import (
	"bytes"
	"fmt"
	"os"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Template defines a prompt template
type Template struct {
	Name                string       `yaml:"name"`
	Version            string       `yaml:"version"`
	Author             string       `yaml:"author"`
	Description        string       `yaml:"description"`
	Tags               []string     `yaml:"tags"`
	CompatibleCommands []string     `yaml:"compatible_commands"`
	SystemPrompt       string       `yaml:"system_prompt"`
	UserPromptTemplate string       `yaml:"user_prompt_template"`
	Variables          []Variable   `yaml:"variables"`
	Output             OutputConfig `yaml:"output"`
	RecommendedModels  []string     `yaml:"recommended_models"`
	Examples           []Example    `yaml:"examples"`
}

// Variable defines a template variable
type Variable struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Default     string `yaml:"default"`
	Required    bool   `yaml:"required"`
}

// OutputConfig defines output formatting preferences
type OutputConfig struct {
	Format   string    `yaml:"format"`
	Sections []Section `yaml:"sections"`
}

// Section defines an output section
type Section struct {
	Title    string `yaml:"title"`
	Required bool   `yaml:"required"`
}

// Example shows how to use the template
type Example struct {
	Description string `yaml:"description"`
	Command     string `yaml:"command"`
}

// Validate validates the template structure
func (t *Template) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if t.UserPromptTemplate == "" {
		return fmt.Errorf("user_prompt_template is required")
	}
	if len(t.CompatibleCommands) == 0 {
		return fmt.Errorf("at least one compatible command is required")
	}
	return nil
}

// Execute executes the template with provided data
func (t *Template) Execute(data map[string]interface{}) (string, string, error) {
	// Validate required variables
	for _, v := range t.Variables {
		if v.Required {
			if _, ok := data[v.Name]; !ok {
				return "", "", fmt.Errorf("required variable %s not provided", v.Name)
			}
		}
	}

	// Apply defaults
	for _, v := range t.Variables {
		if _, ok := data[v.Name]; !ok && v.Default != "" {
			data[v.Name] = v.Default
		}
	}

	// Execute user prompt template
	tmpl, err := template.New("user_prompt").Parse(t.UserPromptTemplate)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse template: %w", err)
	}

	var userPromptBuf bytes.Buffer
	if err := tmpl.Execute(&userPromptBuf, data); err != nil {
		return "", "", fmt.Errorf("failed to execute template: %w", err)
	}

	return t.SystemPrompt, userPromptBuf.String(), nil
}

// LoadTemplate loads a template from a file
func LoadTemplate(path string) (*Template, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tpl Template
	if err := yaml.Unmarshal(data, &tpl); err != nil {
		return nil, err
	}

	if err := tpl.Validate(); err != nil {
		return nil, err
	}

	return &tpl, nil
}

// Save saves the template to a file
func (t *Template) Save(path string) error {
	data, err := yaml.Marshal(t)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// IsCompatibleWith checks if the template is compatible with a command
func (t *Template) IsCompatibleWith(command string) bool {
	for _, cmd := range t.CompatibleCommands {
		if cmd == command {
			return true
		}
	}
	return false
}