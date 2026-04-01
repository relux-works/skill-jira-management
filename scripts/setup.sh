#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

SKILL_NAME="jira-management"
SKILL_CONTENT_DIR="$PROJECT_ROOT/agents/skills/$SKILL_NAME"

AGENTS_DIR="$HOME/.agents/skills"
CLAUDE_DIR="$HOME/.claude/skills"
CODEX_DIR="$HOME/.codex/skills"
BIN_DIR="$HOME/.local/bin"
INSTALLED_SKILL_DIR="$AGENTS_DIR/$SKILL_NAME"
INSTALLED_BIN="$BIN_DIR/jira-mgmt"

ARTIFACT_DIR="$(mktemp -d "${TMPDIR:-/tmp}/jira-management-setup.XXXXXX")"
ARTIFACT_SKILL_DIR="$ARTIFACT_DIR/skill"
ARTIFACT_BIN="$ARTIFACT_DIR/jira-mgmt"

cleanup() {
  rm -rf "$ARTIFACT_DIR"
}

prune_source_artifacts() {
  local target="$1"

  rm -rf \
    "$target/.git" \
    "$target/.task-board" \
    "$target/.planning" \
    "$target/.research" \
    "$target/.spec" \
    "$target/cmd" \
    "$target/internal" \
    "$target/scripts"

  rm -f \
    "$target/README.md" \
    "$target/CLAUDE.md" \
    "$target/LICENSE" \
    "$target/NOTICE" \
    "$target/go.mod" \
    "$target/go.sum" \
    "$target/jira-mgmt" \
    "$target/task-board.config.json"
}

trap cleanup EXIT

echo "=== $SKILL_NAME Setup ==="

# 1. Build binary artifact
echo "Building jira-mgmt binary..."
cd "$PROJECT_ROOT"
go build -o "$ARTIFACT_BIN" ./cmd/jira-mgmt/

# 2. Stage a degitized skill artifact for global installs
mkdir -p "$ARTIFACT_SKILL_DIR"
rsync -a --delete "$SKILL_CONTENT_DIR/" "$ARTIFACT_SKILL_DIR/" --exclude='.git'
prune_source_artifacts "$ARTIFACT_SKILL_DIR"

# 3. Install binary as a standalone artifact (not a source-linked symlink)
mkdir -p "$BIN_DIR"
rm -f "$INSTALLED_BIN"
install -m 0755 "$ARTIFACT_BIN" "$INSTALLED_BIN"
echo "  Binary copied -> $INSTALLED_BIN"

# 4. Verify PATH
if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
  echo "  WARNING: ~/.local/bin is not in your PATH"
fi

# 5. Copy skill into .agents/skills/ as an artifact copy
echo "Installing skill: $SKILL_NAME"
if [ -L "$INSTALLED_SKILL_DIR" ]; then
  rm -f "$INSTALLED_SKILL_DIR"
fi
mkdir -p "$INSTALLED_SKILL_DIR"
rsync -a --delete "$ARTIFACT_SKILL_DIR/" "$INSTALLED_SKILL_DIR/"
echo "  Installed artifact -> $INSTALLED_SKILL_DIR/"

# 6. Symlink from .claude/skills/ -> .agents/skills/
mkdir -p "$CLAUDE_DIR"
rm -f "$CLAUDE_DIR/$SKILL_NAME"
ln -s "$AGENTS_DIR/$SKILL_NAME" "$CLAUDE_DIR/$SKILL_NAME"
echo "  Symlink -> $CLAUDE_DIR/$SKILL_NAME"

# 7. Symlink from .codex/skills/ -> .agents/skills/
mkdir -p "$CODEX_DIR"
rm -f "$CODEX_DIR/$SKILL_NAME"
ln -s "$AGENTS_DIR/$SKILL_NAME" "$CODEX_DIR/$SKILL_NAME"
echo "  Symlink -> $CODEX_DIR/$SKILL_NAME"

echo ""
echo "Done. Installed $(git -C "$PROJECT_ROOT" describe --tags --always 2>/dev/null || echo 'unknown')"
echo ""
echo "Next steps:"
echo "  jira-mgmt auth"
echo "  jira-mgmt config set project YOUR-KEY"
