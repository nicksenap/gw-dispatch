# Changelog

## v0.3.0

### Added

- `--pr <github-pr-url>`: resolve the PR's head branch and repository through `gh` and `gw repos --json`, create the workspace with `--track` and `--source-*` provenance, and start the agent with a review prompt (or your `--prompt`) plus the PR context. `--repos` may add sibling repos; `--branch` and `--preset` are rejected with `--pr`.

## v0.2.0

### Added

- `-P` shorthand for the required `--prompt` flag.
- `-n`/`--no-hooks` forwarding to workspace creation.
- Deterministic prompt-derived branches when `--branch` is omitted.
- Dispatch-local `default_preset` and `default_repos` configuration.

## v0.1.0

### Added

- `gw dispatch` workspace creation followed by an interactive coding-agent launch.
- Built-in Pi, Claude Code, Codex, and OpenCode support.
- Configurable default and custom argv-based agents through `dispatch.toml`.
