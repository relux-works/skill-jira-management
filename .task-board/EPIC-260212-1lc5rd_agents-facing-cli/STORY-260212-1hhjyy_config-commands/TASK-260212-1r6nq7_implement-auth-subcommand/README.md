# TASK-260212-1r6nq7: Implement 'auth' subcommand

## Description
Implement 'jira-mgmt auth' subcommand for interactive authentication setup. Prompt for: Jira instance URL, email, API token. Validate credentials by test API call. Store securely (keychain on macOS, credential manager on Linux/Windows).

## Scope
(define task scope)

## Acceptance Criteria
- Interactive prompts for URL, email, API token
- Validates credentials with test Jira API call
- Stores credentials securely (keychain/credential manager)
- Shows success/error messages clearly
- Handles invalid credentials gracefully
