# Quick Start

Get up and running with scmd in 5 minutes. This guide assumes you've already [installed scmd](installation.md).

## Your First Command

Type `/explain` to explain code or concepts:

```bash
./scmd /explain "what is a goroutine in Go?"
```

!!! note "First Run"
    On first run, scmd will download the default qwen3-4b model (~2.6GB). This takes a few minutes but only happens once.

Output:
```
A goroutine is a lightweight thread managed by the Go runtime. Goroutines
are functions or methods that run concurrently with other goroutines in
the same address space. They are created with the `go` keyword...
```

## Explaining Code

Explain a file directly:

```bash
./scmd /explain main.go
```

Or pipe code to the command:

```bash
cat main.go | ./scmd /explain
```

## Common Use Cases

### 1. Generate Git Commit Messages

```bash
# Stage your changes
git add .

# Generate commit message from diff
git diff --staged | ./scmd /gc
```

Output:
```
feat: add user authentication system

Implemented JWT-based authentication with:
- Login and registration endpoints
- Token validation middleware
- Password hashing with bcrypt
- Session management

Breaking changes: Auth header format changed from
Bearer to JWT.
```

### 2. Review Code for Issues

```bash
# Review a file
./scmd /review main.go

# Or pipe code
git diff | ./scmd /review
```

Output:
```
Code Review:

✓ Good use of error handling
✓ Functions are well-structured
⚠️ Consider adding input validation in CreateUser()
⚠️ SQL query should use parameterized queries to prevent injection
❌ Missing database connection cleanup - add defer db.Close()

Security Concerns:
- Line 45: Potential SQL injection vulnerability
- Line 78: Password stored in plaintext

Suggestions:
1. Use prepared statements for database queries
2. Hash passwords with bcrypt
3. Add proper error messages for user feedback
```

### 3. Summarize Changes

```bash
git diff | ./scmd /sum
```

Output:
```
**Summary of Changes:**

- Added user authentication endpoints (/login, /register)
- Implemented JWT token generation and validation
- Created User model and database schema
- Added password hashing utilities
- Updated middleware to check authentication
- Added tests for auth endpoints (15 new tests)

**Files changed:** 8 modified, 3 new
```

### 4. Fix Errors

```bash
# Pipe error logs
go build 2>&1 | ./scmd /fix

# Or from a file
cat error.log | ./scmd /fix
```

Output:
```
Error Analysis:

Error: "undefined: User.Pasword"
Cause: Typo in field name (should be "Password")
Fix: Line 42, change User.Pasword to User.Password

Error: "cannot use string as int value in assignment"
Cause: Type mismatch in ID field
Fix: Convert string to int using strconv.Atoi()

Suggested fixes:
```go
// Line 42
-user.Pasword = hash
+user.Password = hash

// Line 58
-userID := "123"
+userID, err := strconv.Atoi("123")
```

## Available Slash Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `/explain` | `/e`, `/exp` | Explain code or concepts |
| `/review` | `/r`, `/rev` | Review code for issues |
| `/commit` | `/gc`, `/gitc` | Generate git commit messages |
| `/summarize` | `/s`, `/sum`, `/tldr` | Summarize text |
| `/fix` | `/f`, `/err` | Explain and fix errors |

## Using Different Models

List available models:

```bash
scmd models list
```

Download a specific model:

```bash
scmd models pull qwen2.5-3b
```

Use a specific model:

```bash
scmd -m qwen2.5-3b /explain main.go
```

## Using Different Backends

### Ollama (Local, Alternative)

```bash
# Start Ollama
ollama serve

# Pull a model
ollama pull qwen2.5-coder:1.5b

# Use with scmd
scmd -b ollama /explain main.go
```

### OpenAI (Cloud)

```bash
export OPENAI_API_KEY=sk-...
scmd -b openai -m gpt-4o-mini /review code.py
```

## Inline Prompts

Use custom prompts on the fly:

```bash
# With -p flag
echo "SELECT * FROM users WHERE id = ?" | scmd -p "optimize this SQL query"

# File input
cat config.yaml | scmd -p "convert this to JSON"
```

## Saving Output

```bash
# Save to file
git diff | ./scmd /review -o review.md

# Append to file
cat error.log | ./scmd /fix >> fixes.txt
```

## Chaining Commands

```bash
# Complex pipeline
git log -1 --pretty=format:"%B" | \
  ./scmd /sum | \
  ./scmd -p "translate to Spanish"
```

## Debug Mode

Enable debug output to see what's happening:

```bash
SCMD_DEBUG=1 ./scmd /explain "what is Docker?"
```

Output:
```
[DEBUG] Model path: /Users/you/.scmd/models/qwen3-4b-Q4_K_M.gguf
[DEBUG] Prompt length: 156 chars
[DEBUG] Sending request to http://127.0.0.1:8089/completion
[DEBUG] Response status: 200
[DEBUG] Response length: 842 chars

