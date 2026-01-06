# Documentation Guide for scmd

This document explains the documentation setup and provides guidance for completing and maintaining the docs.

## What's Been Created

### 1. Documentation Infrastructure

✅ **MkDocs Material Setup**
- `mkdocs.yml` - Complete configuration with navigation, theme, plugins
- `requirements.txt` - Python dependencies
- `.github/workflows/docs.yml` - Auto-deployment to GitHub Pages
- Comprehensive directory structure in `docs/`

✅ **Core Documentation Pages**
- `docs/index.md` - Professional landing page with feature showcase
- `docs/getting-started/installation.md` - Multi-platform installation guide
- `docs/getting-started/quick-start.md` - 5-minute tutorial with examples

✅ **New Feature Documentation** (The Key Value-Add)
- `docs/command-authoring/tool-calling.md` - **Comprehensive tool calling guide** (15+ pages)
  - What tool calling is and why it matters
  - All 4 built-in tools documented
  - Agent loop explanation with diagrams
  - Security model
  - Real-world examples
  - Debugging tips
  - Best practices

- `docs/command-authoring/hooks.md` - **Complete hooks guide** (14+ pages)
  - Pre/post execution hooks
  - Shell vs command hooks
  - Conditional hooks
  - 5 common patterns
  - 3 real-world examples
  - Security considerations
  - Troubleshooting

- `docs/command-authoring/composition.md` - **Full composition guide** (16+ pages)
  - All 3 composition types (pipeline, parallel, fallback)
  - Advanced patterns
  - 4 real-world examples
  - Performance considerations
  - Best practices
  - Combining with hooks and tools

## Documentation Structure

```
docs/
├── index.md                          ✅ Done - Professional landing page
├── getting-started/
│   ├── installation.md              ✅ Done - Multi-platform install guide
│   ├── quick-start.md               ✅ Done - 5-minute tutorial
│   ├── first-command.md             📝 TODO - Creating first custom command
│   └── shell-integration.md         📝 TODO - Bash/Zsh/Fish setup
├── user-guide/
│   ├── slash-commands.md            📝 TODO - Using slash commands
│   ├── backends.md                  📝 TODO - Backend configuration
│   ├── models.md                    📝 TODO - Model management
│   ├── repositories.md              📝 TODO - Repository system
│   ├── configuration.md             📝 TODO - Config file reference
│   ├── cli-reference.md             📝 TODO - Complete CLI docs
│   └── troubleshooting.md           📝 TODO - Common issues
├── command-authoring/
│   ├── overview.md                  📝 TODO - Command creation intro
│   ├── yaml-specification.md        📝 TODO - Complete YAML spec
│   ├── prompts-and-templates.md     📝 TODO - Effective prompts
│   ├── tool-calling.md              ✅ Done - Agentic behavior
│   ├── hooks.md                     ✅ Done - Pre/post execution
│   ├── composition.md               ✅ Done - Command chaining
│   ├── dependencies.md              📝 TODO - Dependency management
│   ├── testing-commands.md          📝 TODO - Testing & debugging
│   └── best-practices.md            📝 TODO - Design patterns
├── repository-guide/
│   ├── creating-repositories.md     📝 TODO - Repo creation
│   ├── manifest-format.md           📝 TODO - Manifest reference
│   ├── versioning.md                📝 TODO - Semantic versioning
│   ├── publishing.md                📝 TODO - GitHub/GitLab hosting
│   └── registry-submission.md       📝 TODO - Central registry
├── contributing/
│   ├── development-setup.md         📝 TODO - Dev environment
│   ├── architecture.md              📝 Can leverage scmd-architecture.md
│   ├── codebase-guide.md            📝 TODO - Code organization
│   ├── internal-packages.md         📝 TODO - Package docs
│   ├── adding-backends.md           📝 TODO - New LLM backends
│   ├── adding-tools.md              📝 TODO - Custom tools
│   ├── testing.md                   📝 TODO - Test suite
│   └── pull-requests.md             📝 TODO - PR guidelines
├── reference/
│   ├── command-spec.md              📝 TODO - YAML reference
│   ├── tool-api.md                  📝 TODO - Tool interface
│   ├── backend-api.md               📝 TODO - Backend interface
│   ├── environment-variables.md     📝 TODO - All env vars
│   └── error-codes.md               📝 TODO - Error reference
├── examples/
│   ├── basic-commands.md            📝 TODO - Simple examples
│   ├── advanced-composition.md      📝 TODO - Complex pipelines
│   ├── tool-calling-examples.md     📝 TODO - Agentic examples
│   ├── git-workflow.md              📝 TODO - Git integration
│   └── use-cases.md                 📝 TODO - Real-world scenarios
└── about/
    ├── faq.md                       📝 TODO - FAQs
    ├── changelog.md                 📝 TODO - Version history
    ├── roadmap.md                   📝 TODO - Future plans
    └── license.md                   📝 TODO - MIT license
```

