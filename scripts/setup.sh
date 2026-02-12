#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=== jira-mgmt Setup ==="

# 1. Build binary
echo "Building jira-mgmt binary..."
cd "$PROJECT_ROOT"
go build -o jira-mgmt ./cmd/jira-mgmt/

# 2. Create ~/.local/bin if it doesn't exist
mkdir -p "$HOME/.local/bin"

# 3. Symlink binary to ~/.local/bin
echo "Creating symlink to ~/.local/bin/jira-mgmt..."
ln -sf "$PROJECT_ROOT/jira-mgmt" "$HOME/.local/bin/jira-mgmt"

# 4. Verify PATH
if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
  echo ""
  echo "WARNING: ~/.local/bin is not in your PATH"
  echo "Add this to your ~/.zshrc:"
  echo '  export PATH="$HOME/.local/bin:$PATH"'
  echo ""
fi

# 5. Create skill symlinks
echo "Creating skill symlinks..."

# Claude Code symlink
mkdir -p "$HOME/.claude/skills"
ln -sf "$PROJECT_ROOT/agents/skills/jira-management" "$HOME/.claude/skills/jira-management"

# Codex CLI symlink
mkdir -p "$HOME/.codex/skills"
ln -sf "$PROJECT_ROOT/agents/skills/jira-management" "$HOME/.codex/skills/jira-management"

echo ""
echo "Setup complete!"
echo "Binary: $HOME/.local/bin/jira-mgmt"
echo "Skill (Claude): ~/.claude/skills/jira-management"
echo "Skill (Codex): ~/.codex/skills/jira-management"
echo ""
echo "Next steps:"
echo "  1. Ensure ~/.local/bin is in your PATH"
echo "  2. Run: jira-mgmt auth"
echo "  3. Configure: jira-mgmt config set project YOUR-KEY"
