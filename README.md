# scmd

**AI-powered slash commands for any terminal. Works offline by default.**

scmd brings the power of LLM-based slash commands to your command line. Works offline by default with llama.cpp and Qwen models, or connect to Ollama, OpenAI, and more. Type `/gc` to generate commit messages, `/explain` to understand code, or install new commands from community repositories.

```bash
# Works immediately - no API keys or setup required:
./scmd /explain main.go        # Explain code
./scmd /cmd "find files modified today"  # Generate exact commands
./scmd /gc                      # Generate commit message from staged changes
./scmd /review                  # Review code for issues
git diff | ./scmd /sum          # Summarize changes

# Or use the scmd command directly:
cat main.go | scmd explain
scmd /cmd "search for TODO in all Go files"
git diff | scmd review
```

## 🛡️ Stability & Reliability First

**Core Design Principle**: scmd is designed to be **zero-maintenance** and **self-healing**.

### You Never Need to Manage the LLM Server

- ✅ **Auto-starts** - Server starts automatically when needed
- ✅ **Auto-restarts** - Crashes and failures handled automatically
- ✅ **Self-healing** - Detects issues (OOM, context mismatches) and recovers
- ✅ **Clear feedback** - Every error includes actionable solutions
- ✅ **No manual intervention** - Never need `pkill` or server management commands

### Intelligent Error Handling

When issues occur, scmd:
1. **Detects** the root cause (GPU memory, context size, server crash)
2. **Attempts auto-recovery** (restart server, reduce context, use CPU mode)
3. **Provides clear guidance** with exact commands when manual action needed

**Example:**
```
❌ Input exceeds available context size

What happened:
  Your input (5502 tokens) exceeds GPU memory capacity (4096 tokens)
  Metal GPU cannot allocate enough VRAM for the full context

Solutions:
  1. 💡 Use CPU mode (slower, supports full 32K context):
     export SCMD_CPU_ONLY=1
     scmd /explain <your-input>

  2. Split your input into smaller files

  3. Use cloud backend (fastest):
     export OPENAI_API_KEY=your-key
     scmd -b openai /explain <your-input>
```

Every error message tells you:
- **What went wrong** (in plain English)
- **What was tried** (transparency into auto-recovery attempts)
- **What you can do** (copy-paste solutions)

See [docs/architecture/STABILITY.md](docs/architecture/STABILITY.md) for the complete stability architecture.

## Features

- **Offline-First** - llama.cpp with local Qwen models, no API keys required
- **Interactive Conversations** - Multi-turn chat with context retention and searchable history
- **Beautiful Output** - Markdown rendering with syntax highlighting, multiple themes
- **Template System** - Customizable prompts for security, performance, documentation reviews
- **Interactive Setup Wizard** - Beautiful guided setup on first run (~2 minutes)
- **Man Page Integration** - `/cmd` reads man pages to generate exact commands
- **Smart Model Selection** - Choose Fast (0.5B), Balanced (1.5B), Best (3B), or Premium (7B)
- **Real Slash Commands** - Type `/command` directly (with or without shell integration)
- **Repository-First Architecture** - Commands install from repos like npm packages
- **Multiple LLM Backends** - llama.cpp (default), Ollama, OpenAI, Together.ai, Groq
- **Production-Grade Downloads** - Retry logic, resume support, disk space validation
- **Command Composition** - Chain commands in pipelines, run in parallel, or use fallbacks
- **Shell Integration** - Bash, Zsh, and Fish support with tab completion
- **Local Caching** - Commands and manifests cached locally
- **Lockfiles** - Reproducible installations for teams

## Architecture

scmd uses a **repository-first architecture** similar to package managers like npm or Homebrew:

- **Small Core**: Core commands built-in (`explain`, `review`, `cmd`), keeping the binary lean (~14MB)
- **Repository-Based**: Additional commands install from repositories (official or community)
- **Network Optional**: Core functionality works offline; network needed only for installing new commands
- **Decentralized**: Anyone can create and host command repositories
- **Version Management**: Commands have versions, dependencies, and lockfiles for reproducibility

