# tmx - Tmux Session Manager

`tmx` is a simple Go application that helps you manage tmux sessions. It uses fzf for interactive directory selection and creates tmux sessions with predefined windows based on your configuration.

## Features

- Interactive directory selection using [fzf](https://github.com/junegunn/fzf)
- Configure workspaces with custom names and window layouts
- Attach to existing sessions or create new ones as needed
- Simple and easy-to-use command-line interface
- Accepts an optional path argument to specify search directory

## Installation

### Prerequisites

- Go 1.16 or higher
- tmux installed on your system
- fzf installed on your system

### Build and Install

```bash
# Clone the repository
git clone https://github.com/vbrdnk/tmx.git
cd tmx

# Build and install
go install
```

This will compile the application and place the executable in your `$GOPATH/bin` directory. Make sure this directory is in your `PATH` to access the `tmx` command from anywhere.

## Configuration

Create a configuration file at `~/.config/tmx.toml` with the following structure:

```toml
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

The application will:

1. Present an interactive fzf-based selection menu of directories (from either your home directory or the specified path)
2. After you select a directory, it will check if it matches any configured workspace
3. Create a tmux session with the configured windows if it doesn't exist
4. Attach to the session

If no configuration matches the selected directory, it will create a session named after the directory.

### Subcommands

- `connect`
- `list`
- `kill`

### Prerequisites

- [fzf](https://github.com/junegunn/fzf) must be installed on your system
- [tmux](https://github.com/tmux/tmux/wiki) must be installed on your system

## Example

For a configuration like:

```toml
[[workspace]]
directory = "/git/example"
name = "example session"
windows = ["editor", "server", "lazygit"]
```

When you run `tmx` and select the `/git/example` directory from the fzf menu, it will create a session named `example_session` with three windows: `editor`, `server`, and `lazygit`.

## License

MIT
