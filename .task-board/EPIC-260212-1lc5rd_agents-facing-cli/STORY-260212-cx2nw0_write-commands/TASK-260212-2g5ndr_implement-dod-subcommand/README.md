# TASK-260212-2g5ndr: Implement 'dod' subcommand

## Description
Implement 'jira-mgmt dod' subcommand for Definition of Done management. Args: ISSUE-KEY, --set "criteria" (sets DoD), --show (displays current DoD), --clear (removes DoD). Store in custom field or description section.

## Scope
(define task scope)

## Acceptance Criteria
- --set stores DoD criteria
- --show displays current DoD
- --clear removes DoD
- Handles missing DoD gracefully
- Stores DoD in consistent location (custom field or description)
