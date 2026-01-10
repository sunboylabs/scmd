# README v0.2.0 Additions

## New Sections to Add to README.md

### Section 1: Interactive Chat Mode (Add after "Quick Start" section)

```markdown
## Interactive Chat Mode

scmd now supports multi-turn conversations with full context retention. Have extended discussions with the AI while scmd remembers your entire conversation history.

### Starting a Conversation

```bash
# Start a new interactive chat session
scmd chat

# Use a specific model
scmd chat --model qwen2.5-7b

# Resume a previous conversation
scmd chat --continue abc123
```

### In-Session Commands

While in a chat session, you can use these special commands:

- `/help` - Show available commands
- `/clear` - Clear current context (keeps history)
- `/info` - Show conversation info
- `/save` - Force save conversation
- `/export` - Export conversation to markdown
- `/model <name>` - Switch to a different model
- `/exit` - Exit session (or press Ctrl+D)

### Managing Conversation History

```bash
# List recent conversations
scmd history list

# Show a specific conversation
scmd history show abc123

# Search conversations
scmd history search "docker"

# Delete a conversation
scmd history delete abc123

# Clear all conversations
scmd history clear
```

### Example Workflow

```bash
# Start a conversation about Go programming
$ scmd chat

You: How do I create a REST API in Go?
Assistant: I'll help you create a REST API in Go. Here's a comprehensive guide...

You: Can you show me how to add authentication?
Assistant: Of course! Building on the previous example...

[Ctrl+D to exit]

# Later, resume the conversation
$ scmd history list
  1. [a3f2b1c4] How do I create a REST API...
     Model: qwen2.5-1.5b | Messages: 4 | Updated: Jan 10, 14:23

$ scmd chat --continue a3f2b1c4
Resumed with 4 previous messages

You: What about rate limiting?
Assistant: [Continues with full context of previous discussion...]
```

### Storage

Conversations are stored in `~/.scmd/conversations.db` using SQLite. Each conversation includes:
- Unique ID (UUID)
- Auto-generated title from first message
- Full message history
- Model and backend used
- Timestamps

### Configuration

```yaml
chat:
  max_context_messages: 20    # Max messages to keep in context
  auto_save: true            # Auto-save after each message
  auto_title: true           # Auto-generate titles from first message
```
```

---

### Section 2: Template System (Add after "Repository System" section)

```markdown
## Template System

Create and use customizable prompt templates for specialized workflows. Templates allow you to standardize prompts for security reviews, performance analysis, documentation, and more.

### Quick Start

```bash
# Initialize built-in templates
scmd template init

# List available templates
scmd template list

# View template details
scmd template show security-review

# Use a template
scmd review auth.py --template security-review
scmd explain loop.py --template beginner-explain
```

### Built-in Templates

scmd comes with 6 professional templates:

1. **security-review** - OWASP Top 10 focused security analysis
   ```bash
   scmd review auth.js --template security-review
   ```

2. **performance** - Bottleneck and optimization analysis
   ```bash
   scmd review algorithm.py --template performance
   ```

3. **api-design** - REST API best practices review
   ```bash
   scmd review api.go --template api-design
   ```

4. **testing** - Test coverage and generation
   ```bash
   scmd review service.ts --template testing
   ```

5. **documentation** - Generate or review documentation
   ```bash
   scmd explain utils.rs --template documentation
   ```

6. **beginner-explain** - Explain code to beginners
   ```bash
   scmd explain quicksort.py --template beginner-explain
   ```

### Template Management

```bash
# Search for templates
scmd template search security

# Create a custom template
scmd template create my-review

# Export a template for sharing
scmd template export security-review > security.yaml

# Import a template
scmd template import security.yaml

# Delete a template
scmd template delete my-review
```

### Creating Custom Templates

Templates are YAML files with a simple structure:

```yaml
name: my-custom-review
version: "1.0"
author: "Your Name"
description: "Custom code review for my team"
tags:
  - custom
  - team
compatible_commands:
  - review

system_prompt: |
  You are an expert code reviewer for our team.
  Focus on our coding standards and best practices.

user_prompt_template: |
  Review the following {{.Language}} code:

  ```{{.Language}}
  {{.Code}}
  ```

  {{if .Context}}
  Context: {{.Context}}
  {{end}}

  Check for:
  1. Adherence to team coding standards
  2. Proper error handling
  3. Code clarity and maintainability

variables:
  - name: Language
    description: "Programming language"
    default: "auto-detect"
  - name: Code
    description: "Code to review"
    required: true
  - name: Context
    description: "Additional context"
    required: false

recommended_models:
  - qwen2.5-7b
  - gpt-4
```

### Template Variables

Templates support variable substitution using Go's `text/template` syntax:

- `{{.Code}}` - The code being analyzed
- `{{.Language}}` - Detected programming language
- `{{.Context}}` - Additional context (from --focus or --context flags)
- `{{.FocusOn}}` - Specific area to focus on
- `{{if .Variable}}...{{end}}` - Conditional sections

### Sharing Templates

Templates are perfect for teams:

```bash
# Export team template
scmd template export team-review > team-review.yaml

