# TASK-260212-yj0qqv: Define Jira-specific operations

## Description
Implement operations: get(ISSUE-KEY) for single issue fetch, list(filters) for issue listing, summary() for aggregate data, search(jql="...") for JQL queries. Each operation returns structured data.

## Scope
(define task scope)

## Acceptance Criteria
- get(ISSUE-KEY) fetches single issue from Jira API
- list(filters) returns filtered issue list
- summary() returns aggregate statistics
- search(jql='...') executes JQL query
- All operations return consistent data structure
