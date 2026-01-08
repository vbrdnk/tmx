# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`tmx` is a tmux session manager written in Go that provides interactive directory selection using fzf, with support for nested directory search and zoxide integration for frecency-based suggestions. It manages tmux sessions with configurable workspace layouts defined in TOML files.

## Development Commands

### Building and Running

```bash
# Build and install to $GOPATH/bin
go install

# Build without installing
go build -o tmx

# Run directly
go run main.go

# Run with arguments
go run main.go --depth 3 ~/Projects
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./pkg/config
go test ./pkg/session
go test ./pkg/discovery

# Run a specific test
go test ./pkg/session -run TestNewSessionManager

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter (if golangci-lint is installed)
golangci-lint run

# Vet code
go vet ./...

# Update dependencies
go mod tidy
```

## Architecture Overview

### High-Level Flow

The application follows a clean separation of concerns with distinct layers:

1. **CLI Layer** (`cmd/`): Handles command-line argument parsing and orchestrates the main workflow
2. **Discovery Layer** (`pkg/discovery/`): Orchestrates directory discovery combining zoxide and file search
3. **Search Layer** (`pkg/search/`): Performs directory searches using fd/find and queries zoxide cache
4. **Session Layer** (`pkg/session/`): Manages tmux session lifecycle (create, attach, list, kill)
5. **UI Layer** (`pkg/ui/`): Provides fzf integration for interactive selection
6. **Config Layer** (`pkg/config/`): Handles TOML configuration parsing and validation

### Key Components

#### DirectorySearcher (`pkg/search/searcher.go`)

Handles low-level directory discovery using external tools:
- Detects if `fd` is available, falls back to `find` if not
- Queries zoxide for frecency-based directory suggestions
- Returns lists of directories with configurable depth

#### DirectorySelector (`pkg/discovery/selector.go`)

Orchestrates the directory selection workflow:
- Combines zoxide results (marked with ★) and find/fd results
- Deduplicates paths using a seen-paths map
- Presents unified list to fzf for user selection

#### SessionManager (`pkg/session/manager.go`)

Manages the full tmux session lifecycle:
- Maps selected directories to workspace configurations
- Creates sessions with custom window layouts based on config
- Falls back to default single-window sessions when no config matches
- Uses base path matching: `filepath.Base(dir) == filepath.Base(ws.Directory)`

#### Config (`pkg/config/config.go`)

Handles configuration from `~/.config/tmx/*.toml`:
- Supports multiple TOML files in the config directory
- Merges workspace configs from all files
- Global settings: `search_depth` (default: 1), `use_zoxide` (default: true)
- Validates workspace configs to prevent duplicates and empty values

### Data Flow

```
User runs tmx [path] [--depth N]
    ↓
CLI parses args, loads config from ~/.config/tmx/*.toml
    ↓
DirectorySelector.SelectDirectory()
    ├─→ DirectorySearcher.QueryZoxideCache() → zoxide query --list
    └─→ DirectorySearcher.Search() → fd/find with depth limit
    ↓
Combine and deduplicate results
    ↓
ui.FuzzyFind() → Present to user via fzf
    ↓
User selects directory
    ↓
SessionManager.ResolveSession()
    ├─→ Match directory to workspace config (base path comparison)
    ├─→ Create session if doesn't exist
    └─→ Attach/switch to session
```

### Important Implementation Details

#### Session Name Sanitization

Session names are sanitized in `SessionManager.createSessionName()` by replacing characters that tmux doesn't accept (spaces, dots, colons, special chars) with underscores.

#### Tmux Command Abstraction

The `TmuxCommand` type (`pkg/session/command.go`) provides methods for different execution contexts:
- `Execute()`: Silent execution
- `ExecuteWithIO()`: Connects stdin/stdout/stderr (used for interactive attach)
- `ExecuteVerbose()`: Captures stderr for debugging
- `Output()`: Returns command output

#### Zoxide Integration

Zoxide is optional and degrades gracefully:
- If `use_zoxide = false` in config, skip zoxide queries entirely
- If zoxide is not installed, silently ignore errors
- Frecent directories are marked with ★ prefix and appear first in fzf

#### Depth Handling

Depth priority: CLI flag > config file > default (1)
- `0` means unlimited depth (use with caution on large trees)
- `1` searches only direct subdirectories (fastest, recommended)

## Testing Patterns

Tests use Go's standard testing package with table-driven tests where appropriate:
- Mock tmux commands are not typically used; tests focus on business logic
- Config tests create temporary directories with TOML files
- Session manager tests verify session name sanitization and config matching

## Dependencies

- `github.com/BurntSushi/toml` - TOML configuration parsing
- `github.com/fatih/color` - Colored terminal output
- `github.com/urfave/cli/v3` - CLI framework

External tools (runtime dependencies):
- `tmux` (required)
- `fzf` (required)
- `fd` (optional, falls back to `find`)
- `zoxide` (optional, can be disabled)

## Configuration

Config files are loaded from `~/.config/tmx/` directory:
- All `.toml` files are parsed and merged
- Hidden files (starting with `.`) are ignored
- Workspace configs are appended from all files
- Global settings from the last file win

Example workspace matching:
```toml
[[workspace]]
directory = "/git/example"     # Base path is "example"
name = "example session"
windows = ["editor", "server"]
```

This matches when you select `/git/example` or any path ending in `example` directory.
