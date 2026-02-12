# TASK-260212-14vfyb: Adapt scoped grep from agent-facing-api skill

## Description
Port scoped grep logic from agent-facing-api skill. Implement pattern matching, file filtering, case-insensitive search, context lines (-A/-B/-C flags). Search across cached Jira data.

## Scope
(define task scope)

## Acceptance Criteria
- Pattern matching works for literal strings and regex
- Case-insensitive search (-i) works correctly
- Context lines (-A/-B/-C) display correctly
- File filtering limits search scope
- Searches cached Jira data efficiently
