# TASK-260212-25t9q0: Implement batch execution

## Description
Execute semicolon-separated queries in order. Collect results, handle errors per-query. Return array of results in JSON format or formatted text.

## Scope
(define task scope)

## Acceptance Criteria
- Multiple queries separated by semicolon execute in order
- Each query result collected independently
- Errors in one query don't stop execution
- Results returned as array (JSON) or sequential blocks (text)
- Empty results handled correctly
