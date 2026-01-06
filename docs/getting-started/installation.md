# Installation

scmd works offline by default using llama.cpp and local Qwen models. This guide covers installation on all supported platforms.

## Prerequisites

- **Go 1.21 or later** (for building from source)
- **llama-server** (for offline inference)
- ~3GB disk space for the default model

## Quick Install (Recommended)

=== "macOS"

    ```bash
    # Install Go (if not installed)
    brew install go

    # Install llama-server for offline inference
    brew install llama.cpp

    # Clone and build scmd
    git clone https://github.com/scmd/scmd
    cd scmd
    go build -o scmd ./cmd/scmd

    # Move to PATH (optional)
    sudo mv scmd /usr/local/bin/

    # Verify installation
    scmd --version
    ```

=== "Linux"

    ```bash
    # Install Go (if not installed)
    # Ubuntu/Debian
    sudo apt update && sudo apt install golang-go

    # Fedora/RHEL
    sudo dnf install golang

    # Install llama-server
    # Build from source (recommended for latest version)
    git clone https://github.com/ggerganov/llama.cpp
    cd llama.cpp
    mkdir build && cd build
    cmake .. -DLLAMA_CUDA=ON  # For NVIDIA GPU support
    # cmake ..                # For CPU only
    cmake --build . --config Release
    sudo cp bin/llama-server /usr/local/bin/

    # Clone and build scmd
    git clone https://github.com/scmd/scmd
    cd scmd
    go build -o scmd ./cmd/scmd

    # Move to PATH (optional)
    sudo mv scmd /usr/local/bin/

    # Verify installation
    scmd --version
    ```

=== "Windows"

    ```powershell
    # Install Go from https://go.dev/dl/

    # Install llama.cpp (via CMake or pre-built binaries)
    # Download from https://github.com/ggerganov/llama.cpp/releases

    # Clone and build scmd
    git clone https://github.com/scmd/scmd
    cd scmd
    go build -o scmd.exe ./cmd/scmd

    # Add to PATH or use .\scmd.exe

    # Verify installation
    .\scmd.exe --version
    ```

## Install from Source

```bash
# Install with go install
go install github.com/scmd/scmd/cmd/scmd@latest

# Install llama-server
# macOS
brew install llama.cpp

# Linux - build from source
git clone https://github.com/ggerganov/llama.cpp
cd llama.cpp && mkdir build && cd build
cmake .. && cmake --build . --config Release
sudo cp bin/llama-server /usr/local/bin/
```

## Verify Installation

### Check scmd

```bash
scmd --version
# Output: scmd version 1.0.0
```

### Check Backends

```bash
scmd backends
```

Expected output:

```
Available backends:

✓ llamacpp      Ready (qwen3-4b)
  llama-server running on http://127.0.0.1:8089

✗ ollama        Not running
  Start with: ollama serve

✗ openai        Not configured
  Set OPENAI_API_KEY environment variable

✗ together      Not configured
  Set TOGETHER_API_KEY environment variable

✗ groq          Not configured
  Set GROQ_API_KEY environment variable
```

### First Run (Model Download)

On first use, scmd will automatically download the default model (~2.6GB):

```bash
./scmd /explain "what is a channel in Go?"
```

Output:
```
[INFO] First run detected
[INFO] Downloading qwen3-4b model (2.6 GB)...
[INFO] Progress: ████████████████████ 100%
[INFO] Model downloaded to ~/.scmd/models/qwen3-4b-Q4_K_M.gguf
[INFO] Starting llama-server...

A channel in Go is a typed conduit through which you can send
and receive values with the channel operator <-...
```

## Installation Paths

scmd uses the following directory structure:

```
~/.scmd/
├── config.yaml          # Configuration file
├── slash.yaml           # Slash command mappings
├── repos.json           # Repository list
├── models/              # Downloaded GGUF models
│   ├── qwen3-4b-Q4_K_M.gguf
│   └── qwen2.5-3b-Q4_K_M.gguf
├── commands/            # Installed command specs
│   ├── git-commit.yaml
│   └── explain.yaml
└── cache/               # Cached manifests
    └── official/
        └── manifest.yaml
```

You can customize the data directory:

```bash
export SCMD_DATA_DIR=/path/to/custom/dir
scmd /explain "test"
```

## Optional: Additional LLM Backends

### Ollama

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull a model
ollama pull qwen2.5-coder:1.5b

# Start Ollama server
ollama serve

# Use with scmd
scmd -b ollama /explain main.go
```

### OpenAI

```bash
# Set API key
export OPENAI_API_KEY=sk-...

# Use with scmd
scmd -b openai -m gpt-4o-mini /review code.py
```

### Together.ai (Free Tier Available)

```bash
# Get API key from https://together.ai
export TOGETHER_API_KEY=...

# Use with scmd
scmd -b together /explain main.go
```

### Groq (Free Tier Available)

```bash
# Get API key from https://groq.com
export GROQ_API_KEY=gsk_...

# Use with scmd
scmd -b groq -m llama-3.1-8b-instant /review code.py
```

## Troubleshooting

### "llama-server not found"

Install llama.cpp:

=== "macOS"
    ```bash
    brew install llama.cpp
    ```

=== "Linux"
    ```bash
    # Build from source
    git clone https://github.com/ggerganov/llama.cpp
    cd llama.cpp && mkdir build && cd build
    cmake .. && cmake --build . --config Release
    sudo cp bin/llama-server /usr/local/bin/
    ```

### "Model download failed"

Check your internet connection and try manually downloading:

```bash
# Download qwen3-4b
mkdir -p ~/.scmd/models
cd ~/.scmd/models
wget https://huggingface.co/Qwen/Qwen3-4B-GGUF/resolve/main/qwen3-4b-q4_k_m.gguf
```

### "Port 8089 already in use"

Stop any running llama-server instances:

```bash
# Find process
lsof -i :8089

# Kill process
kill -9 <PID>

# Or use different port
export LLAMA_SERVER_PORT=8090
scmd /explain "test"
```

### GPU Acceleration Not Working

For NVIDIA GPUs on Linux:

```bash
# Rebuild llama.cpp with CUDA support
cd llama.cpp/build
cmake .. -DLLAMA_CUDA=ON
cmake --build . --config Release
sudo cp bin/llama-server /usr/local/bin/
```

For Apple Silicon (M1/M2/M3):

```bash
# Metal is enabled by default on macOS
# Verify with:
llama-server --help | grep metal
```

## Next Steps

- [Quick Start Tutorial](quick-start.md) - Learn basic usage in 5 minutes
- [Your First Command](first-command.md) - Create a custom command
- [Shell Integration](shell-integration.md) - Set up `/command` shortcuts
- [Model Management](../user-guide/models.md) - Download and manage models

## Upgrading

```bash
# Pull latest changes
cd scmd
git pull origin main

# Rebuild
go build -o scmd ./cmd/scmd

# Or with go install
go install github.com/scmd/scmd/cmd/scmd@latest
```

## Uninstall

```bash
# Remove binary
sudo rm /usr/local/bin/scmd

# Remove data directory (optional - deletes all downloaded models)
rm -rf ~/.scmd

# Uninstall llama.cpp (optional)
brew uninstall llama.cpp  # macOS
```
