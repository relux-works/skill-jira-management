# TASK-260212-1reomw: Implement 'create' subcommand

## Description
Implement 'jira-mgmt create' subcommand. Flags: --type (epic/story/task/subtask), --summary (required), --description, --project (or use config), --parent (for subtasks). Call Jira API to create issue. Return created issue key.

## Scope
(define task scope)

## Acceptance Criteria
- Creates epic with --type epic
- Creates story with --type story
- Creates task with --type task
- Creates subtask with --type subtask and --parent
- Returns created issue key
- Validates required fields (summary, type)
- Uses --project or falls back to config
