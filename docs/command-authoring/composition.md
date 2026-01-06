# Command Composition

Command composition allows you to chain multiple commands together in powerful workflows. Build complex AI pipelines by combining simple, focused commands.

!!! tip "Why Composition?"
    - **Modularity**: Reuse existing commands as building blocks
    - **Complexity**: Solve complex problems with simple components
    - **Maintainability**: Update individual commands without changing workflows
    - **Flexibility**: Mix and match commands for different scenarios

## Composition Types

scmd supports three types of composition:

### 1. Pipeline (Sequential)

Chain commands sequentially, passing output as input to the next command.

```yaml
compose:
  pipeline:
    - command: step1
    - command: step2
    - command: step3
```

**Flow:**
```
Input → step1 → output1 → step2 → output2 → step3 → Final Output
```

### 2. Parallel (Concurrent)

Run commands concurrently and merge their results.

```yaml
compose:
  parallel:
    - command1
    - command2
    - command3
```

**Flow:**
```
Input → command1 → output1 ┐
      → command2 → output2 ├─→ Merged Output
      → command3 → output3 ┘
```

### 3. Fallback (Try Until Success)

Try commands in order until one succeeds.

```yaml
compose:
  fallback:
    - primary-command
    - backup-command
    - last-resort
```

**Flow:**
```
Input → primary ✗ → backup ✗ → last-resort ✓ → Output
```

## Pipeline Composition

### Basic Pipeline

```yaml
name: analyze-and-summarize
version: 1.0.0
description: Analyze code then summarize findings

compose:
  pipeline:
    - command: analyze-code
    - command: summarize

prompt:
  template: "Placeholder - using composition instead"
```

**Usage:**
```bash
cat main.go | scmd analyze-and-summarize
```

**What happens:**
1. `analyze-code` receives main.go content
2. Produces detailed analysis
3. `summarize` receives analysis
4. Produces concise summary

### Pipeline with Arguments

Pass custom arguments to each step:

```yaml
compose:
  pipeline:
    - command: explain
      args:
        detail: high
        format: markdown
    - command: summarize
      args:
        length: short
        style: bullet-points
```

### Pipeline with Transforms

Transform output between steps:

```yaml
compose:
  pipeline:
    - command: analyze-code
      transform: trim  # Remove whitespace
    - command: extract-issues
      transform: json.issues  # Extract JSON field
    - command: summarize
      transform: upper  # Convert to uppercase
```

**Available transforms:**
- `trim`: Remove leading/trailing whitespace
- `upper`: Convert to uppercase
- `lower`: Convert to lowercase
- `first`: Get first line
- `last`: Get last non-empty line
- `lines`: Count lines
- `json.field`: Extract JSON field

### Error Handling in Pipelines

```yaml
compose:
  pipeline:
    - command: try-fast-analysis
      on_error: continue  # Continue even if fails
    - command: deep-analysis
      on_error: stop      # Stop pipeline if fails (default)
    - command: summarize
```

**Error handling options:**
- `stop` (default): Stop pipeline on error
- `continue`: Skip failed step, continue with next
- `fallback`: Try alternative command (requires `fallback_command`)

## Parallel Composition

### Basic Parallel

```yaml
name: multi-perspective-review
version: 1.0.0
description: Review code from multiple angles

compose:
  parallel:
    - security-review
    - performance-review
    - style-review
```

**Usage:**
```bash
cat app.js | scmd multi-perspective-review
```

**Output:**
```markdown
## security-review
- Line 42: Potential SQL injection
- Line 78: Missing input validation
...

## performance-review
- Line 15: O(n²) loop can be optimized
- Line 93: Redundant database query
...

## style-review
- Inconsistent indentation
- Missing JSDoc comments
...
```

### Parallel with Dependencies

Commands run in parallel but respect dependencies:

```yaml
compose:
  parallel:
    - command1  # Runs immediately
    - command2  # Runs immediately
    - command3  # Runs immediately

  # All must complete before next step
  then:
    - combine-results
```

## Fallback Composition

### Basic Fallback

```yaml
name: smart-explain
version: 1.0.0
description: Try multiple explanation strategies

compose:
  fallback:
    - explain-with-examples    # Try first (detailed)
    - explain-simple           # Fallback (simpler)
    - explain-minimal          # Last resort (basic)
```

**Usage:**
```bash
cat complex-code.rs | scmd smart-explain
```

**What happens:**
1. Tries `explain-with-examples`
2. If it fails or times out, tries `explain-simple`
3. If that fails, tries `explain-minimal`
4. Returns first successful result

### Fallback with Different Backends

```yaml
compose:
  fallback:
    - local-explain       # Try local model first
    - cloud-explain       # Fallback to cloud if local fails
```

## Advanced Patterns

