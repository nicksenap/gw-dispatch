package dispatch

import (
	"os"
	"os/exec"
)

type ExecRunner struct{}

func (ExecRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func (ExecRunner) Output(name string, args []string) ([]byte, error) {
	command := exec.Command(name, args...)
	command.Stderr = os.Stderr
	return command.Output()
}

func (ExecRunner) Run(name string, args []string, dir string) error {
	command := exec.Command(name, args...)
	command.Dir = dir
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
