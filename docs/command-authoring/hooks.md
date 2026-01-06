# Hooks (Pre/Post Execution)

Hooks let you run shell commands or other scmd commands before and after the main LLM execution. Perfect for automation, validation, and integration with existing workflows.

!!! tip "When to Use Hooks"
    - **Pre-hooks**: Gather context, validate inputs, check prerequisites
    - **Post-hooks**: Process results, trigger actions, cleanup resources

## Basic Example

```yaml
name: git-commit-with-checks
version: 1.0.0
description: Generate commit message with pre/post validation

hooks:
  pre:
    - shell: git status --short
    - shell: git diff --check  # Check for whitespace errors
  post:
    - shell: echo "Commit message generated!"
    - shell: git diff --stat

prompt:
  system: |
    Generate a clear, conventional commit message.
  template: |
    Based on the staged changes, create a commit message.

model:
  temperature: 0.3
```

**Execution flow:**
```
1. Run pre-hooks
   ↓
2. Execute LLM completion
   ↓
3. Run post-hooks
   ↓
4. Return result
```

## Hook Types

### Shell Hooks

Execute shell commands directly:

```yaml
hooks:
  pre:
    - shell: echo "Starting analysis..."
    - shell: mkdir -p ./output
    - shell: date +%Y-%m-%d > ./output/timestamp.txt
```

**Features:**
- Full shell syntax support
- Output is captured but not shown to user
- Errors stop execution (can be configured)
- 30-second timeout per command

### Command Hooks

Execute other scmd commands:

```yaml
hooks:
  pre:
    - command: validate-code
    - command: check-style
  post:
    - command: update-docs
```

**Use cases:**
- Chain multiple AI commands
- Reuse existing command logic
- Build complex workflows

## Hook Specifications

### Full Hook Schema

```yaml
hooks:
  pre:
    - shell: "command"           # Shell command
      if: "condition"            # Optional condition
    - command: "scmd-command"    # scmd command
      if: "condition"
  post:
    - shell: "command"
    - command: "scmd-command"
```

### Conditional Hooks

Run hooks only when conditions are met:

```yaml
hooks:
  pre:
    - shell: git status
      if: "{{.git}}"  # Only if git context enabled
    - shell: npm test
      if: "{{.run_tests}}"  # Only if run_tests flag set
```

**Note:** Conditional hook evaluation is currently basic. Complex conditions coming in future releases.

## Common Patterns

### 1. Git Integration

```yaml
name: smart-commit
hooks:
  pre:
    # Check for unstaged changes
    - shell: git diff --quiet || echo "Warning: Unstaged changes"
    # Check current branch
    - shell: git branch --show-current
    # Run linter
    - shell: npm run lint 2>&1
  post:
    # Show what will be committed
    - shell: git diff --staged --stat
    # Copy commit message to clipboard (macOS)
    - shell: echo "$COMMIT_MSG" | pbcopy

prompt:
  template: |
    Generate a commit message for these changes:
    {{.stdin}}
```

### 2. Test Execution

```yaml
name: test-and-explain
hooks:
  pre:
    # Run tests and capture output
    - shell: go test ./... > /tmp/test-output.txt 2>&1 || true
  post:
    # If tests failed, open results
    - shell: |
        if [ $? -ne 0 ]; then
          cat /tmp/test-output.txt
        fi

prompt:
  template: |
    Analyze the test results and explain failures:
    $(cat /tmp/test-output.txt)
```

### 3. Environment Validation

```yaml
name: deploy-check
hooks:
  pre:
    # Check if kubectl is installed
    - shell: which kubectl || echo "kubectl not found"
    # Verify cluster connection
    - shell: kubectl cluster-info
    # Check current context
    - shell: kubectl config current-context
  post:
    # Log deployment
    - shell: echo "Deployed at $(date)" >> deploy.log

prompt:
  template: |
    Generate a deployment plan for:
    {{.service}}
```

### 4. File Operations

```yaml
name: code-review-with-backup
hooks:
  pre:
    # Create backup
    - shell: cp {{.file}} {{.file}}.backup
    # Get file stats
    - shell: wc -l {{.file}}
  post:
    # Save review to file
    - shell: echo "$REVIEW" > review-$(date +%Y%m%d).md
    # Clean up backup
    - shell: rm {{.file}}.backup

prompt:
  template: |
    Review the code in {{.file}}:
    $(cat {{.file}})
```

### 5. Notification Hooks

```yaml
name: ai-task-with-notifications
hooks:
  pre:
    - shell: osascript -e 'display notification "AI task started"'
  post:
    - shell: osascript -e 'display notification "AI task completed"'
    # Send to Slack
    - shell: |
        curl -X POST $SLACK_WEBHOOK_URL \
          -H 'Content-Type: application/json' \
          -d '{"text": "AI task completed"}'

prompt:
  template: |
    {{.task}}
```

## Advanced Features

### Multi-Line Shell Scripts

```yaml
hooks:
  pre:
    - shell: |
        #!/bin/bash
        set -e  # Exit on error

        echo "Preparing environment..."

        # Create directories
        mkdir -p ./output/{reports,logs,temp}

        # Set variables
        export TIMESTAMP=$(date +%Y%m%d_%H%M%S)
        export OUTPUT_DIR="./output/reports"

        echo "Environment ready"
```

