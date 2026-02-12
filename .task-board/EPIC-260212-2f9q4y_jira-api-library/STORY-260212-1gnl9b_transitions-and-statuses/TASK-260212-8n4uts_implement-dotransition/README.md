# TASK-260212-8n4uts: Implement DoTransition

## Description
Implement DoTransition(issueKey, transitionID) to execute a transition. Handle optional transition fields (resolution, comment)

## Scope
(define task scope)

## Acceptance Criteria
- DoTransition(issueKey, transitionID) function implemented\n- Uses POST /rest/api/3/issue/{issueKey}/transitions\n- Supports optional fields (resolution, comment)\n- Returns error if transition not available or fails\n- No return value on success (204 response)
