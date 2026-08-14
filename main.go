package main

import (
	"fmt"
	"os"

	"github.com/nicksenap/gw-dispatch/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(cmd.ExitCode(err))
	}
}
