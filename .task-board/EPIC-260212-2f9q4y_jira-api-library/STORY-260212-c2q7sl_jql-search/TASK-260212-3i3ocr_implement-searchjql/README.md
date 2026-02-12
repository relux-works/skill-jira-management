# TASK-260212-3i3ocr: Implement SearchJQL

## Description
Implement SearchJQL(query, startAt, maxResults, fields) to execute JQL queries with pagination and field projection. Return SearchResult with issues slice and pagination info

## Scope
(define task scope)

## Acceptance Criteria
- SearchJQL(query, startAt, maxResults, fields) function implemented\n- Uses POST /rest/api/3/search\n- Accepts JQL query string\n- Field projection: specify which fields to return\n- Returns SearchResult struct with issues slice
