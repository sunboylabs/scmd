# Welcome to scmd

**AI-powered slash commands for any terminal. Works offline by default.**

scmd brings the power of LLM-based slash commands to your command line. Type `/explain` to understand code, `/gc` to generate commit messages, or create your own AI-powered commands.

<div class="grid cards" markdown>

-   :material-rocket-launch:{ .lg .middle } __Quick Start__

    ---

    Get up and running in minutes with our quick start guide

    [:octicons-arrow-right-24: Getting Started](getting-started/quick-start.md)

-   :material-book-open-variant:{ .lg .middle } __User Guide__

    ---

    Learn how to use slash commands, manage models, and configure backends

    [:octicons-arrow-right-24: User Guide](user-guide/slash-commands.md)

-   :material-code-braces:{ .lg .middle } __Create Commands__

    ---

    Author powerful AI commands with tools, hooks, and composition

    [:octicons-arrow-right-24: Command Authoring](command-authoring/overview.md)

-   :material-github:{ .lg .middle } __Contribute__

    ---

    Join the development and help build the future of AI commands

    [:octicons-arrow-right-24: Contributing](contributing/development-setup.md)

</div>

## Features

### :fontawesome-solid-bolt: Works Offline

Default llama.cpp backend with local Qwen models. No API keys, no internet required for basic usage.

```bash
./scmd /explain "what is a goroutine"
# Works immediately - model downloads automatically on first run
```

### :fontawesome-solid-terminal: Real Slash Commands

Type `/command` directly in your terminal, with or without shell integration.

```bash
# Direct usage - no setup required
./scmd /explain main.go
./scmd /gc  # Generate commit message

# With shell integration
/explain main.go
cat error.log | /fix
```

### :fontawesome-solid-wand-magic-sparkles: Tool Calling (Agentic Behavior)

LLM can execute shell commands, read files, fetch URLs, and more - automatically.

```yaml
# Commands can use tools for autonomous actions
name: project-analyzer
prompt:
  template: |
    Analyze the project structure. Use tools to:
    1. List files
    2. Read key configuration files
    3. Provide recommendations
```

### :fontawesome-solid-link: Command Composition

Chain commands in pipelines, run in parallel, or use fallbacks.

```yaml
compose:
  pipeline:
    - command: analyze-code
    - command: summarize
      args:
        format: bullet-points
```

### :fontawesome-solid-clock: Pre/Post Hooks

Run shell commands before and after LLM execution.

```yaml
hooks:
  pre:
    - shell: git status --short
  post:
    - shell: git diff --stat
```

### :fontawesome-solid-boxes-stacked: Repository System

Install commands from community repositories. Think Homebrew taps, but for AI prompts.

```bash
scmd repo add community https://github.com/scmd-community/commands
scmd repo install community/git-commit
```

## Quick Example

```bash
# Install scmd
git clone https://github.com/scmd/scmd
cd scmd
go build -o scmd ./cmd/scmd

# Install llama-server for offline inference
brew install llama.cpp  # macOS

# Use it immediately
./scmd /explain main.go
git diff | ./scmd /gc
cat error.log | ./scmd /fix
```

## Supported LLM Backends

| Backend | Local | Free | Setup |
|---------|-------|------|-------|
| **llama.cpp** ⭐ | ✓ | ✓ | `brew install llama.cpp` |
| **Ollama** | ✓ | ✓ | `ollama serve` |
| **OpenAI** | ✗ | ✗ | `export OPENAI_API_KEY=...` |
| **Together.ai** | ✗ | Free tier | `export TOGETHER_API_KEY=...` |
| **Groq** | ✗ | Free tier | `export GROQ_API_KEY=...` |

## Available Models

| Model | Size | Speed | Quality | Use Case |
|-------|------|-------|---------|----------|
| qwen2.5-0.5b | 379 MB | ⚡⚡⚡ | ⭐⭐ | Quick queries |
| qwen2.5-1.5b | 940 MB | ⚡⚡ | ⭐⭐⭐ | Fast, lightweight |
| qwen2.5-3b | 1.9 GB | ⚡⚡ | ⭐⭐⭐⭐ | Good balance |
| **qwen3-4b** ⭐ | 2.5 GB | ⚡ | ⭐⭐⭐⭐⭐ | Default (tool calling) |
| qwen2.5-7b | 4.4 GB | ⚡ | ⭐⭐⭐⭐⭐ | Best quality |

## What's Next?

<div class="grid cards" markdown>

-   :material-download: [**Install scmd**](getting-started/installation.md)

    Get scmd installed on your system

-   :material-play-circle: [**Quick Start Tutorial**](getting-started/quick-start.md)

    5-minute guide to using slash commands

-   :material-puzzle: [**Create Your First Command**](getting-started/first-command.md)

    Build a custom AI command

-   :material-book-multiple: [**Explore Examples**](examples/basic-commands.md)

    Learn from real-world examples

</div>

## Why scmd?

| Traditional AI Tools | scmd |
|---------------------|------|
| API keys required | Works offline by default |
| Web interfaces | Native terminal integration |
| Fixed prompts | Customizable command specifications |
| Isolated tools | Repository system for sharing |
| Text generation only | Tool calling for actions |
| | Hook system for automation |
| | Command composition |

## Community

- **GitHub**: [scmd/scmd](https://github.com/scmd/scmd)
- **Issues**: [Report bugs or request features](https://github.com/scmd/scmd/issues)
- **Discussions**: [Ask questions and share commands](https://github.com/scmd/scmd/discussions)
- **Command Registry**: Browse and share commands

## License

scmd is open source software licensed under the [MIT License](about/license.md).
