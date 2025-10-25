# AGENTS.md

This file provides guidance to AI agents when working with code in this repository.

## Commands

### Build and Development
- **Build**: `task build` or `go build ./cmd/bump`
- **Install**: `task install` or `go install ./cmd/bump`
- **Clean**: `task clean` (removes the `bump` binary)

### Testing
- **Run tests**: `task test` or `go test -cover -coverprofile=coverage.out ./...`
- **Coverage report**: Tests automatically generate `coverage.out` and open HTML coverage report

### Code Quality
- **Lint**: `task lint` or `golangci-lint run`

### Task Runner
This project uses [Task](https://taskfile.dev) as a task runner. Run `task` to see all available tasks.

## Architecture

This is a Go CLI tool for semantic versioning and git tag management with the following structure:

### Core Components
- **`bump.go`**: Core library containing semantic versioning logic, git operations, and tag management
- **`cmd/bump/main.go`**: CLI application using urfave/cli that wraps the core library
- **`bump_test.go`**: Unit tests for the core functionality

### Key Libraries
- **urfave/cli/v2**: Command-line interface framework
- **go-git/go-git/v5**: Git operations (reading tags, opening repositories)
- **charmbracelet/log**: Structured logging
- **gopkg.in/ini.v1**: Configuration file parsing for git config

### Core Functionality
The tool operates on semantic versioning (semver) with these primary operations:
1. **Version Parsing**: Uses regex `^v(\d+)\.(\d+)\.(\d+)(-[0-9A-Za-z-.]+)?$` to parse semantic versions
2. **Tag Management**: Reads existing git tags, determines latest version, calculates next version
3. **Repository Configuration**: Stores per-repo push preferences in `.git/config` under `[bump]` section
4. **File Updates**: Can update Go source files with new development versions using AST parsing

### Command Structure
- `bump patch|minor|major [--suffix] [--push] [--update-file]`: Version bumping commands
- `bump push`: Push existing tags to remote
- `bump config --default-push[=true|false]`: Configure default push behavior per repository

### Testing Patterns
- Uses table-driven tests and mocking patterns
- Mock `execCommand` variable for testing git command execution
- Tests cover version parsing, sorting, and tag operations

## Best Practices
- Use Taskfile tasks where possible

## Issue Tracking with bd (beads)

**IMPORTANT**: This project uses **bd (beads)** for ALL issue tracking. Do NOT use markdown TODOs, task lists, or other tracking methods.

### Why bd?

- Dependency-aware: Track blockers and relationships between issues
- Git-friendly: Auto-syncs to JSONL for version control
- Agent-optimized: JSON output, ready work detection, discovered-from links
- Prevents duplicate tracking systems and confusion

### Quick Start

**Check for ready work:**
```bash
bd ready --json
```

**Create new issues:**
```bash
bd create "Issue title" -t bug|feature|task -p 0-4 --json
bd create "Issue title" -p 1 --deps discovered-from:bd-123 --json
```

**Claim and update:**
```bash
bd update bd-42 --status in_progress --json
bd update bd-42 --priority 1 --json
```

**Complete work:**
```bash
bd close bd-42 --reason "Completed" --json
```

### Issue Types

- `bug` - Something broken
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature with subtasks
- `chore` - Maintenance (dependencies, tooling)

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

### Workflow for AI Agents

1. **Check ready work**: `bd ready` shows unblocked issues
2. **Claim your task**: `bd update <id> --status in_progress`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue:
   - `bd create "Found bug" -p 1 --deps discovered-from:<parent-id>`
5. **Complete**: `bd close <id> --reason "Done"`

### Auto-Sync

bd automatically syncs with git:
- Exports to `.beads/issues.jsonl` after changes (5s debounce)
- Imports from JSONL when newer (e.g., after `git pull`)
- No manual export/import needed!

### MCP Server (Recommended)

If using Claude or MCP-compatible clients, install the beads MCP server:

```bash
pip install beads-mcp
```

Add to MCP config (e.g., `~/.config/claude/config.json`):
```json
{
  "beads": {
    "command": "beads-mcp",
    "args": []
  }
}
```

Then use `mcp__beads__*` functions instead of CLI commands.

### Important Rules

- ✅ Use bd for ALL task tracking
- ✅ Always use `--json` flag for programmatic use
- ✅ Link discovered work with `discovered-from` dependencies
- ✅ Check `bd ready` before asking "what should I work on?"
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT use external issue trackers
- ❌ Do NOT duplicate tracking systems

For more details, see README.md and QUICKSTART.md.
