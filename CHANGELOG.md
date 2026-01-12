# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/sunboylabs/scmd/compare/v0.4.2...HEAD
[0.4.2]: https://github.com/sunboylabs/scmd/releases/tag/v0.4.2
[0.4.1]: https://github.com/sunboylabs/scmd/releases/tag/v0.4.1
[0.4.0]: https://github.com/sunboylabs/scmd/releases/tag/v0.4.0
[0.3.1]: https://github.com/sunboylabs/scmd/releases/tag/v0.3.1
[0.3.0]: https://github.com/sunboylabs/scmd/releases/tag/v0.3.0
[0.2.1]: https://github.com/sunboylabs/scmd/releases/tag/v0.2.1
[0.1.0]: https://github.com/sunboylabs/scmd/releases/tag/v0.1.0
