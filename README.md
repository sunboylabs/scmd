<div align="center">

# scmd

**AI-powered slash commands for any terminal. Works offline by default.**

[![Release](https://img.shields.io/github/v/release/sunboylabs/scmd?color=blue)](https://github.com/sunboylabs/scmd/releases)
[![Build](https://img.shields.io/github/actions/workflow/status/sunboylabs/scmd/release.yml?branch=main)](https://github.com/sunboylabs/scmd/actions)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Go Report](https://goreportcard.com/badge/github.com/sunboylabs/scmd)](https://goreportcard.com/report/github.com/sunboylabs/scmd)

[Features](#-features) • [Installation](#-installation) • [Quick Start](#-quick-start) • [Documentation](#-documentation)

</div>

---

## What is scmd?

**scmd** brings AI superpowers to your terminal—offline, private, and fast. Ask questions in plain English, get exact commands. Review code with security templates. Chat with AI that remembers context. All without API keys or cloud dependencies.

```bash
# Just ask what you want to do
scmd /cmd "find files modified today"
# → find . -type f -mtime -1

# Get instant explanations
scmd /explain main.go

# Review with professional templates
scmd /review auth.js --template security-review

# Chat with full context
scmd chat
You: How do I set up OAuth2 in Go?
🤖 Assistant: [Detailed explanation...]
You: Show me an example with JWT
🤖 Assistant: [Builds on previous context...]
```

**No setup required.** First run downloads everything automatically (~2 min). Works 100% offline after that.

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 🔒 **Privacy First**
- **100% offline** after initial setup
- **Local LLMs** via llama.cpp + Qwen models
- **No telemetry**, no cloud, your code stays yours
- Optional cloud backends (OpenAI, Groq, etc.)

### 💬 **Smart Conversations**
- **Multi-turn chat** with context retention
- **Searchable history** - find past discussions
- **Auto-save** - never lose a conversation
- **Export to markdown** for sharing

</td>
<td width="50%">

### 🎨 **Beautiful Output**
- **Syntax highlighting** for 40+ languages
- **Markdown rendering** with themes
- **Auto-detection** - plain text when piped
- **NO_COLOR** support for accessibility

### 🚀 **Fast & Light**
- **14MB binary** - no bloat
- **0.5B-7B models** - choose speed vs quality
- **GPU acceleration** (Metal/CUDA)
- **Streaming output** - see results instantly

</td>
</tr>
</table>

### 📚 **Man Page Integration**

The `/cmd` command reads your system's man pages to generate **exact, copy-paste ready commands**:

```bash
scmd /cmd "compress directory into tar.gz"
# → tar -czf archive.tar.gz directory/

scmd /cmd "list processes sorted by memory"
# → ps aux --sort=-%mem | head -n 20
```

Detects 60+ common tools automatically. Falls back to general CLI knowledge when needed.

### 🔐 **Security Templates**

Professional security reviews with OWASP Top 10 focus:

```bash
scmd review auth.js --template security-review

# Returns:
# 🔴 CRITICAL: SQL Injection vulnerability (CWE-89)
# 🟡 MEDIUM: Missing input validation
# 🟢 INFO: Consider using prepared statements
# [... with code examples and fixes]
```

Six built-in templates: `security-review`, `performance`, `api-design`, `testing`, `documentation`, `beginner-explain`.

### 📦 **Repository System**

Commands install like npm packages—small core, unlimited extensions:

```bash
# Discover 100+ community commands
scmd repo search docker
scmd repo install official/compose

# Create your own repos
# Share with your team
# Version control with lockfiles
```

---

## 🚀 Installation

<details>
<summary><b>Homebrew</b> (macOS/Linux) — Recommended ⭐</summary>

```bash
brew install sunboylabs/tap/scmd
```

Installs scmd + shell completions. Ready to go!

</details>

<details>
<summary><b>npm</b> (Cross-platform)</summary>

```bash
npm install -g scmd-cli
```

Works on macOS, Linux, and Windows.

</details>

<details>
<summary><b>Shell Script</b> (wget/curl)</summary>

```bash
# Using curl
curl -fsSL https://scmd.sh/install.sh | bash

# Using wget
wget -qO- https://scmd.sh/install.sh | bash
```

Automatically detects platform and installs.

</details>

<details>
<summary><b>Linux Packages</b> (deb/rpm)</summary>

**Debian/Ubuntu:**
```bash
wget https://github.com/sunboylabs/scmd/releases/latest/download/scmd_VERSION_linux_amd64.deb
sudo dpkg -i scmd_VERSION_linux_amd64.deb
```

**Fedora/RHEL/CentOS:**
```bash
wget https://github.com/sunboylabs/scmd/releases/latest/download/scmd_VERSION_linux_amd64.rpm
sudo rpm -i scmd_VERSION_linux_amd64.rpm
```

</details>

<details>
<summary><b>Go Install</b> (Build from source)</summary>

```bash
go install github.com/sunboylabs/scmd/cmd/scmd@latest
```

Requires Go 1.21+ and CGO for SQLite support.

</details>

<details>
<summary><b>Manual Download</b></summary>

Download binaries from [GitHub Releases](https://github.com/sunboylabs/scmd/releases) for your platform. Extract and add to PATH.

</details>

### First Run

```bash
scmd /explain "what is docker"
```

**Beautiful setup wizard appears:**
1. Choose model preset (Fast/Balanced/Best/Premium)
2. Watch download progress (~1-4GB depending on model)
3. Optional quick test
4. Done! (~2 minutes total)

All models run **100% offline** with GPU acceleration when available.

---

## ⚡ Quick Start

### 1. Generate Commands from Natural Language

The killer feature—ask what you want, get exact commands:

```bash
# Natural language → exact command
scmd /cmd "find files modified in last 24 hours"
# → find . -type f -mtime -1

scmd /cmd "search for TODO in all Go files"
# → find . -name "*.go" -exec grep -n "TODO" {} \;

scmd /cmd "download file from URL"
# → curl -O https://example.com/file.zip
```

**Pro tip:** Works with 60+ common CLI tools. Reads man pages for accuracy.

### 2. Explain Code or Concepts

```bash
# Explain code
scmd /explain main.go
cat algorithm.py | scmd explain

# Explain concepts
scmd /explain "what are Go channels?"

# Beginner mode
scmd /explain quicksort.py --template beginner-explain
```

**Beautiful markdown output** with syntax highlighting and clear structure.

### 3. Review Code with Templates

```bash
# Basic review
scmd /review code.py

# Security-focused review (OWASP Top 10)
scmd /review auth.js --template security-review

# Performance optimization review
scmd /review algorithm.go --template performance

# Check test coverage
scmd /review service_test.ts --template testing
```

**Professional reports** with severity levels and actionable fixes.

### 4. Chat with Context

```bash
# Start conversation
scmd chat

You: How do I implement rate limiting in Express?
🤖 Assistant: [Detailed explanation with code]

You: What about Redis-based rate limiting?
🤖 Assistant: [Builds on previous context...]

You: Show me middleware example
🤖 Assistant: [Complete working example...]

/export  # Save to markdown
```

**Resume anytime:**
```bash
scmd history list
scmd chat --continue abc123
```

### 5. Install Community Commands

```bash
# Add official repo (100+ commands)
scmd repo add official https://raw.githubusercontent.com/sunboylabs/commands/main

# Discover commands
scmd repo search git
scmd repo search docker

# Install
scmd repo install official/commit
scmd repo install official/dockerfile

# Use
git diff --staged | scmd /gc  # Generate commit message
```

**Share your own commands** by creating a repo. It's just YAML files!

---

## 📖 Documentation

<details>
<summary><b>Model Management</b></summary>

### Available Models

```bash
scmd models list
```

| Model | Size | Speed | Quality | Best For |
|-------|------|-------|---------|----------|
| **qwen2.5-0.5b** | 379 MB | ⚡⚡⚡⚡ | ⭐⭐⭐ | Quick tasks, fast iteration |
| **qwen2.5-1.5b** | 1.0 GB | ⚡⚡⚡ | ⭐⭐⭐⭐ | Daily work (default) ⭐ |
| **qwen2.5-3b** | 1.9 GB | ⚡⚡ | ⭐⭐⭐⭐⭐ | Complex analysis |
| **qwen2.5-7b** | 3.8 GB | ⚡ | ⭐⭐⭐⭐⭐ | Production code reviews |

**Performance** (M1 Mac, 8GB RAM):
- **0.5B**: ~45 tok/s, 3-5s average response
- **1.5B**: ~30 tok/s, 5-8s average response (recommended)
- **3B**: ~18 tok/s, 8-12s average response
- **7B**: ~8 tok/s, 15-25s average response

All models support:
- ✅ 8192 token context (increased from 1024)
- ✅ GPU acceleration (Metal/CUDA)
- ✅ 4-bit quantization for efficiency
- ✅ Function calling

### Commands

```bash
# Download model
scmd models pull qwen2.5-3b

# Set default
scmd models default qwen2.5-3b

# Model info
scmd models info qwen2.5-1.5b

# Remove model
scmd models remove qwen2.5-7b

# Switch on the fly
scmd --model qwen2.5-7b /review critical.go
```

**Storage:** Models stored in `~/.scmd/models/`

</details>

<details>
<summary><b>Template System</b></summary>

### Built-in Templates

Templates customize prompts for specialized workflows:

| Template | Focus | Example |
|----------|-------|---------|
| **security-review** | OWASP Top 10, vulnerabilities | `scmd review auth.js --template security-review` |
| **performance** | Bottlenecks, Big O analysis | `scmd review sort.py --template performance` |
| **api-design** | REST best practices | `scmd review api.go --template api-design` |
| **testing** | Coverage, edge cases | `scmd review service.ts --template testing` |
| **documentation** | Doc generation | `scmd explain utils.rs --template documentation` |
| **beginner-explain** | ELI5 explanations | `scmd explain recursion.go --template beginner-explain` |

### Template Management

```bash
# List templates
scmd template list

# View details
scmd template show security-review

# Search
scmd template search security

# Create custom
scmd template create team-standards

# Export for sharing
scmd template export team-standards > team.yaml

# Import team template
scmd template import team.yaml
```

### Creating Templates

Templates are simple YAML files:

```yaml
name: team-standards
version: "1.0"
description: "Review against team coding standards"
tags:
  - team
  - standards
compatible_commands:
  - review

system_prompt: |
  You are a code reviewer for our team.
  Focus on our specific coding standards.

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
    default: "auto-detect"
  - name: Code
    required: true

recommended_models:
  - qwen2.5-7b
```

**Storage:** Templates in `~/.scmd/templates/` - version control and share with teams.

</details>

<details>
<summary><b>Repository System</b></summary>

### How It Works

scmd uses a **repository-first architecture** like npm or Homebrew:

- **Small core**: Built-in commands keep binary lean (~14MB)
- **Repository-based**: Additional commands install from repos
- **Decentralized**: Anyone can host repos
- **Version control**: Lockfiles for reproducibility

### Managing Repositories

```bash
# Add repository
scmd repo add official https://raw.githubusercontent.com/sunboylabs/commands/main
scmd repo add team https://github.com/myteam/scmd-commands/raw/main

# List repos
scmd repo list

# Update manifests
scmd repo update

# Remove repo
scmd repo remove team
```

### Discovering Commands

```bash
# Search all repos
scmd repo search git
scmd repo search docker

# Show command details
scmd repo show official/commit

# Install command
scmd repo install official/commit
scmd repo install official/review
scmd repo install official/dockerfile
```

### Creating a Repository

Create your own command repo:

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
    category: utility
```

**commands/my-command.yaml:**
```yaml
name: my-command
version: "1.0.0"
description: Generate optimized Dockerfile
usage: "my-command [options]"

args:
  - name: language
    description: Programming language
    default: "nodejs"

prompt:
  system: "You are an expert at what you do."
  template: |
    Do something useful with {{.language}}

    Input: {{.stdin}}

model:
  temperature: 0.3
  max_tokens: 512
```

Host on GitHub, GitLab, or any HTTP server:

```bash
scmd repo add myrepo https://raw.githubusercontent.com/you/my-commands/main
```

### Lockfiles

Share exact command versions with your team:

```bash
# Generate lockfile
scmd lock generate

# Commit to repo
git add scmd-lock.yaml
git commit -m "Add scmd lockfile"

# Team members install
scmd lock install

# Check for updates
scmd update --check
```

</details>

<details>
<summary><b>LLM Backends</b></summary>

### Supported Backends

| Backend | Local | Free | Setup |
|---------|-------|------|-------|
| **llama.cpp** | ✓ | ✓ | Default - no setup needed |
| **Ollama** | ✓ | ✓ | `ollama serve` |
| **OpenAI** | | | `export OPENAI_API_KEY=...` |
| **Together.ai** | | Free tier | `export TOGETHER_API_KEY=...` |
| **Groq** | | Free tier | `export GROQ_API_KEY=...` |

### Backend Priority

Backends tried in this order:
1. **llama.cpp** - Local, offline (default)
2. **Ollama** - If running
3. **OpenAI** - If API key set
4. **Together.ai** - If API key set
5. **Groq** - If API key set

### Using Backends

```bash
# Use specific backend
scmd -b ollama /explain main.go
scmd -b openai -m gpt-4 /review code.py

# List backends
scmd backends

# Example output:
#   ✓ llamacpp     qwen2.5-1.5b (active)
#   ✓ ollama       qwen2.5-coder-1.5b
#   ✗ openai       (not configured)
#   ✗ groq         (not configured)
```

**Pro tip:** llama.cpp is fast enough for most tasks. Use cloud backends for maximum quality or specialized models.

</details>

<details>
<summary><b>Conversation History</b></summary>

### Managing Conversations

All chat sessions are automatically saved to `~/.scmd/conversations.db`:

```bash
# List conversations
scmd history list

# Example output:
#  1. [a3f2b1c4] How do I create a REST API in Go?
#     Model: qwen2.5-1.5b | Messages: 6 | Jan 10, 14:23
#
#  2. [b7d3e5a1] Docker container optimization
#     Model: qwen2.5-3b | Messages: 12 | Jan 09, 16:45

# Show full conversation
scmd history show a3f2b1c4

# Search history
scmd history search "docker"
scmd history search "authentication"

# Delete conversation
scmd history delete a3f2b1c4

# Clear all history
scmd history clear
```

### Resuming Conversations

```bash
# Resume with full ID
scmd chat --continue a3f2b1c4

# Or partial ID (matches first)
scmd chat --continue a3f2

# Lists matches if ambiguous
scmd chat --continue a3
# → Found multiple matches:
#   - a3f2b1c4: REST API discussion
#   - a37d4e2a: Docker setup
#   Please use more specific ID
```

### Exporting

```bash
# Export from within chat
scmd chat --continue a3f2b1c4
/export
# ✓ Exported to conversation_a3f2b1c4.md

# Export with custom filename
/export my-conversation.md
```

**Markdown format** includes timestamps, model info, and formatted code blocks.

</details>

<details>
<summary><b>Configuration</b></summary>

### Config File Location

Configuration stored in `~/.scmd/config.yaml`:

```yaml
# Default backend and model
default_backend: llamacpp
default_model: qwen2.5-1.5b

# Backend settings
backends:
  llamacpp:
    model: qwen2.5-1.5b
    context_size: 8192
  ollama:
    host: http://localhost:11434
    model: qwen2.5-coder-1.5b
  openai:
    model: gpt-4o-mini
    max_tokens: 4096

# Chat settings
chat:
  max_context_messages: 20    # Messages kept in context
  auto_save: true             # Auto-save after each message
  auto_title: true            # Auto-generate conversation titles

# UI settings
ui:
  format: auto                # auto, markdown, plain
  theme: auto                 # auto, dark, light
  word_wrap: 0                # 0 = terminal width
  streaming: true             # Stream responses
  verbose: false              # Verbose output

# Template settings
templates:
  directory: ~/.scmd/templates

# Repository settings
repos:
  cache_ttl: 3600            # Cache TTL in seconds
```

### Environment Variables

Override config with environment variables:

| Variable | Description |
|----------|-------------|
| `SCMD_CONFIG` | Config file path |
| `SCMD_DATA_DIR` | Data directory (default: `~/.scmd`) |
| `SCMD_DEBUG` | Enable debug logging (set to `1`) |
| `SCMD_CPU_ONLY` | Force CPU-only mode (set to `1`) |
| `SCMD_THEME` | Override theme (dark/light/auto) |
| `NO_COLOR` | Disable colored output (standard) |
| `OLLAMA_HOST` | Ollama server URL |
| `OPENAI_API_KEY` | OpenAI API key |
| `TOGETHER_API_KEY` | Together.ai API key |
| `GROQ_API_KEY` | Groq API key |

### Viewing Config

```bash
# Show current config
scmd config

# Edit config file
$EDITOR ~/.scmd/config.yaml
```

</details>

<details>
<summary><b>Shell Integration</b></summary>

### Setup

Shell integration allows using `/command` directly without `scmd` prefix:

**Bash/Zsh** - add to `~/.bashrc` or `~/.zshrc`:
```bash
eval "$(scmd slash init bash)"
```

**Fish** - add to `~/.config/fish/config.fish`:
```fish
scmd slash init fish | source
```

**Restart shell** or source config file.

### Using Slash Commands

After setup:

```bash
# Direct slash commands
/cmd "find large files"
/explain algorithm.py
/review auth.js --template security-review
/gc  # Generate commit message

# With pipes
cat error.log | /fix
git diff | /gc
curl api.com/data | /sum
```

**Without shell integration**, use `scmd` prefix:

```bash
scmd /cmd "find large files"
cat error.log | scmd /fix
```

Both work identically—shell integration is just convenience.

### Managing Slash Commands

```bash
# List all slash commands
scmd slash list

# Add custom command
scmd slash add mycommand custom-cmd --alias=mc

# Add alias to existing
scmd slash alias explain exp

# Remove command
scmd slash remove mycommand

# Interactive REPL mode
scmd slash interactive
```

</details>

<details>
<summary><b>Advanced Usage</b></summary>

### Command Composition

Chain commands in pipelines:

```bash
# Pipe through multiple commands
git diff | scmd /review | scmd /sum

# Save output
scmd /review code.py -o review.md

# JSON output
scmd /explain main.go -f json > result.json
```

### Custom Prompts

Override prompts inline:

```bash
# Custom prompt
echo "SELECT * FROM users" | scmd -p "optimize this SQL query"

# With specific backend/model
scmd -b openai -m gpt-4 -p "explain design patterns" main.go

# Combine with templates
scmd -p "focus on error handling" /review --template security-review
```

### Batch Processing

Process multiple files:

```bash
# Review all Python files
for file in *.py; do
  scmd /review "$file" -o "reviews/${file%.py}_review.md"
done

# Generate commit for each component
for dir in services/*/; do
  cd "$dir"
  git diff | scmd /gc >> ../../commits.txt
  cd -
done
```

### Performance Tips

1. **Choose the right model:**
   - Quick tasks → `qwen2.5-0.5b`
   - General use → `qwen2.5-1.5b`
   - Complex analysis → `qwen2.5-3b`
   - Production code → `qwen2.5-7b`

2. **Use GPU acceleration:**
   - Enabled by default on macOS (Metal)
   - Install CUDA drivers on Linux
   - 2-3x faster than CPU-only

3. **Reduce context size:**
   - Pass specific files, not entire directories
   - Use `--quiet` to reduce output overhead
   - Stream mode shows results faster

4. **Cache aggressively:**
   - Repo manifests cached for 1 hour
   - Models cached permanently
   - Templates loaded on startup

</details>

---

## 🛡️ Stability & Reliability

**Core Design Principle**: scmd is **zero-maintenance** and **self-healing**.

### You Never Manage the LLM Server

- ✅ **Auto-starts** - Server starts automatically when needed
- ✅ **Auto-restarts** - Crashes handled gracefully
- ✅ **Self-healing** - Detects issues (OOM, context mismatches) and recovers
- ✅ **Clear feedback** - Every error includes actionable solutions
- ✅ **No manual intervention** - Never need `pkill` or restart commands

### Intelligent Error Handling

When issues occur, scmd:
1. **Detects** root cause (GPU memory, context size, server crash)
2. **Attempts auto-recovery** (restart server, reduce context, CPU fallback)
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

  2. Split input into smaller files

  3. Use cloud backend (fastest):
     export OPENAI_API_KEY=your-key
     scmd -b openai /explain <your-input>
```

Every error tells you:
- **What went wrong** (plain English)
- **What was tried** (transparency)
- **What you can do** (copy-paste solutions)

See [docs/architecture/STABILITY.md](docs/architecture/STABILITY.md) for complete stability architecture.

---

## 📊 Performance

### Real-World Benchmarks (M1 Mac, 8GB RAM)

**Response Times:**

| Task | qwen2.5-0.5b | qwen2.5-1.5b | qwen2.5-3b | qwen2.5-7b |
|------|-------------|-------------|-----------|-----------|
| Explain 50-line file | 3.2s | 5.8s | 9.1s | 16.3s |
| Generate commit msg | 2.8s | 4.9s | 7.5s | 14.1s |
| Review 200-line file | 6.5s | 11.2s | 18.7s | 32.4s |
| Generate CLI command | 2.1s | 3.4s | 5.8s | 10.2s |

**Inference Speed:**

| Model | CPU (tok/s) | GPU Metal (tok/s) | Quality |
|-------|------------|------------------|---------|
| qwen2.5-0.5b | ~25 | ~60 | ⭐⭐⭐ |
| qwen2.5-1.5b | ~18 | ~45 | ⭐⭐⭐⭐ |
| qwen2.5-3b | ~12 | ~28 | ⭐⭐⭐⭐⭐ |
| qwen2.5-7b | ~5 | ~12 | ⭐⭐⭐⭐⭐ |

**Optimizations:**
- 4-bit quantization (Q4_K_M/Q3_K_M)
- 8192 token context window
- Flash attention for speed
- Memory locking for consistency
- KV cache optimization (F16)

---

## 🤝 Contributing

Contributions welcome! Here's how to get involved:

### Contributing Code

1. Fork the repo
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Commit with conventional commits (`feat: add amazing feature`)
6. Push and open a PR

### Creating Commands

1. Fork [sunboylabs/commands](https://github.com/sunboylabs/commands)
2. Add your command YAML file
3. Update the manifest
4. Submit a PR

### Creating Templates

Share your templates with the community:

```bash
# Create template
scmd template create my-template

# Export
scmd template export my-template > my-template.yaml

# Share on GitHub, in issues, or PRs
```

### Reporting Issues

Found a bug? Have a feature request?

- [Open an issue](https://github.com/sunboylabs/scmd/issues)
- Include: OS, scmd version, model, command that failed
- Paste error messages
- Describe expected vs actual behavior

### Documentation

Help improve docs:

- Fix typos
- Add examples
- Clarify confusing sections
- Translate to other languages

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## 📜 License

MIT License - see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgments

- **llama.cpp** - High-performance LLM inference
- **Qwen** - Efficient, high-quality language models
- **Charm** - Beautiful terminal UI components
- **Go community** - Amazing ecosystem

---

## 🔗 Links

- **GitHub**: [sunboylabs/scmd](https://github.com/sunboylabs/scmd)
- **Releases**: [Latest releases](https://github.com/sunboylabs/scmd/releases)
- **Commands Repo**: [sunboylabs/commands](https://github.com/sunboylabs/commands)
- **Changelog**: [CHANGELOG.md](CHANGELOG.md)
- **Installation Guide**: [INSTALL.md](INSTALL.md)

---

<div align="center">

**Built with ❤️ using Go**

*Inspired by the Unix philosophy and modern AI tooling*

[⬆ Back to Top](#scmd)

</div>
