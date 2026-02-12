# TASK-260212-1di1gt: setup-deinit-global

## Description
Setup script: build jira-mgmt binary and symlink to ~/.local/bin/jira-mgmt, create skill symlinks to ~/.claude/skills/ and ~/.codex/skills/. Deinit script: remove binary symlink, remove skill symlinks, optionally clean config (~/.config/jira-mgmt/)

## Scope
(define task scope)

## Acceptance Criteria
- scripts/setup.sh: go build cmd/jira-mgmt → symlink to ~/.local/bin/jira-mgmt\n- scripts/setup.sh: symlink agents/skills/jira-management → ~/.claude/skills/jira-management\n- scripts/setup.sh: symlink agents/skills/jira-management → ~/.codex/skills/jira-management\n- scripts/setup.sh: verify PATH includes ~/.local/bin\n- scripts/deinit.sh: remove ~/.local/bin/jira-mgmt symlink\n- scripts/deinit.sh: remove ~/.claude/skills/jira-management symlink\n- scripts/deinit.sh: remove ~/.codex/skills/jira-management symlink\n- scripts/deinit.sh: optionally remove ~/.config/jira-mgmt/ (prompt or --purge flag)\n- Both scripts idempotent (safe to run multiple times)
