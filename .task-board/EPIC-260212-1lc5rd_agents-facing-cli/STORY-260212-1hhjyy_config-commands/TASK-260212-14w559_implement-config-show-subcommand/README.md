# TASK-260212-14w559: Implement 'config show' subcommand

## Description
Implement 'jira-mgmt config show' subcommand. Display current configuration: instance URL (without token), default project, default board, locale. Format as human-readable text or JSON (respect --format flag).

## Scope
(define task scope)

## Acceptance Criteria
- Displays Jira instance URL (without token)
- Displays default project
- Displays default board
- Displays locale setting
- Output respects --format flag (text/json)
- Handles missing config gracefully
