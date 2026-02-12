# TASK-260212-fpvw1r: Write tests for HTTP client and auth

## Description
Create mock HTTP server, test auth header format, test error handling (401, 404, 500), test request/response cycle for all methods

## Scope
(define task scope)

## Acceptance Criteria
- Mock HTTP server created for tests\n- Auth header verified in requests\n- 401 error returns AuthError\n- 404/500 errors return APIError with status code\n- Successful requests properly unmarshal response\n- All HTTP methods (GET, POST, PUT, DELETE) tested
