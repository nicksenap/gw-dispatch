# Spec: gw-dispatch

## Objective

Provide an agent-agnostic Grove plugin that creates a workspace and starts an interactive coding-agent session in that workspace with an initial prompt.

## CLI contract

```text
gw dispatch [--branch <branch>] [--repos <repos> | --preset <preset>] --prompt <prompt> [--no-hooks] [--agent <name>] [--config <path>]
```

Short flags `-b`, `-r`, `-p`, and `-n` mirror `gw create`; `-P` is shorthand for `--prompt`. `-n`/`--no-hooks` is forwarded to the `gw create` subprocess. When `--branch` is omitted, dispatch normalizes the prompt, builds a capped readable slug, and appends the first eight hexadecimal characters of its SHA-256 hash under the `dispatch/` prefix. Pi is the default agent. Built-in agent names are `pi`, `claude`, `codex`, and `opencode`.

## Configuration

The default path is `$GROVE_DIR/dispatch.toml`, falling back to `~/.grove/dispatch.toml` when `GROVE_DIR` is unset.

```toml
default_agent = "pi"
default_preset = "backend"
# Alternatively: default_repos = "api,web"

[agents.aider]
command = ["aider", "--message", "{prompt}"]
```

`default_preset` and `default_repos` belong to dispatch config, are mutually exclusive, and are overridden by an explicit selector flag. Commands are argv arrays and `gw-dispatch` never adds an implicit shell. Every custom command must include `{prompt}` in at least one argument. Built-ins may be overridden by a config entry with the same name. Configuration is trusted executable policy: users can explicitly opt into a shell or interpreter, and must not do so with untrusted prompts.

## Execution

1. Load and validate configuration.
2. Resolve the selected agent and verify both `gw` and the agent executable are on `PATH`.
3. Derive a deterministic branch from the prompt when no branch is supplied.
4. Resolve repo selection from explicit flags or the dispatch config default, then run `gw create`.
5. Derive the workspace name using Grove's branch-to-name rule and read its path from `$GROVE_STATE` or `$GROVE_DIR/state.json`.
6. Start the selected agent in that directory, inheriting stdin, stdout, stderr, and environment.
7. Propagate create or agent failures. If agent launch fails after creation, report the retained workspace path.

## Built-in commands

- Pi: `pi "{prompt}"`
- Claude Code: `claude "{prompt}"`
- Codex: `codex "{prompt}"`
- OpenCode: `opencode --prompt "{prompt}"`

## Testing strategy

- Unit tests for configuration, placeholder expansion, workspace-name derivation, state lookup, and argument construction.
- Workflow tests with a fake process runner to prove create-before-agent ordering, working directory selection, and failure handling.
- CLI tests for required and mutually exclusive flags.
- Full `go test`, `go vet`, and build verification.

## Boundaries

- Always: invoke commands directly without a shell; validate config before creating a workspace; preserve terminal streams.
- Ask first: background execution, lifecycle/session management, or additional Grove create flags.
- Never: add implicit shell evaluation, interpolate prompts into launcher-built shell strings, or silently delete a workspace after an agent launch failure.

## Success criteria

- The documented Pi, Claude, Codex, and OpenCode invocations resolve correctly.
- `--agent` overrides the configured default.
- A user-defined agent works through `dispatch.toml`.
- Exactly one repo selector resolves from explicit `--repos`/`--preset` or dispatch config defaults.
- Failed workspace creation never starts an agent.
- Agent launch happens from the created workspace root.
