# Welcome to scmd

**AI-powered slash commands for any terminal. Works offline by default.**

scmd brings AI superpowers to your command line—offline, private, and fast. Ask questions in plain English, review code with security templates, chat with AI that remembers context. All without API keys or cloud dependencies.

```bash
# Just ask what you want to do
scmd /cmd "find files modified today"
# → find . -type f -mtime -1

# Get instant explanations with beautiful formatting
scmd /explain main.go

# Review with professional security templates
scmd /review auth.js --template security-review

# Chat with full context retention
scmd chat
You: How do I set up OAuth2 in Go?
🤖 Assistant: [Detailed explanation...]
You: Show me an example with JWT
🤖 Assistant: [Builds on previous context...]
```

---

## ✨ Key Features

<div class="grid cards" markdown>

-   :material-shield-lock:{ .lg .middle } **100% Offline & Private**

    ---

    Local LLMs via llama.cpp + Qwen models. No API keys, no telemetry, your code stays on your machine. Optional cloud backends available.

-   :material-chat:{ .lg .middle } **Smart Conversations**

    ---

    Multi-turn chat with context retention, searchable history, auto-save, and markdown export. Pick up conversations anytime.

-   :material-palette:{ .lg .middle } **Beautiful Output** *NEW in v0.4.3*

    ---

    Markdown rendering with syntax highlighting for 40+ languages. Multiple themes, auto-detection, NO_COLOR support.

-   :material-flash:{ .lg .middle } **Fast & Lightweight**

    ---

    14MB binary, 0.5B-7B models, GPU acceleration (Metal/CUDA), streaming output. Choose speed vs quality.

-   :material-book-open-page-variant:{ .lg .middle } **Man Page Integration**

    ---

    `/cmd` reads system man pages to generate exact, copy-paste ready commands. Works with 60+ common CLI tools.

-   :material-security:{ .lg .middle } **Security Templates** *NEW in v0.4.0*

    ---

    Professional code reviews with OWASP Top 10 focus. 6 built-in templates for security, performance, API design, testing, docs, and education.

-   :material-package-variant:{ .lg .middle } **Repository System**

    ---

    Commands install like npm packages. Discover 100+ community commands. Create and share your own.

-   :material-tools:{ .lg .middle } **Zero Maintenance**

    ---

    Auto-starts, auto-restarts, self-healing. Intelligent error handling with actionable solutions. Never need `pkill`.

</div>

---

## 🚀 Quick Start

<div class="grid cards" markdown>

-   :material-download:{ .lg .middle } **1. Install**

    ---

    ```bash
    # Homebrew (macOS/Linux)
    brew install sunboylabs/tap/scmd

    # npm (Cross-platform)
    npm install -g scmd-cli

    # Shell script
    curl -fsSL https://scmd.sh/install.sh | bash
    ```

    [:octicons-arrow-right-24: Installation Guide](getting-started/installation.md)

-   :material-play-circle:{ .lg .middle } **2. First Run**

    ---

    ```bash
    scmd /explain "what is docker"
    ```

    Beautiful setup wizard appears:
    - Choose model preset (Fast/Balanced/Best/Premium)
    - Auto-download (~2 min)
    - Done! Works 100% offline

    [:octicons-arrow-right-24: Quick Start](getting-started/quick-start.md)

-   :material-rocket-launch:{ .lg .middle } **3. Try Commands**

    ---

    ```bash
    # Generate commands
    scmd /cmd "compress directory"

    # Explain code
    scmd /explain algorithm.py

    # Review code
    scmd /review auth.js --template security-review

    # Start chat
    scmd chat
    ```

    [:octicons-arrow-right-24: User Guide](user-guide/slash-commands.md)

-   :material-puzzle:{ .lg .middle } **4. Explore**

    ---

    ```bash
    # Discover 100+ commands
    scmd repo search docker

    # Install commands
    scmd repo install official/commit

    # Create your own
    scmd template create my-template
    ```

    [:octicons-arrow-right-24: Command Authoring](command-authoring/overview.md)

</div>

---

## 🎯 Highlights

### Beautiful Markdown Output (v0.5.1)

AI responses now render with gorgeous markdown formatting, syntax highlighting, and themes:

```bash
scmd /explain quicksort.py

# Output includes:
# - Syntax-highlighted code blocks
# - Formatted headers and lists
# - Tables and links
# - Theme detection (dark/light/auto)
# - Plain text when piped
```

**Performance**: Lazy-loaded Glamour renderer with < 1µs overhead when disabled.

### Interactive Chat Mode (v0.4.0)

Multi-turn conversations with full context retention:

```bash
scmd chat

You: How do I implement rate limiting in Express?
🤖 [Detailed explanation with code]

You: What about Redis-based rate limiting?
🤖 [Builds on previous context...]

/export  # Save to markdown
```