**Example workflow:**
```bash
# Built-in command works immediately
scmd /explain main.go

# Install additional commands from repositories
scmd repo add official https://github.com/scmd/commands/raw/main
scmd repo install official/review
scmd /review code.py  # Now available
```

This design allows:
- ✅ Small binary size and fast installation
- ✅ Community-driven command ecosystem
- ✅ Easy command discovery and sharing
- ✅ Team-specific command repositories
- ✅ Reproducible environments with lockfiles

## Installation

### Quick Install

Choose the method that works best for you:

#### Homebrew (macOS/Linux)

```bash
brew install scmd/tap/scmd
```

#### npm (Cross-Platform)

```bash
npm install -g scmd-cli
```

#### Shell Script (wget/curl)

```bash
# Using curl
curl -fsSL https://scmd.sh/install.sh | bash

# Using wget
wget -qO- https://scmd.sh/install.sh | bash
```

#### Linux Packages

**Debian/Ubuntu:**
```bash
wget https://github.com/scmd/scmd/releases/latest/download/scmd_VERSION_linux_amd64.deb
sudo dpkg -i scmd_VERSION_linux_amd64.deb
```

**Fedora/RHEL:**
```bash
wget https://github.com/scmd/scmd/releases/latest/download/scmd_VERSION_linux_amd64.rpm
sudo rpm -i scmd_VERSION_linux_amd64.rpm
```

### Post-Installation

1. **Verify installation:**
   ```bash
   scmd --version
   ```

2. **Install llama.cpp for offline usage:**
   ```bash
   # macOS
   brew install llama.cpp

   # Linux - build from source
   # See: https://github.com/ggerganov/llama.cpp
   ```

3. **Try it out:**
   ```bash
   scmd /explain "what is a goroutine"
   ```

For detailed installation instructions, platform-specific guides, and troubleshooting, see [INSTALL.md](INSTALL.md).

### Build from Source

```bash
# Clone and build
git clone https://github.com/scmd/scmd
cd scmd
make build

# Install to /usr/local/bin
sudo make install

# Or build with Go directly
go install github.com/scmd/scmd/cmd/scmd@latest
```

## Model Management

scmd uses llama.cpp with efficient Qwen models for offline inference. Models are downloaded automatically on first use.

### Available Models

```bash
# List available models
scmd models list

# Output:
# NAME          SIZE      SPEED         QUALITY    DESCRIPTION
# qwen2.5-0.5b  379 MB    ⚡⚡⚡⚡       ⭐⭐⭐      Fastest - Basic tasks
# qwen2.5-1.5b  1.0 GB    ⚡⚡⚡        ⭐⭐⭐⭐     Balanced - Default (recommended)
# qwen2.5-3b    1.9 GB    ⚡⚡          ⭐⭐⭐⭐⭐   Best - Complex tasks
# qwen2.5-7b    3.8 GB    ⚡            ⭐⭐⭐⭐⭐   Premium - Highest quality
```

**Performance Benchmarks** (M1 Mac / 8GB RAM):

| Model | Avg Response | Tokens/sec | Use Case |
|-------|-------------|------------|----------|
| qwen2.5-0.5b | 3-5s | ~45 tok/s | Quick explanations, simple queries |
| qwen2.5-1.5b | 5-8s | ~30 tok/s | General purpose (default) |
| qwen2.5-3b | 8-12s | ~18 tok/s | Complex code review, detailed analysis |
| qwen2.5-7b | 15-25s | ~8 tok/s | Production code, critical decisions |

All models support:
- ✅ Tool calling and function use
- ✅ Context window: 8192 tokens (increased from 1024)
- ✅ GPU acceleration (Metal/CUDA)
- ✅ 4-bit quantization (Q4_K_M/Q3_K_M)

### Managing Models

```bash
# Download a specific model
scmd models pull qwen2.5-3b

# Show model info
scmd models info qwen3-4b

# Set default model
scmd models default qwen2.5-3b

# Remove a downloaded model
scmd models remove qwen2.5-7b
```

