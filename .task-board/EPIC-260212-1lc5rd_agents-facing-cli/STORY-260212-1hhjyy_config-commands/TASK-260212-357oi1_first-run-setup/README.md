# TASK-260212-357oi1: first-run-setup

## Description
When CLI is launched for the first time without config, run interactive setup wizard: ask for Jira instance URL, email, API token (save to secure storage), locale (ru/en), optionally set active project/board

## Scope
(define task scope)

## Acceptance Criteria
- Detect missing config on any command launch\n- Prompt for instance URL, email, API token\n- Validate credentials against Jira API (test request)\n- Save to secure storage\n- Prompt for locale (ru/en)\n- Optionally list and select active project\n- Optionally list and select active board\n- Skip setup if already configured\n- Allow re-running setup via jira-mgmt auth --reconfigure
