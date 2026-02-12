# TASK-260212-29tq09: Implement AddComment

## Description
Implement AddComment(issueKey, body) to add comment to issue. Accept plain text or ADF format. Return created comment ID

## Scope
(define task scope)

## Acceptance Criteria
- AddComment(issueKey, body) function implemented\n- Uses POST /rest/api/3/issue/{issueKey}/comment\n- Accepts plain text (auto-converts to ADF) or ADF struct\n- Returns created comment ID\n- Handles errors (issue not found, permission denied)