## Quick Start for Documentation Development

### 1. Install MkDocs

```bash
# Install Python dependencies
pip install -r requirements.txt
```

### 2. Preview Documentation Locally

```bash
# Start live-reload server
mkdocs serve

# Open http://127.0.0.1:8000 in browser
# Edit docs/*.md files and see changes instantly
```

### 3. Build Static Site

```bash
# Build to site/ directory
mkdocs build

# Check output
open site/index.html
```

### 4. Deploy to GitHub Pages

```bash
# Manual deployment
mkdocs gh-deploy

# Or push to main branch - GitHub Actions will auto-deploy
git add docs/ mkdocs.yml
git commit -m "docs: update documentation"
git push origin main
```

## How to Complete Remaining Documentation

### Priority 1: User-Facing Documentation (Week 1)

These are the most important for users:

1. **Complete Getting Started**
   ```bash
   # Create these files:
   docs/getting-started/first-command.md
   docs/getting-started/shell-integration.md
   ```
   - Copy patterns from quick-start.md
   - Include code examples and screenshots
   - Test all commands before documenting

2. **Create User Guide**
   ```bash
   docs/user-guide/slash-commands.md     # How to use /commands
   docs/user-guide/backends.md           # Backend setup
   docs/user-guide/models.md             # Model management
   docs/user-guide/configuration.md      # Config file
   ```
   - Reference existing README.md content
   - Add troubleshooting sections
   - Include configuration examples

3. **Create Examples**
   ```bash
   docs/examples/basic-commands.md        # Simple use cases
   docs/examples/tool-calling-examples.md # Showcase new features
   docs/examples/git-workflow.md          # Git integration
   ```
   - Use examples from `examples/` directory
   - Include full YAML specs
   - Show expected output

### Priority 2: Command Author Documentation (Week 2)

1. **YAML Specification Reference**
   ```bash
   docs/command-authoring/yaml-specification.md
   ```
   - Document every field with type, default, examples
   - Reference repos/manager.go for CommandSpec structure
   - Include annotated examples

2. **Complete Command Authoring**
   ```bash
   docs/command-authoring/overview.md
   docs/command-authoring/prompts-and-templates.md
   docs/command-authoring/dependencies.md
   docs/command-authoring/testing-commands.md
   docs/command-authoring/best-practices.md
   ```

3. **Repository Guide**
   ```bash
   docs/repository-guide/*.md
   ```
   - How to create and publish command repos
   - Manifest format reference
   - Versioning strategies

### Priority 3: Contributor Documentation (Week 3)

1. **Architecture**
   ```bash
   docs/contributing/architecture.md
   ```
   - Leverage existing scmd-architecture.md
   - Add diagrams using Mermaid
   - Explain package responsibilities

2. **Development Guides**
   ```bash
   docs/contributing/development-setup.md
   docs/contributing/codebase-guide.md
   docs/contributing/adding-backends.md
   docs/contributing/adding-tools.md
   ```
   - Step-by-step setup
   - Code walkthrough
   - Extension points

3. **Reference Documentation**
   ```bash
   docs/reference/*.md
   ```
   - API specifications
   - Environment variables
   - Error codes

### Priority 4: Polish (Week 4)

1. **About Section**
   ```bash
   docs/about/faq.md
   docs/about/changelog.md
   docs/about/roadmap.md
   docs/about/license.md
   ```

2. **Visual Enhancements**
   - Add Mermaid diagrams to architecture docs
   - Include screenshots in getting started
   - Create flow diagrams for tool calling/composition

3. **Search Optimization**
   - Add meta descriptions to all pages
   - Ensure proper heading hierarchy
   - Add tags and keywords

## Documentation Best Practices

### Writing Style

✅ **Do:**
- Use active voice: "Run the command" not "The command is run"
- Include code examples for every concept
- Add `!!! tip` and `!!! warning` callouts
- Use tables for comparisons
- Include "Next Steps" sections

❌ **Don't:**
- Assume knowledge - explain acronyms
- Write long paragraphs - use lists and headings
- Forget code syntax highlighting
- Leave TODOs in published docs

### Code Examples

```markdown
### Good Example

Show command, expected output, and explanation:

\`\`\`bash
scmd /explain "what is a goroutine?"
\`\`\`

Output:
\`\`\`
A goroutine is a lightweight thread...
\`\`\`

This command explains Go concepts using the default model.
```

### Diagrams

Use Mermaid for diagrams:

