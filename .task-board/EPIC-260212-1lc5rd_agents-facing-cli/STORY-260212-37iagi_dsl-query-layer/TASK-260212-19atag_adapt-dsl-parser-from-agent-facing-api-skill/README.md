# TASK-260212-19atag: Adapt DSL parser from agent-facing-api skill

## Description
Port tokenizer and recursive descent parser from agent-facing-api skill. Adapt for Jira operations. Implement semicolon-separated batch query support.

## Scope
(define task scope)

## Acceptance Criteria
- Tokenizer splits input into tokens correctly
- Parser builds AST from tokens
- Semicolon-separated queries parse as batch
- Parser handles syntax errors gracefully
- Code follows agent-facing-api skill patterns
