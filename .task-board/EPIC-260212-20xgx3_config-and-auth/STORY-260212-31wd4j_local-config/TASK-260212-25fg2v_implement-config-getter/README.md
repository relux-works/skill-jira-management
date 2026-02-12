# TASK-260212-25fg2v: implement-config-getter

## Description
Implement GetConfig method with fallback to defaults. Handle missing config file gracefully. Return merged config (file values override defaults). Cache config in memory to avoid repeated file reads.

## Scope
(define task scope)

## Acceptance Criteria
- GetConfig returns merged config (file + defaults)
- Missing config file handled gracefully
- Config cached in memory
- Cache invalidated on updates
- No performance issues from repeated calls