### Mixed Composition

Combine composition types:

```yaml
name: comprehensive-analysis
compose:
  pipeline:
    # Stage 1: Parallel analysis
    - parallel:
        - static-analysis
        - complexity-check
        - security-scan
      merge: true  # Merge parallel outputs

    # Stage 2: Process combined results
    - command: aggregate-findings

    # Stage 3: Try summaries until one works
    - fallback:
        - detailed-summary
        - basic-summary
```

### Conditional Steps

```yaml
compose:
  pipeline:
    - command: check-type
      output_var: file_type  # Store output in variable

    - command: process-javascript
      if: "{{.file_type}} == 'javascript'"

    - command: process-python
      if: "{{.file_type}} == 'python'"
```

!!! note "Coming Soon"
    Advanced conditional logic is planned for future releases.

### Dynamic Pipeline

Generate pipeline steps dynamically:

```yaml
compose:
  pipeline:
    # First step discovers what to do
    - command: discover-tasks
      output_format: json

    # Remaining steps created from JSON
    dynamic:
      source: "{{.tasks}}"
      template: |
        {% for task in tasks %}
        - command: {{task.command}}
          args: {{task.args}}
        {% endfor %}
```

!!! note "Experimental"
    Dynamic composition is experimental. Schema may change.

## Real-World Examples

### Example 1: Code Review Pipeline

```yaml
name: full-code-review
version: 1.0.0
description: Complete code review workflow

compose:
  pipeline:
    # Stage 1: Run linters and tests
    - command: run-lint
      on_error: continue
      transform: trim

    # Stage 2: Parallel static analysis
    - parallel:
        - security-scan
        - complexity-analysis
        - dependency-check

    # Stage 3: AI review
    - command: ai-review
      args:
        focus: critical-issues

    # Stage 4: Generate report
    - command: format-review-report
      args:
        format: markdown
        include_metrics: true

# Fallback prompt if composition not available
prompt:
  template: "This command requires composition support"
```

**Usage:**
```bash
git diff | scmd full-code-review > review.md
```

### Example 2: Documentation Pipeline

```yaml
name: auto-documentation
version: 1.0.0
description: Generate comprehensive documentation

compose:
  pipeline:
    # Extract code structure
    - command: extract-exports
      transform: json

    # Generate documentation sections in parallel
    - parallel:
        - generate-api-docs
        - generate-examples
        - generate-usage-guide

    # Combine and format
    - command: merge-docs
      args:
        format: markdown
        toc: true

    # Add diagrams
    - command: add-diagrams
      args:
        types: [sequence, class, flow]

# Output to file with post-hook
hooks:
  post:
    - shell: cat $OUTPUT > README.md
```

### Example 3: Multi-Stage Analysis

```yaml
name: deep-code-analysis
version: 1.0.0
description: Progressive code analysis with fallbacks

compose:
  pipeline:
    # Quick scan first
    - command: quick-scan
      transform: trim

    # Decide if deep analysis needed
    - command: triage-issues
      output_var: severity

    # Deep analysis with fallbacks
    - fallback:
        # Try comprehensive analysis
        - command: deep-analysis-with-tools
          timeout: 60

        # Fallback to standard analysis
        - command: standard-analysis
          timeout: 30

        # Last resort: basic analysis
        - command: basic-analysis
          timeout: 10

    # Generate recommendations
    - command: generate-recommendations
      args:
        severity: "{{.severity}}"
```

### Example 4: Multi-Language Project

```yaml
name: polyglot-review
version: 1.0.0
description: Review multi-language projects

compose:
  parallel:
    # Language-specific reviews
    - command: review-javascript
      filter: "*.js"
    - command: review-python
      filter: "*.py"
    - command: review-go
      filter: "*.go"
    - command: review-rust
      filter: "*.rs"

  # Then combine results
  then:
    pipeline:
      - command: merge-reviews
      - command: prioritize-issues
      - command: generate-unified-report
```

## Composition with Hooks

Combine composition with hooks for powerful workflows:

```yaml
name: tested-refactor
version: 1.0.0
description: Refactor code with automated testing

hooks:
  pre:
    # Backup original
    - shell: cp {{.file}} {{.file}}.backup
    # Run tests (baseline)
    - shell: npm test > /tmp/tests-before.txt

compose:
  pipeline:
    # Analyze code
    - command: analyze-for-refactor
      transform: json

    # Generate refactoring plan
    - command: plan-refactor
      args:
        safe_mode: true

    # Apply refactoring
    - command: apply-refactor

hooks:
  post:
    # Run tests again
    - shell: npm test > /tmp/tests-after.txt
    # Compare results
    - shell: diff /tmp/tests-before.txt /tmp/tests-after.txt
    # If tests fail, restore backup
    - shell: |
        if [ $? -ne 0 ]; then
          echo "Tests failed, restoring backup"
          mv {{.file}}.backup {{.file}}
        else
          rm {{.file}}.backup
        fi
```

