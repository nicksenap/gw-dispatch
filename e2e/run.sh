#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM

bin="$tmp/bin"
mkdir -p "$bin" "$tmp/workspace"

cat > "$bin/gw" <<'SCRIPT'
#!/bin/sh
set -eu
printf '%s\n' "$@" > "$CREATE_LOG"
if [ "${GW_EXIT:-0}" -ne 0 ]; then
  exit "$GW_EXIT"
fi
printf '[{"name":"feat-e2e","path":"%s"}]\n' "$WORKSPACE_PATH" > "$GROVE_STATE"
SCRIPT

cat > "$bin/pi" <<'SCRIPT'
#!/bin/sh
set -eu
{
  printf '%s\n' "$PWD"
  printf '%s\n' "$#"
  printf '%s\n' "$1"
} > "$AGENT_LOG"
SCRIPT

chmod +x "$bin/gw" "$bin/pi"
go build -o "$bin/gw-dispatch" "$root"

export PATH="$bin:$PATH"
export GROVE_STATE="$tmp/state.json"
export WORKSPACE_PATH="$tmp/workspace"
export CREATE_LOG="$tmp/create.log"
export AGENT_LOG="$tmp/agent.log"
printf '[]\n' > "$GROVE_STATE"

prompt='fix "quotes"; echo not-evaluated'
"$bin/gw-dispatch" --branch feat/e2e --repos api,web --prompt "$prompt"

cat > "$tmp/want-create.log" <<'EOF'
create
--branch
feat/e2e
--repos
api,web
EOF

diff -u "$tmp/want-create.log" "$CREATE_LOG"
{
  printf '%s\n' "$WORKSPACE_PATH"
  printf '1\n'
  printf '%s\n' "$prompt"
} > "$tmp/want-agent.log"
diff -u "$tmp/want-agent.log" "$AGENT_LOG"

rm "$AGENT_LOG"
set +e
GW_EXIT=23 "$bin/gw-dispatch" --branch feat/e2e --repos api --prompt "should not run" >/dev/null 2>&1
status=$?
set -e
if [ "$status" -ne 23 ]; then
  printf 'expected child exit 23, got %s\n' "$status" >&2
  exit 1
fi
if [ -e "$AGENT_LOG" ]; then
  printf 'agent started after failed workspace creation\n' >&2
  exit 1
fi

printf 'gw-dispatch e2e: ok\n'
