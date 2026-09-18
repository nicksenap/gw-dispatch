# gw-dispatch

Agent-agnostic [Grove](https://github.com/nicksenap/grove) plugin that creates a workspace and starts a coding agent there with an initial prompt.

```bash
gw dispatch -n -r api,web -P "Implement login"
gw dispatch -b feat/login -p backend --agent claude -P "Implement login"
gw dispatch --pr https://github.com/acme/api/pull/42
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
gw dispatch --prompt <prompt> [flags]
```

| Flag | Short | Description |
|---|---:|---|
| `--branch` | `-b` | Branch passed to `gw create` (default: derived from prompt) |
| `--repos` | `-r` | Comma-separated repositories; overrides dispatch config default |
| `--preset` | `-p` | Grove preset; overrides dispatch config default |
| `--prompt` | `-P` | Initial agent prompt |
| `--no-hooks` | `-n` | Pass `--no-hooks` to `gw create` |
| `--pr` | | GitHub pull request URL; see [Reviewing a pull request](#reviewing-a-pull-request) |
| `--agent` | | Override the configured/default agent |
| `--config` | | Override the dispatch config path |

When `--branch` is omitted, the prompt deterministically produces `dispatch/<slug>-<8-char-hash>` using only the standard library. For example, `Fix login redirect` produces `dispatch/fix-login-redirect-98488061`. An explicit branch still follows Grove's normal branch-derived workspace naming, so `feat/login` creates and opens `feat-login`.

## Reviewing a pull request

```bash
gw dispatch --pr https://github.com/acme/api/pull/42
gw dispatch --pr https://github.com/acme/api/pull/42 -r web -P "Fix the failing tests"
```

`--pr` turns a GitHub PR URL into a workspace on the PR's head branch:

1. `gh pr view` resolves the head branch, title, and description (requires the [gh CLI](https://cli.github.com), authenticated).
2. The PR's `owner/repo` is matched against `gw repos --json`; the matching Grove repo becomes the workspace's primary repo. `--repos` may add sibling repos, which get fresh branches from their base.
3. `gw create --branch <head> --track --source-url ... --source-provider github --source-ref <N> --source-title ...` checks the existing remote branch out and records the PR as the workspace source.
4. The agent starts with your `--prompt`, or a default review prompt when omitted. Either way the PR URL, title, head branch, and description are appended.

`--pr` cannot be combined with `--branch` or `--preset`. PRs from forks are rejected because their head branch is not on `origin`.

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
default_preset = "backend"
# Alternatively: default_repos = "api,web"

[agents.aider]
command = ["aider", "--message", "{prompt}"]

[agents.my-agent]
command = ["my-agent", "start", "--task={prompt}"]
```

`default_preset` and `default_repos` are mutually exclusive. Set either one to omit `-p`/`-r` from normal invocations; an explicit selector flag overrides the dispatch default.

Then run:

```bash
gw dispatch -n --agent aider -P "Implement the task"
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
