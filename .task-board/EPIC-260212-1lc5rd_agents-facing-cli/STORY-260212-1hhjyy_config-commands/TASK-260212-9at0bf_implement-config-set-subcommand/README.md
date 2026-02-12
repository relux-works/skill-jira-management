# TASK-260212-9at0bf: Implement 'config set' subcommand

## Description
Implement 'jira-mgmt config set' subcommand. Set config values: --project (default project key), --board (default board ID), --locale (e.g., en_US, ru_RU). Store in config file (~/.jira-mgmt/config.json or similar).

## Scope
(define task scope)

## Acceptance Criteria
- Sets default project with --project
- Sets default board with --board
- Sets locale with --locale (validates locale format)
- Stores settings in config file
- Shows confirmation after setting
- Validates input values
