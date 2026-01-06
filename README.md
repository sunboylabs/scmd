# scmd

**AI-powered slash commands for any terminal.**

scmd brings the power of LLM-based slash commands to your command line. Type `/gc` to generate commit messages, `/explain` to understand code, or install new commands from community repositories.

```bash
# With shell integration, use real slash commands:
/gc                     # Generate commit message from staged changes
/explain main.go        # Explain code
/review                 # Review code for issues
git diff | /sum         # Summarize changes

# Or use the scmd command directly:
cat main.go | scmd explain
git diff | scmd git-commit
```

## Features

- **Real Slash Commands** - Type `/command` directly in your terminal
- **Repository System** - Install commands from community repos or create your own
- **Multiple LLM Backends** - Ollama (local), OpenAI, Together.ai, Groq
- **Command Composition** - Chain commands in pipelines, run in parallel, or use fallbacks
- **Shell Integration** - Bash, Zsh, and Fish support with tab completion
- **Offline Support** - Local caching for commands and manifests
- **Lockfiles** - Reproducible installations for teams

## Installation

```bash
# Build from source
git clone https://github.com/scmd/scmd
cd scmd
make build

# Or with Go
go install github.com/scmd/scmd/cmd/scmd@latest
```

## Quick Start

```bash
# List available backends
scmd backends

# Explain some code
cat myfile.go | scmd explain

# Use with inline prompt
echo "SELECT * FROM users" | scmd -p "optimize this SQL query"

# Save output to file
git diff | scmd review -o review.md
```

## Slash Commands

The core feature of scmd is real slash commands in your terminal.

### Setup Shell Integration

```bash
# For Bash/Zsh - add to your ~/.bashrc or ~/.zshrc:
eval "$(scmd slash init bash)"

# For Fish - add to ~/.config/fish/config.fish:
scmd slash init fish | source
```

### Using Slash Commands

After setup, use slash commands directly:

```bash
/explain main.go           # Explain code
/gc                        # Generate commit message (alias for git-commit)
/review                    # Review code
/sum                       # Summarize (alias for summarize)
/fix                       # Explain errors

# Pipe input to commands
cat error.log | /fix
git diff | /gc
curl api.com/data | /sum
```

### Built-in Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `/explain` | `/e`, `/exp` | Explain code or concepts |
| `/review` | `/r`, `/rev` | Review code for issues |
| `/commit` | `/gc`, `/gitc` | Generate git commit messages |
| `/summarize` | `/s`, `/sum`, `/tldr` | Summarize text |
| `/fix` | `/f`, `/err` | Explain and fix errors |

### Managing Slash Commands

```bash
# List all slash commands
scmd slash list

# Add a new slash command
scmd slash add doc generate-docs --alias=d,docs

# Add an alias to existing command
scmd slash alias commit c

# Remove a slash command
scmd slash remove doc

# Interactive mode (REPL)
scmd slash interactive
```

## Repository System

scmd's killer feature is its repository-based command distribution. Think Homebrew taps, but for AI prompts.

### Installing Commands

```bash
# Add a repository
scmd repo add community https://github.com/scmd-community/commands/raw/main

# Search for commands
scmd repo search git

# Show command details
scmd repo show community/git-commit

# Install a command
scmd repo install community/git-commit

# Use the installed command
git diff | scmd git-commit
```

### Managing Repositories

```bash
# List configured repos
scmd repo list

# Update repo manifests
scmd repo update

# Remove a repo
scmd repo remove community
```

### Central Registry

Discover commands from the central scmd registry:

```bash
# Search the registry
scmd registry search docker

# Browse by category
scmd registry categories

# Show trending commands
scmd registry featured
```

## Command Specification

Commands are defined in YAML files with a powerful specification:

```yaml
name: git-commit
version: "1.0.0"
description: Generate commit messages from diffs
category: git
author: scmd team

args:
  - name: style
    description: Commit style (conventional, simple)
    default: conventional

prompt:
  system: |
    You are a git commit message expert.
    Use conventional commits format.
  template: |
    Generate a commit message for:
    {{.stdin}}

    Style: {{.style}}

model:
  temperature: 0.3
  max_tokens: 256
```

