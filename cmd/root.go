package cmd

import (
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

	command := &cobra.Command{
		Use:   "gw-dispatch",
		Short: "Create a Grove workspace and dispatch a coding agent",
		Long: `Create a Grove workspace, then start an interactive coding agent in it
with the supplied prompt. Pi is used by default; Claude Code, Codex, OpenCode,
and custom configured agents are also supported.`,
		Example: `  gw dispatch -b feat/login -r api,web --prompt "Implement login"
  gw dispatch -b feat/login -p backend --agent claude --prompt "Implement login"`,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			return run(dispatch.Options{
				Branch: branch,
				Repos:  repos,
				Preset: preset,
				Prompt: prompt,
				Agent:  agent,
			}, cfg)
		},
	}

	flags := command.Flags()
	flags.StringVarP(&branch, "branch", "b", "", "Branch name")
	flags.StringVarP(&repos, "repos", "r", "", "Comma-separated repo names")
	flags.StringVarP(&preset, "preset", "p", "", "Use named preset")
	flags.StringVar(&prompt, "prompt", "", "Initial prompt sent to the agent")
	flags.StringVar(&agent, "agent", "", "Agent name (default: configured agent or pi)")
	flags.StringVar(&configPath, "config", config.DefaultPath(), "Dispatch config path")
	_ = command.MarkFlagRequired("branch")
	_ = command.MarkFlagRequired("prompt")
	command.MarkFlagsOneRequired("repos", "preset")
	command.MarkFlagsMutuallyExclusive("repos", "preset")
	command.Version = Version
	return command
}

func Execute() error {
	return newRootCommand(func(opts dispatch.Options, cfg config.Config) error {
		return dispatch.Run(opts, cfg, dispatch.ExecRunner{})
	}).Execute()
}
