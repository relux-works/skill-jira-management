# TASK-260212-3kt4lo: Implement field selector with Jira presets

## Description
Field projection system: minimal (key, summary, status), default (+ assignee, priority, updated), overview (+ description, comments count), full (all fields). Syntax: operation{field1,field2} or operation@preset.

## Scope
(define task scope)

## Acceptance Criteria
- Minimal preset: key, summary, status
- Default preset: + assignee, priority, updated
- Overview preset: + description, comments count
- Full preset: all available fields
- Custom field selector: {field1,field2,...} syntax works
- Invalid preset/field shows error
