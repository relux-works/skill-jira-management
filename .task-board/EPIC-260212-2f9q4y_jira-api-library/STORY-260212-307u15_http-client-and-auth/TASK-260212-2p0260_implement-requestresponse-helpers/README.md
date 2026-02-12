# TASK-260212-2p0260: Implement request/response helpers

## Description
Implement helper methods for GET, POST, PUT, DELETE requests with JSON marshal/unmarshal, handle base URL concatenation, set Content-Type headers

## Scope
(define task scope)

## Acceptance Criteria
- doRequest() method handles GET, POST, PUT, DELETE\n- Request body marshaled to JSON when provided\n- Response body unmarshaled from JSON\n- URL properly constructed from base URL + path\n- Content-Type: application/json header set for requests with body