Models are stored in `~/.scmd/models/` and use GPU acceleration when available (Metal on macOS, CUDA on Linux).

## Quick Start

Get up and running with scmd in under 2 minutes:

### 1. First Run - Interactive Setup Wizard

On your first run, scmd launches a beautiful interactive setup wizard:

```bash
scmd /explain "what is docker"
```

The wizard guides you through:
1. **Model Selection** - Choose your preset:
   - ⚡ Fast (0.5B) - Lightning fast, basic tasks
   - ⚙️  Balanced (1.5B) - Recommended for most users
   - 🎯 Best (3B) - High quality, complex tasks
   - 💎 Premium (7B) - Maximum quality, slower

2. **Download** - Clean progress bar shows download status (happens once)
3. **Quick Test** - Optional test query to verify everything works

**Time to setup**: Under 2 minutes (including download on fast connection)

That's it! Your AI assistant is now ready, 100% offline and private.

**Features:**
- ✅ Production-grade downloads with retry logic
- ✅ Resume support for interrupted downloads
- ✅ Disk space validation before download
- ✅ Beautiful single-line progress indicator
- ✅ Post-setup quick test option

### 2. Start Using Commands

```bash
# Generate exact commands from natural language
scmd /cmd "how do I find files modified in the last 24 hours?"
scmd /cmd "search for TODO in all Go files"
scmd /cmd "compress a directory into tar.gz"
scmd /cmd "download a file from URL"

# Explain any code or concept
scmd /explain main.go
scmd /explain "what is a goroutine"
cat myfile.go | scmd explain

# Review code for issues
scmd /review main.go
git diff | scmd /review

# Install additional commands from repositories
scmd repo add official https://github.com/scmd/commands/raw/main
scmd repo install official/commit

# Use the installed commands
git diff --staged | scmd /gc  # Generate commit message

# Use with inline prompts
echo "SELECT * FROM users" | scmd -p "optimize this SQL query"

# Save output to file
git diff | scmd review -o review.md

# Use specific backend/model
scmd -b openai -m gpt-4 explain main.go
```

### 3. Explore More

```bash
# Discover commands
scmd repo search git
scmd repo search docker

# List available models
scmd models list

# Change your model anytime
scmd setup --force
```

**Pro Tip:** All models run 100% offline with GPU acceleration when available. No API keys needed!

## Interactive Chat Mode

Have extended AI conversations with full context retention. scmd remembers your entire chat history, letting you ask follow-up questions and build on previous discussions.

### Quick Start

```bash
# Start a new chat
scmd chat

# Use a specific model
scmd chat --model qwen2.5-7b

# Resume a previous conversation
scmd chat --continue abc123
```

### Example Session

```bash
$ scmd chat
💬 Conversation: a3f2b1c4
🔧 Model: qwen2.5-1.5b

You: How do I create a REST API in Go?
🤖 Assistant: I'll help you create a REST API in Go...

You: Can you show me how to add authentication?
🤖 Assistant: Of course! Building on the previous example...
[Full context from previous messages maintained]

You: What about rate limiting?
🤖 Assistant: [Continues with complete context...]

[Press Ctrl+D to exit]
👋 Conversation saved. Use 'scmd chat --continue a3f2b1c4' to resume.
```

### Managing Conversations

```bash
# List all conversations
scmd history list

# Example output:
#  1. [a3f2b1c4] How do I create a REST API...
#     Model: qwen2.5-1.5b | Messages: 6 | Updated: Jan 10, 14:23
#  2. [b7d3e5a1] Docker container basics
#     Model: qwen2.5-3b | Messages: 12 | Updated: Jan 09, 16:45

# Show conversation details
scmd history show a3f2b1c4

# Search conversations
scmd history search "docker"

# Delete a conversation
scmd history delete a3f2b1c4

# Clear all history
scmd history clear
```

### In-Chat Commands

