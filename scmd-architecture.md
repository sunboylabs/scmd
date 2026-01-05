# scmd - Slash Command CLI

## Complete Architecture & Development Guide

**Version:** 1.0.0  
**Last Updated:** January 2025  
**Language:** Go  
**Codename:** scmd (slash command)

---

# Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Multi-Agent Development Structure](#2-multi-agent-development-structure)
3. [Technical Architecture](#3-technical-architecture)
4. [Project Structure](#4-project-structure)
5. [Core Interfaces](#5-core-interfaces)
6. [Command System](#6-command-system)
7. [Pipe & Streaming I/O](#7-pipe--streaming-io)
8. [Backend System](#8-backend-system)
9. [Plugin & Repository System](#9-plugin--repository-system)
10. [Repository Distribution Platform](#10-repository-distribution-platform)
11. [Tool Calling Framework](#11-tool-calling-framework)
12. [Permission System](#12-permission-system)
13. [Configuration](#13-configuration)
14. [Distribution](#14-distribution)
15. [Testing Strategy](#15-testing-strategy)
16. [Security Guidelines](#16-security-guidelines)
17. [Quality Standards](#17-quality-standards)
18. [Development Workflow](#18-development-workflow)
19. [Roadmap](#19-roadmap)

---

# 1. Executive Summary

## What is scmd?

scmd (slash command) is a lightweight, terminal-agnostic CLI tool that brings AI-powered slash commands to any developer's workflow. It runs local LLMs by default with zero configuration required.

## Core Principles

| Principle | Description |
|-----------|-------------|
| **Local First** | Works offline with local models, cloud is optional |
| **Zero Config** | Works out of the box, first command in < 60 seconds |
| **Terminal Agnostic** | Works in any terminal, any shell, any OS |
| **Lightweight** | < 10MB binary, < 1ms startup, minimal memory |
| **Secure by Default** | Explicit permissions, sandboxed execution |
| **Extensible** | Plugin system with multi-repo support |

## Why Go?

| Factor | Benefit |
|--------|---------|
| Startup Time | ~0.5ms vs ~100ms for Node.js |
| Binary Size | ~5MB single binary, no runtime needed |
| Distribution | Cross-compile trivially, single file deployment |
| Memory | ~10MB vs ~50MB for Node.js |
| CLI Ecosystem | Cobra, Bubbletea, Lipgloss - best in class |
| Development Speed | Fast iteration, large contributor pool |

---

# 2. Multi-Agent Development Structure

This project uses specialized Claude sub-agents for development. Each agent has a specific role, responsibilities, and quality gates.

## Agent Roles

### 2.1 Project Manager Agent (PM)

**Role:** Coordinates all agents, manages priorities, tracks progress.

**Responsibilities:**
- Break down features into tasks
- Assign tasks to appropriate agents
- Track dependencies between tasks
- Ensure milestones are met
- Resolve conflicts between agent decisions
- Maintain project documentation

**Artifacts:**
- `docs/PROJECT_STATUS.md` - Current status
- `docs/TASKS.md` - Task breakdown
- `docs/DECISIONS.md` - Architecture decisions log

**Quality Gates:**
- All tasks have clear acceptance criteria
- Dependencies are documented
- No blocking issues unresolved > 24h

---

### 2.2 Product Manager Agent (ProdM)

**Role:** Defines requirements, user experience, and feature specifications.

**Responsibilities:**
- Write user stories and requirements
- Define command UX and behavior
- Specify error messages and help text
- Design first-run experience
- Define success metrics
- Gather and incorporate feedback

**Artifacts:**
- `docs/REQUIREMENTS.md` - Feature requirements
- `docs/UX_SPEC.md` - User experience specifications
- `docs/COMMANDS.md` - Command specifications

**Quality Gates:**
- All features have user stories
- Error messages are helpful and actionable
- Help text is clear and complete
- UX is consistent across commands

---

### 2.3 Architect Agent (Arch)

**Role:** Designs system architecture and ensures technical coherence.

**Responsibilities:**
- Define interfaces and contracts
- Design module boundaries
- Make technology decisions
- Review architectural changes
- Ensure scalability and maintainability
- Document architecture decisions (ADRs)

**Artifacts:**
- `docs/ARCHITECTURE.md` - This document
- `docs/ADR/` - Architecture Decision Records
- `internal/*/interfaces.go` - Interface definitions

**Quality Gates:**
- All modules have clear interfaces
- No circular dependencies
- Changes don't break existing contracts
- ADRs for significant decisions

---

### 2.4 Developer Agent (Dev)

**Role:** Implements features according to specifications.

**Responsibilities:**
- Write clean, idiomatic Go code
- Implement features per specifications
- Write inline documentation
- Follow coding standards
- Create pull request descriptions

**Coding Standards:**
```go
// All exported functions must have documentation
// Function names are verb-first: GetConfig, LoadPlugin, ExecuteCommand
// Error messages are lowercase, no punctuation
// Use context.Context for cancellation
// Prefer composition over inheritance
// Keep functions < 50 lines
// Keep files < 500 lines
```

**Quality Gates:**
- Code passes `go vet` and `golangci-lint`
- All exported symbols documented
- No hardcoded strings (use constants)
- Error handling is complete
- Code is formatted with `gofmt`

---

### 2.5 Security Agent (Sec)

**Role:** Reviews code for security vulnerabilities and ensures secure design.

**Responsibilities:**
- Review all code for security issues
- Design permission system
- Audit shell command execution
- Review plugin sandboxing
- Check for injection vulnerabilities
- Validate input sanitization

**Security Checklist:**
```
□ Shell commands are parameterized, not concatenated
□ File paths are validated and sandboxed
□ User input is sanitized before use
□ Permissions are checked before operations
□ Sensitive data is not logged
□ Dependencies are audited
□ No hardcoded secrets
□ TLS is enforced for network calls
```

**Artifacts:**
- `docs/SECURITY.md` - Security guidelines
- `docs/THREAT_MODEL.md` - Threat analysis
- Security review comments on all PRs

**Quality Gates:**
- No high/critical vulnerabilities
- All user input validated
- Shell execution uses whitelist
- Audit log for sensitive operations

---

### 2.6 Quality Assurance Agent (QA)

**Role:** Ensures overall product quality and user experience.

**Responsibilities:**
- Define quality metrics
- Review feature completeness
- Test user workflows end-to-end
- Verify documentation accuracy
- Check cross-platform behavior
- Validate error handling UX

**Quality Metrics:**
```
- Command response time < 100ms (excluding LLM)
- Memory usage < 50MB during operation
- Binary size < 15MB
- Zero crashes in normal operation
- All commands have help text
- All errors have resolution suggestions
```

**Artifacts:**
- `docs/QUALITY_REPORT.md` - Quality status
- `tests/e2e/` - End-to-end test scenarios

**Quality Gates:**
- All quality metrics met
- Cross-platform testing passed
- Documentation matches behavior
- No P0/P1 bugs open

---

### 2.7 Unit Test Agent (UnitTest)

**Role:** Writes and maintains unit tests for all modules.

**Responsibilities:**
- Write unit tests for all packages
- Maintain > 80% code coverage
- Test edge cases and error paths
- Mock external dependencies
- Keep tests fast (< 10s total)

**Testing Standards:**
```go
// Test file naming: *_test.go
// Test function naming: Test<Function>_<Scenario>
// Use table-driven tests for multiple cases
// Mock interfaces, not implementations
// Test one thing per test
// Use testify for assertions
```

**Example Test Structure:**
```go
func TestCommandExecute_ValidArgs(t *testing.T) {
    // Arrange
    cmd := NewExplainCommand()
    args := &CommandArgs{Positional: []string{"file.go"}}
    
    // Act
    result, err := cmd.Execute(context.Background(), args)
    
    // Assert
    require.NoError(t, err)
    assert.True(t, result.Success)
}
```

**Coverage Requirements:**
```
internal/cli/       >= 80%
internal/backend/   >= 85%
internal/config/    >= 90%
internal/context/   >= 80%
internal/plugins/   >= 80%
internal/tools/     >= 85%
pkg/                >= 90%
```

**Quality Gates:**
- Coverage thresholds met
- All tests pass
- No flaky tests
- Tests run in < 30 seconds

---

### 2.8 Integration Test Agent (IntTest)

**Role:** Tests interactions between modules and external systems.

**Responsibilities:**
- Test module integration
- Test LLM backend integration
- Test plugin loading and execution
- Test repository operations
- Test shell command execution
- Verify configuration loading

**Integration Test Categories:**
```
tests/integration/
├── backend_test.go      # LLM backend tests
├── cli_test.go          # Full CLI command tests
├── config_test.go       # Config loading tests
├── git_test.go          # Git operations tests
├── plugin_test.go       # Plugin system tests
├── repo_test.go         # Repository tests
└── tools_test.go        # Tool execution tests
```

**Test Environment:**
- Use Docker for reproducible environments
- Mock LLM responses for speed
- Use temp directories for file operations
- Clean up after each test

**Quality Gates:**
- All integration tests pass
- Tests are isolated and reproducible
- No external network calls in CI
- Tests complete in < 2 minutes

---

## Agent Collaboration Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                    Feature Request / Bug                         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Project Manager (PM)                          │
│    • Creates task breakdown                                      │
│    • Assigns to agents                                           │
│    • Sets priorities                                             │
└─────────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                 ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│  Product Mgr    │ │   Architect     │ │   Security      │
│  • Requirements │ │   • Design      │ │   • Threat      │
│  • UX spec      │ │   • Interfaces  │ │     model       │
└────────┬────────┘ └────────┬────────┘ └────────┬────────┘
         │                   │                   │
         └─────────┬─────────┴─────────┬─────────┘
                   ▼                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                       Developer (Dev)                            │
│    • Implements feature                                          │
│    • Follows specs from ProdM, Arch, Sec                        │
└─────────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                 ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│  Unit Tests     │ │ Integration     │ │   Security      │
│  • Coverage     │ │   Tests         │ │   Review        │
│  • Edge cases   │ │ • E2E flows     │ │ • Audit         │
└────────┬────────┘ └────────┬────────┘ └────────┬────────┘
         │                   │                   │
         └─────────┬─────────┴─────────┬─────────┘
                   ▼                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                       QA Review                                  │
│    • Feature complete?                                           │
│    • Quality metrics met?                                        │
│    • Documentation updated?                                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Release                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

# 3. Technical Architecture

## Technology Stack

```
Language:        Go 1.22+
CLI Framework:   Cobra (github.com/spf13/cobra)
Configuration:   Viper (github.com/spf13/viper)
TUI Components:  Bubbletea, Lipgloss, Bubbles (Charm)
LLM Bindings:    Ollama client, llama.cpp CGO bindings
YAML Parsing:    gopkg.in/yaml.v3
Git Operations:  go-git (github.com/go-git/go-git/v5)
HTTP Client:     net/http (stdlib)
Testing:         testify, gomock
Linting:         golangci-lint
Release:         GoReleaser
```

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         scmd CLI                                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │    CLI      │  │    REPL     │  │   Config    │             │
│  │   (Cobra)   │  │  (Bubble)   │  │   (Viper)   │             │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘             │
│         │                │                │                     │
│         └────────────────┼────────────────┘                     │
│                          ▼                                      │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Command Router                         │  │
│  │   • Built-in commands (/explain, /review, /commit...)    │  │
│  │   • Plugin commands (loaded from repos)                   │  │
│  └──────────────────────────┬───────────────────────────────┘  │
│                             │                                   │
│         ┌───────────────────┼───────────────────┐              │
│         ▼                   ▼                   ▼              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐        │
│  │  Context    │    │   Tools     │    │  Backends   │        │
│  │  Gatherer   │    │  Executor   │    │  Manager    │        │
│  └─────────────┘    └─────────────┘    └─────────────┘        │
│         │                   │                   │              │
│         ▼                   ▼                   ▼              │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐        │
│  │ • Git       │    │ • Shell     │    │ • Local     │        │
│  │ • Files     │    │ • HTTP      │    │   (llama)   │        │
│  │ • Project   │    │ • Files     │    │ • Ollama    │        │
│  │ • Env       │    │ • Chain     │    │ • Claude    │        │
│  └─────────────┘    └─────────────┘    │ • OpenAI    │        │
│                                        └─────────────┘        │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│                      Plugin System                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │   Repos     │  │   Loader    │  │ Permissions │             │
│  │   Manager   │  │   & Cache   │  │   Manager   │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

## Module Dependencies

```
                    ┌─────────┐
                    │   cmd   │
                    └────┬────┘
                         │
                    ┌────▼────┐
                    │   cli   │
                    └────┬────┘
                         │
    ┌────────────────────┼────────────────────┐
    │                    │                    │
┌───▼───┐          ┌─────▼─────┐         ┌───▼───┐
│backend│          │  command  │         │ config│
└───┬───┘          └─────┬─────┘         └───┬───┘
    │                    │                   │
    │              ┌─────┼─────┐             │
    │              │     │     │             │
    │          ┌───▼──┐ ┌▼───┐ ┌▼────┐      │
    │          │tools │ │ctx │ │plugin│      │
    │          └───┬──┘ └─┬──┘ └──┬──┘      │
    │              │      │       │          │
    │              └──────┼───────┘          │
    │                     │                  │
    └─────────────────────┼──────────────────┘
                          │
                     ┌────▼────┐
                     │   pkg   │  (shared utilities)
                     └─────────┘

RULE: Dependencies flow downward only. No circular imports.
```

---

# 4. Project Structure

```
scmd/
├── cmd/
│   └── scmd/
│       └── main.go                 # Entry point
│
├── internal/                       # Private packages
│   ├── cli/                        # CLI setup
│   │   ├── root.go                 # Root command
│   │   ├── repl.go                 # Interactive REPL
│   │   ├── version.go              # Version command
│   │   └── completion.go           # Shell completions
│   │
│   ├── command/                    # Command definitions
│   │   ├── interface.go            # Command interface
│   │   ├── registry.go             # Command registry
│   │   ├── parser.go               # Argument parser
│   │   ├── builtin/                # Built-in commands
│   │   │   ├── explain.go
│   │   │   ├── review.go
│   │   │   ├── commit.go
│   │   │   ├── fix.go
│   │   │   ├── test.go
│   │   │   ├── help.go
│   │   │   ├── config.go
│   │   │   ├── repos.go
│   │   │   ├── install.go
│   │   │   └── search.go
│   │   └── builtin_test.go
│   │
│   ├── backend/                    # LLM backends
│   │   ├── interface.go            # Backend interface
│   │   ├── registry.go             # Backend registry
│   │   ├── local/                  # Local llama.cpp
│   │   │   ├── backend.go
│   │   │   ├── model.go
│   │   │   └── download.go
│   │   ├── ollama/                 # Ollama client
│   │   │   ├── backend.go
│   │   │   └── client.go
│   │   ├── claude/                 # Claude API
│   │   │   └── backend.go
│   │   ├── openai/                 # OpenAI API
│   │   │   └── backend.go
│   │   └── mock/                   # Mock for testing
│   │       └── backend.go
│   │
│   ├── context/                    # Context gathering
│   │   ├── interface.go
│   │   ├── gatherer.go
│   │   ├── git.go                  # Git operations
│   │   ├── project.go              # Project detection
│   │   ├── files.go                # File operations
│   │   └── env.go                  # Environment
│   │
│   ├── tools/                      # Tool execution
│   │   ├── interface.go            # Tool interface
│   │   ├── executor.go             # Execution engine
│   │   ├── shell.go                # Shell commands
│   │   ├── http.go                 # HTTP requests
│   │   ├── files.go                # File operations
│   │   ├── sandbox.go              # Sandboxed execution
│   │   └── toolcall.go             # LLM tool calling
│   │
│   ├── plugins/                    # Plugin system
│   │   ├── interface.go
│   │   ├── loader.go               # Plugin loader
│   │   ├── parser.go               # YAML parser
│   │   ├── validator.go            # Plugin validation
│   │   └── executor.go             # Plugin execution
│   │
│   ├── repos/                      # Repository management
│   │   ├── interface.go
│   │   ├── manager.go              # Repo manager
│   │   ├── resolver.go             # URL resolution
│   │   ├── fetcher.go              # Manifest fetching
│   │   ├── cache.go                # Local caching
│   │   └── auth.go                 # Authentication
│   │
│   ├── permissions/                # Permission system
│   │   ├── interface.go
│   │   ├── manager.go              # Permission manager
│   │   ├── grants.go               # Grant storage
│   │   ├── validator.go            # Validation
│   │   └── prompts.go              # User prompts
│   │
│   ├── config/                     # Configuration
│   │   ├── config.go               # Config struct
│   │   ├── loader.go               # Config loading
│   │   ├── defaults.go             # Default values
│   │   └── migrate.go              # Config migrations
│   │
│   ├── ui/                         # UI components
│   │   ├── spinner.go              # Loading spinner
│   │   ├── progress.go             # Progress bar
│   │   ├── stream.go               # Streaming output
│   │   ├── prompt.go               # User prompts
│   │   └── colors.go               # Color utilities
│   │
│   └── models/                     # Model management
│       ├── registry.go             # Model registry
│       ├── downloader.go           # Download manager
│       └── selector.go             # Model selection
│
├── pkg/                            # Public packages
│   ├── version/
│   │   └── version.go              # Version info
│   ├── errors/
│   │   └── errors.go               # Error types
│   └── utils/
│       ├── strings.go
│       ├── files.go
│       └── platform.go
│
├── configs/                        # Default configs
│   ├── default.yaml                # Default config
│   └── models.json                 # Model registry
│
├── scripts/                        # Build scripts
│   ├── install.sh                  # Curl installer
│   ├── publish-npm.sh              # npm publisher
│   └── build-all.sh                # Multi-platform build
│
├── npm/                            # npm distribution
│   ├── package.json
│   ├── install.js
│   └── platforms/
│       ├── darwin-arm64/
│       ├── darwin-x64/
│       ├── linux-x64/
│       ├── linux-arm64/
│       └── win32-x64/
│
├── tests/                          # Test suites
│   ├── unit/                       # Unit tests (mirrors internal/)
│   ├── integration/                # Integration tests
│   │   ├── backend_test.go
│   │   ├── cli_test.go
│   │   ├── plugin_test.go
│   │   └── tools_test.go
│   ├── e2e/                        # End-to-end tests
│   │   ├── commands_test.go
│   │   └── workflows_test.go
│   └── fixtures/                   # Test data
│       ├── projects/
│       ├── plugins/
│       └── responses/
│
├── docs/                           # Documentation
│   ├── ARCHITECTURE.md             # This document
│   ├── COMMANDS.md                 # Command reference
│   ├── PLUGINS.md                  # Plugin development
│   ├── CONTRIBUTING.md             # Contribution guide
│   ├── SECURITY.md                 # Security policy
│   └── ADR/                        # Architecture decisions
│       ├── 001-use-go.md
│       ├── 002-local-first.md
│       └── 003-plugin-format.md
│
├── .github/
│   └── workflows/
│       ├── ci.yml                  # CI pipeline
│       ├── release.yml             # Release pipeline
│       └── security.yml            # Security scanning
│
├── .goreleaser.yaml                # Release config
├── .golangci.yaml                  # Linter config
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile                      # For testing
├── LICENSE
└── README.md
```

---

# 5. Core Interfaces

## 5.1 Command Interface

```go
// internal/command/interface.go

package command

import (
    "context"
)

// Command defines the interface for all scmd commands
type Command interface {
    // Metadata
    Name() string
    Aliases() []string
    Description() string
    Usage() string
    Examples() []string
    Category() Category

    // Execution
    Execute(ctx context.Context, args *Args, execCtx *ExecContext) (*Result, error)

    // Validation
    Validate(args *Args) error

    // Requirements
    RequiresBackend() bool
    SuggestedBackend() BackendType
    Complexity() Complexity
}

// Category classifies commands
type Category string

const (
    CategoryCore   Category = "core"
    CategoryCode   Category = "code"
    CategoryGit    Category = "git"
    CategoryConfig Category = "config"
    CategoryPlugin Category = "plugin"
)

// Complexity hints at model selection
type Complexity string

const (
    ComplexitySimple   Complexity = "simple"
    ComplexityModerate Complexity = "moderate"
    ComplexityComplex  Complexity = "complex"
)

// Args represents parsed command arguments
type Args struct {
    Positional []string
    Flags      map[string]bool
    Options    map[string]string
    Raw        string
}

// ExecContext provides execution dependencies
type ExecContext struct {
    Backend Backend
    Config  *Config
    Project *ProjectContext
    UI      UI
    Tools   ToolExecutor
}

// Result represents command execution result
type Result struct {
    Success     bool
    Output      string
    Error       string
    Suggestions []string
    ExitCode    int
}
```

## 5.2 Backend Interface

```go
// internal/backend/interface.go

package backend

import (
    "context"
)

// Backend defines the LLM backend interface
type Backend interface {
    // Identity
    Name() string
    Type() Type

    // Lifecycle
    Initialize(ctx context.Context) error
    IsAvailable(ctx context.Context) (bool, error)
    Shutdown(ctx context.Context) error

    // Inference
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error)

    // Tool calling (optional)
    SupportsToolCalling() bool
    CompleteWithTools(ctx context.Context, req *ToolRequest) (*ToolResponse, error)

    // Info
    ModelInfo() *ModelInfo
    EstimateTokens(text string) int
}

// Type identifies backend type
type Type string

const (
    TypeLocal   Type = "local"
    TypeOllama  Type = "ollama"
    TypeClaude  Type = "claude"
    TypeOpenAI  Type = "openai"
)

// CompletionRequest for inference
type CompletionRequest struct {
    Prompt        string
    SystemPrompt  string
    MaxTokens     int
    Temperature   float64
    StopSequences []string
}

// CompletionResponse from inference
type CompletionResponse struct {
    Content      string
    TokensUsed   int
    FinishReason FinishReason
    Timing       *Timing
}

// StreamChunk for streaming responses
type StreamChunk struct {
    Content string
    Done    bool
    Error   error
}

// FinishReason why generation stopped
type FinishReason string

const (
    FinishComplete FinishReason = "complete"
    FinishLength   FinishReason = "length"
    FinishStop     FinishReason = "stop"
)

// Timing information
type Timing struct {
    PromptMS      int64
    CompletionMS  int64
    TokensPerSec  float64
}

// ModelInfo describes the loaded model
type ModelInfo struct {
    Name          string
    Size          string
    Quantization  string
    ContextLength int
    Capabilities  []string
}
```

## 5.3 Tool Interface

```go
// internal/tools/interface.go

package tools

import (
    "context"
)

// Tool defines an executable tool
type Tool struct {
    Name                string
    Type                Type
    Description         string
    Command             string            // For shell tools
    Parameters          map[string]Param
    Timeout             int               // Seconds
    RequiresConfirmation bool
}

// Type of tool
type Type string

const (
    TypeShell     Type = "shell"
    TypeFileRead  Type = "file_read"
    TypeFileWrite Type = "file_write"
    TypeFileList  Type = "file_list"
    TypeHTTP      Type = "http"
    TypeScmd      Type = "scmd"  // Chain to another command
)

// Param defines a tool parameter
type Param struct {
    Type        string   // string, int, bool
    Description string
    Required    bool
    Default     any
    Enum        []string
}

// Call represents a tool invocation
type Call struct {
    Tool       string
    Parameters map[string]any
}

// Result from tool execution
type Result struct {
    Success   bool
    Output    string
    Error     string
    ExitCode  int
    Duration  int64  // Milliseconds
}

// Executor executes tools
type Executor interface {
    Execute(ctx context.Context, tool *Tool, params map[string]any) (*Result, error)
    ValidatePermissions(tool *Tool) error
}
```

## 5.4 Plugin Interface

```go
// internal/plugins/interface.go

package plugins

// Plugin represents a loaded plugin command
type Plugin struct {
    Metadata    Metadata
    Command     CommandDef
    Permissions Permissions
    Context     ContextDef
    Tools       []Tool
    Prompt      PromptDef
    Output      OutputDef
    Execution   ExecutionDef
}

// Metadata about the plugin
type Metadata struct {
    Name        string
    Version     string
    Description string
    Author      string
    License     string
    Homepage    string
    Repository  string  // Source repo
}

// CommandDef defines the command interface
type CommandDef struct {
    Name        string
    Aliases     []string
    Usage       string
    Description string
    Args        []ArgDef
    Flags       []FlagDef
    Options     []OptionDef
}

// Permissions required by the plugin
type Permissions struct {
    Files   FilePerms
    Shell   ShellPerms
    Git     GitPerms
    Network NetworkPerms
    Env     EnvPerms
}

// Loader loads plugins from various sources
type Loader interface {
    LoadFromFile(path string) (*Plugin, error)
    LoadFromRepo(repo, name string) (*Plugin, error)
    ListInstalled() ([]*Plugin, error)
}
```

## 5.5 Repository Interface

```go
// internal/repos/interface.go

package repos

import (
    "context"
)

// Repository represents a plugin repository
type Repository struct {
    Name        string
    URL         string
    Auth        AuthConfig
    Enabled     bool
    LastUpdated string
}

// Manifest is the repo's slashdev-repo.yaml
type Manifest struct {
    Name        string
    Version     string
    Description string
    Maintainers []Maintainer
    Requires    Requirements
    Auth        AuthDef
    Commands    map[string]CommandRef
    Dependencies map[string]string
}

// CommandRef references a command in the repo
type CommandRef struct {
    Version     string
    Path        string
    Description string
    Tags        []string
    Deprecated  bool
}

// Manager manages repositories
type Manager interface {
    // Repo operations
    Add(ctx context.Context, name, url string, auth *AuthConfig) error
    Remove(name string) error
    Update(ctx context.Context) error
    List() ([]*Repository, error)

    // Command operations
    Search(ctx context.Context, query string) ([]SearchResult, error)
    Install(ctx context.Context, repo, command, version string) error
    Uninstall(repo, command string) error

    // Info
    GetManifest(repo string) (*Manifest, error)
    GetCommand(repo, command string) (*CommandDef, error)
}
```

---

# 6. Command System

## 6.1 Built-in Commands

| Command | Description | Complexity |
|---------|-------------|------------|
| `/explain <file>` | Explain code in plain English | Simple |
| `/review [--staged]` | Review code changes | Moderate |
| `/commit [--conventional]` | Generate commit message | Simple |
| `/fix <error>` | Fix an error or bug | Moderate |
| `/test <file>` | Generate tests | Moderate |
| `/help [command]` | Show help | - |
| `/config [action]` | Manage configuration | - |
| `/repos [action]` | Manage repositories | - |
| `/install <plugin>` | Install a plugin | - |
| `/search <query>` | Search for plugins | - |

## 6.2 Command Implementation Example

```go
// internal/command/builtin/explain.go

package builtin

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/scmd/scmd/internal/command"
)

const explainSystemPrompt = `You are a senior software engineer explaining code to a colleague.
Be concise but thorough. Focus on:
1. What the code does (high-level purpose)
2. How it works (key mechanisms)
3. Why it might be written this way (design decisions)
Use plain English. Avoid jargon unless necessary.`

// ExplainCommand implements /explain
type ExplainCommand struct{}

// Ensure interface compliance
var _ command.Command = (*ExplainCommand)(nil)

func NewExplainCommand() *ExplainCommand {
    return &ExplainCommand{}
}

func (c *ExplainCommand) Name() string        { return "explain" }
func (c *ExplainCommand) Aliases() []string   { return []string{"e", "exp"} }
func (c *ExplainCommand) Description() string { return "Explain code in plain English" }
func (c *ExplainCommand) Usage() string       { return "/explain <file> [--verbose]" }
func (c *ExplainCommand) Category() command.Category { return command.CategoryCode }
func (c *ExplainCommand) RequiresBackend() bool      { return true }
func (c *ExplainCommand) SuggestedBackend() command.BackendType { return command.BackendLocal }
func (c *ExplainCommand) Complexity() command.Complexity { return command.ComplexitySimple }

func (c *ExplainCommand) Examples() []string {
    return []string{
        "/explain src/auth.go",
        "/explain ./utils.py --verbose",
    }
}

func (c *ExplainCommand) Validate(args *command.Args) error {
    if len(args.Positional) == 0 {
        return fmt.Errorf("file path required")
    }
    return nil
}

func (c *ExplainCommand) Execute(
    ctx context.Context,
    args *command.Args,
    execCtx *command.ExecContext,
) (*command.Result, error) {
    // Validate
    if err := c.Validate(args); err != nil {
        return &command.Result{
            Success:     false,
            Error:       err.Error(),
            Suggestions: []string{"/explain <filename>", "/help explain"},
        }, nil
    }

    // Resolve file path
    filePath := args.Positional[0]
    absPath, err := filepath.Abs(filePath)
    if err != nil {
        return nil, fmt.Errorf("invalid path: %w", err)
    }

    // Check file exists
    if _, err := os.Stat(absPath); os.IsNotExist(err) {
        return &command.Result{
            Success:     false,
            Error:       fmt.Sprintf("file not found: %s", filePath),
            Suggestions: []string{"check the file path", "use tab completion"},
        }, nil
    }

    // Read file
    content, err := os.ReadFile(absPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }

    // Detect language
    lang := detectLanguage(absPath)

    // Build prompt
    prompt := fmt.Sprintf("Explain this %s code:\n\n```%s\n%s\n```", lang, lang, string(content))

    // Stream response
    execCtx.UI.WriteLine("")
    
    chunks, err := execCtx.Backend.Stream(ctx, &backend.CompletionRequest{
        Prompt:       prompt,
        SystemPrompt: explainSystemPrompt,
        MaxTokens:    2000,
        Temperature:  0.3,
    })
    if err != nil {
        return nil, fmt.Errorf("backend error: %w", err)
    }

    for chunk := range chunks {
        if chunk.Error != nil {
            return nil, chunk.Error
        }
        execCtx.UI.Write(chunk.Content)
    }

    execCtx.UI.WriteLine("")

    return &command.Result{Success: true}, nil
}

func detectLanguage(path string) string {
    ext := filepath.Ext(path)
    langMap := map[string]string{
        ".go":   "go",
        ".py":   "python",
        ".js":   "javascript",
        ".ts":   "typescript",
        ".rs":   "rust",
        ".java": "java",
        ".rb":   "ruby",
        ".c":    "c",
        ".cpp":  "cpp",
        ".h":    "c",
    }
    if lang, ok := langMap[ext]; ok {
        return lang
    }
    return "code"
}
```

## 6.3 Command Registry

```go
// internal/command/registry.go

package command

import (
    "fmt"
    "sync"
)

// Registry manages all available commands
type Registry struct {
    mu       sync.RWMutex
    commands map[string]Command
    aliases  map[string]string
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
    return &Registry{
        commands: make(map[string]Command),
        aliases:  make(map[string]string),
    }
}

// Register adds a command to the registry
func (r *Registry) Register(cmd Command) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    name := cmd.Name()
    if _, exists := r.commands[name]; exists {
        return fmt.Errorf("command already registered: %s", name)
    }

    r.commands[name] = cmd

    // Register aliases
    for _, alias := range cmd.Aliases() {
        if existing, exists := r.aliases[alias]; exists {
            return fmt.Errorf("alias %s already used by %s", alias, existing)
        }
        r.aliases[alias] = name
    }

    return nil
}

// Get retrieves a command by name or alias
func (r *Registry) Get(nameOrAlias string) (Command, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    // Check direct name
    if cmd, ok := r.commands[nameOrAlias]; ok {
        return cmd, true
    }

    // Check alias
    if name, ok := r.aliases[nameOrAlias]; ok {
        return r.commands[name], true
    }

    return nil, false
}

// List returns all registered commands
func (r *Registry) List() []Command {
    r.mu.RLock()
    defer r.mu.RUnlock()

    cmds := make([]Command, 0, len(r.commands))
    for _, cmd := range r.commands {
        cmds = append(cmds, cmd)
    }
    return cmds
}

// ListByCategory returns commands filtered by category
func (r *Registry) ListByCategory(cat Category) []Command {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var cmds []Command
    for _, cmd := range r.commands {
        if cmd.Category() == cat {
            cmds = append(cmds, cmd)
        }
    }
    return cmds
}
```

---

# 7. Pipe & Streaming I/O

scmd is designed as a Unix-native tool that composes with pipes and redirects.

## 7.1 Usage Patterns

```bash
# Pipe input with inline prompt
cat foo.md | scmd -p "summarize this" > summary.md

# Pipe input to specific command
cat error.log | scmd fix

# Pipe code for review
git diff | scmd review > review.md

# Chain commands
cat api.go | scmd explain | scmd -p "convert to bullet points" > notes.md

# Multiple files via cat
cat src/*.go | scmd -p "find security issues" > audit.md

# From clipboard (macOS)
pbpaste | scmd -p "improve this email"

# Process and save
curl -s https://api.example.com/data | scmd -p "analyze this JSON" -o analysis.md

# With context files
scmd -p "explain the relationship" -c models.py -c views.py

# Heredoc input
scmd -p "convert to Python" << 'EOF'
function add(a, b) {
    return a + b;
}
EOF
```

## 7.2 CLI Flags for I/O

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--prompt` | `-p` | Inline prompt (required for pipe mode without command) | - |
| `--output` | `-o` | Output file (alternative to redirect) | stdout |
| `--context` | `-c` | Additional context files (repeatable) | - |
| `--format` | `-f` | Output format: `text`, `json`, `markdown` | `text` |
| `--no-stream` | - | Buffer output instead of streaming | `false` |
| `--quiet` | `-q` | Suppress progress/status messages | auto in pipe |
| `--raw` | `-r` | Raw output (no formatting/colors) | auto in pipe |

## 7.3 Mode Detection

scmd automatically detects its execution context:

```go
// internal/cli/mode.go

package cli

import (
    "os"

    "golang.org/x/term"
)

// IOMode represents the input/output mode
type IOMode struct {
    // Input detection
    HasStdin      bool  // Data is being piped in
    StdinIsTTY    bool  // Stdin is a terminal
    
    // Output detection
    StdoutIsTTY   bool  // Stdout is a terminal
    StderrIsTTY   bool  // Stderr is a terminal
    
    // Derived modes
    Interactive   bool  // Full interactive mode (both TTY)
    PipeIn        bool  // Receiving piped input
    PipeOut       bool  // Output is being piped/redirected
    ScriptMode    bool  // Non-interactive (no TTY)
}

// DetectIOMode determines how scmd is being invoked
func DetectIOMode() *IOMode {
    stdinIsTTY := term.IsTerminal(int(os.Stdin.Fd()))
    stdoutIsTTY := term.IsTerminal(int(os.Stdout.Fd()))
    stderrIsTTY := term.IsTerminal(int(os.Stderr.Fd()))
    
    // Check if stdin has data (non-blocking check)
    hasStdin := !stdinIsTTY
    
    return &IOMode{
        HasStdin:    hasStdin,
        StdinIsTTY:  stdinIsTTY,
        StdoutIsTTY: stdoutIsTTY,
        StderrIsTTY: stderrIsTTY,
        Interactive: stdinIsTTY && stdoutIsTTY,
        PipeIn:      hasStdin,
        PipeOut:     !stdoutIsTTY,
        ScriptMode:  !stdinIsTTY && !stdoutIsTTY,
    }
}

// Behavior adjustments based on mode
func (m *IOMode) ShouldStream() bool {
    // Stream in interactive mode, buffer in pipe mode for clean output
    return m.StdoutIsTTY
}

func (m *IOMode) ShouldShowProgress() bool {
    // Show spinners/progress only in interactive mode
    return m.StdoutIsTTY && m.StderrIsTTY
}

func (m *IOMode) ShouldUseColors() bool {
    // Colors only when outputting to terminal
    return m.StdoutIsTTY
}

func (m *IOMode) ProgressWriter() *os.File {
    // Write progress to stderr so it doesn't pollute piped output
    if m.PipeOut {
        return os.Stderr
    }
    return os.Stdout
}
```

## 7.4 Stdin Reader

```go
// internal/cli/stdin.go

package cli

import (
    "bufio"
    "context"
    "io"
    "os"
    "time"
)

// StdinReader handles piped input
type StdinReader struct {
    timeout time.Duration
    maxSize int64
}

// NewStdinReader creates a stdin reader
func NewStdinReader() *StdinReader {
    return &StdinReader{
        timeout: 30 * time.Second,
        maxSize: 10 * 1024 * 1024, // 10MB max
    }
}

// Read reads all available stdin with timeout
func (r *StdinReader) Read(ctx context.Context) (string, error) {
    // Create timeout context
    ctx, cancel := context.WithTimeout(ctx, r.timeout)
    defer cancel()
    
    // Read in goroutine
    resultCh := make(chan readResult, 1)
    go func() {
        reader := io.LimitReader(os.Stdin, r.maxSize)
        data, err := io.ReadAll(reader)
        resultCh <- readResult{data: data, err: err}
    }()
    
    select {
    case result := <-resultCh:
        if result.err != nil {
            return "", result.err
        }
        return string(result.data), nil
    case <-ctx.Done():
        return "", ctx.Err()
    }
}

// ReadLines reads stdin line by line (for streaming processing)
func (r *StdinReader) ReadLines(ctx context.Context) (<-chan string, <-chan error) {
    lines := make(chan string)
    errs := make(chan error, 1)
    
    go func() {
        defer close(lines)
        defer close(errs)
        
        scanner := bufio.NewScanner(os.Stdin)
        // Increase buffer for long lines
        buf := make([]byte, 1024*1024)
        scanner.Buffer(buf, len(buf))
        
        for scanner.Scan() {
            select {
            case lines <- scanner.Text():
            case <-ctx.Done():
                errs <- ctx.Err()
                return
            }
        }
        
        if err := scanner.Err(); err != nil {
            errs <- err
        }
    }()
    
    return lines, errs
}

type readResult struct {
    data []byte
    err  error
}
```

## 7.5 Output Writer

```go
// internal/cli/output.go

package cli

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "sync"
)

// OutputWriter handles output to stdout/file
type OutputWriter struct {
    mu       sync.Mutex
    writer   io.Writer
    buffered *bufio.Writer
    file     *os.File
    format   OutputFormat
    mode     *IOMode
}

// OutputFormat for structured output
type OutputFormat string

const (
    FormatText     OutputFormat = "text"
    FormatJSON     OutputFormat = "json"
    FormatMarkdown OutputFormat = "markdown"
)

// OutputConfig for writer configuration
type OutputConfig struct {
    FilePath string
    Format   OutputFormat
    Mode     *IOMode
    NoBuffer bool
}

// NewOutputWriter creates an output writer
func NewOutputWriter(cfg *OutputConfig) (*OutputWriter, error) {
    var writer io.Writer = os.Stdout
    var file *os.File
    
    // Output to file if specified
    if cfg.FilePath != "" {
        f, err := os.Create(cfg.FilePath)
        if err != nil {
            return nil, fmt.Errorf("failed to create output file: %w", err)
        }
        writer = f
        file = f
    }
    
    ow := &OutputWriter{
        writer: writer,
        format: cfg.Format,
        mode:   cfg.Mode,
        file:   file,
    }
    
    // Buffer output in pipe mode for efficiency
    if !cfg.NoBuffer && (cfg.Mode.PipeOut || cfg.FilePath != "") {
        ow.buffered = bufio.NewWriter(writer)
        ow.writer = ow.buffered
    }
    
    return ow, nil
}

// Write writes raw string
func (w *OutputWriter) Write(s string) error {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    _, err := w.writer.Write([]byte(s))
    return err
}

// WriteLine writes a line
func (w *OutputWriter) WriteLine(s string) error {
    return w.Write(s + "\n")
}

// WriteJSON writes JSON output
func (w *OutputWriter) WriteJSON(v interface{}) error {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    encoder := json.NewEncoder(w.writer)
    encoder.SetIndent("", "  ")
    return encoder.Encode(v)
}

// Stream writes streaming content (no newline)
func (w *OutputWriter) Stream(s string) error {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    _, err := w.writer.Write([]byte(s))
    
    // Flush immediately in interactive mode
    if w.mode.StdoutIsTTY {
        if f, ok := w.writer.(*os.File); ok {
            f.Sync()
        }
    }
    
    return err
}

// Flush flushes any buffered output
func (w *OutputWriter) Flush() error {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    if w.buffered != nil {
        return w.buffered.Flush()
    }
    return nil
}

// Close closes the output writer
func (w *OutputWriter) Close() error {
    if err := w.Flush(); err != nil {
        return err
    }
    if w.file != nil {
        return w.file.Close()
    }
    return nil
}
```

## 7.6 Prompt Command

The `-p` flag creates an ad-hoc prompt command:

```go
// internal/command/builtin/prompt.go

package builtin

import (
    "context"
    "fmt"
    "strings"

    "github.com/scmd/scmd/internal/command"
)

// PromptCommand handles inline prompts with piped input
type PromptCommand struct{}

func NewPromptCommand() *PromptCommand {
    return &PromptCommand{}
}

func (c *PromptCommand) Name() string        { return "prompt" }
func (c *PromptCommand) Aliases() []string   { return []string{"p"} }
func (c *PromptCommand) Description() string { return "Run an inline prompt on input" }
func (c *PromptCommand) Usage() string       { return "scmd -p \"<prompt>\" or echo \"input\" | scmd -p \"<prompt>\"" }
func (c *PromptCommand) Category() command.Category { return command.CategoryCore }
func (c *PromptCommand) RequiresBackend() bool      { return true }

func (c *PromptCommand) Examples() []string {
    return []string{
        `cat README.md | scmd -p "summarize this"`,
        `git diff | scmd -p "review these changes"`,
        `scmd -p "write a haiku about coding"`,
        `cat data.json | scmd -p "extract all email addresses" -f json`,
    }
}

func (c *PromptCommand) Validate(args *command.Args) error {
    prompt := args.Options["prompt"]
    if prompt == "" {
        return fmt.Errorf("prompt required: use -p \"your prompt\"")
    }
    return nil
}

func (c *PromptCommand) Execute(
    ctx context.Context,
    args *command.Args,
    execCtx *command.ExecContext,
) (*command.Result, error) {
    prompt := args.Options["prompt"]
    stdin := args.Options["stdin"] // Piped input if any
    
    // Build the full prompt
    var fullPrompt strings.Builder
    
    if stdin != "" {
        fullPrompt.WriteString("Input:\n```\n")
        fullPrompt.WriteString(stdin)
        fullPrompt.WriteString("\n```\n\n")
    }
    
    fullPrompt.WriteString("Task: ")
    fullPrompt.WriteString(prompt)
    
    // Add context files if provided
    if contextFiles, ok := args.Options["context"]; ok && contextFiles != "" {
        // Context files already loaded and appended
        fullPrompt.WriteString("\n\nAdditional context:\n")
        fullPrompt.WriteString(contextFiles)
    }
    
    // Execute with backend
    req := &backend.CompletionRequest{
        Prompt:      fullPrompt.String(),
        MaxTokens:   4000,
        Temperature: 0.7,
    }
    
    // Stream or buffer based on mode
    if execCtx.Mode.ShouldStream() {
        chunks, err := execCtx.Backend.Stream(ctx, req)
        if err != nil {
            return nil, err
        }
        
        for chunk := range chunks {
            if chunk.Error != nil {
                return nil, chunk.Error
            }
            execCtx.Output.Stream(chunk.Content)
        }
    } else {
        // Buffer mode for pipes
        resp, err := execCtx.Backend.Complete(ctx, req)
        if err != nil {
            return nil, err
        }
        execCtx.Output.Write(resp.Content)
    }
    
    // Final newline
    execCtx.Output.WriteLine("")
    
    return &command.Result{Success: true}, nil
}
```

## 7.7 CLI Integration

```go
// internal/cli/root.go (updated)

var (
    promptFlag   string
    outputFlag   string
    contextFlags []string
    formatFlag   string
    quietFlag    bool
    rawFlag      bool
    noStreamFlag bool
)

func init() {
    // I/O flags
    rootCmd.PersistentFlags().StringVarP(&promptFlag, "prompt", "p", "", "inline prompt")
    rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "output file")
    rootCmd.PersistentFlags().StringArrayVarP(&contextFlags, "context", "c", nil, "context files")
    rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", "text", "output format: text, json, markdown")
    rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "suppress progress messages")
    rootCmd.PersistentFlags().BoolVarP(&rawFlag, "raw", "r", false, "raw output without formatting")
    rootCmd.PersistentFlags().BoolVar(&noStreamFlag, "no-stream", false, "buffer output instead of streaming")
}

// Main execution flow with pipe support
func runCommand(cmd *cobra.Command, args []string) error {
    ctx := context.Background()
    
    // Detect I/O mode
    mode := DetectIOMode()
    
    // Read stdin if available
    var stdinContent string
    if mode.PipeIn {
        reader := NewStdinReader()
        content, err := reader.Read(ctx)
        if err != nil {
            return fmt.Errorf("failed to read stdin: %w", err)
        }
        stdinContent = content
    }
    
    // Setup output writer
    outputCfg := &OutputConfig{
        FilePath: outputFlag,
        Format:   OutputFormat(formatFlag),
        Mode:     mode,
        NoBuffer: noStreamFlag,
    }
    output, err := NewOutputWriter(outputCfg)
    if err != nil {
        return err
    }
    defer output.Close()
    
    // Handle -p flag (inline prompt)
    if promptFlag != "" {
        return runPromptCommand(ctx, promptFlag, stdinContent, mode, output)
    }
    
    // Handle piped input to specific command
    if mode.PipeIn && len(args) > 0 {
        return runCommandWithStdin(ctx, args[0], args[1:], stdinContent, mode, output)
    }
    
    // Interactive mode
    if mode.Interactive {
        return runREPL()
    }
    
    // No input, no command - show help
    return cmd.Help()
}
```

## 7.8 JSON Output Mode

```go
// For -f json flag

// JSONResponse is the structured output format
type JSONResponse struct {
    Success   bool            `json:"success"`
    Command   string          `json:"command,omitempty"`
    Input     *JSONInput      `json:"input,omitempty"`
    Output    string          `json:"output"`
    Metadata  *JSONMetadata   `json:"metadata,omitempty"`
    Error     string          `json:"error,omitempty"`
}

type JSONInput struct {
    Source string `json:"source"` // "stdin", "file", "arg"
    Size   int    `json:"size"`   // bytes
}

type JSONMetadata struct {
    Model       string  `json:"model"`
    TokensUsed  int     `json:"tokens_used"`
    DurationMS  int64   `json:"duration_ms"`
}

// Usage:
// cat data.csv | scmd -p "convert to json" -f json
// {
//   "success": true,
//   "input": {"source": "stdin", "size": 1234},
//   "output": "[{\"name\": \"John\", ...}]",
//   "metadata": {"model": "qwen2.5-coder-1.5b", "tokens_used": 450, "duration_ms": 2340}
// }
```

## 7.9 Progress to Stderr

When output is piped, progress goes to stderr:

```go
// internal/ui/progress.go

func (u *UI) ShowProgress(message string) func() {
    // In pipe mode, write to stderr
    if u.mode.PipeOut {
        fmt.Fprintf(os.Stderr, "⏳ %s...\n", message)
        return func() {
            fmt.Fprintf(os.Stderr, "✓ %s done\n", message)
        }
    }
    
    // Interactive mode - show spinner on stdout
    spinner := NewSpinner(message)
    spinner.Start()
    return spinner.Stop
}
```

## 7.10 Pipe Examples with Expected Behavior

```bash
# Simple summarize
$ cat long_article.md | scmd -p "summarize in 3 bullet points"
• Point one about the main topic
• Point two with supporting details  
• Point three with conclusion

# Code review to file
$ git diff HEAD~3 | scmd review -o review.md
⏳ Reviewing changes...  # (to stderr)
✓ Review saved to review.md

# JSON extraction
$ cat logs.txt | scmd -p "extract all IP addresses" -f json
{
  "success": true,
  "output": ["192.168.1.1", "10.0.0.1", "172.16.0.1"]
}

# Chain with other tools
$ find . -name "*.go" -exec cat {} \; | scmd -p "find TODO comments" | grep -i "urgent"

# With additional context
$ cat error.log | scmd fix -c src/server.go -c src/handler.go

# Quiet mode for scripting
$ RESULT=$(cat data.json | scmd -p "extract emails" -q)
```

---

# 7. Backend System

## 7.1 Local Backend (llama.cpp)

```go
// internal/backend/local/backend.go

package local

import (
    "context"
    "fmt"
    "sync"

    "github.com/scmd/scmd/internal/backend"
    "github.com/scmd/scmd/internal/models"
)

// Backend implements local llama.cpp inference
type Backend struct {
    mu        sync.Mutex
    config    *Config
    model     *Model
    modelInfo *backend.ModelInfo
}

// Config for local backend
type Config struct {
    ModelPath     string
    ModelName     string
    ContextLength int
    GPULayers     int
    Threads       int
}

// NewBackend creates a new local backend
func NewBackend(cfg *Config) *Backend {
    return &Backend{
        config: cfg,
    }
}

func (b *Backend) Name() string         { return "local" }
func (b *Backend) Type() backend.Type   { return backend.TypeLocal }
func (b *Backend) SupportsToolCalling() bool { return false }

func (b *Backend) Initialize(ctx context.Context) error {
    b.mu.Lock()
    defer b.mu.Unlock()

    // Check if model exists, download if needed
    modelPath := b.config.ModelPath
    if modelPath == "" {
        // Use default model
        mgr := models.NewManager()
        var err error
        modelPath, err = mgr.EnsureModel(ctx, b.config.ModelName)
        if err != nil {
            return fmt.Errorf("failed to ensure model: %w", err)
        }
    }

    // Load model
    model, err := LoadModel(modelPath, &LoadOptions{
        ContextLength: b.config.ContextLength,
        GPULayers:     b.config.GPULayers,
        Threads:       b.config.Threads,
    })
    if err != nil {
        return fmt.Errorf("failed to load model: %w", err)
    }

    b.model = model
    b.modelInfo = &backend.ModelInfo{
        Name:          b.config.ModelName,
        ContextLength: b.config.ContextLength,
    }

    return nil
}

func (b *Backend) IsAvailable(ctx context.Context) (bool, error) {
    // Check if llama.cpp is working
    return b.model != nil, nil
}

func (b *Backend) Shutdown(ctx context.Context) error {
    b.mu.Lock()
    defer b.mu.Unlock()

    if b.model != nil {
        return b.model.Close()
    }
    return nil
}

func (b *Backend) Stream(
    ctx context.Context, 
    req *backend.CompletionRequest,
) (<-chan backend.StreamChunk, error) {
    b.mu.Lock()
    defer b.mu.Unlock()

    if b.model == nil {
        return nil, fmt.Errorf("model not loaded")
    }

    ch := make(chan backend.StreamChunk)

    go func() {
        defer close(ch)

        err := b.model.Generate(ctx, req.Prompt, req.SystemPrompt, func(token string) {
            select {
            case ch <- backend.StreamChunk{Content: token}:
            case <-ctx.Done():
                return
            }
        })

        if err != nil {
            ch <- backend.StreamChunk{Error: err}
        }
        ch <- backend.StreamChunk{Done: true}
    }()

    return ch, nil
}

func (b *Backend) Complete(
    ctx context.Context,
    req *backend.CompletionRequest,
) (*backend.CompletionResponse, error) {
    chunks, err := b.Stream(ctx, req)
    if err != nil {
        return nil, err
    }

    var content string
    for chunk := range chunks {
        if chunk.Error != nil {
            return nil, chunk.Error
        }
        content += chunk.Content
    }

    return &backend.CompletionResponse{
        Content:      content,
        FinishReason: backend.FinishComplete,
    }, nil
}

func (b *Backend) CompleteWithTools(
    ctx context.Context,
    req *backend.ToolRequest,
) (*backend.ToolResponse, error) {
    return nil, fmt.Errorf("tool calling not supported by local backend")
}

func (b *Backend) ModelInfo() *backend.ModelInfo {
    return b.modelInfo
}

func (b *Backend) EstimateTokens(text string) int {
    // Rough estimate: ~4 chars per token
    return len(text) / 4
}
```

## 7.2 Ollama Backend

```go
// internal/backend/ollama/backend.go

package ollama

import (
    "context"
    "fmt"

    "github.com/scmd/scmd/internal/backend"
)

// Backend implements Ollama inference
type Backend struct {
    client    *Client
    config    *Config
    modelInfo *backend.ModelInfo
}

// Config for Ollama backend
type Config struct {
    Host  string
    Model string
}

// NewBackend creates a new Ollama backend
func NewBackend(cfg *Config) *Backend {
    if cfg.Host == "" {
        cfg.Host = "http://localhost:11434"
    }
    return &Backend{
        config: cfg,
        client: NewClient(cfg.Host),
    }
}

func (b *Backend) Name() string       { return "ollama" }
func (b *Backend) Type() backend.Type { return backend.TypeOllama }
func (b *Backend) SupportsToolCalling() bool { return true }

func (b *Backend) Initialize(ctx context.Context) error {
    // Check connection
    if err := b.client.Ping(ctx); err != nil {
        return fmt.Errorf("cannot connect to Ollama at %s: %w", b.config.Host, err)
    }

    // Get model info
    info, err := b.client.ModelInfo(ctx, b.config.Model)
    if err != nil {
        return fmt.Errorf("failed to get model info: %w", err)
    }

    b.modelInfo = &backend.ModelInfo{
        Name:          b.config.Model,
        ContextLength: info.ContextLength,
    }

    return nil
}

func (b *Backend) IsAvailable(ctx context.Context) (bool, error) {
    err := b.client.Ping(ctx)
    return err == nil, err
}

func (b *Backend) Shutdown(ctx context.Context) error {
    return nil // Ollama is external
}

func (b *Backend) Stream(
    ctx context.Context,
    req *backend.CompletionRequest,
) (<-chan backend.StreamChunk, error) {
    return b.client.ChatStream(ctx, &ChatRequest{
        Model:    b.config.Model,
        Messages: buildMessages(req),
        Options: Options{
            Temperature: req.Temperature,
            NumPredict:  req.MaxTokens,
        },
    })
}

func (b *Backend) Complete(
    ctx context.Context,
    req *backend.CompletionRequest,
) (*backend.CompletionResponse, error) {
    resp, err := b.client.Chat(ctx, &ChatRequest{
        Model:    b.config.Model,
        Messages: buildMessages(req),
        Stream:   false,
        Options: Options{
            Temperature: req.Temperature,
            NumPredict:  req.MaxTokens,
        },
    })
    if err != nil {
        return nil, err
    }

    return &backend.CompletionResponse{
        Content:      resp.Message.Content,
        FinishReason: backend.FinishComplete,
    }, nil
}

func (b *Backend) CompleteWithTools(
    ctx context.Context,
    req *backend.ToolRequest,
) (*backend.ToolResponse, error) {
    // Ollama supports function calling
    return b.client.ChatWithTools(ctx, req)
}

func (b *Backend) ModelInfo() *backend.ModelInfo {
    return b.modelInfo
}

func (b *Backend) EstimateTokens(text string) int {
    return len(text) / 4
}

func buildMessages(req *backend.CompletionRequest) []Message {
    msgs := []Message{}
    if req.SystemPrompt != "" {
        msgs = append(msgs, Message{Role: "system", Content: req.SystemPrompt})
    }
    msgs = append(msgs, Message{Role: "user", Content: req.Prompt})
    return msgs
}
```

## 7.3 Backend Registry

```go
// internal/backend/registry.go

package backend

import (
    "context"
    "fmt"
    "sync"
)

// Registry manages available backends
type Registry struct {
    mu       sync.RWMutex
    backends map[Type]Backend
    active   Backend
}

// NewRegistry creates a new backend registry
func NewRegistry() *Registry {
    return &Registry{
        backends: make(map[Type]Backend),
    }
}

// Register adds a backend
func (r *Registry) Register(b Backend) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.backends[b.Type()] = b
}

// Get retrieves a backend by type
func (r *Registry) Get(t Type) (Backend, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    b, ok := r.backends[t]
    return b, ok
}

// SetActive sets the active backend
func (r *Registry) SetActive(ctx context.Context, t Type) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    b, ok := r.backends[t]
    if !ok {
        return fmt.Errorf("backend not registered: %s", t)
    }

    if err := b.Initialize(ctx); err != nil {
        return fmt.Errorf("failed to initialize backend: %w", err)
    }

    r.active = b
    return nil
}

// Active returns the active backend
func (r *Registry) Active() Backend {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.active
}

// Available returns all available backends
func (r *Registry) Available(ctx context.Context) []Type {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var available []Type
    for t, b := range r.backends {
        if ok, _ := b.IsAvailable(ctx); ok {
            available = append(available, t)
        }
    }
    return available
}
```

---

# 8. Plugin & Repository System

## 8.1 Repository Manifest Format

```yaml
# scmd-repo.yaml

name: acme-corp/internal
version: "1.0.0"
description: ACME Corporation internal developer commands
homepage: https://github.com/acme-corp/scmd-plugins

maintainers:
  - name: DevTools Team
    email: devtools@acme.com

requires:
  scmd: ">=0.2.0"

auth:
  type: token  # none | token | oauth | basic
  env_var: ACME_SCMD_TOKEN

commands:
  django-migrate:
    version: "1.2.0"
    path: commands/django-migrate
    description: Generate Django migrations
    tags: [django, python, database]
    
  k8s-deploy:
    version: "1.0.0"
    path: commands/k8s-deploy
    description: Deploy to Kubernetes
    tags: [kubernetes, devops]
```

## 8.2 Plugin Command Format

```yaml
# commands/k8s-deploy/command.yaml

name: k8s-deploy
version: "1.0.0"
description: Deploy to Kubernetes with validation and rollback

command:
  name: k8s-deploy
  aliases: [deploy, k8s]
  usage: /k8s-deploy <environment> [--image=tag] [--dry-run]
  
  args:
    - name: environment
      type: string
      required: true
      enum: [dev, staging, production]
      description: Target environment
      
  flags:
    - name: dry-run
      short: d
      description: Show what would happen
      
  options:
    - name: image
      description: Docker image tag
      default: latest

permissions:
  files:
    read:
      - "k8s/*.yaml"
      - ".env.*"
    write: []
    
  shell:
    commands:
      - "kubectl get *"
      - "kubectl apply *"
      - "kubectl rollout *"
    allow_arbitrary: false
    
  network:
    allowed_hosts:
      - "*.k8s.internal"
      
  env:
    read:
      - KUBECONFIG

context:
  project:
    detect:
      - "k8s/*.yaml"
      
  files:
    patterns:
      - "k8s/{{args.environment}}/*.yaml"

tools:
  - name: kubectl_get_pods
    type: shell
    description: Get pods in namespace
    command: "kubectl get pods -n {{namespace}}"
    parameters:
      namespace:
        type: string
        required: true
        
  - name: kubectl_apply
    type: shell
    description: Apply manifest
    command: "kubectl apply -f -"
    requires_confirmation: true
    parameters:
      manifest:
        type: string
        required: true

system_prompt: |
  You are a Kubernetes deployment expert. Deploy applications safely.
  Always check status before changes, validate, and verify after.

prompt: |
  Deploy to {{args.environment}} environment.
  
  {{#if options.image}}
  Image: {{options.image}}
  {{/if}}
  
  ## Manifests
  {{files}}
  
  {{#if flags.dry-run}}
  Show deployment plan without applying.
  {{else}}
  1. Check current status
  2. Apply deployment
  3. Wait for rollout
  4. Verify health
  {{/if}}

output:
  format: text

execution:
  timeout: 300
  max_tokens: 4000
  temperature: 0.2
  suggested_backend: cloud
  complexity: complex
```

## 8.3 Repository Manager

```go
// internal/repos/manager.go

package repos

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "sync"

    "gopkg.in/yaml.v3"
)

// Manager handles repository operations
type Manager struct {
    mu          sync.RWMutex
    configPath  string
    cachePath   string
    repos       map[string]*Repository
    fetcher     *Fetcher
    auth        *AuthManager
}

// NewManager creates a new repository manager
func NewManager(configPath, cachePath string) *Manager {
    return &Manager{
        configPath: configPath,
        cachePath:  cachePath,
        repos:      make(map[string]*Repository),
        fetcher:    NewFetcher(),
        auth:       NewAuthManager(),
    }
}

// Load loads repositories from config
func (m *Manager) Load() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    data, err := os.ReadFile(m.configPath)
    if os.IsNotExist(err) {
        return nil // No repos configured
    }
    if err != nil {
        return err
    }

    var repos []*Repository
    if err := yaml.Unmarshal(data, &repos); err != nil {
        return err
    }

    for _, repo := range repos {
        m.repos[repo.Name] = repo
    }
    return nil
}

// Add adds a new repository
func (m *Manager) Add(ctx context.Context, name, url string, auth *AuthConfig) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // Resolve URL if short name
    resolvedURL, err := m.fetcher.ResolveURL(name, url)
    if err != nil {
        return fmt.Errorf("failed to resolve URL: %w", err)
    }

    // Fetch manifest to validate
    manifest, err := m.fetcher.FetchManifest(ctx, resolvedURL, auth)
    if err != nil {
        return fmt.Errorf("failed to fetch manifest: %w", err)
    }

    // Store repo
    repo := &Repository{
        Name:        name,
        URL:         resolvedURL,
        Auth:        auth,
        Enabled:     true,
        LastUpdated: manifest.Version,
    }
    m.repos[name] = repo

    // Cache manifest
    if err := m.cacheManifest(name, manifest); err != nil {
        return fmt.Errorf("failed to cache manifest: %w", err)
    }

    return m.save()
}

// Remove removes a repository
func (m *Manager) Remove(name string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    if _, exists := m.repos[name]; !exists {
        return fmt.Errorf("repository not found: %s", name)
    }

    delete(m.repos, name)

    // Remove cache
    cachePath := filepath.Join(m.cachePath, name)
    os.RemoveAll(cachePath)

    return m.save()
}

// Update refreshes all repository manifests
func (m *Manager) Update(ctx context.Context) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    for name, repo := range m.repos {
        if !repo.Enabled {
            continue
        }

        manifest, err := m.fetcher.FetchManifest(ctx, repo.URL, &repo.Auth)
        if err != nil {
            // Log error but continue
            continue
        }

        if err := m.cacheManifest(name, manifest); err != nil {
            continue
        }
    }

    return nil
}

// Search searches for commands across all repos
func (m *Manager) Search(ctx context.Context, query string) ([]SearchResult, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    var results []SearchResult

    for name, repo := range m.repos {
        if !repo.Enabled {
            continue
        }

        manifest, err := m.getManifest(name)
        if err != nil {
            continue
        }

        for cmdName, cmdRef := range manifest.Commands {
            if matchesQuery(cmdName, cmdRef.Description, cmdRef.Tags, query) {
                results = append(results, SearchResult{
                    Repo:        name,
                    Command:     cmdName,
                    Version:     cmdRef.Version,
                    Description: cmdRef.Description,
                    Tags:        cmdRef.Tags,
                })
            }
        }
    }

    return results, nil
}

// Install installs a command from a repository
func (m *Manager) Install(ctx context.Context, repo, command, version string) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    r, exists := m.repos[repo]
    if !exists {
        return fmt.Errorf("repository not found: %s", repo)
    }

    manifest, err := m.getManifest(repo)
    if err != nil {
        return err
    }

    cmdRef, exists := manifest.Commands[command]
    if !exists {
        return fmt.Errorf("command not found: %s/%s", repo, command)
    }

    // Use specified version or latest
    if version == "" {
        version = cmdRef.Version
    }

    // Fetch command definition
    cmdDef, err := m.fetcher.FetchCommand(ctx, r.URL, cmdRef.Path, &r.Auth)
    if err != nil {
        return fmt.Errorf("failed to fetch command: %w", err)
    }

    // Cache command
    cmdPath := filepath.Join(m.cachePath, "commands", repo, command)
    if err := os.MkdirAll(cmdPath, 0755); err != nil {
        return err
    }

    data, _ := yaml.Marshal(cmdDef)
    return os.WriteFile(filepath.Join(cmdPath, "command.yaml"), data, 0644)
}

// List returns all configured repositories
func (m *Manager) List() []*Repository {
    m.mu.RLock()
    defer m.mu.RUnlock()

    repos := make([]*Repository, 0, len(m.repos))
    for _, r := range m.repos {
        repos = append(repos, r)
    }
    return repos
}

func (m *Manager) save() error {
    repos := make([]*Repository, 0, len(m.repos))
    for _, r := range m.repos {
        repos = append(repos, r)
    }

    data, err := yaml.Marshal(repos)
    if err != nil {
        return err
    }

    return os.WriteFile(m.configPath, data, 0644)
}

func (m *Manager) cacheManifest(name string, manifest *Manifest) error {
    path := filepath.Join(m.cachePath, name, "manifest.yaml")
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return err
    }

    data, err := yaml.Marshal(manifest)
    if err != nil {
        return err
    }

    return os.WriteFile(path, data, 0644)
}

func (m *Manager) getManifest(name string) (*Manifest, error) {
    path := filepath.Join(m.cachePath, name, "manifest.yaml")
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var manifest Manifest
    if err := yaml.Unmarshal(data, &manifest); err != nil {
        return nil, err
    }

    return &manifest, nil
}
```

---

# 10. Repository Distribution Platform

scmd solves not just the CLI runtime, but also the **distribution problem** for AI command repos. This is where **OneSkill** fits in.

## 10.1 The Distribution Problem

| Problem | How scmd Solves It |
|---------|-------------------|
| **Discovery** | Central registry + search across all repos |
| **Trust** | Verified publishers, permission transparency |
| **Versioning** | Semantic versioning, pinned dependencies |
| **Updates** | `scmd repos update` fetches latest manifests |
| **Portability** | YAML-based commands work across teams |
| **Private repos** | Auth support for enterprise/internal repos |

## 10.2 Repository Ecosystem Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    OneSkill Registry                             │
│                  (registry.oneskill.dev)                         │
│                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │  Official   │  │  Verified   │  │  Community  │             │
│  │   Repos     │  │  Publishers │  │    Repos    │             │
│  └─────────────┘  └─────────────┘  └─────────────┘             │
│                                                                  │
│  • Discovery & Search API                                        │
│  • Publisher verification                                        │
│  • Download analytics                                            │
│  • Security scanning                                             │
│  • Version management                                            │
└─────────────────────────────────────────────────────────────────┘
          │                    │                    │
          ▼                    ▼                    ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   GitHub/GL     │  │   Self-Hosted   │  │   Enterprise    │
│   Public Repos  │  │   Repos         │  │   Private Repos │
└─────────────────┘  └─────────────────┘  └─────────────────┘
          │                    │                    │
          └────────────────────┼────────────────────┘
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                      scmd CLI Client                             │
│                                                                  │
│  scmd repos add oneskill/django-tools                           │
│  scmd search "kubernetes deploy"                                 │
│  scmd install acme-corp/internal/k8s-deploy                     │
└─────────────────────────────────────────────────────────────────┘
```

## 10.3 OneSkill Registry API

```yaml
# OneSkill provides these endpoints for scmd

# Search commands across all registered repos
GET /api/v1/search?q=kubernetes&category=devops
Response:
  results:
    - repo: oneskill/official
      command: k8s-deploy
      version: "1.2.0"
      description: Deploy to Kubernetes with validation
      downloads: 12450
      verified: true
      
    - repo: devops-tools/helm-commands  
      command: helm-upgrade
      version: "2.0.1"
      description: Smart Helm upgrades with rollback
      downloads: 5230
      verified: true

# Get repo manifest (cached/proxied)
GET /api/v1/repos/{owner}/{name}/manifest
Response:
  name: oneskill/official
  version: "3.2.0"
  commands:
    k8s-deploy: {version: "1.2.0", ...}
    docker-build: {version: "1.0.0", ...}

# Get command definition
GET /api/v1/repos/{owner}/{name}/commands/{command}
Response:
  <full command.yaml content>

# Analytics (for publishers)
GET /api/v1/repos/{owner}/{name}/stats
Response:
  total_downloads: 45000
  weekly_downloads: 1200
  command_breakdown:
    k8s-deploy: 12450
    docker-build: 8900

# Publish (authenticated)
POST /api/v1/repos/{owner}/{name}/publish
Authorization: Bearer <oneskill_token>
Body: <repo tarball>
```

## 10.4 Repository Types

### Official Repository (`oneskill/official`)

Pre-installed, curated commands maintained by OneSkill team:

```yaml
# Built into scmd, always available
scmd explain file.go      # Built-in
scmd review               # Built-in
scmd commit               # Built-in

# Official plugins (auto-suggested based on project type)
scmd install oneskill/official/django-migrate
scmd install oneskill/official/react-component
scmd install oneskill/official/terraform-plan
```

### Verified Publisher Repos

Organizations/individuals verified by OneSkill:

```bash
# Verified badge shown in search results
$ scmd search docker

  oneskill/official:
    ✓ docker-build      Build optimized Docker images
    
  docker/cli-plugins:     # Verified: Docker Inc.
    ✓ docker-debug      Debug container issues
    ✓ docker-slim       Optimize image size
    
  devops-community/tools:  # Verified: DevOps Community
    ✓ docker-compose-ai  AI-assisted compose files
```

### Community Repos

Unverified but searchable:

```bash
$ scmd search "game dev"

  community/game-tools:   # Community (unverified)
    ⚠ unity-scripts     Generate Unity boilerplate
    ⚠ godot-helper      GDScript assistance
```

### Enterprise/Private Repos

Internal repos with authentication:

```bash
# Add private repo
$ scmd repos add acme-corp/internal \
    --url https://git.acme.com/scmd-commands \
    --auth token \
    --token-env ACME_GIT_TOKEN

# Private repos don't appear in public search
# But are searchable locally
$ scmd search deploy
  acme-corp/internal:
    🔒 k8s-deploy       Deploy to ACME K8s clusters
    🔒 db-migrate       Run ACME database migrations
```

## 10.5 Publishing Workflow

### Via GitHub (Recommended)

```yaml
# .github/workflows/publish.yml

name: Publish to OneSkill

on:
  release:
    types: [published]

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Validate commands
        uses: oneskill/validate-action@v1
        
      - name: Publish to OneSkill
        uses: oneskill/publish-action@v1
        with:
          token: ${{ secrets.ONESKILL_TOKEN }}
```

### Via CLI

```bash
# Login to OneSkill
$ scmd publish login
Opening browser for authentication...
✓ Logged in as @sandeep

# Validate repo
$ scmd publish validate
Validating scmd-repo.yaml... ✓
Validating commands/k8s-deploy/command.yaml... ✓
Validating commands/docker-build/command.yaml... ✓
✓ All validations passed

# Publish
$ scmd publish
Publishing acme-tools v1.2.0 to OneSkill...
✓ Published successfully

View at: https://oneskill.dev/repos/sandeep/acme-tools
```

## 10.6 Repo Manifest for OneSkill

```yaml
# scmd-repo.yaml (with OneSkill metadata)

name: sandeep/devtools
version: "1.2.0"
description: Sandeep's developer productivity commands

# OneSkill-specific metadata
oneskill:
  # Visibility: public, unlisted, or private
  visibility: public
  
  # Categories for discovery
  categories:
    - devops
    - python
    - automation
    
  # Keywords for search
  keywords:
    - django
    - kubernetes
    - docker
    - ci-cd
    
  # Funding/sponsorship links
  funding:
    github: sandeep
    
  # Required for verified publisher status
  verification:
    domain: sandeep.dev
    github: sandeep

maintainers:
  - name: Sandeep
    email: sandeep@example.com
    url: https://sandeep.dev

requires:
  scmd: ">=0.2.0"

commands:
  django-migrate:
    version: "1.2.0"
    path: commands/django-migrate
    description: Generate Django migrations intelligently
    tags: [django, python, database]
    
  k8s-deploy:
    version: "1.0.0"
    path: commands/k8s-deploy
    description: Deploy to Kubernetes with validation
    tags: [kubernetes, devops]
```

## 10.7 Installation UX

```bash
# Discover
$ scmd search django
  oneskill/official:
    ✓ django-migrate    Generate Django migrations (v1.2.0)
    ✓ django-admin      Generate admin configs (v1.0.0)
    
  sandeep/devtools:     # Community
    django-test         Generate Django tests (v0.5.0)

# Install with permission prompt
$ scmd install oneskill/official/django-migrate

Installing django-migrate v1.2.0 from oneskill/official...

⚠ Permission Request

This command requests:
  📁 Read files:
     */models.py
     */migrations/*.py
  🖥️  Execute commands:
     python manage.py makemigrations --dry-run
     python manage.py showmigrations

Grant permissions? [Yes/Ask each time/No]: Yes

✓ Installed django-migrate v1.2.0

# Use immediately
$ scmd django-migrate users
```

## 10.8 Enterprise Features

For organizations using scmd at scale:

```yaml
# ~/.scmd/config.yaml (enterprise)

repositories:
  # Corporate registry (mirrors + private)
  - name: acme/registry
    url: https://scmd-registry.acme.com
    enabled: true
    auth:
      type: oauth
      provider: okta
      
# Policy: only allow verified repos
policies:
  allow_unverified: false
  allowed_repos:
    - "oneskill/*"
    - "acme/*"
    - "docker/*"
  blocked_repos:
    - "untrusted/*"
    
# Audit logging
audit:
  enabled: true
  destination: https://logs.acme.com/scmd
```

## 10.9 OneSkill Business Model Integration

| Tier | Features |
|------|----------|
| **Free** | Public repos, community support, basic analytics |
| **Pro** | Private repos, priority support, advanced analytics |
| **Team** | Shared private repos, team management, SSO |
| **Enterprise** | Self-hosted registry, audit logs, compliance features |

scmd drives adoption → OneSkill monetizes on distribution/enterprise features.

---

# 9. Tool Calling Framework

## 9.1 Tool Executor

```go
// internal/tools/executor.go

package tools

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"
    "strings"
    "time"

    "github.com/scmd/scmd/internal/permissions"
)

// Executor executes tools with permission checking
type Executor struct {
    perms  *permissions.Manager
    ui     UI
    config *Config
}

// Config for tool executor
type Config struct {
    DefaultTimeout     int
    RequireConfirmation bool
    Sandbox            bool
}

// NewExecutor creates a new tool executor
func NewExecutor(perms *permissions.Manager, ui UI, cfg *Config) *Executor {
    return &Executor{
        perms:  perms,
        ui:     ui,
        config: cfg,
    }
}

// Execute runs a tool
func (e *Executor) Execute(
    ctx context.Context,
    tool *Tool,
    params map[string]any,
) (*Result, error) {
    // Check permissions
    if err := e.perms.Check(tool); err != nil {
        return nil, fmt.Errorf("permission denied: %w", err)
    }

    // Confirm if required
    if tool.RequiresConfirmation || e.config.RequireConfirmation {
        if !e.confirm(tool, params) {
            return &Result{
                Success: false,
                Error:   "cancelled by user",
            }, nil
        }
    }

    // Execute based on type
    start := time.Now()
    var result *Result
    var err error

    switch tool.Type {
    case TypeShell:
        result, err = e.executeShell(ctx, tool, params)
    case TypeFileRead:
        result, err = e.executeFileRead(ctx, params)
    case TypeFileWrite:
        result, err = e.executeFileWrite(ctx, params)
    case TypeFileList:
        result, err = e.executeFileList(ctx, params)
    case TypeHTTP:
        result, err = e.executeHTTP(ctx, tool, params)
    case TypeScmd:
        result, err = e.executeScmd(ctx, params)
    default:
        return nil, fmt.Errorf("unknown tool type: %s", tool.Type)
    }

    if result != nil {
        result.Duration = time.Since(start).Milliseconds()
    }

    return result, err
}

func (e *Executor) executeShell(
    ctx context.Context,
    tool *Tool,
    params map[string]any,
) (*Result, error) {
    // Interpolate command
    command := tool.Command
    for key, value := range params {
        placeholder := fmt.Sprintf("{{%s}}", key)
        command = strings.ReplaceAll(command, placeholder, fmt.Sprint(value))
    }

    // Show what's being executed
    e.ui.WriteLine(fmt.Sprintf("$ %s", command))

    // Set timeout
    timeout := tool.Timeout
    if timeout == 0 {
        timeout = e.config.DefaultTimeout
    }
    ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
    defer cancel()

    // Execute
    cmd := exec.CommandContext(ctx, "sh", "-c", command)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()

    exitCode := 0
    if err != nil {
        if exitErr, ok := err.(*exec.ExitError); ok {
            exitCode = exitErr.ExitCode()
        }
    }

    return &Result{
        Success:  exitCode == 0,
        Output:   stdout.String(),
        Error:    stderr.String(),
        ExitCode: exitCode,
    }, nil
}

func (e *Executor) executeFileRead(
    ctx context.Context,
    params map[string]any,
) (*Result, error) {
    path, ok := params["path"].(string)
    if !ok {
        return nil, fmt.Errorf("path parameter required")
    }

    // Validate path is allowed
    if err := e.perms.CheckFilePath(path, "read"); err != nil {
        return nil, err
    }

    content, err := os.ReadFile(path)
    if err != nil {
        return &Result{
            Success: false,
            Error:   err.Error(),
        }, nil
    }

    return &Result{
        Success: true,
        Output:  string(content),
    }, nil
}

func (e *Executor) executeFileWrite(
    ctx context.Context,
    params map[string]any,
) (*Result, error) {
    path, _ := params["path"].(string)
    content, _ := params["content"].(string)

    // Validate path is allowed
    if err := e.perms.CheckFilePath(path, "write"); err != nil {
        return nil, err
    }

    if err := os.WriteFile(path, []byte(content), 0644); err != nil {
        return &Result{
            Success: false,
            Error:   err.Error(),
        }, nil
    }

    return &Result{
        Success: true,
        Output:  fmt.Sprintf("wrote %d bytes to %s", len(content), path),
    }, nil
}

func (e *Executor) confirm(tool *Tool, params map[string]any) bool {
    e.ui.WriteLine("")
    e.ui.WriteLine("⚠ Tool execution requested:")
    e.ui.WriteLine(fmt.Sprintf("  Tool: %s", tool.Name))
    e.ui.WriteLine(fmt.Sprintf("  Type: %s", tool.Type))

    if tool.Type == TypeShell {
        cmd := tool.Command
        for k, v := range params {
            cmd = strings.ReplaceAll(cmd, fmt.Sprintf("{{%s}}", k), fmt.Sprint(v))
        }
        e.ui.WriteLine(fmt.Sprintf("  Command: %s", cmd))
    }

    e.ui.WriteLine("")
    return e.ui.Confirm("Execute? [y/N]")
}
```

## 9.2 Tool Calling Engine (Agentic)

```go
// internal/tools/toolcall.go

package tools

import (
    "context"
    "fmt"

    "github.com/scmd/scmd/internal/backend"
)

// ToolCallEngine handles agentic tool calling loops
type ToolCallEngine struct {
    backend  backend.Backend
    executor *Executor
    config   *ToolCallConfig
}

// ToolCallConfig configures tool calling behavior
type ToolCallConfig struct {
    MaxIterations      int
    AutoApprove        []string
    RequireConfirmation []string
}

// Event from tool calling loop
type Event struct {
    Type       EventType
    Content    string
    ToolName   string
    Parameters map[string]any
    Result     *Result
    Error      error
}

// EventType of tool calling event
type EventType string

const (
    EventText           EventType = "text"
    EventToolCallStart  EventType = "tool_call_start"
    EventToolCallResult EventType = "tool_call_result"
    EventComplete       EventType = "complete"
    EventMaxIterations  EventType = "max_iterations"
)

// NewToolCallEngine creates a new engine
func NewToolCallEngine(
    b backend.Backend,
    exec *Executor,
    cfg *ToolCallConfig,
) *ToolCallEngine {
    if cfg.MaxIterations == 0 {
        cfg.MaxIterations = 10
    }
    return &ToolCallEngine{
        backend:  b,
        executor: exec,
        config:   cfg,
    }
}

// Run executes an agentic loop with tools
func (e *ToolCallEngine) Run(
    ctx context.Context,
    prompt string,
    systemPrompt string,
    tools []*Tool,
) (<-chan Event, error) {
    if !e.backend.SupportsToolCalling() {
        return nil, fmt.Errorf("backend does not support tool calling")
    }

    ch := make(chan Event)

    go func() {
        defer close(ch)

        messages := []backend.Message{
            {Role: "user", Content: prompt},
        }

        toolDefs := e.convertTools(tools)

        for iteration := 0; iteration < e.config.MaxIterations; iteration++ {
            // Call LLM with tools
            resp, err := e.backend.CompleteWithTools(ctx, &backend.ToolRequest{
                Messages:     messages,
                SystemPrompt: systemPrompt,
                Tools:        toolDefs,
            })
            if err != nil {
                ch <- Event{Type: EventComplete, Error: err}
                return
            }

            // Emit text content
            if resp.Content != "" {
                ch <- Event{Type: EventText, Content: resp.Content}
            }

            // Check for tool calls
            if len(resp.ToolCalls) == 0 {
                ch <- Event{Type: EventComplete}
                return
            }

            // Execute tool calls
            for _, tc := range resp.ToolCalls {
                ch <- Event{
                    Type:       EventToolCallStart,
                    ToolName:   tc.Name,
                    Parameters: tc.Parameters,
                }

                // Find tool
                tool := e.findTool(tools, tc.Name)
                if tool == nil {
                    ch <- Event{
                        Type:     EventToolCallResult,
                        ToolName: tc.Name,
                        Error:    fmt.Errorf("unknown tool: %s", tc.Name),
                    }
                    continue
                }

                // Execute
                result, err := e.executor.Execute(ctx, tool, tc.Parameters)
                if err != nil {
                    ch <- Event{
                        Type:     EventToolCallResult,
                        ToolName: tc.Name,
                        Error:    err,
                    }
                    continue
                }

                ch <- Event{
                    Type:     EventToolCallResult,
                    ToolName: tc.Name,
                    Result:   result,
                }

                // Add result to messages
                resultContent := result.Output
                if !result.Success {
                    resultContent = fmt.Sprintf("Error: %s", result.Error)
                }
                messages = append(messages, backend.Message{
                    Role:       "tool",
                    ToolCallID: tc.ID,
                    Content:    resultContent,
                })
            }
        }

        ch <- Event{Type: EventMaxIterations}
    }()

    return ch, nil
}

func (e *ToolCallEngine) convertTools(tools []*Tool) []backend.ToolDef {
    defs := make([]backend.ToolDef, len(tools))
    for i, t := range tools {
        defs[i] = backend.ToolDef{
            Name:        t.Name,
            Description: t.Description,
            Parameters:  t.Parameters,
        }
    }
    return defs
}

func (e *ToolCallEngine) findTool(tools []*Tool, name string) *Tool {
    for _, t := range tools {
        if t.Name == name {
            return t
        }
    }
    return nil
}
```

---

# 10. Permission System

## 10.1 Permission Manager

```go
// internal/permissions/manager.go

package permissions

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"

    "gopkg.in/yaml.v3"
)

// Manager handles permission checking and grants
type Manager struct {
    mu        sync.RWMutex
    grants    map[string]*Grant
    trustLevels map[string]TrustLevel
    grantPath string
}

// Grant represents a permission grant for a command
type Grant struct {
    Command     string      `yaml:"command"`
    Repo        string      `yaml:"repo"`
    Permissions Permissions `yaml:"permissions"`
    GrantedAt   string      `yaml:"granted_at"`
    GrantedBy   string      `yaml:"granted_by"`
}

// Permissions for a command
type Permissions struct {
    Files   FilePerms   `yaml:"files"`
    Shell   ShellPerms  `yaml:"shell"`
    Git     GitPerms    `yaml:"git"`
    Network NetworkPerms `yaml:"network"`
    Env     EnvPerms    `yaml:"env"`
}

// FilePerms for file system access
type FilePerms struct {
    Read  []string `yaml:"read"`
    Write []string `yaml:"write"`
}

// ShellPerms for shell execution
type ShellPerms struct {
    Commands       []string `yaml:"commands"`
    AllowArbitrary bool     `yaml:"allow_arbitrary"`
}

// GitPerms for git operations
type GitPerms struct {
    Read  bool `yaml:"read"`
    Write bool `yaml:"write"`
}

// NetworkPerms for network access
type NetworkPerms struct {
    AllowedHosts []string `yaml:"allowed_hosts"`
}

// EnvPerms for environment variables
type EnvPerms struct {
    Read  []string `yaml:"read"`
    Write []string `yaml:"write"`
}

// TrustLevel for repositories
type TrustLevel string

const (
    TrustFull   TrustLevel = "full"
    TrustHigh   TrustLevel = "high"
    TrustMedium TrustLevel = "medium"
    TrustLow    TrustLevel = "low"
)

// NewManager creates a new permission manager
func NewManager(grantPath string) *Manager {
    return &Manager{
        grants:      make(map[string]*Grant),
        trustLevels: make(map[string]TrustLevel),
        grantPath:   grantPath,
    }
}

// Load loads grants from disk
func (m *Manager) Load() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    data, err := os.ReadFile(m.grantPath)
    if os.IsNotExist(err) {
        return nil
    }
    if err != nil {
        return err
    }

    var grants []*Grant
    if err := yaml.Unmarshal(data, &grants); err != nil {
        return err
    }

    for _, g := range grants {
        key := fmt.Sprintf("%s/%s", g.Repo, g.Command)
        m.grants[key] = g
    }
    return nil
}

// Check verifies permission for a tool
func (m *Manager) Check(tool *Tool) error {
    // Built-in tools are always allowed
    if tool.BuiltIn {
        return nil
    }

    key := fmt.Sprintf("%s/%s", tool.Repo, tool.Command)

    m.mu.RLock()
    grant, exists := m.grants[key]
    m.mu.RUnlock()

    if !exists {
        return fmt.Errorf("no permission grant for %s", key)
    }

    // Check specific permission based on tool type
    switch tool.Type {
    case TypeShell:
        if !m.checkShellPerm(tool.Command, grant.Permissions.Shell) {
            return fmt.Errorf("shell command not permitted: %s", tool.Command)
        }
    case TypeFileRead:
        // Checked per-call in executor
    case TypeFileWrite:
        // Checked per-call in executor
    case TypeHTTP:
        // Checked per-call based on host
    }

    return nil
}

// CheckFilePath validates file path against grants
func (m *Manager) CheckFilePath(path, mode string) error {
    m.mu.RLock()
    defer m.mu.RUnlock()

    absPath, err := filepath.Abs(path)
    if err != nil {
        return err
    }

    // Check against all grants
    for _, grant := range m.grants {
        patterns := grant.Permissions.Files.Read
        if mode == "write" {
            patterns = grant.Permissions.Files.Write
        }

        for _, pattern := range patterns {
            matched, err := filepath.Match(pattern, absPath)
            if err != nil {
                continue
            }
            if matched {
                return nil
            }
        }
    }

    return fmt.Errorf("file access not permitted: %s (%s)", path, mode)
}

// RequestGrant prompts user for permission
func (m *Manager) RequestGrant(
    command string,
    repo string,
    perms Permissions,
    ui UI,
) (bool, error) {
    ui.WriteLine("")
    ui.WriteLine("⚠ Permission Request")
    ui.WriteLine(fmt.Sprintf("Command: %s (%s)", command, repo))
    ui.WriteLine("")
    ui.WriteLine("This command requests the following permissions:")
    ui.WriteLine("")

    if len(perms.Files.Read) > 0 {
        ui.WriteLine("  📁 Read files:")
        for _, p := range perms.Files.Read {
            ui.WriteLine(fmt.Sprintf("     %s", p))
        }
    }

    if len(perms.Files.Write) > 0 {
        ui.WriteLine("  ✏️  Write files:")
        for _, p := range perms.Files.Write {
            ui.WriteLine(fmt.Sprintf("     %s", p))
        }
    }

    if len(perms.Shell.Commands) > 0 || perms.Shell.AllowArbitrary {
        ui.WriteLine("  🖥️  Execute commands:")
        if perms.Shell.AllowArbitrary {
            ui.WriteLine("     ⚠ ANY shell command")
        } else {
            for _, c := range perms.Shell.Commands {
                ui.WriteLine(fmt.Sprintf("     %s", c))
            }
        }
    }

    if len(perms.Network.AllowedHosts) > 0 {
        ui.WriteLine("  🌐 Network access:")
        for _, h := range perms.Network.AllowedHosts {
            ui.WriteLine(fmt.Sprintf("     %s", h))
        }
    }

    ui.WriteLine("")

    choice := ui.Select("Grant permissions?", []string{
        "Yes, grant all",
        "Yes, ask each time",
        "No, cancel",
    })

    if choice == "No, cancel" {
        return false, nil
    }

    // Store grant
    m.mu.Lock()
    key := fmt.Sprintf("%s/%s", repo, command)
    m.grants[key] = &Grant{
        Command:     command,
        Repo:        repo,
        Permissions: perms,
        GrantedAt:   time.Now().Format(time.RFC3339),
        GrantedBy:   "user",
    }
    m.mu.Unlock()

    return true, m.save()
}

func (m *Manager) checkShellPerm(command string, perms ShellPerms) bool {
    if perms.AllowArbitrary {
        return true
    }

    for _, pattern := range perms.Commands {
        if matchCommand(command, pattern) {
            return true
        }
    }
    return false
}

func (m *Manager) save() error {
    m.mu.RLock()
    defer m.mu.RUnlock()

    grants := make([]*Grant, 0, len(m.grants))
    for _, g := range m.grants {
        grants = append(grants, g)
    }

    data, err := yaml.Marshal(grants)
    if err != nil {
        return err
    }

    return os.WriteFile(m.grantPath, data, 0600)
}
```

---

# 11. Configuration

## 11.1 Config Structure

```go
// internal/config/config.go

package config

import (
    "os"
    "path/filepath"

    "github.com/spf13/viper"
)

// Config represents scmd configuration
type Config struct {
    Version string `mapstructure:"version"`

    // Backend configuration
    Backends BackendsConfig `mapstructure:"backends"`

    // Repository configuration
    Repositories []RepoConfig `mapstructure:"repositories"`

    // Tool calling configuration
    Tools ToolsConfig `mapstructure:"tools"`

    // Permissions configuration
    Permissions PermissionsConfig `mapstructure:"permissions"`

    // UI preferences
    UI UIConfig `mapstructure:"ui"`

    // Model management
    Models ModelsConfig `mapstructure:"models"`
}

// BackendsConfig for LLM backends
type BackendsConfig struct {
    Default string             `mapstructure:"default"`
    Local   LocalBackendConfig `mapstructure:"local"`
    Ollama  OllamaConfig       `mapstructure:"ollama"`
    Claude  ClaudeConfig       `mapstructure:"claude"`
    OpenAI  OpenAIConfig       `mapstructure:"openai"`
}

// LocalBackendConfig for local llama.cpp
type LocalBackendConfig struct {
    Model         string `mapstructure:"model"`
    ModelPath     string `mapstructure:"model_path"`
    ContextLength int    `mapstructure:"context_length"`
    GPULayers     int    `mapstructure:"gpu_layers"`
    Threads       int    `mapstructure:"threads"`
}

// OllamaConfig for Ollama backend
type OllamaConfig struct {
    Host  string `mapstructure:"host"`
    Model string `mapstructure:"model"`
}

// ClaudeConfig for Claude API
type ClaudeConfig struct {
    Model string `mapstructure:"model"`
    // API key via ANTHROPIC_API_KEY env var
}

// OpenAIConfig for OpenAI API
type OpenAIConfig struct {
    Model   string `mapstructure:"model"`
    BaseURL string `mapstructure:"base_url"`
    // API key via OPENAI_API_KEY env var
}

// RepoConfig for a repository
type RepoConfig struct {
    Name    string     `mapstructure:"name"`
    URL     string     `mapstructure:"url"`
    Enabled bool       `mapstructure:"enabled"`
    Auth    AuthConfig `mapstructure:"auth"`
}

// AuthConfig for repository authentication
type AuthConfig struct {
    Type   string `mapstructure:"type"` // none, token, oauth, basic
    EnvVar string `mapstructure:"env_var"`
}

// ToolsConfig for tool execution
type ToolsConfig struct {
    Enabled       bool              `mapstructure:"enabled"`
    MaxIterations int               `mapstructure:"max_iterations"`
    Confirmation  ConfirmationConfig `mapstructure:"confirmation"`
    Timeouts      TimeoutsConfig    `mapstructure:"timeouts"`
}

// ConfirmationConfig for tool confirmation
type ConfirmationConfig struct {
    AlwaysConfirm  []string `mapstructure:"always_confirm"`
    NeverConfirm   []string `mapstructure:"never_confirm"`
    RememberChoice bool     `mapstructure:"remember_choice"`
}

// TimeoutsConfig for tool timeouts
type TimeoutsConfig struct {
    Shell int `mapstructure:"shell"`
    HTTP  int `mapstructure:"http"`
    File  int `mapstructure:"file"`
}

// PermissionsConfig for trust levels
type PermissionsConfig struct {
    TrustLevels map[string]string `mapstructure:"trust_levels"`
}

// UIConfig for UI preferences
type UIConfig struct {
    Streaming     bool `mapstructure:"streaming"`
    Colors        bool `mapstructure:"colors"`
    Verbose       bool `mapstructure:"verbose"`
    ShowToolCalls bool `mapstructure:"show_tool_calls"`
    ConfirmWrites bool `mapstructure:"confirm_writes"`
}

// ModelsConfig for model management
type ModelsConfig struct {
    Directory    string `mapstructure:"directory"`
    AutoDownload bool   `mapstructure:"auto_download"`
}

// Default returns default configuration
func Default() *Config {
    home, _ := os.UserHomeDir()
    return &Config{
        Version: "1.0",
        Backends: BackendsConfig{
            Default: "local",
            Local: LocalBackendConfig{
                Model:         "qwen2.5-coder-1.5b",
                ContextLength: 8192,
                GPULayers:     0,
                Threads:       0, // Auto-detect
            },
            Ollama: OllamaConfig{
                Host:  "http://localhost:11434",
                Model: "codellama",
            },
            Claude: ClaudeConfig{
                Model: "claude-sonnet-4-20250514",
            },
            OpenAI: OpenAIConfig{
                Model: "gpt-4o",
            },
        },
        Repositories: []RepoConfig{
            {
                Name:    "oneskill/official",
                URL:     "https://github.com/oneskill/scmd-plugins",
                Enabled: true,
            },
        },
        Tools: ToolsConfig{
            Enabled:       true,
            MaxIterations: 10,
            Confirmation: ConfirmationConfig{
                AlwaysConfirm:  []string{"shell:*", "files:write"},
                NeverConfirm:   []string{"files:read", "git:read"},
                RememberChoice: true,
            },
            Timeouts: TimeoutsConfig{
                Shell: 30,
                HTTP:  10,
                File:  5,
            },
        },
        Permissions: PermissionsConfig{
            TrustLevels: map[string]string{
                "built-in":           "full",
                "oneskill/official":  "high",
                "*":                  "low",
            },
        },
        UI: UIConfig{
            Streaming:     true,
            Colors:        true,
            Verbose:       false,
            ShowToolCalls: true,
            ConfirmWrites: true,
        },
        Models: ModelsConfig{
            Directory:    filepath.Join(home, ".scmd", "models"),
            AutoDownload: true,
        },
    }
}
```

## 11.2 Default Config File

```yaml
# ~/.scmd/config.yaml

version: "1.0"

backends:
  default: local
  
  local:
    model: qwen2.5-coder-1.5b
    context_length: 8192
    # gpu_layers: 0          # GPU layers to offload (0 = CPU only)
    # threads: 4             # CPU threads (auto if not set)
  
  # ollama:
  #   host: http://localhost:11434
  #   model: codellama
  
  # claude:
  #   model: claude-sonnet-4-20250514
  
  # openai:
  #   model: gpt-4o

repositories:
  - name: oneskill/official
    url: https://github.com/oneskill/scmd-plugins
    enabled: true

tools:
  enabled: true
  max_iterations: 10
  confirmation:
    always_confirm:
      - "shell:*"
      - "files:write"
    never_confirm:
      - "files:read"
      - "git:read"
    remember_choice: true
  timeouts:
    shell: 30
    http: 10
    file: 5

permissions:
  trust_levels:
    built-in: full
    oneskill/official: high
    "*": low

ui:
  streaming: true
  colors: true
  verbose: false
  show_tool_calls: true
  confirm_writes: true

models:
  directory: ~/.scmd/models
  auto_download: true
```

---

# 12. Distribution

## 12.1 GoReleaser Configuration

```yaml
# .goreleaser.yaml

project_name: scmd

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - id: scmd
    main: ./cmd/scmd
    binary: scmd
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X github.com/scmd/scmd/pkg/version.Version={{.Version}}
      - -X github.com/scmd/scmd/pkg/version.Commit={{.Commit}}
      - -X github.com/scmd/scmd/pkg/version.Date={{.Date}}

archives:
  - id: default
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    name_template: "{{ .ProjectName }}-{{ .Os }}-{{ .Arch }}"
    files:
      - LICENSE
      - README.md

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^ci:'

brews:
  - repository:
      owner: scmd
      name: homebrew-tap
    homepage: "https://github.com/scmd/scmd"
    description: "AI-powered slash commands in your terminal"
    license: "MIT"
    install: |
      bin.install "scmd"
    test: |
      system "#{bin}/scmd", "--version"

# npm distribution via custom script
publishers:
  - name: npm
    cmd: ./scripts/publish-npm.sh {{ .Version }}
    env:
      - NPM_TOKEN={{ .Env.NPM_TOKEN }}

# curl installer
release:
  extra_files:
    - glob: ./scripts/install.sh
```

## 12.2 npm Package Structure

```json
// npm/package.json
{
  "name": "scmd",
  "version": "0.1.0",
  "description": "AI-powered slash commands in your terminal",
  "bin": {
    "scmd": "bin/scmd"
  },
  "scripts": {
    "postinstall": "node install.js"
  },
  "optionalDependencies": {
    "@scmd/darwin-arm64": "0.1.0",
    "@scmd/darwin-x64": "0.1.0",
    "@scmd/linux-x64": "0.1.0",
    "@scmd/linux-arm64": "0.1.0",
    "@scmd/win32-x64": "0.1.0"
  },
  "keywords": [
    "ai",
    "llm",
    "cli",
    "developer-tools",
    "slash-commands"
  ],
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/scmd/scmd"
  }
}
```

```javascript
// npm/install.js
const os = require('os');
const path = require('path');
const fs = require('fs');

const PLATFORM_MAP = {
  'darwin-arm64': '@scmd/darwin-arm64',
  'darwin-x64': '@scmd/darwin-x64',
  'linux-x64': '@scmd/linux-x64',
  'linux-arm64': '@scmd/linux-arm64',
  'win32-x64': '@scmd/win32-x64',
};

function install() {
  const platform = `${os.platform()}-${os.arch()}`;
  const pkg = PLATFORM_MAP[platform];

  if (!pkg) {
    console.error(`Unsupported platform: ${platform}`);
    console.error('Please install manually from: https://github.com/scmd/scmd/releases');
    process.exit(1);
  }

  try {
    const binaryName = os.platform() === 'win32' ? 'scmd.exe' : 'scmd';
    const sourcePath = require.resolve(`${pkg}/bin/${binaryName}`);
    const targetDir = path.join(__dirname, 'bin');
    const targetPath = path.join(targetDir, binaryName);

    if (!fs.existsSync(targetDir)) {
      fs.mkdirSync(targetDir, { recursive: true });
    }

    fs.copyFileSync(sourcePath, targetPath);
    fs.chmodSync(targetPath, 0o755);

    console.log(`✓ scmd installed successfully`);
  } catch (err) {
    console.error(`Failed to install scmd: ${err.message}`);
    process.exit(1);
  }
}

install();
```

## 12.3 Homebrew Formula

```ruby
# Formula/scmd.rb
class Scmd < Formula
  desc "AI-powered slash commands in your terminal"
  homepage "https://github.com/scmd/scmd"
  version "0.1.0"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/scmd/scmd/releases/download/v#{version}/scmd-darwin-arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"
    end
    on_intel do
      url "https://github.com/scmd/scmd/releases/download/v#{version}/scmd-darwin-x64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/scmd/scmd/releases/download/v#{version}/scmd-linux-arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"
    end
    on_intel do
      url "https://github.com/scmd/scmd/releases/download/v#{version}/scmd-linux-x64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256"
    end
  end

  def install
    bin.install "scmd"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/scmd --version")
  end
end
```

## 12.4 Install Script

```bash
#!/bin/sh
# scripts/install.sh

set -e

REPO="scmd/scmd"
INSTALL_DIR="${SCMD_INSTALL_DIR:-/usr/local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() { echo "${GREEN}$1${NC}"; }
warn() { echo "${YELLOW}$1${NC}"; }
error() { echo "${RED}$1${NC}" >&2; }

# Detect platform
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) error "Unsupported architecture: $ARCH"; exit 1 ;;
    esac

    case "$OS" in
        linux|darwin) ;;
        mingw*|msys*|cygwin*) OS="windows" ;;
        *) error "Unsupported OS: $OS"; exit 1 ;;
    esac

    PLATFORM="${OS}-${ARCH}"
}

# Get latest version
get_latest_version() {
    VERSION=$(curl -sI "https://github.com/$REPO/releases/latest" | grep -i "location:" | sed 's/.*tag\/\(.*\)\r/\1/')
    if [ -z "$VERSION" ]; then
        error "Failed to get latest version"
        exit 1
    fi
}

# Download and install
install() {
    detect_platform
    get_latest_version

    EXT="tar.gz"
    [ "$OS" = "windows" ] && EXT="zip"

    URL="https://github.com/$REPO/releases/download/$VERSION/scmd-${PLATFORM}.${EXT}"
    
    info "Downloading scmd $VERSION for $PLATFORM..."

    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    if ! curl -sL "$URL" -o "$TMP_DIR/scmd.$EXT"; then
        error "Failed to download scmd"
        exit 1
    fi

    if [ "$EXT" = "zip" ]; then
        unzip -q "$TMP_DIR/scmd.$EXT" -d "$TMP_DIR"
    else
        tar -xzf "$TMP_DIR/scmd.$EXT" -C "$TMP_DIR"
    fi

    BINARY="scmd"
    [ "$OS" = "windows" ] && BINARY="scmd.exe"

    if [ ! -w "$INSTALL_DIR" ]; then
        warn "Installing to $INSTALL_DIR requires sudo..."
        sudo mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/"
        sudo chmod +x "$INSTALL_DIR/$BINARY"
    else
        mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/$BINARY"
    fi

    info "✓ scmd $VERSION installed to $INSTALL_DIR/$BINARY"
    echo ""
    echo "Run 'scmd' to get started"
}

install
```

---

# 13. Testing Strategy

## 13.1 Test Structure

```
tests/
├── unit/                       # Unit tests (in-package)
│   ├── command/
│   │   ├── registry_test.go
│   │   ├── parser_test.go
│   │   └── builtin/
│   │       ├── explain_test.go
│   │       ├── review_test.go
│   │       └── commit_test.go
│   ├── backend/
│   │   ├── registry_test.go
│   │   └── mock/
│   │       └── backend_test.go
│   ├── context/
│   │   ├── git_test.go
│   │   ├── project_test.go
│   │   └── files_test.go
│   ├── tools/
│   │   ├── executor_test.go
│   │   └── sandbox_test.go
│   ├── plugins/
│   │   ├── loader_test.go
│   │   ├── parser_test.go
│   │   └── validator_test.go
│   ├── repos/
│   │   ├── manager_test.go
│   │   └── fetcher_test.go
│   ├── permissions/
│   │   └── manager_test.go
│   └── config/
│       └── loader_test.go
│
├── integration/                # Integration tests
│   ├── backend_test.go         # LLM backend integration
│   ├── cli_test.go             # CLI command integration
│   ├── plugin_test.go          # Plugin system integration
│   ├── repo_test.go            # Repository integration
│   └── tools_test.go           # Tool execution integration
│
├── e2e/                        # End-to-end tests
│   ├── commands_test.go        # Full command workflows
│   ├── install_test.go         # Installation flows
│   └── workflows_test.go       # User workflow scenarios
│
├── fixtures/                   # Test data
│   ├── projects/               # Sample projects
│   │   ├── nodejs/
│   │   ├── python/
│   │   └── go/
│   ├── plugins/                # Sample plugins
│   │   └── test-plugin/
│   ├── repos/                  # Sample repo manifests
│   │   └── test-repo/
│   └── responses/              # Mock LLM responses
│       ├── explain.json
│       ├── review.json
│       └── commit.json
│
└── testutil/                   # Test utilities
    ├── mock_backend.go
    ├── mock_ui.go
    ├── fixtures.go
    └── assertions.go
```

## 13.2 Unit Test Example

```go
// internal/command/builtin/explain_test.go

package builtin

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/scmd/scmd/internal/command"
    "github.com/scmd/scmd/tests/testutil"
)

func TestExplainCommand_Metadata(t *testing.T) {
    cmd := NewExplainCommand()

    assert.Equal(t, "explain", cmd.Name())
    assert.Contains(t, cmd.Aliases(), "e")
    assert.Contains(t, cmd.Aliases(), "exp")
    assert.True(t, cmd.RequiresBackend())
    assert.Equal(t, command.CategoryCode, cmd.Category())
}

func TestExplainCommand_Validate(t *testing.T) {
    cmd := NewExplainCommand()

    tests := []struct {
        name    string
        args    *command.Args
        wantErr bool
    }{
        {
            name:    "valid file path",
            args:    &command.Args{Positional: []string{"file.go"}},
            wantErr: false,
        },
        {
            name:    "missing file path",
            args:    &command.Args{Positional: []string{}},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := cmd.Validate(tt.args)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestExplainCommand_Execute(t *testing.T) {
    // Create temp file
    tmpDir := t.TempDir()
    testFile := filepath.Join(tmpDir, "test.go")
    err := os.WriteFile(testFile, []byte("package main\n\nfunc main() {}"), 0644)
    require.NoError(t, err)

    tests := []struct {
        name       string
        args       *command.Args
        wantSuccess bool
        wantError  string
    }{
        {
            name: "explain existing file",
            args: &command.Args{
                Positional: []string{testFile},
                Raw:        "/explain " + testFile,
            },
            wantSuccess: true,
        },
        {
            name: "explain non-existent file",
            args: &command.Args{
                Positional: []string{"/nonexistent/file.go"},
                Raw:        "/explain /nonexistent/file.go",
            },
            wantSuccess: false,
            wantError:   "not found",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := NewExplainCommand()

            // Create mock context
            mockBackend := testutil.NewMockBackend()
            mockBackend.SetResponse("This is a simple Go main function...")
            
            mockUI := testutil.NewMockUI()

            execCtx := &command.ExecContext{
                Backend: mockBackend,
                UI:      mockUI,
                Config:  testutil.DefaultConfig(),
            }

            result, err := cmd.Execute(context.Background(), tt.args, execCtx)

            require.NoError(t, err)
            assert.Equal(t, tt.wantSuccess, result.Success)

            if tt.wantError != "" {
                assert.Contains(t, result.Error, tt.wantError)
            }
        })
    }
}

func TestDetectLanguage(t *testing.T) {
    tests := []struct {
        path string
        want string
    }{
        {"file.go", "go"},
        {"file.py", "python"},
        {"file.js", "javascript"},
        {"file.ts", "typescript"},
        {"file.rs", "rust"},
        {"file.unknown", "code"},
    }

    for _, tt := range tests {
        t.Run(tt.path, func(t *testing.T) {
            got := detectLanguage(tt.path)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## 13.3 Integration Test Example

```go
// tests/integration/cli_test.go

package integration

import (
    "bytes"
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCLI_Help(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "scmd", "--help")
    output, err := cmd.Output()

    require.NoError(t, err)
    assert.Contains(t, string(output), "scmd")
    assert.Contains(t, string(output), "explain")
    assert.Contains(t, string(output), "review")
    assert.Contains(t, string(output), "commit")
}

func TestCLI_Version(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "scmd", "--version")
    output, err := cmd.Output()

    require.NoError(t, err)
    assert.Contains(t, string(output), "scmd")
}

func TestCLI_ExplainCommand(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // Create test file
    tmpDir := t.TempDir()
    testFile := filepath.Join(tmpDir, "test.go")
    err := os.WriteFile(testFile, []byte(`
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
`), 0644)
    require.NoError(t, err)

    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "scmd", "explain", testFile)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err = cmd.Run()

    require.NoError(t, err, "stderr: %s", stderr.String())
    assert.NotEmpty(t, stdout.String())
}

func TestCLI_ConfigCommand(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "scmd", "config", "show")
    output, err := cmd.Output()

    require.NoError(t, err)
    assert.Contains(t, string(output), "backends")
}
```

## 13.4 Test Coverage Requirements

| Package | Minimum Coverage |
|---------|-----------------|
| `internal/command` | 80% |
| `internal/command/builtin` | 85% |
| `internal/backend` | 80% |
| `internal/backend/local` | 75% |
| `internal/context` | 80% |
| `internal/tools` | 85% |
| `internal/plugins` | 80% |
| `internal/repos` | 80% |
| `internal/permissions` | 90% |
| `internal/config` | 90% |
| `pkg/*` | 90% |

---

# 14. Security Guidelines

## 14.1 Security Checklist

```markdown
## Code Review Security Checklist

### Input Validation
- [ ] All user input is validated before use
- [ ] File paths are sanitized and bounded
- [ ] Shell commands use parameterization, not concatenation
- [ ] URLs are validated before HTTP requests

### Shell Execution
- [ ] Commands are whitelisted, not arbitrary
- [ ] Parameters are escaped properly
- [ ] Working directory is validated
- [ ] Timeout is enforced

### File Operations
- [ ] Paths are resolved to absolute paths
- [ ] Path traversal attacks are prevented
- [ ] Write operations require confirmation
- [ ] Temp files are securely created

### Network
- [ ] TLS is enforced for all connections
- [ ] Hosts are whitelisted per permission
- [ ] Timeouts are enforced
- [ ] Response size is limited

### Secrets
- [ ] API keys are read from env vars only
- [ ] No secrets in logs or error messages
- [ ] Config files have proper permissions (0600)
- [ ] Credentials are not stored in plain text

### Permissions
- [ ] Principle of least privilege
- [ ] All operations check permissions first
- [ ] Grants are stored securely
- [ ] Trust levels are enforced
```

## 14.2 Threat Model

```markdown
## Threat Model

### Assets
1. User's source code
2. User's API keys and credentials
3. User's system (shell access)
4. User's network access

### Threat Actors
1. Malicious plugin authors
2. Compromised plugin repositories
3. Man-in-the-middle attackers
4. Local privilege escalation

### Attack Vectors

#### T1: Malicious Plugin
- **Description**: Plugin contains malicious code
- **Mitigation**: 
  - Permission system with explicit grants
  - Shell command whitelisting
  - Network host whitelisting
  - Code review for official repo

#### T2: Command Injection
- **Description**: User input used in shell command
- **Mitigation**:
  - Parameterized commands
  - Input sanitization
  - Whitelist validation

#### T3: Path Traversal
- **Description**: File access outside allowed directories
- **Mitigation**:
  - Absolute path resolution
  - Prefix checking
  - Permission grants per path

#### T4: Credential Theft
- **Description**: Plugin steals API keys
- **Mitigation**:
  - Env var permissions
  - No arbitrary env access
  - Audit logging

#### T5: Supply Chain Attack
- **Description**: Compromised dependency
- **Mitigation**:
  - Dependency pinning
  - Checksum verification
  - Regular audits
```

---

# 15. Quality Standards

## 15.1 Code Style

```yaml
# .golangci.yaml

run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gofmt
    - goimports
    - misspell
    - unconvert
    - dupl
    - goconst
    - gosec
    - prealloc

linters-settings:
  errcheck:
    check-type-assertions: true
    check-blank: true
  govet:
    check-shadowing: true
  gofmt:
    simplify: true
  dupl:
    threshold: 100
  goconst:
    min-len: 3
    min-occurrences: 3
  gosec:
    excludes:
      - G104 # Unhandled errors (handled by errcheck)

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - dupl
        - gosec
```

## 15.2 Documentation Standards

```markdown
## Documentation Requirements

### Code Documentation
- All exported functions must have a doc comment
- Doc comments start with the function name
- Complex logic must have inline comments
- Package-level doc in doc.go

### Example:
```go
// Execute runs the explain command with the given arguments.
// It reads the specified file, sends it to the LLM backend,
// and streams the explanation to the UI.
//
// Execute returns an error if the file cannot be read or
// the backend fails to respond.
func (c *ExplainCommand) Execute(
    ctx context.Context,
    args *Args,
    execCtx *ExecContext,
) (*Result, error) {
    // ...
}
```

### User Documentation
- README.md with quick start
- Command reference in docs/COMMANDS.md
- Plugin development guide in docs/PLUGINS.md
- All error messages include resolution hints
```

## 15.3 Performance Standards

```markdown
## Performance Requirements

### Startup Time
- Cold start: < 50ms
- With config load: < 100ms

### Memory Usage
- Idle: < 20MB
- During operation: < 50MB (excluding model)

### Response Time
- Command parsing: < 1ms
- Context gathering: < 100ms
- LLM excluded from response time

### Binary Size
- Uncompressed: < 15MB
- Compressed: < 5MB
```

---

# 16. Development Workflow

## 16.1 Makefile

```makefile
# Makefile

VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -s -w -X github.com/scmd/scmd/pkg/version.Version=$(VERSION)

.PHONY: all build test lint clean install release dev

all: lint test build

# Build
build:
	go build -ldflags "$(LDFLAGS)" -o bin/scmd ./cmd/scmd

build-all:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-darwin-arm64 ./cmd/scmd
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-darwin-x64 ./cmd/scmd
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-linux-x64 ./cmd/scmd
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-linux-arm64 ./cmd/scmd
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/scmd-win32-x64.exe ./cmd/scmd

# Test
test:
	go test -race -coverprofile=coverage.out ./...

test-unit:
	go test -short -race ./internal/...

test-integration:
	go test -race ./tests/integration/...

test-e2e:
	go test -race ./tests/e2e/...

coverage:
	go tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

# Development
dev:
	go run ./cmd/scmd

install: build
	cp bin/scmd /usr/local/bin/

# Release
release:
	goreleaser release --clean

release-dry:
	goreleaser release --clean --snapshot

# Clean
clean:
	rm -rf bin/ dist/ coverage.out coverage.html

# Dependencies
deps:
	go mod tidy
	go mod verify

# Generate
generate:
	go generate ./...
```

## 16.2 CI Pipeline

```yaml
# .github/workflows/ci.yml

name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Run tests
        run: make test
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: coverage.out

  build:
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Build
        run: make build-all
      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: binaries
          path: dist/

  integration:
    runs-on: ubuntu-latest
    needs: [build]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Download binaries
        uses: actions/download-artifact@v4
        with:
          name: binaries
          path: dist/
      - name: Run integration tests
        run: make test-integration
```

---

# 19. Roadmap

## Phase 1: MVP (Weeks 1-2)
- [ ] Project setup and structure
- [ ] Core CLI framework (Cobra)
- [ ] Config loading (Viper)
- [ ] **Pipe/stdin support with -p flag**
- [ ] **Mode detection (interactive vs pipe)**
- [ ] Built-in commands: /explain, /help
- [ ] Mock backend for testing
- [ ] Basic REPL

## Phase 2: Local Backend (Weeks 3-4)
- [ ] Local llama.cpp integration
- [ ] Model download with progress
- [ ] First-run experience
- [ ] Built-in commands: /review, /commit, /fix
- [ ] Streaming output (stdout + stderr separation)
- [ ] **JSON output format (-f json)**

## Phase 3: Tool Calling (Weeks 5-6)
- [ ] Tool executor framework
- [ ] Shell command execution
- [ ] File operations
- [ ] Permission system
- [ ] Tool calling engine (agentic)

## Phase 4: Plugins (Weeks 7-8)
- [ ] Plugin YAML parser
- [ ] Plugin loader
- [ ] Repository manager
- [ ] /install, /repos, /search commands
- [ ] **OneSkill as default registry**

## Phase 5: Cloud Backends (Weeks 9-10)
- [ ] Ollama backend
- [ ] Claude API backend
- [ ] OpenAI API backend
- [ ] Backend routing

## Phase 6: Distribution (Weeks 11-12)
- [ ] GoReleaser setup
- [ ] GitHub releases
- [ ] Homebrew tap
- [ ] npm package
- [ ] curl installer
- [ ] Documentation site

## Phase 7: OneSkill Integration (Weeks 13-14)
- [ ] **OneSkill Registry API integration**
- [ ] **Publisher verification flow**
- [ ] **`scmd publish` command**
- [ ] **Analytics integration**
- [ ] **Enterprise auth (OAuth/SAML)**

---

# Appendix A: Agent Task Template

```markdown
## Task: [Task Name]

### Assigned Agent: [Agent Role]

### Description
[What needs to be done]

### Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

### Dependencies
- [Dependency 1]
- [Dependency 2]

### Files to Create/Modify
- `path/to/file.go`

### Test Requirements
- Unit tests for [X]
- Integration tests for [Y]

### Security Considerations
- [Security note]

### Documentation Required
- [ ] Code comments
- [ ] README update
- [ ] Command reference update
```

---

*Document Version: 1.0.0*
*Last Updated: January 2025*
