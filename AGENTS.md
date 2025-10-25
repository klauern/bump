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

## Task Management

This project uses [Beads](https://github.com/thoughtrealm/beads) for issue tracking and task management.

### Getting Started

**First time?** Run `bd onboard` to learn beads interactively.

All tasks are stored in `.beads/` directory. Common commands:

```bash
bd ready              # Find tasks ready to work on
bd list               # List all tasks
bd show task-<id>     # View task details
bd stats              # View project statistics
```

For full command reference: `bd --help`
