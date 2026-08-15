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
branch=
next_is_branch=false
for arg in "$@"; do
  if [ "$next_is_branch" = true ]; then
    branch=$arg
    break
  fi
  if [ "$arg" = "--branch" ]; then
    next_is_branch=true
  fi
done
name=$(printf '%s' "$branch" | sed 's#[ /]#-#g')
printf '[{"name":"%s","path":"%s"}]\n' "$name" "$WORKSPACE_PATH" > "$GROVE_STATE"
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
printf 'default_repos = "api,web"\n' > "$tmp/dispatch.toml"

prompt='fix "quotes"; echo not-evaluated'
"$bin/gw-dispatch" -P "$prompt" -n --config "$tmp/dispatch.toml"

cat > "$tmp/want-create.log" <<'EOF'
create
--branch
dispatch/fix-quotes-echo-not-evaluated-fd0ad53d
--repos
api,web
--no-hooks
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
