# TASK-260212-2m0qi1: Implement ListComments

## Description
Implement ListComments(issueKey) to fetch all comments for an issue. Return slice of Comment structs with ID, author, body, created timestamp

## Scope
(define task scope)

## Acceptance Criteria
- ListComments(issueKey) function implemented\n- Uses GET /rest/api/3/issue/{issueKey}/comment\n- Returns slice of Comment structs (ID, author, body, created)\n- Parses ADF body to plain text for display\n- Handles empty comment list