```markdown
\`\`\`mermaid
sequenceDiagram
    User->>scmd: /command
    scmd->>LLM: prompt
    LLM->>Tool: execute
    Tool->>LLM: result
    LLM->>scmd: response
\`\`\`
```

### Callouts

```markdown
!!! tip "Performance Tip"
    Use smaller models for faster responses

!!! warning "Security Warning"
    Never commit API keys to git

!!! note "Coming Soon"
    This feature is planned for v2.0
```

## Leveraging Existing Content

### From README.md

The current README.md is excellent and can be repurposed:

- CLI Reference section → `docs/user-guide/cli-reference.md`
- Backend table → `docs/user-guide/backends.md`
- Model table → `docs/user-guide/models.md`
- Configuration example → `docs/user-guide/configuration.md`

### From scmd-architecture.md

Use for:
- `docs/contributing/architecture.md`
- `docs/contributing/internal-packages.md`

### From examples/

The example commands provide great content for:
- `docs/examples/tool-calling-examples.md`
- `docs/examples/advanced-composition.md`
- `docs/command-authoring/best-practices.md`

## Auto-Generating Documentation

### CLI Reference

Generate from code:

```go
// Add to cmd/scmd/main.go
if *genDocs {
    docs.GenerateCLIReference("docs/user-guide/cli-reference.md")
}
```

### Environment Variables

Extract from config package:

```bash
grep -r "os.Getenv" internal/ | # Find all env vars
  sed 's/.*Getenv("\(.*\)").*/\1/' | # Extract names
  sort -u > docs/reference/environment-variables.md
```

### Error Codes

Extract from code:

```bash
grep -r "fmt.Errorf\|errors.New" internal/ | # Find errors
  extract_and_document
```

## Deployment

### Automatic Deployment

Documentation auto-deploys on push to `main`:

```bash
git add docs/ mkdocs.yml
git commit -m "docs: add user guide"
git push origin main

# GitHub Actions deploys to:
# https://scmd.github.io/scmd/
```

### Manual Deployment

```bash
mkdocs gh-deploy --force
```

### Custom Domain (Optional)

1. Add `CNAME` file:
   ```bash
   echo "docs.scmd.dev" > docs/CNAME
   ```

2. Configure DNS:
   ```
   CNAME docs.scmd.dev -> scmd.github.io
   ```

3. Enable HTTPS in GitHub Pages settings

## Maintenance

### Updating Documentation

When code changes:

1. **Update relevant docs** in same PR
2. **Run mkdocs build** to verify
3. **Test code examples** to ensure accuracy
4. **Update version** in about/changelog.md

### Documentation Review

- Review docs monthly for accuracy
- Check for broken links
- Update examples with new features
- Respond to documentation issues

### Community Contributions

Encourage docs contributions:

```markdown
<!-- At bottom of each page -->
**Found an issue?** [Edit this page on GitHub](...)
```

## Publishing Checklist

Before announcing documentation:

- [ ] All Priority 1 pages complete
- [ ] All code examples tested
- [ ] No broken internal links
- [ ] Search works properly
- [ ] Mobile-friendly checked
- [ ] GitHub Pages deployed
- [ ] README updated with docs link
- [ ] Announcement ready

## Getting Help

- **MkDocs Material Docs**: https://squidfunk.github.io/mkdocs-material/
- **Markdown Guide**: https://www.markdownguide.org/
- **Mermaid Diagrams**: https://mermaid.js.org/
- **Example Sites**:
  - FastAPI: https://fastapi.tiangolo.com/
  - Pydantic: https://docs.pydantic.dev/

## Summary

### ✅ What's Complete

1. **Documentation Infrastructure** - MkDocs, CI/CD, directory structure
2. **Critical New Features** - Tool calling, hooks, composition (45+ pages)
3. **Getting Started Foundation** - Installation, quick start (20+ pages)
4. **Professional Landing Page** - Feature showcase, navigation

### 📝 What's Next

1. Complete remaining Getting Started pages (2 pages)
2. Create User Guide (7 pages)
3. Create Command Authoring basics (6 pages)
4. Create examples (5 pages)
5. Polish and deploy

### 🎯 Estimated Completion

- **Priority 1** (User docs): 2-3 days
- **Priority 2** (Author docs): 2-3 days
- **Priority 3** (Contributor docs): 3-4 days
- **Priority 4** (Polish): 1-2 days

**Total**: 8-12 days for comprehensive documentation

### 🚀 Immediate Next Steps

1. Run `mkdocs serve` to preview current docs
2. Complete Priority 1 pages (user-facing)
3. Test all code examples
4. Deploy to GitHub Pages
5. Add docs badge to README

The foundation is solid. The most challenging and valuable documentation (tool calling, hooks, composition) is complete. The remaining work is more straightforward content creation following the established patterns.
