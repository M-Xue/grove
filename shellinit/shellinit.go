// Package shellinit produces the shell integration snippets emitted by
// `grove shell-init <shell>`. The snippet wraps the grove binary so that the
// path grove prints on stdout is turned into a directory change in the user's
// shell.
//
// The package is pure: the caller resolves the binary path (via
// os.Executable()) and passes it in, keeping process/global state out of here.
package shellinit

import (
	"fmt"
	"path/filepath"
	"strings"
)

// IsSupported reports whether grove can emit a shell-init script for shell.
func IsSupported(shell string) bool {
	switch shell {
	case "bash", "zsh", "powershell":
		return true
	default:
		return false
	}
}

// Script returns the shell integration snippet for shell, wiring the `grove`
// command to cd into the path grove prints on stdout. binaryPath is the
// absolute path to the grove executable, resolved by the caller. It returns an
// error if shell is not supported.
func Script(shell, binaryPath string) (string, error) {
	if !IsSupported(shell) {
		return "", fmt.Errorf("unsupported shell %q", shell)
	}
	quotedPath := shellQuote(binaryPath)
	switch shell {
	case "powershell":
		return fmt.Sprintf(`function Invoke-Grove {
    param(
        [Parameter(ValueFromRemainingArguments = $true)]
        [string[]]$Arguments
    )

    $output = & "%s" @Arguments
    if ($LASTEXITCODE -ne 0) {
        return $LASTEXITCODE
    }

    if (-not [string]::IsNullOrWhiteSpace($output)) {
        Set-Location $output
    }
}

Set-Alias grove Invoke-Grove
`, filepath.Clean(binaryPath)), nil
	default:
		return fmt.Sprintf(`grove() {
		local output
		output="$(%s "$@")"
		local status=$?
		if [ $status -ne 0 ]; then
			return $status
		fi
		if [ -n "$output" ]; then
			cd "$output" || return 1
		fi
	}
	`, quotedPath), nil
	}
}

func shellQuote(path string) string {
	return "'" + strings.ReplaceAll(path, "'", "'\\''") + "'"
}
