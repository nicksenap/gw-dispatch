package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/nicksenap/gw-dispatch/internal/config"
	"github.com/nicksenap/gw-dispatch/internal/dispatch"
	"github.com/spf13/cobra"
)

var Version = "dev"

type dispatchFunc func(dispatch.Options, config.Config) error

func newRootCommand(run dispatchFunc) *cobra.Command {
	var branch string
	var repos string
	var preset string
	var prompt string
	var agent string
	var configPath string
	var noHooks bool
	var prURL string

	command := &cobra.Command{
		Use:   "gw-dispatch",
		Short: "Create a Grove workspace and dispatch a coding agent",
		Long: `Create a Grove workspace, then start an interactive coding agent in it
with the supplied prompt. Pi is used by default; Claude Code, Codex, OpenCode,
and custom configured agents are also supported.`,
		Example: `  gw dispatch -n -P "Implement login"  # uses dispatch config selector
  gw dispatch -b feat/login -p backend --agent claude -P "Implement login"
  gw dispatch --pr https://github.com/acme/api/pull/42            # review the PR
  gw dispatch --pr https://github.com/acme/api/pull/42 -P "Fix the failing tests"`,
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("config") {
				if _, err := os.Stat(configPath); err != nil {
					return fmt.Errorf("read config %s: %w", configPath, err)
				}
			}
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			if prompt == "" && prURL == "" {
				return fmt.Errorf("required flag(s) \"prompt\" not set")
			}
			return run(dispatch.Options{
				Branch:  branch,
				Repos:   repos,
				Preset:  preset,
				Prompt:  prompt,
				Agent:   agent,
				NoHooks: noHooks,
				PR:      prURL,
			}, cfg)
		},
	}

	flags := command.Flags()
	flags.StringVarP(&branch, "branch", "b", "", "Branch name (default: generated from prompt)")
	flags.StringVarP(&repos, "repos", "r", "", "Comma-separated repo names")
	flags.StringVarP(&preset, "preset", "p", "", "Use named preset")
	flags.StringVarP(&prompt, "prompt", "P", "", "Initial prompt sent to the agent")
	flags.StringVar(&agent, "agent", "", "Agent name (default: configured agent or pi)")
	flags.StringVar(&configPath, "config", config.DefaultPath(), "Dispatch config path")
	flags.BoolVarP(&noHooks, "no-hooks", "n", false, "Skip Grove lifecycle hooks")
	flags.StringVar(&prURL, "pr", "", "GitHub pull request URL: check out its head branch (--track) and default the prompt to a review")
	command.MarkFlagsMutuallyExclusive("repos", "preset")
	command.MarkFlagsMutuallyExclusive("pr", "branch")
	command.MarkFlagsMutuallyExclusive("pr", "preset")
	command.Version = Version
	return command
}

func Execute() error {
	return newRootCommand(func(opts dispatch.Options, cfg config.Config) error {
		return dispatch.Run(opts, cfg, dispatch.ExecRunner{})
	}).Execute()
}

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() > 0 {
		return exitErr.ExitCode()
	}
	return 1
}
