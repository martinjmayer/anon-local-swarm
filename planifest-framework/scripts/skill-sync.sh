#!/usr/bin/env bash
set -euo pipefail
# Planifest skill-sync — install, remove, and sync external skills (REQ-024, REQ-025)
#
# Usage:
#   skill-sync.sh add     <skill-name> <tool> [--from <url>] [--authorized]
#   skill-sync.sh install <skill-name> <tool>
#   skill-sync.sh remove  <skill-name> <tool>
#   skill-sync.sh sync    <tool>
#
# <tool> is optional on all commands. When omitted, the active tool is
# auto-detected from the first recognised skills directory found.
#
# Note: manifest is stored as JSON (external-skills.json) for native node
# compatibility — the .yml name in the spec is a display convention.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
FRAMEWORK_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PROJECT_ROOT="$(cd "$FRAMEWORK_DIR/.." && pwd)"
SETUP_DIR="$FRAMEWORK_DIR/setup"

MANIFEST="$FRAMEWORK_DIR/external-skills.json"
PLAN_SKILLS_DIR="$PROJECT_ROOT/plan/current/external-skills"
PRESERVED_SKILLS_DIR="$FRAMEWORK_DIR/external-skills"
ANTHROPIC_RAW_BASE="https://raw.githubusercontent.com/anthropics/skills/main/skills"

# ── Helpers ────────────────────────────────────────────────────────────────

die() { echo "  [skill-sync] Error: $*" >&2; exit 1; }
info() { echo "  [skill-sync] $*"; }

resolve_tool_skills_dir() {
  local tool="$1"
  local config="$SETUP_DIR/${tool}.sh"
  [ -f "$config" ] || die "Unknown tool '$tool'. Run: setup.sh <tool> to see valid tools."
  # Source only TOOL_SKILLS_DIR from the config (safe — we grep it out)
  local skills_rel
  skills_rel=$(grep '^TOOL_SKILLS_DIR=' "$config" | head -1 | cut -d= -f2 | tr -d '"')
  [ -n "$skills_rel" ] || die "TOOL_SKILLS_DIR not set in $config"
  echo "$PROJECT_ROOT/$skills_rel"
}