### Error Handling

By default, hook errors stop execution. Override with `|| true`:

```yaml
hooks:
  pre:
    # This failure won't stop execution
    - shell: npm run optional-check || true

    # This failure WILL stop execution
    - shell: npm run critical-check
```

### Accessing Hook Output

Currently, hook output is captured but not passed to the LLM. Workaround using files:

```yaml
hooks:
  pre:
    # Write to temporary file
    - shell: git log -1 --pretty=%B > /tmp/last-commit.txt

prompt:
  template: |
    Previous commit message:
    $(cat /tmp/last-commit.txt)

    Generate a new commit message for:
    {{.stdin}}
```

### Template Variables in Hooks

Use command arguments in hooks:

```yaml
args:
  - name: file
    required: true
  - name: output_dir
    default: ./output

hooks:
  pre:
    - shell: echo "Processing {{.file}}"
    - shell: mkdir -p {{.output_dir}}
  post:
    - shell: cp {{.file}} {{.output_dir}}/
```

## Real-World Examples

### Example 1: Automated Code Review

```yaml
name: auto-review
version: 1.0.0
description: Review code with automated checks

args:
  - name: file
    description: File to review
    required: true

hooks:
  pre:
    # Run linter
    - shell: eslint {{.file}} > /tmp/lint-results.txt || true
    # Check complexity
    - shell: npx complexity-report {{.file}} > /tmp/complexity.txt || true
    # Run tests
    - shell: npm test -- {{.file}} > /tmp/test-results.txt || true
  post:
    # Save complete review
    - shell: |
        cat /tmp/lint-results.txt \
            /tmp/complexity.txt \
            /tmp/test-results.txt \
            > review-{{.file}}.txt
    # Clean up
    - shell: rm /tmp/{lint,complexity,test}-results.txt

prompt:
  system: |
    You are a senior code reviewer.
    Consider linting, complexity, and test results in your review.
  template: |
    Review {{.file}}:

    Linter output: $(cat /tmp/lint-results.txt)
    Complexity: $(cat /tmp/complexity.txt)
    Test results: $(cat /tmp/test-results.txt)

    Code:
    $(cat {{.file}})
```

**Usage:**
```bash
scmd auto-review src/app.js
```

### Example 2: Documentation Generator

```yaml
name: doc-gen
version: 1.0.0
description: Generate documentation from code

args:
  - name: module
    required: true

hooks:
  pre:
    # Extract exports
    - shell: |
        grep -E "^export|^function|^class" {{.module}} \
          > /tmp/exports.txt
    # Get git info
    - shell: |
        echo "Last modified: $(git log -1 --format=%cd {{.module}})" \
          > /tmp/git-info.txt
  post:
    # Save to docs directory
    - shell: mkdir -p ./docs/api
    - shell: |
        echo "# {{.module}} API Documentation" > ./docs/api/{{.module}}.md
        echo "" >> ./docs/api/{{.module}}.md
        cat $AI_OUTPUT >> ./docs/api/{{.module}}.md
    # Update index
    - shell: |
        echo "- [{{.module}}](./api/{{.module}}.md)" \
          >> ./docs/index.md

prompt:
  template: |
    Generate API documentation for {{.module}}:

    Exports:
    $(cat /tmp/exports.txt)

    Git info:
    $(cat /tmp/git-info.txt)

    Code:
    $(cat {{.module}})
```

### Example 3: Database Migration Helper

```yaml
name: migration-helper
version: 1.0.0
description: Generate database migration from schema changes

hooks:
  pre:
    # Dump current schema
    - shell: pg_dump --schema-only mydb > /tmp/old-schema.sql
    # Check migration directory
    - shell: mkdir -p ./migrations
    # Get next migration number
    - shell: |
        LAST=$(ls ./migrations | tail -1 | cut -d_ -f1)
        NEXT=$((LAST + 1))
        echo $NEXT > /tmp/migration-number.txt
  post:
    # Save migration file
    - shell: |
        NUM=$(cat /tmp/migration-number.txt)
        DATE=$(date +%Y%m%d)
        cp $AI_OUTPUT ./migrations/${NUM}_${DATE}_migration.sql
    # Test migration (dry run)
    - shell: |
        psql mydb_test < ./migrations/${NUM}_${DATE}_migration.sql || \
        echo "Warning: Migration test failed"

prompt:
  template: |
    Generate a migration from old to new schema:

    Old schema:
    $(cat /tmp/old-schema.sql)

    New schema:
    $(cat {{.new_schema_file}})
```

## Security Considerations

### Safe Commands Only

Hooks run with your shell permissions. Be careful with:

- `rm -rf` (deletion)
- `chmod` (permissions)
- `curl -X POST` (write operations)
- `sudo` (privilege escalation)

### Validating User Input

Always validate arguments used in hooks:

