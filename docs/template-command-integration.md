# Template and Command Integration (v0.4.2)

This document describes the integration between slash commands and templates in scmd v0.4.2, enabling powerful template-based command execution with variable mapping.

## Overview

Commands can now reference templates in three ways:
1. **Direct prompts** (existing): Uses the `Prompt` field directly
2. **Template reference**: References an existing template by name
3. **Inline template**: Defines a template inline within the command spec

This integration allows:
- Reusable templates across multiple commands
- Variable mapping from command context to template variables
- Automatic context population (file content, language detection, etc.)
- Repository-based template distribution
- Full backward compatibility with existing prompt-based commands

## CommandSpec Changes

### New Fields

```go
type CommandSpec struct {
    // ... existing fields ...

    // Template integration (NEW in v0.4.2)
    Template *TemplateRef `yaml:"template,omitempty" json:"template,omitempty"`
}

type TemplateRef struct {
    // Name of the template to use (references ~/.scmd/templates/<name>.yaml)
    Name string `yaml:"name,omitempty" json:"name,omitempty"`

    // Variables to map from command context to template variables
    // Keys are template variable names, values are command context references
    // e.g., "Language": "{{.file_extension}}", "Code": "{{.file_content}}"
    Variables map[string]string `yaml:"variables,omitempty" json:"variables,omitempty"`

    // Inline template definition (alternative to referencing by name)
    Inline *InlineTemplate `yaml:"inline,omitempty" json:"inline,omitempty"`
}

type InlineTemplate struct {
    SystemPrompt       string            `yaml:"system_prompt,omitempty"`
    UserPromptTemplate string            `yaml:"user_prompt_template"`
    Variables          []TemplateVariable `yaml:"variables,omitempty"`
}
```

### Manifest Changes

Repository manifests can now include templates:

```yaml
name: my-repo
version: 1.0.0
description: My command repository
commands:
  - path: commands/review.yaml
  - path: commands/analyze.yaml
templates:
  - path: templates/security-review.yaml
  - path: templates/performance-analysis.yaml
```

## Usage Examples

### Example 1: Command with Template Reference

```yaml
# security-audit.yaml
name: security-audit
version: 1.0.0
description: Security audit using OWASP template
category: security

args:
  - name: file
    description: File to audit
    required: true

flags:
  - name: additional-context
    description: Additional context for the audit

template:
  name: security-review  # References ~/.scmd/templates/security-review.yaml
  variables:
    Language: "{{.file_extension}}"  # Auto-detected from file argument
    Code: "{{.file_content}}"        # Populated from stdin
    Context: "{{.additional_context}}"  # Maps flag to template variable

model:
  preferred: qwen2.5-7b
  temperature: 0.3
```

Usage:
```bash
# Audit a file
cat auth.js | scmd security-audit auth.js

# With additional context
cat auth.js | scmd security-audit auth.js --additional-context "JWT authentication system"
```

### Example 2: Command with Inline Template

```yaml
# quick-review.yaml
name: quick-review
version: 1.0.0
description: Quick code review with inline template
category: review

args:
  - name: file
    description: File to review
    required: true

template:
  inline:
    system_prompt: |
      You are an experienced code reviewer.
      Focus on code quality, best practices, and potential bugs.

    user_prompt_template: |
      Review this {{.Language}} code:

      ```{{.Language}}
      {{.Code}}
      ```

      Provide:
      1. Overall code quality assessment
      2. Potential bugs or issues
      3. Suggestions for improvement

    variables:
      - name: Language
        description: Programming language
        default: unknown

      - name: Code
        description: Code to review
        required: true
```

Usage:
```bash
# Review a Python file
cat script.py | scmd quick-review script.py
```

### Example 3: Repository with Templates

**Repository structure:**
```
my-repo/
├── scmd-repo.yaml
├── commands/
│   ├── security-audit.yaml
│   └── perf-analyze.yaml
└── templates/
    ├── owasp-security.yaml
    └── performance.yaml
```

