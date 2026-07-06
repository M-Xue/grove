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

Requirements:

- Go 1.24.2+

`grove` can print the selected worktree path, but changing your current shell directory requires shell integration. The installer handles both the binary install and shell config wiring. Run the installer for your shell, then reload your shell config.

zsh on Linux/macOS:

```bash
sh scripts/install.sh zsh
```

bash on Linux/macOS:

```bash
sh scripts/install.sh bash
```

If you omit the argument, the script falls back to your `$SHELL` value. The script itself still runs under `sh`. This builds `grove`, installs it to `~/.local/bin/grove`, writes the shell wrapper to `~/.local/share/grove/init.sh` (honoring `$XDG_DATA_HOME`), and adds a single line sourcing that file to either `~/.zshrc` or `~/.bashrc`.

PowerShell:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install.ps1
```

This builds `grove.exe`, installs it to `$HOME\AppData\Local\Programs\grove\grove.exe`, writes the shell wrapper to `…\Programs\grove\init.ps1`, and adds a single line sourcing that file to `$PROFILE`.

Reload your shell config afterwards (`source ~/.zshrc`, `. ~/.bashrc`, or `. $PROFILE`).

## Updating

The wrapper is re-evaluated from the binary on each shell start, so your rc file / `$PROFILE` never needs editing again. To update after pulling a new version, just re-run the installer for your shell. It rebuilds the binary into the install location and is idempotent, so it won't duplicate the source line in your config.

## Uninstall

grove only ever writes to a couple of fixed locations, so removing it by hand is straightforward: delete the directories it installed to and remove the single source line from your shell config.

zsh and bash:

```bash
rm ~/.local/bin/grove
rm -rf ~/.local/share/grove
```

Then delete the grove source line from `~/.zshrc` (zsh) or `~/.bashrc` (bash). It looks like this, with your home directory expanded to an absolute path:

```sh
[ -f "$HOME/.local/share/grove/init.sh" ] && . "$HOME/.local/share/grove/init.sh"
```

PowerShell:

```powershell
Remove-Item -Recurse -Force "$HOME\AppData\Local\Programs\grove"
```

Then delete the grove source line from your profile at `$PROFILE`. It looks like this, with your home directory expanded to an absolute path:

```powershell
. "$HOME\AppData\Local\Programs\grove\init.ps1"
```

Reload your shell (or open a new session) afterwards so the wrapper is no longer defined.
