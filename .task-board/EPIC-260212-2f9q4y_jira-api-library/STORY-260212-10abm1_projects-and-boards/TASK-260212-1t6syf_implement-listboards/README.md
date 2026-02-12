# TASK-260212-1t6syf: Implement ListBoards

## Description
Implement ListBoards(projectKey) to fetch all boards for a project, return slice of Board structs with ID, name, type

## Scope
(define task scope)

## Acceptance Criteria
- ListBoards(projectKey) function implemented\n- Uses GET /rest/agile/1.0/board with projectKeyOrId filter\n- Returns slice of Board structs (ID, name, type)\n- Handles empty result gracefully
