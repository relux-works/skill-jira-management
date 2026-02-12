# TASK-260212-3vrt9a: Add 'q' subcommand to CLI

## Description
Implement 'jira-mgmt q' subcommand. Accept query string as argument. Wire up parser, executor, formatter. Respect --format flag (json/text).

## Scope
(define task scope)

## Acceptance Criteria
- 'jira-mgmt q "<query>"' executes query
- Output respects --format flag (json/text)
- Parser errors show clear messages
- Query results display correctly
- Subcommand integrates with global flags