### Advanced Features

**Dependencies** - Commands can depend on other commands:
```yaml
dependencies:
  - command: official/explain
    version: ">=1.0.0"
  - command: official/summarize
    optional: true
```

**Composition** - Chain commands together:
```yaml
compose:
  pipeline:
    - command: explain
    - command: summarize
      args:
        length: short
```

**Hooks** - Run shell commands before/after:
```yaml
hooks:
  pre:
    - shell: "git status --porcelain"
      if: "{{.git}}"
  post:
    - shell: "echo 'Done!'"
```

**Context** - Auto-include files and environment:
```yaml
context:
  files:
    - "*.go"
    - "go.mod"
  git: true
  env:
    - GOPATH
```

## Lockfiles

Share exact command versions with your team:

```bash
# Generate lockfile from installed commands
scmd lock generate

# Install from lockfile
scmd lock install

# Check for updates
scmd update --check

# Update all commands
scmd update --all
```

## LLM Backends

scmd supports multiple LLM backends:

| Backend | Local | Free | Setup |
|---------|-------|------|-------|
| Ollama | Yes | Yes | `ollama serve` |
| OpenAI | No | No | `OPENAI_API_KEY` |
| Together.ai | No | Free tier | `TOGETHER_API_KEY` |
| Groq | No | Free tier | `GROQ_API_KEY` |

```bash
# Use specific backend
scmd -b ollama explain main.go

# Use specific model
scmd -b openai -m gpt-4 review code.py

# List available backends
scmd backends
```

## Creating a Repository

Create your own command repository:

```
my-commands/
├── scmd-repo.yaml          # Repository manifest
└── commands/
    ├── my-command.yaml
    └── another-command.yaml
```

**scmd-repo.yaml:**
```yaml
name: my-commands
version: "1.0.0"
description: My custom scmd commands
author: Your Name

commands:
  - name: my-command
    description: Does something useful
    file: commands/my-command.yaml
```

Host on GitHub, GitLab, or any HTTP server, then:
```bash
scmd repo add myrepo https://raw.githubusercontent.com/you/my-commands/main
```

## Configuration

Configuration is stored in `~/.scmd/config.yaml`:

```yaml
default_backend: ollama
default_model: llama3.2

backends:
  ollama:
    host: http://localhost:11434
  openai:
    model: gpt-4o-mini

ui:
  color: true
  spinner: true
```

## CLI Reference

```
scmd [command] [flags]

Commands:
  explain     Explain code or concepts
  review      Review code for issues
  config      View/modify configuration
  backends    List available backends

  slash       Slash command management
    run       Run a slash command
    list      List slash commands
    add       Add a slash command
    remove    Remove a slash command
    alias     Add an alias
    init      Generate shell integration
    interactive  Start REPL mode

  repo        Manage repositories
    add       Add a repository
    remove    Remove a repository
    list      List repositories
    update    Update manifests
    search    Search for commands
    show      Show command details
    install   Install a command

  registry    Central registry
    search    Search registry
    featured  Trending commands
    categories List categories

  update      Check for updates
  lock        Manage lockfiles
  cache       Manage local cache

Flags:
  -b, --backend   Backend to use
  -m, --model     Model to use
  -p, --prompt    Inline prompt
  -o, --output    Output file
  -f, --format    Output format (text, json, markdown)
  -q, --quiet     Suppress progress
  -v, --verbose   Verbose output
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `OLLAMA_HOST` | Ollama server URL |
| `OPENAI_API_KEY` | OpenAI API key |
| `TOGETHER_API_KEY` | Together.ai API key |
| `GROQ_API_KEY` | Groq API key |
| `SCMD_CONFIG` | Config file path |

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Creating Commands

1. Fork the [scmd-community/commands](https://github.com/scmd-community/commands) repo
2. Add your command YAML file
3. Update the manifest
4. Submit a PR

## License

MIT License - see [LICENSE](LICENSE) for details.

---

Built with Go. Inspired by the Unix philosophy and modern AI tooling.