**Features**: Auto-save, searchable history, resume anytime, markdown export.

### Professional Templates (v0.4.0)

Standardized code reviews with 6 built-in templates:

| Template | Focus | Example |
|----------|-------|---------|
| **security-review** | OWASP Top 10 | `scmd review auth.js --template security-review` |
| **performance** | Bottlenecks, Big O | `scmd review sort.py --template performance` |
| **api-design** | REST practices | `scmd review api.go --template api-design` |
| **testing** | Coverage, edges | `scmd review test.ts --template testing` |
| **documentation** | Doc generation | `scmd explain utils.rs --template documentation` |
| **beginner-explain** | ELI5 mode | `scmd explain recursion.go --template beginner-explain` |

Create custom templates for your team. Share as YAML files.

### Man Page Integration

The `/cmd` command reads system man pages for exact commands:

```bash
scmd /cmd "find files modified in last 24 hours"
# → find . -type f -mtime -1

scmd /cmd "list processes sorted by memory"
# → ps aux --sort=-%mem | head -n 20
```

Detects 60+ common tools. Falls back to general CLI knowledge.

### Command Discovery (v0.5.0)

100+ community commands now properly discoverable:

```bash
# Search commands
scmd repo search git
scmd repo search docker

# Install instantly
scmd repo install official/commit
scmd repo install official/dockerfile

# Use immediately
git diff --staged | scmd /gc
```

Repository system fixed in v0.5.0 with legacy manifest support and parallel fetching.

---

## 📊 Performance

**Real-World Benchmarks** (M1 Mac, 8GB RAM):

| Task | qwen2.5-0.5b | qwen2.5-1.5b | qwen2.5-3b | qwen2.5-7b |
|------|-------------|-------------|-----------|-----------|
| Explain 50-line file | 3.2s | 5.8s | 9.1s | 16.3s |
| Generate commit | 2.8s | 4.9s | 7.5s | 14.1s |
| Review 200-line file | 6.5s | 11.2s | 18.7s | 32.4s |
| CLI command | 2.1s | 3.4s | 5.8s | 10.2s |

**Available Models**:

| Model | Size | Speed | Quality | Best For |
|-------|------|-------|---------|----------|
| qwen2.5-0.5b | 379 MB | ⚡⚡⚡⚡ | ⭐⭐⭐ | Quick tasks |
| **qwen2.5-1.5b** ⭐ | 1.0 GB | ⚡⚡⚡ | ⭐⭐⭐⭐ | Daily work (default) |
| qwen2.5-3b | 1.9 GB | ⚡⚡ | ⭐⭐⭐⭐⭐ | Complex analysis |
| qwen2.5-7b | 3.8 GB | ⚡ | ⭐⭐⭐⭐⭐ | Production code |

All models: 8192 token context, GPU acceleration, 4-bit quantization, function calling.

---

## 🛡️ Stability & Reliability

**Zero Maintenance Design**:

- ✅ **Auto-starts** - Server starts when needed
- ✅ **Auto-restarts** - Crashes handled gracefully
- ✅ **Self-healing** - Detects OOM, context mismatches, recovers automatically
- ✅ **Clear feedback** - Every error includes actionable solutions
- ✅ **No manual intervention** - Never need server management

**Example Error Handling**:

```
❌ Input exceeds available context size

What happened:
  Your input (5502 tokens) exceeds GPU capacity (4096 tokens)

Solutions:
  1. Use CPU mode: export SCMD_CPU_ONLY=1
  2. Split input into smaller files
  3. Use cloud backend: export OPENAI_API_KEY=...
```

See [Stability Architecture](architecture/STABILITY.md) for details.

---

## 🌟 Use Cases

<div class="grid cards" markdown>

-   :material-code-braces: **Code Explanation**

    ---

    ```bash
    scmd /explain algorithm.py
    cat main.go | scmd explain
    scmd /explain "what are channels?"
    ```

-   :material-magnify: **Code Review**

    ---

    ```bash
    scmd /review code.py
    scmd /review auth.js --template security-review
    git diff | scmd /review
    ```

-   :material-terminal: **Command Generation**

    ---

    ```bash
    scmd /cmd "find large files"
    scmd /cmd "compress directory"
    scmd /cmd "list processes by memory"
    ```

-   :material-git: **Git Workflows**

    ---

    ```bash
    git diff --staged | scmd /gc
    git log --oneline | scmd /sum
    scmd repo install official/commit
    ```

-   :material-school: **Learning & Exploration**

    ---

    ```bash
    scmd chat
    scmd /explain recursion.go --template beginner-explain
    scmd history search "docker"
    ```