While in a chat session:

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/clear` | Clear context (keeps history) |
| `/info` | Show conversation info |
| `/save` | Force save conversation |
| `/export` | Export to markdown |
| `/model <name>` | Switch model |
| `/exit` or `Ctrl+D` | Exit session |

### Features

- ✅ **Context Retention** - Full conversation history maintained
- ✅ **Auto-Save** - Every message saved automatically
- ✅ **Resume Anytime** - Pick up where you left off
- ✅ **Search History** - Find past conversations
- ✅ **Export to Markdown** - Share conversations
- ✅ **Model Switching** - Change models mid-conversation

Conversations stored in `~/.scmd/conversations.db` (SQLite).

## Slash Commands

The core feature of scmd is slash commands that work directly in your terminal.

### Direct Usage (No Setup Required)

You can use slash commands immediately without any shell integration:

```bash
# Direct invocation
./scmd /explain main.go
./scmd /review code.py
./scmd /gc
./scmd /e "what are channels?"

# With pipes
cat error.log | ./scmd /fix
git diff | ./scmd /gc
curl api.com/data | ./scmd /sum
```

### Setup Shell Integration (Optional)

For even better ergonomics, set up shell integration to use `/command` without the `./scmd` prefix:

```bash
# For Bash/Zsh - add to your ~/.bashrc or ~/.zshrc:
eval "$(scmd slash init bash)"

# For Fish - add to ~/.config/fish/config.fish:
scmd slash init fish | source
```

After setup, use slash commands directly:

```bash
/cmd "find files modified today"  # Generate exact commands
/explain main.go                   # Explain code
/review code.py                    # Review code
/gc                                # Generate commit message
/sum article.md                    # Summarize
/fix                               # Explain errors

# Pipe input to commands
cat error.log | /fix
git diff | /gc
curl api.com/data | /sum
```

### Built-in Commands

Core commands come built-in with scmd - others install from repositories:

| Command | Aliases | Description | Source |
|---------|---------|-------------|--------|
| `/cmd` | `/command`, `/howto` | Generate exact commands from questions | Built-in ✓ |
| `/explain` | `/e`, `/exp` | Explain code or concepts | Built-in ✓ |
| `/review` | `/r`, `/rev` | Review code for issues | Built-in ✓ |

**New: /cmd with Man Page Integration** 🔥

The `/cmd` command reads system man pages and generates exact commands:

```bash
# Ask natural language questions, get exact commands
scmd /cmd "how do I find files modified in the last 24 hours?"
# → find . -type f -mtime -1

scmd /cmd "search for TODO in all .go files"
# → find . -name "*.go" -exec grep -n "TODO" {} \;

scmd /cmd "compress directory into tar.gz"
# → tar -czf archive.tar.gz directory/

scmd /cmd "list processes sorted by memory"
# → ps aux --sort=-%mem | head -n 20
```

Features:
- ✅ Reads system man pages for accurate commands
- ✅ Detects relevant commands automatically (60+ common tools)
- ✅ Returns exact, copy-paste ready commands
- ✅ Includes clear explanations
- ✅ Falls back to general CLI knowledge when man pages unavailable

### Popular Community Commands

Install additional commands from the official repository:

```bash
# Install popular commands
scmd repo add official https://raw.githubusercontent.com/scmd/commands/main
scmd repo install official/review        # Code review
scmd repo install official/commit        # Git commit messages
scmd repo install official/summarize     # Summarize text
scmd repo install official/fix           # Explain and fix errors
```

This repository-first architecture keeps the scmd binary small while allowing the community to build and share commands.

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

scmd's repository system lets you distribute and install AI commands. Think Homebrew taps, but for AI prompts.

### Installing Commands

```bash
# Add a repository
scmd repo add community https://raw.githubusercontent.com/scmd-community/commands/main

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

scmd supports multiple LLM backends. llama.cpp is used by default for offline inference.

| Backend | Local | Free | Default | Setup |
|---------|-------|------|---------|-------|
| **llama.cpp** | ✓ | ✓ | ✓ | `brew install llama.cpp` |
| **Ollama** | ✓ | ✓ | | `ollama serve` |
| **OpenAI** | | | | `export OPENAI_API_KEY=...` |
| **Together.ai** | | Free tier | | `export TOGETHER_API_KEY=...` |
| **Groq** | | Free tier | | `export GROQ_API_KEY=...` |

