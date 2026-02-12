# TASK-260212-bnwl6a: Add global flags

## Description
Implement global flags: --project (string), --board (string), --format (json/text with text default). Flags available to all subcommands. Use Cobra persistent flags.

## Scope
(define task scope)

## Acceptance Criteria
- --project flag accepts string value
- --board flag accepts string value
- --format flag accepts 'json' or 'text', defaults to 'text'
- Flags are persistent and available to all subcommands
- Invalid --format value shows error
