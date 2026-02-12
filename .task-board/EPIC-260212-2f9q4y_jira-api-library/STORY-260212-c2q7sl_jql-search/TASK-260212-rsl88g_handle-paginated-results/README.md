# TASK-260212-rsl88g: Handle paginated results

## Description
Implement pagination handling: accept startAt/maxResults parameters, return total count, next page indicator. Helper functions for auto-pagination if needed

## Scope
(define task scope)

## Acceptance Criteria
- SearchResult struct includes total count, startAt, maxResults\n- Pagination info allows calculating hasNextPage\n- Optional: SearchAllJQL() helper that auto-paginates through all results\n- Handles empty results gracefully
