## Status
done

## Assigned To
agent-cli

## Created
2026-02-12T11:41:36Z

## Last Update
2026-02-12T12:13:35Z

## Blocked By
- (none)

## Blocks
- TASK-260212-yj0qqv
- TASK-260212-25t9q0

## Checklist
(empty)

## Notes
Implemented: internal/query/parser.go — recursive descent DSL parser adapted from agent-facing-api. Operations: get, list, summary, search. Field presets: minimal, default, overview, full. Batch support via semicolons. 17 tests in parser_test.go, all passing.