## Composition with Tool Calling

Tool calling works within composed commands:

```yaml
name: autonomous-project-setup
compose:
  pipeline:
    # Analyze project requirements (uses tools to read files)
    - command: analyze-requirements

    # Generate project structure (uses tools to write files)
    - command: generate-structure

    # Install dependencies (uses shell tool)
    - command: install-deps

    # Initialize git (uses shell tool)
    - command: init-git

    # Generate docs (uses write_file tool)
    - command: generate-docs
```

Each command in the pipeline can use tools independently.

## Performance Considerations

### Pipeline Performance

Pipelines run sequentially:

```
Time = time(cmd1) + time(cmd2) + time(cmd3) + ...
```

**Optimization:**
- Use parallel composition where possible
- Minimize number of steps
- Use transforms instead of full commands when possible

### Parallel Performance

Parallel commands run concurrently:

```
Time ≈ max(time(cmd1), time(cmd2), time(cmd3))
```

**Optimization:**
- Balance command complexity
- Be aware of resource limits (CPU, memory, API rate limits)
- Don't overuse (3-5 parallel commands is usually optimal)

### Fallback Performance

Best case: first command succeeds
Worst case: all commands fail

```
Time = time(cmd1) + time(cmd2) + ... + time(cmdN)
```

**Optimization:**
- Order commands by success probability (most likely first)
- Set appropriate timeouts
- Use cheaper models for fallbacks

## Best Practices

### 1. Keep Commands Focused

Each command should do one thing well:

**Good:**
```yaml
compose:
  pipeline:
    - command: extract-functions  # Single responsibility
    - command: analyze-complexity # Single responsibility
    - command: suggest-improvements # Single responsibility
```

**Bad:**
```yaml
compose:
  pipeline:
    - command: do-everything  # Too broad
```

### 2. Handle Errors Gracefully

```yaml
compose:
  pipeline:
    - command: optional-lint
      on_error: continue  # Non-critical

    - command: critical-security-scan
      on_error: stop      # Critical

    - command: generate-report
```

### 3. Use Meaningful Transforms

```yaml
compose:
  pipeline:
    - command: analyze
      transform: json.critical_issues  # Clear intent

    - command: format
      transform: trim  # Remove clutter
```

### 4. Document Complex Compositions

```yaml
# Multi-stage analysis pipeline
# 1. Quick scan for obvious issues
# 2. Parallel deep analysis (security, performance, style)
# 3. Aggregate findings
# 4. Generate prioritized report
compose:
  pipeline:
    - command: quick-scan
    - parallel:
        - security-analysis
        - performance-analysis
        - style-analysis
    - command: aggregate
    - command: prioritize
```

### 5. Test Compositions

Test individual commands first:

```bash
# Test each command
cat test.js | scmd analyze-code
cat test.js | scmd summarize

# Then test composition
cat test.js | scmd analyze-and-summarize
```

## Troubleshooting

### "Command not found" in Pipeline

Command in pipeline doesn't exist:

```bash
# List available commands
scmd slash list

# Install missing command
scmd repo install official/missing-command
```

### Parallel Commands Timing Out

Too many parallel commands or commands too slow:

1. Reduce parallel count
2. Increase timeouts
3. Use faster models
4. Simplify commands

### Pipeline Stops Unexpectedly

A command failed with `on_error: stop` (default):

```bash
# Run with debug to see which command failed
SCMD_DEBUG=1 scmd your-pipeline
```

### Transforms Not Working

Transform name is incorrect or not supported:

**Supported transforms:**
- `trim`, `upper`, `lower`, `first`, `last`, `lines`
- `json.fieldname`

**Custom transforms:** Not yet supported, coming in future release.

## Limitations

### Current Limitations

1. **Maximum 20 steps** per pipeline
2. **No nested composition** (compose within compose)
3. **Limited transforms** (fixed set, no custom)
4. **No conditional logic** (planned for future)
5. **No loops** (planned for future)
6. **Sequential only within parallel** (can't have pipeline within parallel)

### Future Enhancements

- [ ] Nested composition support
- [ ] Custom transforms
- [ ] Conditional steps (`if/else`)
- [ ] Loops (`for each file`)
- [ ] Variables and state passing
- [ ] Composition templates
- [ ] Visual pipeline editor

## Next Steps

- [Tool Calling Guide](tool-calling.md) - Use tools in composed commands
- [Hooks Guide](hooks.md) - Add automation to compositions
- [Advanced Examples](../examples/advanced-composition.md) - Complex real-world pipelines
- [Best Practices](best-practices.md) - Command design tips
