#!/usr/bin/env bash
set -euo pipefail

PURGE=false

# Parse flags
while [[ $# -gt 0 ]]; do
  case $1 in
    --purge)
      PURGE=true
      shift
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--purge]"
      echo "  --purge: Remove config directory (~/.config/jira-mgmt/)"
      exit 1
      ;;
  esac
done

echo "=== jira-mgmt Deinit ==="

# 1. Remove binary symlink
if [[ -L "$HOME/.local/bin/jira-mgmt" ]]; then
  echo "Removing binary symlink..."
  rm "$HOME/.local/bin/jira-mgmt"
else
  echo "Binary symlink not found (already removed or never created)"
fi

# 2. Remove Claude Code skill symlink
if [[ -L "$HOME/.claude/skills/jira-management" ]]; then
  echo "Removing Claude Code skill symlink..."
  rm "$HOME/.claude/skills/jira-management"
else
  echo "Claude Code skill symlink not found"
fi

# 3. Remove Codex CLI skill symlink
if [[ -L "$HOME/.codex/skills/jira-management" ]]; then
  echo "Removing Codex CLI skill symlink..."
  rm "$HOME/.codex/skills/jira-management"
else
  echo "Codex CLI skill symlink not found"
fi

# 4. Optionally remove config
if [[ "$PURGE" == "true" ]]; then
  if [[ -d "$HOME/.config/jira-mgmt" ]]; then
    echo "Removing config directory..."
    rm -rf "$HOME/.config/jira-mgmt"
  else
    echo "Config directory not found"
  fi
else
  echo ""
  echo "Config directory preserved: ~/.config/jira-mgmt/"
  echo "To remove config, run: $0 --purge"
fi

echo ""
echo "Deinit complete!"
