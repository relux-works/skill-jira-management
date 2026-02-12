# TASK-260212-1ltck2: Define Jira client interface and types

## Description
Define core Client struct, Config struct (base URL, email, API token), and error types (AuthError, NetworkError, APIError with status codes)

## Scope
(define task scope)

## Acceptance Criteria
- Client struct defined with http.Client field\n- Config struct with BaseURL, Email, APIToken fields\n- Error types defined: AuthError, NetworkError, APIError\n- APIError includes HTTP status code and response body
