# TASK-260212-3nvoti: Implement 'comment' subcommand

## Description
Implement 'jira-mgmt comment' subcommand. Args: ISSUE-KEY, --body "text". Add comment to issue via Jira API. Support multiline comments.

## Scope
(define task scope)

## Acceptance Criteria
- Adds comment to specified issue
- Supports single-line comments
- Supports multiline comments
- Shows error for invalid issue key
- Confirms comment added successfully
