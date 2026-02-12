# TASK-260212-rrrc8o: Create symlinks script

## Description
Create script to symlink agents/skills/jira-management/ to .claude/skills/ and .codex/skills/

## Scope
(define task scope)

## Acceptance Criteria
- Script creates .claude/skills/ directory if missing
- Script creates .codex/skills/ directory if missing
- Script creates symlink .claude/skills/jira-management → agents/skills/jira-management/
- Script creates symlink .codex/skills/jira-management → agents/skills/jira-management/
- Script is idempotent (can run multiple times safely)
- Script reports success/failure clearly