-   :material-shield-check: **Security Analysis**

    ---

    ```bash
    scmd /review auth.js --template security-review
    scmd /review api.go --template api-design
    scmd template create team-security
    ```

</div>

---

## 🔧 LLM Backends

| Backend | Local | Free | Setup |
|---------|-------|------|-------|
| **llama.cpp** ⭐ | ✓ | ✓ | Default - no setup |
| **Ollama** | ✓ | ✓ | `ollama serve` |
| **OpenAI** | ✗ | ✗ | `export OPENAI_API_KEY=...` |
| **Together.ai** | ✗ | Free tier | `export TOGETHER_API_KEY=...` |
| **Groq** | ✗ | Free tier | `export GROQ_API_KEY=...` |

**Backend Priority**: llama.cpp (default) → Ollama → OpenAI → Together.ai → Groq

```bash
# Use specific backend
scmd -b ollama /explain main.go
scmd -b openai -m gpt-4 /review code.py

# List backends
scmd backends
```

---

## 📚 Documentation

<div class="grid cards" markdown>

-   :material-book-open-variant: **User Guide**

    ---

    Learn slash commands, model management, chat mode, templates, and repositories.

    [:octicons-arrow-right-24: User Guide](user-guide/slash-commands.md)

-   :material-code-braces: **Command Authoring**

    ---

    Create custom AI commands with tools, hooks, composition, and context gathering.

    [:octicons-arrow-right-24: Command Authoring](command-authoring/overview.md)

-   :material-cog: **Architecture**

    ---

    Understand scmd's design: stability, backends, repository system, and more.

    [:octicons-arrow-right-24: Architecture](architecture.md)

-   :material-help-circle: **Troubleshooting**

    ---

    Common issues and solutions for installation, models, and backends.

    [:octicons-arrow-right-24: Troubleshooting](troubleshooting.md)

</div>

---

## 🚀 Recent Releases

### v0.5.1 (2026-01-12)
- **Fixed**: Streaming AI responses now use Glamour markdown renderer
- Beautiful code blocks with syntax highlighting
- Markdown headers, lists, and formatting properly rendered

### v0.5.0 (2026-01-12)
- **Fixed**: Repository manifest parsing bug (100+ commands now discoverable)
- Added legacy manifest format support
- Enhanced command discovery UI
- Parallel manifest normalization (10 concurrent requests)

### v0.4.3 (2026-01-12)
- **Added**: Beautiful markdown output with Glamour
- Lazy rendering (< 1ns overhead when disabled)
- Syntax highlighting for 40+ languages
- Theme detection and NO_COLOR support

### v0.4.2 (2026-01-11)
- **Added**: Template-command integration
- Official commands repository (100+ commands)
- Unified command specification

### v0.4.1 (2026-01-11)
- **Fixed**: Replaced CGO-dependent SQLite with pure Go
- Improved cross-platform compatibility

### v0.4.0 (2026-01-10)
- **Added**: Interactive conversation mode with SQLite persistence
- Beautiful markdown output with syntax highlighting
- Template/pattern system with 6 built-in templates
- Conversation history management

See [Changelog](about/changelog.md) for full history.

---

## 🌍 Community

- **GitHub**: [sunboylabs/scmd](https://github.com/sunboylabs/scmd)
- **Issues**: [Report bugs or request features](https://github.com/sunboylabs/scmd/issues)
- **Discussions**: [Ask questions and share commands](https://github.com/sunboylabs/scmd/discussions)
- **Commands Repository**: [sunboylabs/commands](https://github.com/sunboylabs/commands) (100+ commands)

---

## 📖 What's Next?

<div class="grid cards" markdown>

-   :material-download: [**Installation**](getting-started/installation.md)

    Get scmd installed via Homebrew, npm, or packages

-   :material-play-circle: [**Quick Start**](getting-started/quick-start.md)

    5-minute tutorial to using slash commands

-   :material-puzzle: [**First Command**](getting-started/first-command.md)

    Create your first custom AI command

-   :material-book-multiple: [**Examples**](examples/basic-commands.md)

    Learn from real-world command examples

</div>

---

## 🆚 Why scmd?

| Traditional AI Tools | scmd |
|---------------------|------|
| API keys required | ✓ Works offline by default |
| Web interfaces | ✓ Native terminal integration |
| Fixed prompts | ✓ Customizable command specs |
| Isolated tools | ✓ Repository system for sharing |
| Text generation only | ✓ Tool calling for actions |
| No automation | ✓ Hook system |
| Single-turn | ✓ Multi-turn chat with history |
| Plain text | ✓ Beautiful markdown output |

---

## 📜 License

scmd is open source software licensed under the [MIT License](about/license.md).

---

**Built with ❤️ using Go** • *Inspired by the Unix philosophy and modern AI tooling*
