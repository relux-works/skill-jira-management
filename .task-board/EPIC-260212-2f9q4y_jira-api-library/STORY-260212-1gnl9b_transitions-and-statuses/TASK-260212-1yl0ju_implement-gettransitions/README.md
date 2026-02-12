# TASK-260212-1yl0ju: Implement GetTransitions

## Description
Implement GetTransitions(issueKey) to fetch available transitions for an issue. Return slice of Transition structs with ID, name, to-status

## Scope
(define task scope)

## Acceptance Criteria
- GetTransitions(issueKey) function implemented\n- Uses GET /rest/api/3/issue/{issueKey}/transitions\n- Returns slice of Transition structs (ID, name, to-status)\n- Includes target status for each transition
