# tmx - Tmux Session Manager

`tmx` is a simple Go application that helps you manage tmux sessions. It uses fzf for interactive directory selection and creates tmux sessions with predefined windows based on your configuration.

## Features

- Interactive directory selection using [fzf](https://github.com/junegunn/fzf)
- **Nested directory search** with configurable depth
- **Zoxide integration** for frecency-based directory suggestions
- **Fast file discovery** using `fd` (with fallback to `find`)
- Configure workspaces with custom names and window layouts
- Attach to existing sessions or create new ones as needed
- Simple and easy-to-use command-line interface
- Accepts an optional path argument to specify search directory

## Installation

### Prerequisites

**Required:**
- Go 1.16 or higher
- [tmux](https://github.com/tmux/tmux/wiki) installed on your system
- [fzf](https://github.com/junegunn/fzf) installed on your system

**Optional (but recommended):**
- [fd](https://github.com/sharkdp/fd) - Fast alternative to `find` (automatically detected and used if available)
- [zoxide](https://github.com/ajeetdsouza/zoxide) - Frecency-based directory jumper for smarter directory suggestions

### Build and Install

```bash
# Clone the repository
git clone https://github.com/vbrdnk/tmx.git
cd tmx

# Build and install
go install

# Install directly from the repository
go install github.com/vbrdnk/tmx@latest
```

This will compile the application and place the executable in your `$GOPATH/bin` directory. Make sure this directory is in your `PATH` to access the `tmx` command from anywhere.

## Shell Completions

`tmx` supports shell completions for bash, zsh, and fish. This enables tab completion for subcommands, flags, and aliases.

### Bash

Add to your `~/.bashrc` or `~/.bash_profile`:

```bash
# Load tmx completions
eval "$(tmx completion bash)"
```

Or install system-wide:

```bash
tmx completion bash | sudo tee /etc/bash_completion.d/tmx
```

### Zsh

Add to your `~/.zshrc`:

```bash
# Load tmx completions
eval "$(tmx completion zsh)"
```

Or install to your fpath (requires a directory in `$fpath`):

```bash
tmx completion zsh > /usr/local/share/zsh/site-functions/_tmx
```

### Fish

Add to your `~/.config/fish/config.fish`:

```fish
# Load tmx completions
tmx completion fish | source
```

Or install to the completions directory:

```bash
tmx completion fish > ~/.config/fish/completions/tmx.fish
```

After installing completions, restart your shell or source your configuration file for the changes to take effect.

## Configuration

Create a configuration file at `~/.config/tmx/tmx.toml` (or any `.toml` file in `~/.config/tmx/`) with the following structure:

```toml
# Global settings (optional)
search_depth = 1        # Search depth for nested directories (1 = direct subdirectories, 0 = unlimited)
use_zoxide = true       # Use zoxide for frecency-based directory suggestions

# Workspace configurations
[[workspace]]
directory = "/path/to/your/project"
name = "project-name"
windows = ["editor", "server", "terminal"]

[[workspace]]
directory = "/another/project"
name = "another-project"
windows = ["code", "build", "logs"]
```

### Configuration Options

#### Global Settings

- `search_depth` (optional, default: `1`): Controls how deep the directory search goes
  - `1`: Only search direct subdirectories (fastest, default)
  - `2-5`: Search nested directories up to N levels deep
  - `0`: Unlimited depth (use with caution on large directory trees)
- `use_zoxide` (optional, default: `true`): Enable integration with [zoxide](https://github.com/ajeetdsouza/zoxide) for frecency-based directory suggestions
  - When enabled, frequently/recently accessed directories appear at the top of the fzf menu (marked with ★)
  - Gracefully falls back if zoxide is not installed

#### Workspace Settings

- `directory`: The directory path that will trigger this workspace configuration. The app uses base path comparison to check it against the directory selected with fzf.
- `name`: A friendly name for the tmux session
- `windows`: A list of window names to create in the session

## Usage

Run without arguments to search from your home directory:

```bash
tmx
```

Or specify a starting directory for the search:

```bash
tmx /path/to/search/from
```

Override the search depth with the `--depth` (or `-d`) flag:

```bash
# Search 3 levels deep from home directory
tmx --depth 3

# Search unlimited depth from a specific directory
tmx --depth 0 ~/Projects

# Search only direct subdirectories (same as default)
tmx -d 1 /git
```

The application will:

1. Present an interactive fzf-based selection menu of directories
   - If zoxide is enabled, frequently accessed directories appear first (marked with ★)
   - Remaining directories are listed alphabetically
2. After you select a directory, it will check if it matches any configured workspace
3. Create a tmux session with the configured windows if it doesn't exist
4. Attach to the session

If no configuration matches the selected directory, it will create a session named after the directory.

### Performance Tips

- **Default depth (1)** is fastest and works well when you organize projects in a flat structure (e.g., `~/Git/project1`, `~/Git/project2`)
- **Moderate depth (2-3)** is suitable for nested project structures (e.g., `~/Git/org/team/project`)
- **Unlimited depth (0)** can be slow on large directory trees - use with specific paths
- **Zoxide integration** helps you quickly access frequently-used directories without deep searches

### Subcommands

- `connect` (aliases: `c`, `conn`) - Connect to an existing tmux session
- `list` (aliases: `l`, `ls`) - List all active tmux sessions
- `kill` (aliases: `k`) - Kill a tmux session

## Example

For a configuration like:

```toml
search_depth = 2
use_zoxide = true

[[workspace]]
directory = "/git/example"
name = "example session"
windows = ["editor", "server", "lazygit"]
```

**Scenario 1:** You run `tmx` from your home directory:
- The tool searches 2 levels deep for directories
- Zoxide-tracked directories appear first with a ★ marker
- When you select `/git/example`, it creates a session named `example_session` with three windows: `editor`, `server`, and `lazygit`

**Scenario 2:** You run `tmx --depth 0 ~/Projects`:
- The tool searches unlimited depth under `~/Projects`
- Useful when you can't remember the exact nesting level of your project
- Zoxide helps prioritize frequently-used projects at the top of the list

## License

MIT
