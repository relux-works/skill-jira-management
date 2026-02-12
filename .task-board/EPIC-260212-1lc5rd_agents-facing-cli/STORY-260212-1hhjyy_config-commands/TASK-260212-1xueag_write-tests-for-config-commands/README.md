# TASK-260212-1xueag: Write tests for config commands

## Description
Unit tests for: auth (credential validation, storage), config set (all settings), config show (output format). Mock credential storage and file I/O. Test error handling for invalid inputs.

## Scope
(define task scope)

## Acceptance Criteria
- Tests auth credential validation and storage
- Tests config set for all settings
- Tests config show output formatting
- Tests error handling for invalid inputs
- Tests missing config scenarios
- All tests use mocked storage and I/O
- All tests pass with 'go test'
