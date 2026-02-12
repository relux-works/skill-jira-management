# TASK-260212-11al5r: Write tests for JQL search

## Description
Test SearchJQL with various queries, test pagination (multiple pages), test field projection, verify result parsing and pagination info

## Scope
(define task scope)

## Acceptance Criteria
- Test SearchJQL with simple queries (project=KEY)\n- Test SearchJQL with complex queries (status IN (Open, In Progress))\n- Test pagination: multiple pages, correct startAt/maxResults\n- Test field projection: verify only requested fields returned\n- Mock server returns paginated responses\n- Test auto-pagination helper if implemented