**scmd-repo.yaml:**
```yaml
name: security-tools
version: 1.0.0
description: Security audit commands and templates

commands:
  - name: security-audit
    description: Security audit with OWASP focus
    file: commands/security-audit.yaml

  - name: perf-analyze
    description: Performance analysis
    file: commands/perf-analyze.yaml

templates:
  - path: templates/owasp-security.yaml
  - path: templates/performance.yaml
```

Installation:
```bash
# Add repository
scmd repo add security https://github.com/org/security-tools/raw/main

# Install command (automatically installs referenced templates)
scmd repo install security/security-audit

# Use the command
cat auth.js | scmd security-audit auth.js
```

## Automatic Variable Mapping

The template executor automatically provides these context variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `{{.arg_name}}` | Command argument by name | `{{.file}}` |
| `{{.flag_name}}` | Command flag by name | `{{.verbose}}` |
| `{{.stdin}}` | Content from stdin | Full file content |
| `{{.input}}` | Alias for stdin | Same as `{{.stdin}}` |
| `{{.file_content}}` | Alias for stdin (commonly used for code) | Same as `{{.stdin}}` |
| `{{.file_extension}}` | Auto-detected from file argument | `go`, `py`, `js` |
| `{{.Language}}` | Full language name from extension | `Go`, `Python`, `JavaScript` |
| `{{.args}}` | Array of all positional arguments | `["file1", "file2"]` |
| `{{.all_args}}` | All args as space-separated string | `file1 file2` |

### Language Detection

The executor automatically detects the programming language from file extensions:

| Extension | Language |
|-----------|----------|
| `.go` | Go |
| `.py` | Python |
| `.js` | JavaScript |
| `.ts` | TypeScript |
| `.java` | Java |
| `.c` | C |
| `.cpp`, `.cc`, `.cxx` | C++ |
| `.rs` | Rust |
| `.rb` | Ruby |
| `.php` | PHP |
| `.sh`, `.bash` | Bash |
| `.sql` | SQL |
| `.yaml`, `.yml` | YAML |
| `.json` | JSON |
| `.html` | HTML |
| `.css` | CSS |
| `.md` | Markdown |

## Template Variable Resolution

Variables can be:

1. **Direct values**: `"Language": "Python"`
2. **Template expressions**: `"Code": "{{.file_content}}"`
3. **Chained references**: `"Language": "{{.Language}}"` (uses auto-detected)

Example:
```yaml
template:
  name: security-review
  variables:
    Language: "{{.file_extension}}"   # Auto-detected: "py"
    Code: "{{.file_content}}"         # From stdin
    Context: "{{.context}}"           # From --context flag
    Framework: "Django"               # Direct value
```

## Validation

Template references are validated on command installation:

1. **Mutual exclusivity**: Must specify either `name` OR `inline`, not both
2. **Required fields**: Inline templates must have `user_prompt_template`
3. **Variable names**: Inline template variables must have names
4. **Template existence**: Named templates must exist when referenced (warning only)

Example validation:
```yaml
# INVALID - both name and inline
template:
  name: security-review
  inline:
    user_prompt_template: "..."  # Error: cannot specify both

# INVALID - no user prompt
template:
  inline:
    system_prompt: "..."  # Error: user_prompt_template required

# VALID - named template
template:
  name: security-review
  variables:
    Code: "{{.file_content}}"

# VALID - inline template
template:
  inline:
    user_prompt_template: "Review: {{.Code}}"
    variables:
      - name: Code
        required: true
```

## Backward Compatibility

Commands without templates continue to work as before:

```yaml
# Old-style command (still works)
name: old-command
version: 1.0.0
description: Legacy command
prompt:
  system: "You are a helpful assistant"
  template: "Answer: {{.input}}"
```

The executor checks for templates first, then falls back to the `prompt` field:
1. If `template` is present → use template executor
2. If `compose` is present → use composer
3. Otherwise → use prompt-based execution (legacy)

## Best Practices

### 1. Template Organization

