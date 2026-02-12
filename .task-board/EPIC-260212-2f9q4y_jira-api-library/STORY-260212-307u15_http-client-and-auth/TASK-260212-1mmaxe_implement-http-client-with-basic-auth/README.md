# TASK-260212-1mmaxe: Implement HTTP client with Basic Auth

## Description
Implement HTTP client initialization with Basic Auth header (email:token base64), set up default timeout, user-agent header

## Scope
(define task scope)

## Acceptance Criteria
- NewClient() constructor accepts Config and returns Client\n- Authorization header set as 'Basic base64(email:apiToken)'\n- Default timeout configured (30s)\n- User-Agent header set to identifiable value
