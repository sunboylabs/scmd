package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/scmd/scmd/internal/config"
	"gopkg.in/yaml.v3"
)

// Manager manages templates
type Manager struct {
	templatesDir string
	cache        map[string]*Template
}

// NewManager creates a new template manager
func NewManager() (*Manager, error) {
	dataDir := config.GetDataDir()
	templatesDir := filepath.Join(dataDir, "templates")

	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return nil, err
	}

	return &Manager{
		templatesDir: templatesDir,
		cache:        make(map[string]*Template),
	}, nil
}

// Load loads a template by name
func (m *Manager) Load(name string) (*Template, error) {
	// Check cache
	if t, ok := m.cache[name]; ok {
		return t, nil
	}

	// Load from file
	path := filepath.Join(m.templatesDir, name+".yaml")
	tpl, err := LoadTemplate(path)
	if err != nil {
		// Try without extension
		path = filepath.Join(m.templatesDir, name)
		tpl, err = LoadTemplate(path)
		if err != nil {
			return nil, fmt.Errorf("template %s not found: %w", name, err)
		}
	}

	// Cache it
	m.cache[name] = tpl
	return tpl, nil
}

// List returns all available templates
func (m *Manager) List() ([]*Template, error) {
	pattern := filepath.Join(m.templatesDir, "*.yaml")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Also check for .yml files
	pattern = filepath.Join(m.templatesDir, "*.yml")
	ymlFiles, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	files = append(files, ymlFiles...)

	var templates []*Template
	for _, file := range files {
		tpl, err := LoadTemplate(file)
		if err != nil {
			// Skip invalid templates
			continue
		}
		templates = append(templates, tpl)
	}

	return templates, nil
}

// Create creates a new template
func (m *Manager) Create(tpl *Template) error {
	if err := tpl.Validate(); err != nil {
		return err
	}

	path := filepath.Join(m.templatesDir, tpl.Name+".yaml")

	// Check if exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("template %s already exists", tpl.Name)
	}

	return tpl.Save(path)
}

// Update updates an existing template
func (m *Manager) Update(tpl *Template) error {
	if err := tpl.Validate(); err != nil {
		return err
	}

	path := filepath.Join(m.templatesDir, tpl.Name+".yaml")

	// Clear from cache
	delete(m.cache, tpl.Name)

	return tpl.Save(path)
}

// Delete removes a template
func (m *Manager) Delete(name string) error {
	path := filepath.Join(m.templatesDir, name+".yaml")

	if err := os.Remove(path); err != nil {
		// Try without extension
		path = filepath.Join(m.templatesDir, name)
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to delete template %s: %w", name, err)
		}
	}

	delete(m.cache, name)
	return nil
}

// Search finds templates matching a query
func (m *Manager) Search(query string) ([]*Template, error) {
	templates, err := m.List()
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var matches []*Template

	for _, tpl := range templates {
		// Search in name, description, and tags
		if strings.Contains(strings.ToLower(tpl.Name), query) ||
			strings.Contains(strings.ToLower(tpl.Description), query) {
			matches = append(matches, tpl)
			continue
		}

		for _, tag := range tpl.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				matches = append(matches, tpl)
				break
			}
		}
	}

	return matches, nil
}

// Execute executes a template with data
func (m *Manager) Execute(name string, data map[string]interface{}) (string, string, error) {
	tpl, err := m.Load(name)
	if err != nil {
		return "", "", err
	}

	return tpl.Execute(data)
}

// Export exports a template to a string
func (m *Manager) Export(name string) (string, error) {
	tpl, err := m.Load(name)
	if err != nil {
		return "", err
	}

	data, err := yaml.Marshal(tpl)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Import imports a template from YAML data
func (m *Manager) Import(data []byte, overwrite bool) error {
	var tpl Template
	if err := yaml.Unmarshal(data, &tpl); err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	if err := tpl.Validate(); err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}

	path := filepath.Join(m.templatesDir, tpl.Name+".yaml")

	// Check if exists
	if _, err := os.Stat(path); err == nil && !overwrite {
		return fmt.Errorf("template %s already exists (use --force to overwrite)", tpl.Name)
	}

	return tpl.Save(path)
}

// GetTemplateDir returns the templates directory
func (m *Manager) GetTemplateDir() string {
	return m.templatesDir
}

// InitBuiltinTemplates creates the default built-in templates
func (m *Manager) InitBuiltinTemplates() error {
	builtins := getBuiltinTemplates()

	for _, tpl := range builtins {
		path := filepath.Join(m.templatesDir, tpl.Name+".yaml")

		// Skip if already exists
		if _, err := os.Stat(path); err == nil {
			continue
		}

		if err := tpl.Save(path); err != nil {
			return fmt.Errorf("failed to create builtin template %s: %w", tpl.Name, err)
		}
	}

	return nil
}