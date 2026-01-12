# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.1] - 2026-01-12

### Fixed
- **Markdown Rendering**: Fixed streaming AI responses to use Glamour markdown renderer
  - Streaming responses (TTY mode) now buffer content and apply markdown formatting
  - Code blocks now display with syntax highlighting
  - Markdown headers, lists, and formatting properly rendered
  - Non-streaming fallback also uses markdown renderer
  - Resolves issue where AI-generated markdown was displayed as plain text

### Changed
- Streaming response path now buffers chunks and applies post-stream markdown rendering
- Terminal cursor manipulation added to replace plain streamed text with formatted output
- Non-streaming paths consistently use `WriteMarkdown()` method

### Technical
- Updated `internal/cli/root.go::runPrompt()` to buffer streaming chunks
- Added ANSI escape sequences for cursor positioning and screen clearing
- Streaming output now checked with `looksLikeMarkdown()` before rendering
- All response paths (streaming, non-streaming, fallback) now support markdown formatting

## [0.5.0] - 2026-01-12

### Fixed
- **CRITICAL**: Fixed repository manifest parsing bug preventing command discovery
  - Added support for legacy manifest format with `path` field instead of `name`/`description`
  - Implemented automatic metadata fetching from command files when manifest lacks details
  - Added parallel fetching (10 concurrent requests) for efficient manifest normalization
  - Commands from official repository (100+ commands) now properly load with names and descriptions
  - Fixed `scmd repo search` showing blank command names and descriptions
  - Fixed `scmd repo show` and `scmd repo install` command lookup failures

### Added
- **Enhanced Command Discovery UI**: Major improvements to help users find the 100+ available commands
  - Added prominent "💡 Discover 100+ Commands" section to root `scmd --help` output
  - Enhanced `scmd registry` help text with category overview and examples
  - Updated `scmd repo` help text with quick start guide
  - Improved `/help` command output to highlight command discovery features
  - Added registry discovery tips to interactive REPL welcome message
- **Legacy Manifest Support**: Backward compatibility for different repository manifest schemas
  - Automatically detects and handles `path` field as alternative to `file`
  - Fetches missing metadata (name, description, category) from individual command files
  - Graceful degradation: skips commands that fail to fetch without breaking entire manifest
- **Test Coverage**: Added comprehensive tests for manifest schema validation
  - New test: `TestManager_FetchManifestLegacyFormat` validates legacy format handling
  - Verifies parallel fetching and metadata normalization
  - Ensures backward compatibility with existing manifests

### Changed
- Repository manager now normalizes manifests after fetching to populate missing metadata
- Help text across all commands now emphasizes the 100+ command registry
- Discovery features promoted to top-level help output for better visibility

### Technical
- Updated `internal/repos/manager.go`:
  - Added `Path` field to `Command` struct for legacy format support
  - Implemented `normalizeManifest()` method with parallel command file fetching
  - Added semaphore-based concurrency control (10 concurrent requests)
- Updated `internal/cli/root.go`: Enhanced root command and REPL help text
- Updated `internal/cli/repo.go`: Added quick start guide to repo command help
- Updated `internal/cli/registry.go`: Enhanced registry command help with category showcase
- Updated `internal/command/builtin/help.go`: Promoted discovery section in help output
- Added `internal/repos/manager_test.go::TestManager_FetchManifestLegacyFormat`

### Impact
- Users can now discover and install all 100+ official commands
- Repository search functionality fully operational
- Improved first-time user experience with better command discoverability
- Official repository at `github.com/sunboylabs/commands` now works seamlessly

## [0.4.3] - 2026-01-12

### Added
- **Beautiful Markdown Output**: Lightweight, performance-focused formatting with Glamour
  - `--format` flag with options: `auto` (default), `markdown`, `plain`
  - Automatic TTY detection - formatted output in terminal, plain text when piped
  - Syntax highlighting for 40+ programming languages via Chroma
  - Theme detection (dark/light/auto) with SCMD_THEME environment variable
  - NO_COLOR environment variable support for accessibility
  - Configuration file support: `ui.format`, `ui.theme`, `ui.word_wrap`
- **Lazy Rendering**: Zero overhead when formatting is disabled
  - Glamour renderer initialized only on first use (< 1ns overhead)
  - Formatter creation: ~1.9µs
  - Rendering: < 10ms for typical responses