# Commit to team repo
git add team-review.yaml
git commit -m "Add team code review template"

# Team members import it
scmd template import team-review.yaml
```

### Template Directory

Templates are stored in `~/.scmd/templates/` as YAML files. You can:
- Edit them manually
- Version control them
- Share them with your team
- Create template repositories

### Advanced Usage

Combine templates with other features:

```bash
# Use template in a chat session
scmd chat
You: /model qwen2.5-7b
You: Review this authentication code using the security-review template

# Use template with piped input
git diff | scmd review --template security-review

# Use template with multiple files
scmd review src/**/*.go --template performance
```
```

---

### Section 3: Beautiful Output (Add to "Features" section)

Add this bullet point to the Features section:

```markdown
- **Beautiful Output** - Markdown rendering with syntax highlighting, multiple themes, and terminal detection
```

And add this subsection after "Configuration":

```markdown
### Output Formatting

scmd automatically renders markdown output with beautiful formatting:

- **Syntax Highlighting** - Code blocks with language-specific syntax coloring
- **Theme Detection** - Automatically adapts to dark or light terminals
- **Smart Formatting** - Headers, lists, tables, and links beautifully rendered
- **Plain Text Mode** - Respects `NO_COLOR` environment variable

```bash
# Beautiful formatted output (default)
scmd explain main.go

# Plain text for piping
NO_COLOR=1 scmd explain main.go > explanation.txt

# Force specific theme
SCMD_THEME=light scmd review code.py
```

### Theme Configuration

```yaml
ui:
  colors: true              # Enable colored output
  style: auto               # Theme: auto, dark, light, notty
```

Supported output formats:
- `--format text` - Beautiful markdown rendering (default)
- `--format markdown` - Raw markdown
- `--format json` - JSON output for scripting
```

---

### Section 4: Update CLI Reference

Add to the Commands section:

```markdown
Commands:
  ...existing commands...

  chat        Start interactive conversation
  history     Manage conversation history
    list      List recent conversations
    show      Show conversation details
    search    Search conversations
    delete    Delete a conversation
    clear     Clear all conversations

  template    Manage prompt templates
    init      Initialize built-in templates
    list      List available templates
    show      Show template details
    create    Create custom template
    delete    Delete template
    search    Search templates
    export    Export template to YAML
    import    Import template from file

Flags:
  ...existing flags...
      --template string   Use a prompt template (for explain/review)
```

---

### Section 5: Update Examples

Add these examples to the README:

```markdown
### Advanced Examples

```bash
# Interactive conversation with context retention
scmd chat --model qwen2.5-7b
You: Explain dependency injection
Assistant: [Explanation]
You: Show me an example in Go
Assistant: [Example with full context]

# Security-focused code review
scmd review auth.js --template security-review

# Beginner-friendly explanation
scmd explain algorithm.py --template beginner-explain

# Resume previous conversation
scmd history list
scmd chat --continue abc123

# Search past conversations
scmd history search "docker containers"

# Export conversation
scmd chat
[discussion about Rust]
/export
✓ Exported to conversation_abc123.md

# Create and use custom template
scmd template create team-standards
[interactive prompts]
scmd review --template team-standards src/*.go

# Beautiful output with syntax highlighting
scmd explain main.go
[Formatted output with colors, headers, highlighted code blocks]
```
```

---

### Section 6: Configuration Reference

Update the Configuration section with:

```yaml
# Complete ~/.scmd/config.yaml reference

default_backend: llamacpp
default_model: qwen2.5-1.5b

backends:
  llamacpp:
    model: qwen2.5-1.5b
  ollama:
    host: http://localhost:11434
  openai:
    model: gpt-4o-mini

# NEW: Chat configuration
chat:
  max_context_messages: 20    # Max messages to keep in context
  auto_save: true             # Auto-save after each message
  auto_title: true            # Auto-generate conversation titles

# NEW: UI configuration
ui:
  color: true                 # Enable colored output
  style: auto                 # Theme: auto, dark, light, notty
  spinner: true               # Show progress spinners

# NEW: Template configuration
templates:
  directory: ~/.scmd/templates  # Template storage location
```

---

## Quick Reference Card

Add this quick reference at the end of README:

```markdown
## Quick Reference

### Chat Commands
```bash
scmd chat                        # Start conversation
scmd chat --continue abc123      # Resume conversation
scmd history list                # List conversations
scmd history search "query"      # Search conversations
```

### Template Commands
```bash
scmd template init               # Initialize built-in templates
scmd template list               # List templates
scmd template show <name>        # View template details
scmd review --template <name>    # Use template
```

### In-Chat Commands
```
/help     Show help
/clear    Clear context
/info     Show conversation info
/export   Export to markdown
/exit     Exit (or Ctrl+D)
```

### Output Control
```bash
scmd explain file.go             # Beautiful formatted output
NO_COLOR=1 scmd explain file.go  # Plain text
scmd explain --format json       # JSON output
```
```
