# gw-dispatch

Agent-agnostic [Grove](https://github.com/nicksenap/grove) plugin that creates a workspace and starts a coding agent there with an initial prompt.

```bash
gw dispatch -b feat/login -r api,web --prompt "Implement login"
gw dispatch -b feat/login -p backend --agent claude --prompt "Implement login"
```

Pi is the default. Claude Code, Codex, OpenCode, and user-defined agents are supported.

## Install

```bash
gw plugin install nicksenap/gw-dispatch
```

Or build from source:

```bash
go install github.com/nicksenap/gw-dispatch@latest
```

Requirements:

- macOS or Linux
- `gw` on `PATH`
- The selected coding-agent executable on `PATH`

## Usage

```text
gw dispatch --branch <branch> (--repos <repos> | --preset <preset>) --prompt <prompt> [flags]
```

| Flag | Short | Description |
|---|---:|---|
| `--branch` | `-b` | Branch passed to `gw create` |
| `--repos` | `-r` | Comma-separated repositories |
| `--preset` | `-p` | Grove preset; mutually exclusive with `--repos` |
| `--prompt` | | Initial agent prompt |
| `--agent` | | Override the configured/default agent |
| `--config` | | Override the dispatch config path |

The workspace name follows Grove's normal branch-derived naming. For example, `feat/login` creates and opens `feat-login`.

## Built-in agents

| Name | Invocation |
|---|---|
| `pi` | `pi "{prompt}"` |
| `claude` | `claude "{prompt}"` |
| `codex` | `codex "{prompt}"` |
| `opencode` | `opencode --prompt "{prompt}"` |

These interactive initial-prompt forms follow the current CLI references for [Pi](https://github.com/earendil-works/pi/tree/main/packages/coding-agent), [Claude Code](https://code.claude.com/docs/en/cli-reference), [Codex](https://developers.openai.com/codex/cli/reference), and [OpenCode](https://opencode.ai/docs/cli/).

## Configuration

Configuration is optional. The default path is `~/.grove/dispatch.toml` (or `$GROVE_DIR/dispatch.toml`).

```toml
default_agent = "pi"

[agents.aider]
command = ["aider", "--message", "{prompt}"]

[agents.my-agent]
command = ["my-agent", "start", "--task={prompt}"]
```

Then run:

```bash
gw dispatch -b feat/task -p backend --agent aider --prompt "Implement the task"
```

### Custom-command rules

- `command` is an argv array; `gw-dispatch` never adds an implicit shell.
- At least one argument must contain `{prompt}`.
- `{prompt}` cannot appear in the executable (`command[0]`).
- Built-in names can be overridden in config.
- Configuration is trusted executable policy. A command such as `["sh", "-c", "{prompt}"]` explicitly opts into shell evaluation and is unsafe for untrusted prompts.
- Put `{prompt}` in a data-bearing argument (for example, after `--message`) rather than where an agent could parse it as another control flag.

## Behavior and failures

`gw-dispatch` validates and resolves the agent executable before creating anything, runs `gw create`, reads the resulting workspace path from Grove state, and starts that exact executable with the workspace as its working directory. The agent inherits the current terminal and environment and remains interactive. Use the selected agent's normal project-trust, sandbox, approval, and tool controls when dispatching into untrusted repositories.

If workspace creation fails, no agent starts. If the agent exits unsuccessfully, the workspace is retained and its path is included in the error.

## Development

```bash
go test ./...
go vet ./...
go build -o gw-dispatch .
./e2e/run.sh
```

## License

MIT
