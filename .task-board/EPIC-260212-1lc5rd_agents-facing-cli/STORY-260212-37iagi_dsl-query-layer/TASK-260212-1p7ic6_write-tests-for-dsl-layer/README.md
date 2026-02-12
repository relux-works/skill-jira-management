# TASK-260212-1p7ic6: Write tests for DSL layer

## Description
Unit tests for: parser (tokenizer, operations, field selectors, batch queries), operations (get, list, summary, search), field selector presets, batch execution, error handling.

## Scope
(define task scope)

## Acceptance Criteria
- Parser tests cover all operations and syntax
- Operation tests verify correct Jira API calls
- Field selector tests verify all presets and custom selectors
- Batch execution tests verify ordering and error handling
- All tests pass with 'go test'