### Backend Priority

Backends are tried in this order:
1. **llama.cpp** - Local, offline, no setup required (default)
2. **Ollama** - Local, if running
3. **OpenAI** - If API key set
4. **Together.ai** - If API key set
5. **Groq** - If API key set

### Using Backends

```bash
# Use specific backend
scmd -b ollama explain main.go

# Use specific model
scmd -b openai -m gpt-4 review code.py

# List available backends
scmd backends

# Example output:
#   ✓ llamacpp     qwen3-4b
#   ✗ ollama       qwen2.5-coder-1.5b
#   ✗ openai       (not configured)
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

## Template System

Customize prompts for specialized workflows. Templates standardize reviews for security, performance, documentation, and more.

### Quick Start

```bash
# List built-in templates
scmd template list

#  📋 security-review (v1.0)
#     OWASP Top 10 focused security analysis
#     Tags: security, owasp
#     Compatible: review, explain

#  📋 performance (v1.0)
#     Performance optimization and bottleneck analysis
#     Tags: performance, optimization
#     Compatible: review, explain

#  📋 beginner-explain (v1.0)
#     Explain code to beginners with simple language
#     Tags: education, beginner
#     Compatible: explain

# Use a template
scmd review auth.js --template security-review
scmd explain quicksort.py --template beginner-explain
```

### Built-in Templates

scmd includes 6 professional templates:

| Template | Use Case | Example |
|----------|----------|---------|
| **security-review** | OWASP Top 10, vulnerability scanning | `scmd review auth.js --template security-review` |
| **performance** | Bottlenecks, Big O analysis | `scmd review algorithm.py --template performance` |
| **api-design** | REST best practices, HTTP methods | `scmd review api.go --template api-design` |
| **testing** | Test coverage, edge cases | `scmd review service.ts --template testing` |
| **documentation** | Doc generation and review | `scmd explain utils.rs --template documentation` |
| **beginner-explain** | Beginner-friendly explanations | `scmd explain recursion.py --template beginner-explain` |

### Template Details

```bash
# View template details
scmd template show security-review

#  📋 Template: security-review (v1.0)
#
#  Author: scmd
#  Description: OWASP Top 10 focused security analysis
#
#  Tags: security, owasp, review
#  Compatible Commands: review, explain
#
#  Variables:
#    - Language: Programming language (auto-detect)
#    - Code: Code to review (required)
#    - Context: Additional context
#
#  Recommended Models: qwen2.5-7b, gpt-4
#
#  Examples:
#    Review authentication code
#    $ scmd review login.js --template security-review

# Search templates
scmd template search security

# Export template for sharing
scmd template export security-review > security.yaml
```

### Creating Custom Templates

Templates are YAML files with a simple structure:

```yaml
name: team-standards
version: "1.0"
author: "Your Team"
description: "Team coding standards review"
tags:
  - team
  - standards
compatible_commands:
  - review

system_prompt: |
  You are a code reviewer for our team.
  Focus on our coding standards and best practices.

user_prompt_template: |
  Review this {{.Language}} code:

  ```{{.Language}}
  {{.Code}}
  ```

  Check for:
  1. Team coding standards
  2. Error handling patterns
  3. Code clarity

variables:
  - name: Language
    description: "Programming language"
    default: "auto-detect"
  - name: Code
    description: "Code to review"
    required: true

recommended_models:
  - qwen2.5-7b
```

### Template Management

```bash
# Create interactive template
scmd template create my-template

# Import from file
scmd template import team-standards.yaml

# Delete template
scmd template delete my-template

# Export for sharing
scmd template export team-standards > share.yaml
```

### Advanced Usage

```bash
# Use in chat sessions
scmd chat
You: Review this auth code using the security-review template
[Detailed security analysis with OWASP focus]

# Pipe input with templates
git diff | scmd review --template security-review

