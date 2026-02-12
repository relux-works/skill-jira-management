# TASK-260212-kh90zn: Implement GetBoard and ListSprints

## Description
Implement GetBoard(boardID) for board details and ListSprints(boardID) for sprint listing. Sprint struct includes ID, name, state (active/closed/future)

## Scope
(define task scope)

## Acceptance Criteria
- GetBoard(boardID) function implemented using GET /rest/agile/1.0/board/{boardID}\n- ListSprints(boardID) function implemented using GET /rest/agile/1.0/board/{boardID}/sprint\n- Sprint struct includes ID, name, state (active/closed/future), start/end dates\n- Both functions return appropriate errors if not found
