package cmd

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"testing"
)

func TestExitCodePreservesChildStatus(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-only")
	}
	childErr := exec.Command("sh", "-c", "exit 23").Run()
	if childErr == nil {
		t.Fatal("child error = nil")
	}
	if got := ExitCode(fmt.Errorf("wrapped: %w", childErr)); got != 23 {
		t.Fatalf("ExitCode() = %d, want 23", got)
	}
}

func TestExitCodeDefaults(t *testing.T) {
	if got := ExitCode(nil); got != 0 {
		t.Fatalf("ExitCode(nil) = %d, want 0", got)
	}
	if got := ExitCode(errors.New("failure")); got != 1 {
		t.Fatalf("ExitCode(error) = %d, want 1", got)
	}
}
