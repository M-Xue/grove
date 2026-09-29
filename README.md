# grove 🌳

`grove` is a terminal UI for browsing, filtering, creating, and removing Git worktrees.

It helps developers jump between parallel branches quickly from a keyboard-first interface, while preserving a clean shell-integration contract for directory switching.

## Features

- Browse all worktrees in a fast terminal UI
- Filter worktrees by path
- Jump directly into a selected worktree
- Create worktrees from existing or new branches
- Remove worktrees without leaving the TUI

## Install

Requirements:

- Go 1.24.2+

`grove` can print the selected worktree path, but changing your current shell directory requires shell integration. The installer handles both the binary install and shell config wiring. Run the installer for your shell, then reload your shell config. The scripts locate the repo themselves, so you can run them from any directory.

zsh on Linux/macOS:

```bash
sh scripts/install.sh zsh
```

bash on Linux/macOS:

```bash
sh scripts/install.sh bash
```

If you omit the argument, the script falls back to your `$SHELL` value (only bash and zsh are supported). The script itself still runs under `sh`. This builds `grove`, installs it to `~/.local/bin/grove` (honoring `$BIN_DIR`), writes the shell wrapper to `~/.local/share/grove/init.sh` (honoring `$XDG_DATA_HOME`), and adds a single line sourcing that file to your shell's startup file: `~/.zshrc` for zsh, `~/.bashrc` for bash on Linux. For bash on macOS — where terminals start login shells that never read `~/.bashrc` — the line goes to the startup file bash actually reads: the first existing of `~/.bash_profile`, `~/.bash_login`, or `~/.profile` (creating `~/.bash_profile` if none exist).

PowerShell (Windows only):

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install.ps1
```

This builds `grove.exe`, installs it to `$HOME\AppData\Local\Programs\grove\grove.exe` (honoring `$env:GROVE_INSTALL_DIR`), writes the shell wrapper to `…\Programs\grove\init.ps1`, and adds a single line sourcing that file to `$PROFILE`. Because `$PROFILE` is specific to the PowerShell host, run the installer from the PowerShell you actually use — substitute `pwsh` for `powershell` above if you use PowerShell 7.

Reload your shell config afterwards (`source ~/.zshrc`, `. ~/.bashrc`, or `. $PROFILE`).

## Updating

The wrapper is re-evaluated from the binary on each shell start, so your rc file / `$PROFILE` never needs editing again. To update after pulling a new version, just re-run the installer for your shell. It rebuilds the binary into the install location and is idempotent, so it won't duplicate the source line in your config.

## Uninstall

zsh and bash:

```bash
sh scripts/uninstall.sh
```

This removes everything `install.sh` created — the binary, `~/.local/share/grove`, the source line from your shell startup files, and grove's worktree cache (`~/Library/Caches/grove` on macOS, `~/.cache/grove` on Linux). It honors the same `$BIN_DIR` and `$XDG_DATA_HOME` overrides as the installer, so export them again here if you set them at install time; if it finds a grove source line pointing at a different install location, it leaves it alone and tells you.

To uninstall by hand instead: delete the binary and directories above, then remove the grove source line from your shell startup file. It looks like this, with your home directory expanded to an absolute path:

```sh
[ -f "$HOME/.local/share/grove/init.sh" ] && . "$HOME/.local/share/grove/init.sh"
```

PowerShell:

```powershell
Remove-Item -Recurse -Force "$HOME\AppData\Local\Programs\grove"
Remove-Item -Recurse -Force "$env:LOCALAPPDATA\grove"   # worktree cache
```

(Adjust the first path if you installed with `$env:GROVE_INSTALL_DIR`.) Then delete the grove source line from your profile at `$PROFILE`. It looks like this, with your home directory expanded to an absolute path:

```powershell
. "$HOME\AppData\Local\Programs\grove\init.ps1"
```

Reload your shell (or open a new session) afterwards so the wrapper is no longer defined.