- **Terminal Detection**: Smart environment detection
  - TTY capability detection
  - Color support verification
  - Terminal dimensions and word wrap
  - Image support detection

### Changed
- Updated output system with three-layer architecture (Detector → Renderer → Formatter)
- Enhanced all command outputs to support markdown formatting
- Commands automatically use formatted output in terminals
- Plain text preserved for piping and redirection

### Technical
- Added `internal/output/detector.go` - Terminal environment detection
- Added `internal/output/renderer.go` - Lazy-loaded Glamour wrapper
- Updated `internal/output/formatter.go` - Markdown/plain text formatting
- Extended `internal/config/config.go` - UI configuration (format, theme, word_wrap)
- Updated `internal/cli/root.go` - Added --format flag
- Updated `internal/cli/output.go` - OutputWriter with formatter integration
- Comprehensive test coverage (1503 lines across 5 test files)
- All performance benchmarks exceed targets

### Dependencies
- github.com/charmbracelet/glamour v0.8.0 (markdown rendering)

## [0.4.2] - 2026-01-12

### Added
- **Template-Command Integration**: Commands can now reference templates for flexible, reusable prompt patterns
  - Commands support three modes: direct prompts, template references, or inline templates
  - Automatic variable mapping from command args/flags to template variables
  - Template validation at command installation time
- **Unified Command Specification**: Extended CommandSpec with optional Template field
  - `TemplateRef` struct supports named templates and inline definitions
  - Variable mapping with Go template syntax (`{{.variable}}`)
  - Automatic context population (file extension, language detection, stdin)
- **Official Commands Repository**: Created github.com/scmd/commands with 100+ categorized commands
  - 8 categories: git (15), code (15), devops (15), docs (10), debug (10), data (10), security (15), shell (10)
  - 10 reusable templates for common workflows
  - Comprehensive documentation (README, CONTRIBUTING, REPO_STRUCTURE)
- **Template Executor**: New `TemplateExecutor` for handling template-based command execution
  - Language detection for 20+ programming languages
  - Context building with automatic file content/extension mapping
  - Support for both named and inline templates

### Changed
- Updated repository manager to handle both commands and templates in manifests
- Enhanced command executor with template execution path
- Improved repository manifest format to include templates list

