# TASK-260212-19mg84: Implement UpdateIssue

## Description
Implement UpdateIssue() to update summary, description, and DoD custom field. Support partial updates

## Scope
(define task scope)

## Acceptance Criteria
- UpdateIssue() function implemented\n- Uses PUT /rest/api/3/issue/{key}\n- Can update summary, description, DoD custom field\n- Supports partial updates (only changed fields)\n- Returns error on failure
