# grove 🌳

`grove` is a terminal UI for browsing, filtering, creating, and removing Git worktrees.

It helps developers jump between parallel branches quickly from a keyboard-first interface, while preserving a clean shell-integration contract for directory switching.

## Features

- Browse all worktrees in a fast terminal UI
- Filter worktrees by path
- Jump directly into a selected worktree
- Switch directly to a local or remote-tracking branch
- Preview recent commits for the selected branch
- Fetch, delete, and bulk-delete local branches
- Create worktrees from existing or new branches
- Remove worktrees without leaving the TUI

## Install

The simplest install path is the repo-managed installer.

1. Run the installer for your shell.
2. Reload your shell config.

`grove` can print the selected worktree path, but changing your current shell directory requires shell integration. The installers handle both the binary install and shell config wiring.

### Quick Install

Requirements:

- Go 1.24.2+

zsh on Linux/macOS:

```bash
sh scripts/install.sh zsh
```

bash on Linux/macOS:

```bash
sh scripts/install.sh bash
```

If you omit the argument, the script falls back to your `$SHELL` value. The script itself still runs under `sh`.

This builds `grove`, installs it to `~/.local/bin/grove`, writes the shell wrapper to `~/.local/share/grove/init.sh` (honoring `$XDG_DATA_HOME`), and adds a single line sourcing that file to either `~/.zshrc` or `~/.bashrc`. The wrapper is re-evaluated from the binary on each shell start, so updating grove never requires editing your rc file again.

PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install.ps1
```

This builds `grove.exe`, installs it to `$HOME\AppData\Local\Programs\grove\grove.exe`, writes the shell wrapper to `…\Programs\grove\init.ps1`, and adds a single line sourcing that file to `$PROFILE`. The wrapper is re-evaluated from the binary on each shell start, so updating grove never requires editing your profile again.

### Build The Binary

Build for your current machine:

```bash
go build -o grove .
```

Go automatically builds for your current OS and CPU architecture. You can inspect what it will use with:

```bash
go env GOOS GOARCH
```

### Manual Install

If you want to wire it up yourself, build the binary and move it to a stable location first.

```bash
go build -o grove .
```

Examples below assume the binary lives at `~/.local/bin/grove` on Unix-like systems and `$HOME\AppData\Local\Programs\grove\grove.exe` on Windows.

### shell-init

`grove shell-init <shell>` prints the wrapper code for a supported shell:

- `bash`
- `zsh`
- `powershell`

The generated wrapper runs the real `grove` binary, captures its `stdout`, and changes your shell directory when a path is returned.

### zsh and bash

Add this line to `~/.zshrc`:

```sh
eval "$(~/.local/bin/grove shell-init zsh)"
```

For bash, use:

```sh
eval "$(~/.local/bin/grove shell-init bash)"
```

Reload your shell:

```bash
source ~/.zshrc
```

or:

```bash
. ~/.bashrc
```

### PowerShell

Add this to your PowerShell profile at `$PROFILE`:

```powershell
Invoke-Expression (& "$HOME\AppData\Local\Programs\grove\grove.exe" shell-init powershell)
```

Reload PowerShell:

```powershell
. $PROFILE
```

## Run From Source

This runs the TUI only. Without shell integration, selecting a worktree prints its path to `stdout` instead of changing your shell directory.

```bash
go run .
```

## Build

```bash
go build ./...
```

## Test

```bash
go test ./...
```