```yaml
args:
  - name: file
    required: true

hooks:
  pre:
    # Bad - vulnerable to injection
    - shell: cat {{.file}}

    # Better - validate extension
    - shell: |
        if [[ "{{.file}}" == *.js ]]; then
          cat {{.file}}
        else
          echo "Error: Only .js files allowed"
          exit 1
        fi
```

### Secrets in Hooks

Never hardcode secrets:

```yaml
hooks:
  post:
    # Bad
    - shell: curl -H "Token: sk-abc123" api.com/notify

    # Good - use environment variables
    - shell: curl -H "Token: $API_TOKEN" api.com/notify
```

## Debugging Hooks

Enable debug mode to see hook execution:

```bash
SCMD_DEBUG=1 scmd git-commit-with-checks
```

Output:
```
[DEBUG] Executing pre-hook 1: git status --short
[DEBUG] Hook output: M  src/main.go
[DEBUG] Hook completed successfully

[DEBUG] Executing pre-hook 2: git diff --check
[DEBUG] Hook output: (no output)
[DEBUG] Hook completed successfully

[DEBUG] Executing LLM completion...
[DEBUG] Completion successful

[DEBUG] Executing post-hook 1: echo "Commit message generated!"
[DEBUG] Hook output: Commit message generated!
```

## Limitations

### Current Limitations

1. **No hook output to LLM**: Hook output isn't automatically passed to the prompt (workaround: use files)
2. **Basic conditionals**: `if` condition support is limited
3. **No hook chaining**: Can't pass output from one hook to another directly
4. **Fixed timeout**: All hooks have 30-second timeout
5. **No async hooks**: All hooks run synchronously

### Future Enhancements

- [ ] Pass hook output to prompt via `{{.hook_output}}`
- [ ] Advanced conditionals (regex, comparisons)
- [ ] Hook dependencies (`depends_on: hook-name`)
- [ ] Configurable timeouts
- [ ] Parallel hook execution
- [ ] Hook templates/reusable hooks

## Best Practices

### 1. Keep Hooks Simple

Each hook should do one thing:

**Good:**
```yaml
hooks:
  pre:
    - shell: git status
    - shell: npm run lint
    - shell: npm test
```

**Bad:**
```yaml
hooks:
  pre:
    - shell: |
        git status && npm run lint && npm test && \
        do-many-other-things...
```

### 2. Use Descriptive Comments

```yaml
hooks:
  pre:
    # Validate git repository
    - shell: git rev-parse --git-dir

    # Check for uncommitted changes
    - shell: git diff --quiet || echo "Warning: uncommitted changes"

    # Run linter
    - shell: npm run lint
```

### 3. Handle Errors Gracefully

```yaml
hooks:
  pre:
    # Optional check - don't fail if missing
    - shell: which prettier || echo "prettier not installed"

    # Critical check - fail if missing
    - shell: which git || exit 1
```

### 4. Clean Up After Yourself

```yaml
hooks:
  pre:
    - shell: mkdir -p /tmp/scmd-work
  post:
    - shell: rm -rf /tmp/scmd-work  # Clean up
```

### 5. Log Important Actions

```yaml
hooks:
  post:
    - shell: echo "[$(date)] Command executed" >> ~/.scmd/command.log
```

## Combining with Other Features

### Hooks + Tool Calling

```yaml
hooks:
  pre:
    # Prepare environment for tools
    - shell: mkdir -p ./tool-workspace
    - shell: export TOOL_DIR=./tool-workspace

prompt:
  system: |
    You have access to tools.
    Write any files to ./tool-workspace directory.
  template: |
    {{.task}}

hooks:
  post:
    # Process tool-generated files
    - shell: cat ./tool-workspace/* > combined-output.txt
    - shell: rm -rf ./tool-workspace
```

### Hooks + Composition

```yaml
compose:
  pipeline:
    - command: analyze
    - command: summarize

hooks:
  pre:
    - shell: echo "Starting pipeline..."
  post:
    - shell: echo "Pipeline complete!"
    - shell: cat $PIPELINE_OUTPUT > results.txt
```

## Troubleshooting

### Hook Fails Silently

Hooks don't show output by default. Enable debug:

```bash
SCMD_DEBUG=1 scmd your-command
```

### "Command not found" in Hook

Hook can't find your command. Options:

1. Use full path: `/usr/local/bin/mycommand`
2. Add to PATH in hook: `export PATH=$PATH:/custom/path`
3. Use shell wrapper: `/bin/bash -c "source ~/.bashrc && mycommand"`

### Hook Timeout

30-second timeout is too short. Options:

1. Optimize the command
2. Run long tasks in background: `long-command & `
3. Use post-hook instead of pre-hook

### Variables Not Expanding

Template variables only work in YAML strings:

**Works:**
```yaml
hooks:
  pre:
    - shell: echo "File: {{.file}}"
```

**Doesn't work:**
```yaml
hooks:
  pre:
    - shell: MYVAR={{.file}}  # Variables in assignments don't expand
```

## Next Steps

- [Composition Guide](composition.md) - Chain commands with hooks
- [Tool Calling Guide](tool-calling.md) - Combine hooks with tools
- [Examples](../examples/git-workflow.md) - Real-world hook examples
- [Best Practices](best-practices.md) - Command authoring tips