Docker is a platform for developing, shipping, and running
applications in containers...
```

## Performance Tips

### Faster Responses

Use a smaller, faster model for quick queries:

```bash
scmd models pull qwen2.5-0.5b  # Smallest, fastest
scmd -m qwen2.5-0.5b /sum article.md
```

### GPU Acceleration

scmd automatically uses GPU if available:

- **macOS**: Metal (M1/M2/M3 chips)
- **Linux**: CUDA (NVIDIA GPUs)
- **Windows**: CUDA or CPU

Verify GPU usage:

```bash
# Check if GPU is being used
llama-server --help | grep -E "metal|cuda"
```

### Model Comparison

| Model | Size | Speed (tokens/sec) | Best For |
|-------|------|-------------------|----------|
| qwen2.5-0.5b | 379 MB | ~50 (GPU), ~10 (CPU) | Quick summaries |
| qwen2.5-1.5b | 940 MB | ~30 (GPU), ~7 (CPU) | Fast queries |
| qwen2.5-3b | 1.9 GB | ~20 (GPU), ~5 (CPU) | Balanced |
| qwen3-4b ⭐ | 2.5 GB | ~15 (GPU), ~4 (CPU) | Default, best quality |
| qwen2.5-7b | 4.4 GB | ~10 (GPU), ~2 (CPU) | Complex tasks |

## Workflow Examples

### Code Review Workflow

```bash
# 1. Make changes
git add .

# 2. Review changes
git diff --staged | ./scmd /review -o review.md

# 3. Generate commit message
git diff --staged | ./scmd /gc > commit.txt

# 4. Commit with generated message
git commit -F commit.txt
```

### Error Debugging Workflow

```bash
# 1. Run tests and capture errors
go test ./... 2>&1 | tee errors.txt

# 2. Analyze errors
cat errors.txt | ./scmd /fix > fixes.md

# 3. Explain specific error
cat errors.txt | ./scmd /explain
```

### Documentation Workflow

```bash
# 1. Explain code
./scmd /explain src/auth.go > docs/auth.md

# 2. Summarize changes
git log --since="1 week ago" --pretty=format:"%s" | ./scmd /sum > CHANGELOG.md

# 3. Generate README sections
ls -la | ./scmd -p "describe this project structure" >> README.md
```

## Common Patterns

### Process Multiple Files

```bash
# Explain all Go files
for file in *.go; do
  echo "=== $file ===" >> explanations.md
  ./scmd /explain "$file" >> explanations.md
done
```

### Watch for Changes

```bash
# Review changes on save (requires fswatch)
fswatch -o src/*.go | xargs -n1 -I{} git diff | ./scmd /review
```

### Interactive Mode

```bash
# REPL-style interaction
scmd slash interactive
```

```
scmd> /explain what is a mutex?
A mutex (mutual exclusion) is a synchronization primitive...

scmd> /review
Paste code (Ctrl+D when done):
func main() {
    x := 1
    x = x + 1
}
^D

Review: Simple code. Consider using x++ instead of x = x + 1
```

## Troubleshooting

### Slow Responses

- Use a smaller model: `scmd -m qwen2.5-1.5b /explain`
- Reduce max_tokens in command spec
- Enable GPU acceleration (check installation)

### Model Not Found

```bash
# List downloaded models
scmd models list

# Download missing model
scmd models pull qwen3-4b
```

### Command Not Found

```bash
# List available commands
scmd slash list

# Install missing command
scmd repo install official/explain
```

## Next Steps

<div class="grid cards" markdown>

-   :material-puzzle: **[Create Your First Command](first-command.md)**

    Build a custom AI command

-   :material-console: **[Set Up Shell Integration](shell-integration.md)**

    Use `/command` directly without `./scmd` prefix

-   :material-book-multiple: **[Explore Advanced Features](../command-authoring/tool-calling.md)**

    Learn about tool calling, hooks, and composition

-   :material-download: **[Install Community Commands](../user-guide/repositories.md)**

    Browse and install commands from repositories

</div>

## Quick Reference

```bash
# Basic usage
./scmd /COMMAND [args]
cat file | ./scmd /COMMAND

# With options
scmd -b BACKEND -m MODEL /COMMAND
scmd -p "custom prompt" /COMMAND

# Management
scmd models list              # List models
scmd backends                 # Check backends
scmd slash list               # List commands
scmd repo search QUERY        # Search commands

# Help
scmd --help
scmd COMMAND --help
```