# Combine with other features
scmd review --template performance --model qwen2.5-7b --format json
```

Templates stored in `~/.scmd/templates/` as YAML files - easily version controlled and shared with teams.

## Configuration

Configuration is stored in `~/.scmd/config.yaml`:

```yaml
default_backend: llamacpp
default_model: qwen2.5-1.5b

backends:
  llamacpp:
    model: qwen2.5-1.5b
  ollama:
    host: http://localhost:11434
  openai:
    model: gpt-4o-mini

# Chat configuration (v0.2.0+)
chat:
  max_context_messages: 20    # Max messages in context
  auto_save: true             # Auto-save after each message
  auto_title: true            # Auto-generate titles

# UI configuration (v0.2.0+)
ui:
  colors: true                # Enable colored output
  style: auto                 # Theme: auto, dark, light, notty
  streaming: true             # Enable streaming output
  verbose: false              # Verbose mode

# Template configuration (v0.2.0+)
templates:
  directory: ~/.scmd/templates  # Template storage
```

## CLI Reference

```
scmd [command] [flags]

Commands:
  explain     Explain code or concepts
  review      Review code for issues
  config      View/modify configuration
  backends    List available backends

  chat        Start interactive conversation         [v0.2.0+]
    --continue <id>  Resume conversation
    --model <name>   Use specific model

  history     Manage conversation history             [v0.2.0+]
    list      List recent conversations
    show      Show conversation details
    search    Search conversations
    delete    Delete a conversation
    clear     Clear all conversations

  template    Manage prompt templates                 [v0.2.0+]
    list      List available templates
    show      Show template details
    create    Create custom template
    delete    Delete template
    search    Search templates
    export    Export template to YAML
    import    Import template from file

  models      Manage local LLM models
    list      List available models
    pull      Download a model
    remove    Remove a model
    info      Show model information
    default   Set default model

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
      --template  Use a prompt template (for explain/review)  [v0.2.0+]
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `OLLAMA_HOST` | Ollama server URL (default: http://localhost:11434) |
| `OPENAI_API_KEY` | OpenAI API key |
| `TOGETHER_API_KEY` | Together.ai API key |
| `GROQ_API_KEY` | Groq API key |
| `SCMD_CONFIG` | Config file path (default: ~/.scmd/config.yaml) |
| `SCMD_DATA_DIR` | Data directory (default: ~/.scmd) |
| `SCMD_DEBUG` | Enable debug logging (set to 1) |

## Performance

### Benchmarks (M1 Mac, 8GB RAM)

**Real-World Response Times:**

| Task | qwen2.5-0.5b | qwen2.5-1.5b | qwen2.5-3b | qwen2.5-7b |
|------|-------------|-------------|-----------|-----------|
| Explain 50-line file | 3.2s | 5.8s | 9.1s | 16.3s |
| Generate git commit | 2.8s | 4.9s | 7.5s | 14.1s |
| Review 200-line file | 6.5s | 11.2s | 18.7s | 32.4s |
| Generate CLI command | 2.1s | 3.4s | 5.8s | 10.2s |

**Inference Speed:**

| Model | CPU (tok/s) | GPU Metal (tok/s) | Quality Score |
|-------|------------|------------------|---------------|
| qwen2.5-0.5b | ~25 | ~60 | ⭐⭐⭐ (Good) |
| qwen2.5-1.5b | ~18 | ~45 | ⭐⭐⭐⭐ (Excellent) |
| qwen2.5-3b | ~12 | ~28 | ⭐⭐⭐⭐⭐ (Outstanding) |
| qwen2.5-7b | ~5 | ~12 | ⭐⭐⭐⭐⭐ (Best) |

**Optimizations:**
- ✅ 4-bit quantization (Q4_K_M/Q3_K_M) for optimal size/quality
- ✅ Context size: 8192 tokens (increased from 1024)
- ✅ Flash attention enabled for faster processing
- ✅ Continuous batching for multiple requests
- ✅ Memory locking for consistent performance
- ✅ KV cache type optimization (F16)

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