detect_tool() {
  # Returns the first tool whose skills directory exists under PROJECT_ROOT
  for tool_config in "$SETUP_DIR"/*.sh; do
    [ -f "$tool_config" ] || continue
    local tool; tool="$(basename "$tool_config" .sh)"
    local skills_rel
    skills_rel=$(grep '^TOOL_SKILLS_DIR=' "$tool_config" 2>/dev/null | head -1 | cut -d= -f2 | tr -d '"')
    [ -n "$skills_rel" ] || continue
    if [ -d "$PROJECT_ROOT/$skills_rel" ]; then
      echo "$tool"
      return 0
    fi
  done
  die "No tool skills directory detected. Run setup.sh <tool> first."
}

ensure_manifest() {
  if [ ! -f "$MANIFEST" ]; then
    printf '{"skills":[]}\n' > "$MANIFEST"
  fi
}

skill_in_manifest() {
  local name="$1"
  node -e "
    const fs = require('fs');
    const m = JSON.parse(fs.readFileSync('$MANIFEST','utf8'));
    process.exit(m.skills.some(s => s.name === '$name') ? 0 : 1);
  " 2>/dev/null
}

get_skill_scope() {
  local name="$1"
  node -e "
    const fs = require('fs');
    const m = JSON.parse(fs.readFileSync('$MANIFEST','utf8'));
    const s = m.skills.find(s => s.name === '$name');
    if (s) process.stdout.write(s.scope);
  " 2>/dev/null
}

add_to_manifest() {
  local name="$1" source="$2" trusted="$3" scope="$4" feature_id="${5:-}"
  ensure_manifest
  SKILL_NAME="$name" SKILL_SOURCE="$source" SKILL_TRUSTED="$trusted" \
  SKILL_SCOPE="$scope" SKILL_FEATURE="$feature_id" SKILL_MANIFEST="$MANIFEST" \
  node -e "
    const fs = require('fs');
    const m = JSON.parse(fs.readFileSync(process.env.SKILL_MANIFEST,'utf8').replace(/^\uFEFF/,''));
    m.skills = m.skills.filter(s => s.name !== process.env.SKILL_NAME);
    const entry = {
      name:        process.env.SKILL_NAME,
      source:      process.env.SKILL_SOURCE,
      trusted:     process.env.SKILL_TRUSTED === 'true',
      installedAt: new Date().toISOString().slice(0,10),
      scope:       process.env.SKILL_SCOPE,
    };
    if (process.env.SKILL_FEATURE) entry.featureId = process.env.SKILL_FEATURE;
    m.skills.push(entry);
    fs.writeFileSync(process.env.SKILL_MANIFEST, JSON.stringify(m,null,2)+'\n');
  "
}

remove_from_manifest() {
  local name="$1"
  [ -f "$MANIFEST" ] || return 0
  SKILL_NAME="$name" SKILL_MANIFEST="$MANIFEST" node -e "
    const fs = require('fs');
    const m = JSON.parse(fs.readFileSync(process.env.SKILL_MANIFEST,'utf8').replace(/^\uFEFF/,''));
    m.skills = m.skills.filter(s => s.name !== process.env.SKILL_NAME);
    if (m.skills.length === 0) { fs.unlinkSync(process.env.SKILL_MANIFEST); }
    else fs.writeFileSync(process.env.SKILL_MANIFEST, JSON.stringify(m,null,2)+'\n');
  "
}

update_scope() {
  local name="$1" scope="$2"
  SKILL_NAME="$name" SKILL_SCOPE="$scope" SKILL_MANIFEST="$MANIFEST" node -e "
    const fs = require('fs');
    const m = JSON.parse(fs.readFileSync(process.env.SKILL_MANIFEST,'utf8').replace(/^\uFEFF/,''));
    const s = m.skills.find(s => s.name === process.env.SKILL_NAME);
    if (s) s.scope = process.env.SKILL_SCOPE;
    fs.writeFileSync(process.env.SKILL_MANIFEST, JSON.stringify(m,null,2)+'\n');
  "
}

# ── Operations ─────────────────────────────────────────────────────────────

cmd_install() {
  local name="$1" tool="$2"
  local tool_skills_dir; tool_skills_dir="$(resolve_tool_skills_dir "$tool")"
  local dest="$tool_skills_dir/$name"

  # Find skill source dir (preserved takes priority over plan-scoped)
  local src=""
  if [ -d "$PRESERVED_SKILLS_DIR/$name" ]; then
    src="$PRESERVED_SKILLS_DIR/$name"
  elif [ -d "$PLAN_SKILLS_DIR/$name" ]; then
    src="$PLAN_SKILLS_DIR/$name"
  else
    die "Skill '$name' not found in either storage tier. Run: skill-sync.sh add $name $tool"
  fi

  mkdir -p "$dest"
  cp -r "$src"/. "$dest/"
  info "Installed: $name → $tool_skills_dir/$name"
}

cmd_remove() {
  local name="$1" tool="$2"
  local tool_skills_dir; tool_skills_dir="$(resolve_tool_skills_dir "$tool")"
  local dest="$tool_skills_dir/$name"

  if [ -d "$dest" ]; then
    rm -rf "$dest"
    info "Removed from tool skills: $name"
  else
    info "Warning: $name not found in $tool_skills_dir — skipping"
  fi
}

cmd_sync() {
  local tool="$1"
  [ -f "$MANIFEST" ] || { info "No manifest found — nothing to sync."; return 0; }

  local names
  names=$(node -e "
    const fs = require('fs');
    const m = JSON.parse(fs.readFileSync('$MANIFEST','utf8').replace(/^\uFEFF/,''));
    m.skills.forEach(s => console.log(s.name));
  " 2>/dev/null || true)

  if [ -z "$names" ]; then
    info "Manifest is empty — nothing to sync."
    return 0
  fi

  while IFS= read -r name; do
    [ -n "$name" ] || continue
    # Re-fetch plan-scoped skills if their files are missing (e.g. after git clone)
    local scope; scope="$(get_skill_scope "$name")"
    if [ "$scope" = "plan" ] && [ ! -d "$PLAN_SKILLS_DIR/$name" ]; then
      local source
      source=$(node -e "
        const fs = require('fs');
        const m = JSON.parse(fs.readFileSync('$MANIFEST','utf8').replace(/^\uFEFF/,''));
        const s = m.skills.find(s => s.name === '$name');
        if (s) process.stdout.write(s.source);
      " 2>/dev/null || true)
      if [ -n "$source" ]; then
        info "Re-fetching plan-scoped skill: $name"
        _fetch_skill "$name" "$source" "plan"
      else
        info "Warning: Cannot re-fetch $name — source not in manifest. Skipping."
        continue
      fi
    fi
    cmd_install "$name" "$tool"
  done <<< "$names"
}

_fetch_skill() {
  local name="$1" source_url="$2" scope="$3"
  local dest_dir
  if [ "$scope" = "preserved" ]; then
    dest_dir="$PRESERVED_SKILLS_DIR/$name"
  else
    dest_dir="$PLAN_SKILLS_DIR/$name"
  fi

  mkdir -p "$dest_dir"

  # Fetch SKILL.md
  local raw_url="$ANTHROPIC_RAW_BASE/$name/SKILL.md"
  # If a custom --from URL was provided, derive raw URL from it
  if [[ "$source_url" != *"raw.githubusercontent.com"* && "$source_url" != *"anthropics/skills"* ]]; then
    raw_url="$source_url/SKILL.md"
  fi

  if curl -fsSL "$raw_url" -o "$dest_dir/SKILL.md" 2>/dev/null; then
    info "Fetched: $raw_url"
  else
    rm -rf "$dest_dir"
    die "Failed to fetch skill '$name' from $raw_url — check the skill name and your network connection."
  fi
}

cmd_add() {
  local name="$1" tool="$2" from_url="" authorized=false

  # Parse flags
  shift 2
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --from)       from_url="$2"; shift 2 ;;
      --authorized) authorized=true; shift ;;
      *)            die "Unknown flag: $1" ;;
    esac
  done

  # Already installed?
  if skill_in_manifest "$name" 2>/dev/null; then
    info "Warning: '$name' is already installed — skipping."
    return 0
  fi

  local trusted=false source_url=""
  if [ -z "$from_url" ]; then
    # Default: Anthropic repo — trusted, no approval needed
    trusted=true
    source_url="https://github.com/anthropics/skills/tree/main/skills/$name"
  else
    # Non-Anthropic source — require --authorized flag
    source_url="$from_url"
    if [ "$authorized" != "true" ]; then
      die "Non-Anthropic source requires human approval.
  Source: $from_url
  Skill:  $name
  The agent must present this URL to the human and receive confirmation before installing.
  Re-run with --authorized once confirmed."
    fi
  fi

  # Get current feature ID from plan/current if available
  local feature_id=""
  local feature_id_file="$PROJECT_ROOT/plan/current/.feature-id"
  if [ -f "$feature_id_file" ]; then
    feature_id=$(cat "$feature_id_file")
  fi

  info "Fetching skill: $name"
  _fetch_skill "$name" "$source_url" "plan"

  ensure_manifest
  add_to_manifest "$name" "$source_url" "$trusted" "plan" "$feature_id"
  info "Recorded in manifest: $MANIFEST"

  cmd_install "$name" "$tool"
  info "Skill '$name' installed for tool '$tool'."
}

cmd_preserve() {
  local name="$1" tool="$2"
  skill_in_manifest "$name" || die "Skill '$name' not found in manifest."

  local scope; scope="$(get_skill_scope "$name")"
  if [ "$scope" = "preserved" ]; then
    info "'$name' is already preserved."
    return 0
  fi

  # Move from plan-scoped to preserved tier
  if [ -d "$PLAN_SKILLS_DIR/$name" ]; then
    mkdir -p "$PRESERVED_SKILLS_DIR"
    cp -r "$PLAN_SKILLS_DIR/$name" "$PRESERVED_SKILLS_DIR/$name"
    rm -rf "$PLAN_SKILLS_DIR/$name"
    info "Moved '$name' to preserved tier: $PRESERVED_SKILLS_DIR/$name"
  fi

  update_scope "$name" "preserved"
  info "'$name' is now preserved — will survive P7 archive."
}

cmd_unpreserve() {
  local name="$1" tool="$2"
  skill_in_manifest "$name" || die "Skill '$name' not found in manifest."

  local scope; scope="$(get_skill_scope "$name")"
  if [ "$scope" = "plan" ]; then
    info "'$name' is already plan-scoped."
    return 0
  fi

  # Move from preserved to plan-scoped tier
  if [ -d "$PRESERVED_SKILLS_DIR/$name" ]; then
    mkdir -p "$PLAN_SKILLS_DIR"
    cp -r "$PRESERVED_SKILLS_DIR/$name" "$PLAN_SKILLS_DIR/$name"
    rm -rf "$PRESERVED_SKILLS_DIR/$name"
    info "Moved '$name' to plan-scoped tier."
  fi

  update_scope "$name" "plan"
  info "'$name' will be removed at P7 archive."
}

# ── Main ───────────────────────────────────────────────────────────────────

OPERATION="${1:-}"
[ -n "$OPERATION" ] || { echo "Usage: skill-sync.sh <add|install|remove|sync|preserve|unpreserve> <skill> [tool]"; exit 1; }
shift

SKILL_NAME="${1:-}"
TOOL="${2:-}"

# Auto-detect tool if not provided (not needed for sync which takes tool as $1)
if [ "$OPERATION" = "sync" ]; then
  TOOL="${1:-}"
  [ -n "$TOOL" ] || TOOL="$(detect_tool)"
  cmd_sync "$TOOL"
  exit 0
fi

[ -n "$SKILL_NAME" ] || die "Skill name required."
[ -n "$TOOL" ] || TOOL="$(detect_tool)"

case "$OPERATION" in
  add)        cmd_add        "$SKILL_NAME" "$TOOL" "${@:3}" ;;
  install)    cmd_install    "$SKILL_NAME" "$TOOL" ;;
  remove)     cmd_remove     "$SKILL_NAME" "$TOOL" ;;
  preserve)   cmd_preserve   "$SKILL_NAME" "$TOOL" ;;
  unpreserve) cmd_unpreserve "$SKILL_NAME" "$TOOL" ;;
  *) die "Unknown operation '$OPERATION'. Valid: add install remove sync preserve unpreserve" ;;
esac
