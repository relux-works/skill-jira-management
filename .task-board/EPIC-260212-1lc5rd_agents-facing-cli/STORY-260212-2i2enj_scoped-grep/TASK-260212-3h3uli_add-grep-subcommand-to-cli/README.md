# TASK-260212-3h3uli: Add 'grep' subcommand to CLI

## Description
Implement 'jira-mgmt grep <pattern>' subcommand. Support flags: -i (case-insensitive), -A/-B/-C (context lines), --file (filter by file pattern). Wire up to grep implementation. Output matches with context.

## Scope
(define task scope)

## Acceptance Criteria
- 'jira-mgmt grep <pattern>' executes search
- -i flag enables case-insensitive search
- -A/-B/-C flags control context lines
- --file flag filters by file pattern
- Output shows matches with file:line:content format
- Matches highlighted or clearly marked
