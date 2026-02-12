# TASK-260212-9owy9a: Implement ListStatuses

## Description
Implement ListStatuses(projectKey) to list all statuses available in a project. Return slice of Status structs with ID, name, category

## Scope
(define task scope)

## Acceptance Criteria
- ListStatuses(projectKey) function implemented\n- Uses GET /rest/api/3/project/{projectKey}/statuses\n- Returns slice of Status structs (ID, name, category)\n- Groups statuses by issue type if needed
