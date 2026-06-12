# grove

`grove` is a terminal UI for browsing and managing git worktrees.

Press `enter` in the TUI to select the highlighted worktree. The binary prints the selected path to `stdout` and exits. A shell wrapper captures that output and changes your current shell directory.

## Install

The simplest install path is:

1. Build `grove` locally with `go build`.
2. Move the binary to a stable location.
3. Add a small shell function to your shell config.

The binary alone cannot change your current shell directory.

### Build The Binary

Requirements:

- Go 1.24.2+

Build for your current machine:

```bash
go build -o grove .
```

Go automatically builds for your current OS and CPU architecture. You can inspect what it will use with:

```bash
go env GOOS GOARCH
```

### zsh and bash

1. Build `grove` locally with `go build -o grove .`.
2. Move it to `~/.local/bin/grove`.
3. Make it executable:

```bash
chmod +x ~/.local/bin/grove
```

4. Add this function to your shell config:

```sh
grove() {
	local output
	output="$($HOME/.local/bin/grove "$@")"
	local status=$?
	if [ $status -ne 0 ]; then
		return $status
	fi
	if [ -n "$output" ]; then
		cd "$output" || return 1
	fi
}
```

Use `~/.zshrc` for `zsh` or `~/.bashrc` for `bash`.

5. Reload your shell:

```bash
source ~/.zshrc
```

or:

```bash
. ~/.bashrc
```

### PowerShell

1. Build `grove.exe` locally with `go build -o grove.exe .`.
2. Move it to `$HOME\AppData\Local\Programs\grove\grove.exe`.
3. Add this to your PowerShell profile at `$PROFILE`:

```powershell
function Invoke-Grove {
    param(
        [Parameter(ValueFromRemainingArguments = $true)]
        [string[]]$Arguments
    )

    $output = & "$HOME\AppData\Local\Programs\grove\grove.exe" @Arguments
    if ($LASTEXITCODE -ne 0) {
        return $LASTEXITCODE
    }

    if (-not [string]::IsNullOrWhiteSpace($output)) {
        Set-Location $output
    }
}

Set-Alias grove Invoke-Grove
```

4. Reload PowerShell:

```powershell
. $PROFILE
```

## Run From Source

This runs the TUI only. Without the shell wrapper, selecting a worktree prints its path to `stdout` instead of changing your shell directory.

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
