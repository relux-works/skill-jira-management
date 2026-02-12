# TASK-260212-2f3j16: Implement ListIssues

## Description
Implement ListIssues with filters: by project, by board. Support pagination (startAt, maxResults). Return slice of Issue structs

## Scope
(define task scope)

## Acceptance Criteria
- ListIssues() function implemented\n- Filters: by project key, by board ID\n- Pagination parameters: startAt, maxResults\n- Returns slice of Issue structs and total count\n- Uses JQL internally for filtering
