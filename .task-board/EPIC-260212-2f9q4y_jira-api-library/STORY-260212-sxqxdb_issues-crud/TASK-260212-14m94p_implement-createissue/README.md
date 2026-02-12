# TASK-260212-14m94p: Implement CreateIssue

## Description
Implement CreateIssue() with parameters: project key, issue type (epic/story/task/subtask/bug), summary, description, parent/epic link. Handle custom fields (DoD)

## Scope
(define task scope)

## Acceptance Criteria
- CreateIssue() function implemented\n- Uses POST /rest/api/3/issue\n- Supports all issue types: epic, story, task, subtask, bug\n- Sets parent link for subtasks\n- Sets epic link for stories/tasks under epic\n- Returns created issue key
