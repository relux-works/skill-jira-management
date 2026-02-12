# TASK-260212-37czh6: implement-config-setters

## Description
Implement SetActiveProject, SetActiveBoard, SetLocale methods. Validate inputs (locale must be ru or en, project/board IDs must be non-empty if set). Persist changes to config file immediately.

## Scope
(define task scope)

## Acceptance Criteria
- SetActiveProject, SetActiveBoard, SetLocale implemented
- Locale validation (must be ru or en)
- Project/board ID validation (non-empty if set)
- Changes persisted immediately to file
- Clear error messages for invalid inputs