### Documentation
- **docs/template-command-integration.md**: Complete integration guide with examples
- **examples/commands/**: Example command specs using templates
- **examples/templates/**: Example template definitions

### Technical
- Added `internal/repos/template_executor.go` for template-based execution
- Added `internal/repos/template_integration_test.go` with comprehensive test coverage
- Updated `internal/repos/manager.go` with TemplateRef and InlineTemplate structs
- Updated `internal/templates/manager.go` with NewManagerWithDir for testing
- All existing tests pass, new tests cover template integration

## [0.4.1] - 2026-01-11

### Fixed
- **CRITICAL**: Replaced CGO-dependent SQLite (github.com/mattn/go-sqlite3) with pure Go implementation (modernc.org/sqlite)
  - Fixes "Binary was compiled with 'CGO_ENABLED=0'" error
  - Enables chat and history features to work in all environments
  - Removes external C library dependencies
  - Improves cross-platform compatibility
- Fixed SQLite driver name from "sqlite3" to "sqlite" for modernc.org/sqlite compatibility

### Changed
- Binary size optimized (25.6 MB vs 43.8 MB in initial v0.4.0)
- Improved error messages for chat/history commands
- Enhanced database connection reliability

### Verified
- Chat feature: Multi-turn conversations with context retention ✓
- History management: list, show, search, continue commands ✓
- Template system: All 6 templates working correctly ✓
- Shareable repos: Slash commands and repository system intact ✓
- Markdown output: Glamour-based rendering working as designed ✓

## [0.4.0] - 2026-01-10

### Added
- **Interactive Conversation Mode**: Multi-turn AI conversations with SQLite persistence
  - New `scmd chat` command for interactive REPL sessions
  - Conversation history management with `scmd history` commands (list, show, search, delete, clear)
  - Context retention across sessions with automatic saving
  - In-session commands: /help, /clear, /info, /save, /export, /model, /exit
- **Beautiful Markdown Output**: Syntax-highlighted code and formatted text
  - Chroma v2 integration for syntax highlighting (40+ languages)
  - Glamour integration for markdown rendering
  - Automatic theme detection (dark/light/auto)
  - Customizable color schemes
- **Template/Pattern System**: Reusable, customizable prompt templates
  - 6 built-in professional templates (security-review, performance, api-design, testing, documentation, beginner-explain)
  - YAML-based template format with variable substitution
  - Template management commands: list, show, create, delete, search, import, export
  - Template integration with existing commands via --template flag
- Conversation export to Markdown
- Token usage tracking
- Model switching within chat sessions

### Changed
- Updated README with comprehensive documentation for new features
- Enhanced `/explain` and `/review` commands to support templates
- Improved output formatting with color and styling
- Added 7 new dependencies for UI/chat features

### Dependencies
- github.com/charmbracelet/glamour v0.6.0 (markdown rendering)
- github.com/charmbracelet/lipgloss v0.9.1 (terminal styling)
- github.com/alecthomas/chroma/v2 v2.12.0 (syntax highlighting)
- github.com/muesli/termenv v0.15.2 (color support)
- github.com/briandowns/spinner v1.23.0 (progress indicators)
- github.com/mattn/go-sqlite3 v1.14.18 (conversation persistence)
- github.com/google/uuid v1.5.0 (conversation IDs)

## [0.3.1] - 2026-01-09

### Fixed
- Critical bug fixes based on dogfooding feedback

## [0.3.0] - 2026-01-09

### Added
- GoReleaser configuration for automated multi-platform releases
- Homebrew tap support for easy macOS/Linux installation
- npm package wrapper for cross-platform distribution
- Shell install script for wget/curl installation
- Native Linux packages (deb, rpm, apk) via nfpm
- GitHub Actions workflow for automated releases
- Shell completion support (bash, zsh, fish)
- Comprehensive installation documentation (INSTALL.md)
- Makefile targets for release management
- Docker image support (multi-arch)
- Checksum verification for downloads
- Post-install and pre-remove scripts for package managers

### Changed
- Updated Makefile with release and distribution targets
- Enhanced documentation with multiple installation methods

### Fixed
- CLI flag parsing and llama-server bundling
- llama-server download script to use latest release
- llama-server bundling for goreleaser
- Use bash instead of sh for download script in goreleaser

## [0.1.0] - 2025-01-06

### Added
- Initial release
- Offline-first AI-powered slash commands
- llama.cpp integration with Qwen models
- Auto-download of models on first use
- Built-in `/explain` command
- Repository system for command distribution
- Shell integration for bash, zsh, and fish
- Multi-backend support (llama.cpp, Ollama, OpenAI, Together.ai, Groq)
- Command composition and chaining
- Configuration management
- Model management (list, pull, remove, info)
- Slash command system
- Command lockfiles for reproducibility
- Context gathering and caching

### Security
- Input validation and sanitization
- Path traversal prevention
- Resource limits for LLM inference
- Secure model downloads with checksum verification
- Comprehensive security documentation

### Documentation
- README with quick start guide
- Architecture documentation
- Security documentation
- Troubleshooting guide
- API documentation
- Contributing guidelines

## Release Process

To create a new release:

1. Update version in relevant files
2. Update this CHANGELOG with release notes
3. Create and push a git tag:
   ```bash
   make tag VERSION=v1.0.0
   git push origin v1.0.0
   ```
4. GitHub Actions will automatically:
   - Build binaries for all platforms
   - Create GitHub release with notes
   - Publish to Homebrew tap
   - Publish to npm registry
   - Build and push Docker images

## Version History

[Unreleased]: https://github.com/sunboylabs/scmd/compare/v0.5.1...HEAD
[0.5.1]: https://github.com/sunboylabs/scmd/releases/tag/v0.5.1
[0.5.0]: https://github.com/sunboylabs/scmd/releases/tag/v0.5.0
[0.4.3]: https://github.com/sunboylabs/scmd/releases/tag/v0.4.3
[0.4.2]: https://github.com/sunboylabs/scmd/releases/tag/v0.4.2
[0.4.1]: https://github.com/sunboylabs/scmd/releases/tag/v0.4.1
[0.4.0]: https://github.com/sunboylabs/scmd/releases/tag/v0.4.0
[0.3.1]: https://github.com/sunboylabs/scmd/releases/tag/v0.3.1
[0.3.0]: https://github.com/sunboylabs/scmd/releases/tag/v0.3.0
[0.2.1]: https://github.com/sunboylabs/scmd/releases/tag/v0.2.1
[0.1.0]: https://github.com/sunboylabs/scmd/releases/tag/v0.1.0
