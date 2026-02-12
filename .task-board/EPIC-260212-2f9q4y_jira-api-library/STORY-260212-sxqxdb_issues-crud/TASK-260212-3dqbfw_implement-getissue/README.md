# TASK-260212-3dqbfw: Implement GetIssue

## Description
Implement GetIssue(key string) to fetch issue by key, parse response into Issue struct with all relevant fields

## Scope
(define task scope)

## Acceptance Criteria
- GetIssue(key) function implemented\n- Uses GET /rest/api/3/issue/{key}\n- Returns Issue struct with all fields populated\n- Returns error if issue not found (404)
