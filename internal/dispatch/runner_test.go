package dispatch

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExecRunnerPreservesArgumentsAndWorkingDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-only")
	}
	dir := t.TempDir()
	output := filepath.Join(t.TempDir(), "result")
	script := filepath.Join(t.TempDir(), "agent")
	contents := "#!/bin/sh\nprintf '%s\\n' \"$PWD\" \"$1\" > \"$OUTPUT\"\n"
	if err := os.WriteFile(script, []byte(contents), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("OUTPUT", output)
	prompt := `fix "quotes"; echo unsafe`

	runner := ExecRunner{}
	if err := runner.Run(script, []string{prompt}, dir); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	want := dir + "\n" + prompt
	if got := strings.TrimSpace(string(data)); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
