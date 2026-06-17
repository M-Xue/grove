package shellinit

import (
	"strings"
	"testing"
)

func TestScript(t *testing.T) {
	bashScript, err := Script("bash", "/tmp/grove")
	if err != nil {
		t.Fatalf("Script(bash) returned error: %v", err)
	}
	if bashScript == "" {
		t.Fatal("expected bash shell script")
	}
	if want := "'/tmp/grove'"; !strings.Contains(bashScript, want) {
		t.Fatalf("expected bash script to contain %q", want)
	}

	powershellScript, err := Script("powershell", `C:\grove\grove.exe`)
	if err != nil {
		t.Fatalf("Script(powershell) returned error: %v", err)
	}
	if want := `Set-Alias grove Invoke-Grove`; !strings.Contains(powershellScript, want) {
		t.Fatalf("expected powershell script to contain %q", want)
	}
}

func TestScriptRejectsUnsupportedShell(t *testing.T) {
	if _, err := Script("tcsh", "/tmp/grove"); err == nil {
		t.Fatal("expected error for unsupported shell")
	}
}

func TestIsSupported(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "powershell"} {
		if !IsSupported(shell) {
			t.Fatalf("expected %q to be supported", shell)
		}
	}
	if IsSupported("tcsh") {
		t.Fatal("expected tcsh to be unsupported")
	}
}
