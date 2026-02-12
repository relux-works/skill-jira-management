# TASK-260212-1yux1u: Write tests for write commands

## Description
Unit tests for: create (all issue types), transition (valid/invalid statuses), comment (single/multiline), dod (set/show/clear), locale-aware content. Mock Jira API calls.

## Scope
(define task scope)

## Acceptance Criteria
- Tests create for all issue types
- Tests transition for valid and invalid statuses
- Tests comment for single and multiline
- Tests dod set/show/clear operations
- Tests locale-aware content generation
- All tests use mocked Jira API
- All tests pass with 'go test'
