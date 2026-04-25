#!/usr/bin/env bash
# OpenCode - tool configuration (Tier 2: Bun/TS plugin shim)
# https://opencode.ai
#
# Registers the @planifest/opencode-hooks plugin in opencode.json.
# Requires: Bun runtime (bundled with OpenCode).

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
OPENCODE_JSON="$PROJECT_ROOT/opencode.json"
PLUGIN_SRC="$SCRIPT_DIR/hooks/adapters/opencode"
PLUGIN_DEST="$PROJECT_ROOT/.opencode/plugins/@planifest/opencode-hooks"

echo ""
echo "  Setting up OpenCode (Tier 2: plugin shim)"

# Verify Bun is available
if ! command -v bun &>/dev/null; then
  echo "  ! Warning: bun not found in PATH. The @planifest/opencode-hooks plugin"
  echo "    requires Bun (bundled with OpenCode). Ensure opencode is installed and"
  echo "    bun is in your PATH before using enforcement hooks."
fi

# Copy plugin source
mkdir -p "$PLUGIN_DEST"
cp "$PLUGIN_SRC/index.ts" "$PLUGIN_DEST/index.ts"
cp "$PLUGIN_SRC/package.json" "$PLUGIN_DEST/package.json"
echo "  + .opencode/plugins/@planifest/opencode-hooks/"

# Copy shared hook scripts for the plugin to delegate to
HOOKS_DEST="$PROJECT_ROOT/.opencode/hooks"
mkdir -p "$HOOKS_DEST/enforcement" "$HOOKS_DEST/telemetry"

for script in "$SCRIPT_DIR/hooks/enforcement/"*.mjs; do
  [ -f "$script" ] || continue
  cp "$script" "$HOOKS_DEST/enforcement/$(basename "$script")"
  echo "  + .opencode/hooks/enforcement/$(basename "$script")"
done

for script in "$SCRIPT_DIR/hooks/telemetry/emit-phase-"*.mjs; do
  [ -f "$script" ] || continue
  cp "$script" "$HOOKS_DEST/telemetry/$(basename "$script")"
  echo "  + .opencode/hooks/telemetry/$(basename "$script")"
done

# Register plugin in opencode.json (idempotent)
PLUGIN_REF=".opencode/plugins/@planifest/opencode-hooks/index.ts"

if [ -f "$OPENCODE_JSON" ]; then
  # Check if already registered
  if jq -e --arg ref "$PLUGIN_REF" '.plugins | index($ref)' "$OPENCODE_JSON" > /dev/null 2>&1; then
    echo "  - opencode.json (plugin already registered)"
  else
    local merged
    merged=$(jq --arg ref "$PLUGIN_REF" '.plugins //= [] | .plugins += [$ref]' "$OPENCODE_JSON")
    printf '%s\n' "$merged" > "$OPENCODE_JSON"
    echo "  ~ opencode.json (plugin registered)"
  fi
else
  jq -n --arg ref "$PLUGIN_REF" '{"plugins":[$ref]}' > "$OPENCODE_JSON"
  echo "  + opencode.json (created with plugin registration)"
fi

echo "  [Planifest] OpenCode hooks installed."
echo "  Note: Run 'bun build' in .opencode/plugins/@planifest/opencode-hooks/ if"
echo "  OpenCode requires a compiled dist/index.js rather than the TypeScript source."
