# TASK-260212-1gpytz: Implement 'transition' subcommand

## Description
Implement 'jira-mgmt transition' subcommand. Args: ISSUE-KEY, --to "Status Name". Fetch available transitions, match status name, execute transition. Handle invalid status names gracefully.

## Scope
(define task scope)

## Acceptance Criteria
- Fetches available transitions for issue
- Matches status name case-insensitively
- Executes transition successfully
- Shows error for invalid status names
- Shows error for invalid issue key
- Confirms transition completion