```
~/.scmd/
├── config.yaml
├── templates/
│   ├── security/
│   │   ├── owasp-review.yaml
│   │   └── crypto-analysis.yaml
│   ├── performance/
│   │   ├── complexity-analysis.yaml
│   │   └── profiling.yaml
│   └── documentation/
│       ├── api-docs.yaml
│       └── readme-gen.yaml
```

### 2. Variable Naming Conventions

- Use `PascalCase` for template variables: `Language`, `Code`, `Context`
- Use `snake_case` for command context: `file_content`, `file_extension`
- Document expected variables in template metadata

### 3. Template Reusability

Create generic templates that work across multiple commands:

```yaml
# templates/code-review-base.yaml
name: code-review-base
version: 1.0.0
description: Base template for code reviews
compatible_commands: ["review", "audit", "analyze"]

system_prompt: |
  You are an experienced code reviewer.

user_prompt_template: |
  Review this {{.Language}} code:

  {{.Code}}

  {{if .Focus}}
  Focus on: {{.Focus}}
  {{end}}

  {{if .Context}}
  Additional context: {{.Context}}
  {{end}}

variables:
  - name: Language
    required: true
  - name: Code
    required: true
  - name: Focus
    required: false
  - name: Context
    required: false
```

### 4. Command Composition

Combine template-based commands with composition:

```yaml
name: full-security-audit
version: 1.0.0
description: Complete security audit with multiple checks

template:
  name: security-review
  variables:
    Code: "{{.file_content}}"
    Language: "{{.file_extension}}"

compose:
  pipeline:
    - command: vulnerability-scan
    - command: dependency-check
    - command: security-review

hooks:
  post:
    - shell: "echo 'Audit complete'"
```

## Troubleshooting

### Template Not Found

```
Error: template security-review not found
```

**Solution:**
1. Check template exists: `scmd template list`
2. Verify template name matches exactly
3. Ensure template is in `~/.scmd/templates/`

### Required Variable Missing

```
Error: required variable Code not provided
```

**Solution:**
1. Check template variables: `scmd template show security-review`
2. Ensure command provides all required variables
3. Add variable mapping in command spec

### Invalid Template Reference

```
Error: template must specify either 'name' or 'inline'
```

**Solution:**
- Choose one: either reference by name OR define inline
- Cannot use both in same command

## Migration Guide

### Migrating Existing Commands to Templates

**Before (v0.4.1):**
```yaml
name: security-review
version: 1.0.0
description: Security code review
prompt:
  system: "You are a security expert..."
  template: |
    Review this code for security issues:
    {{.stdin}}
```

**After (v0.4.2):**
```yaml
name: security-review
version: 1.0.0
description: Security code review
template:
  name: owasp-security  # Reusable template
  variables:
    Code: "{{.file_content}}"
    Language: "{{.file_extension}}"
```

Benefits:
- Template can be shared across multiple commands
- Centralized template updates
- Better organization and discoverability
- Variable mapping makes intent clearer

## API Reference

### TemplateExecutor

```go
type TemplateExecutor struct {
    templateManager *templates.Manager
    dataDir         string
}

// NewTemplateExecutor creates a new template executor
func NewTemplateExecutor(dataDir string) (*TemplateExecutor, error)

// ExecuteTemplateCommand executes a command that uses a template
func (te *TemplateExecutor) ExecuteTemplateCommand(
    ctx context.Context,
    spec *CommandSpec,
    args *command.Args,
    execCtx *command.ExecContext,
) (*command.Result, error)
```

### Template Manager

```go
// NewManagerWithDir creates a template manager with custom directory
func NewManagerWithDir(templatesDir string) (*Manager, error)

// Load loads a template by name
func (m *Manager) Load(name string) (*Template, error)

// Execute executes a template with data
func (m *Manager) Execute(name string, data map[string]interface{}) (string, string, error)
```

## Examples Repository

See the official examples repository for more template and command examples:
- https://github.com/scmd/commands/tree/main/templates
- https://github.com/scmd/commands/tree/main/commands

## See Also

- [Template System Documentation](./template-system.md)
- [Command Specification](./command-spec.md)
- [Repository Management](./repository-management.md)
- [Command Composition](./command-composition.md)
