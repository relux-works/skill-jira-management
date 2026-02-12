# DSL Query Examples

Comprehensive examples of `jira-mgmt` DSL queries for common agent patterns.

---

## Basic Queries

### Get Single Issue

**Minimal (key, summary, status):**
```bash
jira-mgmt q 'get(PROJ-123)'
```
**Output:**
```
PROJ-123: User Authentication
Status: In Progress
```

**Default (+ assignee, priority, type):**
```bash
jira-mgmt q 'get(PROJ-123){default}'
```
**Output:**
```
PROJ-123: User Authentication
Status: In Progress
Assignee: alice@example.com
Priority: High
Type: Story
```

**Overview (+ description, parent, epic):**
```bash
jira-mgmt q 'get(PROJ-123){overview}'
```
**Output:**
```
PROJ-123: User Authentication
Status: In Progress
Assignee: alice@example.com
Priority: High
Type: Story
Parent: PROJ-100 (User Management Epic)

Description:
Implement OAuth2 authentication with Google and GitHub providers.
```

**Full (all fields including subtasks):**
```bash
jira-mgmt q 'get(PROJ-123){full}'
```
**Output (JSON):**
```json
{
  "key": "PROJ-123",
  "summary": "User Authentication",
  "status": "In Progress",
  "assignee": "Alice Jones",
  "type": "Story",
  "priority": "High",
  "parent": "PROJ-100",
  "description": "Implement OAuth2 authentication with Google and GitHub providers.",
  "labels": ["auth", "security"],
  "reporter": "Bob Smith",
  "created": "2026-02-10T10:00:00.000+0000",
  "updated": "2026-02-11T14:30:00.000+0000",
  "project": "PROJ",
  "subtasks": [
    {"key": "PROJ-130", "summary": "Google OAuth2 setup", "status": "Done"},
    {"key": "PROJ-131", "summary": "GitHub OAuth2 setup", "status": "In Progress"}
  ]
}
```

**Subtasks field:** Only present in `full` preset. Shows key, summary, status for each subtask. Useful for tracking progress of parent stories/tasks.

---

## List Queries

### By Sprint

**Current sprint:**
```bash
jira-mgmt q 'list(sprint=current){default}'
```
**Output:**
```
Sprint: Sprint 24 (8 issues)

PROJ-120: Epic: User Management [In Progress]
PROJ-123: Story: User Authentication [In Progress] @alice
PROJ-124: Story: Password Reset [To Do]
PROJ-125: Task: Write tests [In Progress] @bob
PROJ-126: Bug: Login fails [To Do]
PROJ-127: Task: Add validation [To Do] @charlie
PROJ-128: Story: OAuth2 setup [Done] @alice
PROJ-129: Task: Documentation [In Review] @bob
```

**Specific sprint by ID:**
```bash
jira-mgmt q 'list(sprint=123){default}'
```

---

### By Assignee

**My issues:**
```bash
jira-mgmt q 'list(assignee=me){default}'
```

**Specific user:**
```bash
jira-mgmt q 'list(assignee=alice@example.com){default}'
```

**My open issues (exclude done):**
```bash
jira-mgmt q 'list(assignee=me,status=!done){default}'
```

---

### By Status

**In progress only:**
```bash
jira-mgmt q 'list(status=in-progress){default}'
```

**To Do only:**
```bash
jira-mgmt q 'list(status=todo){default}'
```

**Exclude done:**
```bash
jira-mgmt q 'list(status=!done){default}'
```

---

### By Type

**Epics only:**
```bash
jira-mgmt q 'list(type=epic){overview}'
```

**Stories only:**
```bash
jira-mgmt q 'list(type=story){default}'
```

**Tasks only:**
```bash
jira-mgmt q 'list(type=task){default}'
```

**Bugs only:**
```bash
jira-mgmt q 'list(type=bug){default}'
```

**Subtasks only:**
```bash
jira-mgmt q 'list(type=subtask){default}'
```

---

### Combined Filters

**Current sprint, my issues, in progress:**
```bash
jira-mgmt q 'list(sprint=current,assignee=me,status=in-progress){default}'
```

**Current sprint epics:**
```bash
jira-mgmt q 'list(sprint=current,type=epic){overview}'
```

**My open bugs:**
```bash
jira-mgmt q 'list(assignee=me,type=bug,status=!done){default}'
```

**Current sprint tasks to do:**
```bash
jira-mgmt q 'list(sprint=current,type=task,status=todo){default}'
```

---

## Summary Queries

**Board summary:**
```bash
jira-mgmt q 'summary()'
```
**Output:**
```
Board: ACME Development (board ID: 456)
Active Sprint: Sprint 24

Status Breakdown:
- To Do: 12 issues
- In Progress: 8 issues
- In Review: 4 issues
- Done: 26 issues

Type Breakdown:
- Epic: 3
- Story: 18
- Task: 22
- Bug: 7

Top Assignees:
- alice@example.com: 15 issues
- bob@example.com: 12 issues
- charlie@example.com: 8 issues
```

---

## JQL Search Queries

### By Project

**All issues in project:**
```bash
jira-mgmt q 'search(jql="project=PROJ"){default}'
```

**Multiple projects:**
```bash
jira-mgmt q 'search(jql="project in (PROJ, ACME, PLATFORM)"){default}'
```

---

### By Issue Type

**All epics:**
```bash
jira-mgmt q 'search(jql="project=PROJ AND issuetype=Epic"){overview}'
```

**Stories and tasks:**
```bash
jira-mgmt q 'search(jql="issuetype in (Story, Task)"){default}'
```

---

### By Status

**All in progress:**
```bash
jira-mgmt q 'search(jql="statusCategory=In Progress"){default}'
```

**Open issues (not done):**
```bash
jira-mgmt q 'search(jql="statusCategory!=Done"){default}'
```

---

### By Assignee

**My issues:**
```bash
jira-mgmt q 'search(jql="assignee=currentUser()"){default}'
```

**Specific user:**
```bash
jira-mgmt q 'search(jql="assignee=alice@example.com"){default}'
```

**Unassigned:**
```bash
jira-mgmt q 'search(jql="assignee is EMPTY"){default}'
```

---

### By Date Range

**Created in last 7 days:**
```bash
jira-mgmt q 'search(jql="created>=-7d"){default}'
```

**Updated today:**
```bash
jira-mgmt q 'search(jql="updated>=startOfDay()"){default}'
```

**Resolved this week:**
```bash
jira-mgmt q 'search(jql="resolved>=startOfWeek()"){default}'
```

**Overdue:**
```bash
jira-mgmt q 'search(jql="due<now() AND statusCategory!=Done"){default}'
```

---

### Combined JQL

**My open issues in current sprint:**
```bash
jira-mgmt q 'search(jql="assignee=currentUser() AND sprint in openSprints() AND statusCategory!=Done"){default}'
```

**Recent bugs in project:**
```bash
jira-mgmt q 'search(jql="project=PROJ AND issuetype=Bug AND created>=-7d"){default}'
```

**Blocked issues:**
```bash
jira-mgmt q 'search(jql="status=Blocked"){default}'
```

**In review this week:**
```bash
jira-mgmt q 'search(jql="status=In Review AND updated>=startOfWeek()"){default}'
```

---

## Batch Queries

**Summary + current sprint work:**
```bash
jira-mgmt q 'summary(); list(sprint=current){default}'
```

**Multiple issues:**
```bash
jira-mgmt q 'get(PROJ-123){minimal}; get(PROJ-124){minimal}; get(PROJ-125){minimal}'
```

**Summary + my work + blockers:**
```bash
jira-mgmt q 'summary(); list(assignee=me,status=in-progress){default}; search(jql="status=Blocked"){default}'
```

**Epic + all stories under it:**
```bash
jira-mgmt q 'get(PROJ-100){overview}; search(jql="parent=PROJ-100"){default}'
```

---

## Agent Use Cases

### Daily Standup

**Yesterday's completed work:**
```bash
jira-mgmt q 'search(jql="assignee=currentUser() AND statusCategoryChangedDate>=-1d AND statusCategory=Done"){default}'
```

**Today's in progress:**
```bash
jira-mgmt q 'list(assignee=me,status=in-progress){default}'
```

**Blockers:**
```bash
jira-mgmt q 'search(jql="assignee=currentUser() AND status=Blocked"){default}'
```

**All at once:**
```bash
jira-mgmt q 'search(jql="assignee=currentUser() AND statusCategoryChangedDate>=-1d AND statusCategory=Done"){default}; list(assignee=me,status=in-progress){default}; search(jql="status=Blocked"){default}'
```

---

### Sprint Planning

**Backlog items (not in sprint):**
```bash
jira-mgmt q 'search(jql="sprint is EMPTY AND statusCategory=To Do"){default}'
```

**Current sprint scope:**
```bash
jira-mgmt q 'list(sprint=current){default}'
```

**Sprint summary:**
```bash
jira-mgmt q 'summary()'
```

---

### Sprint Review

**Completed this sprint:**
```bash
jira-mgmt q 'list(sprint=current,status=done){overview}'
```

**Carry-over (not done):**
```bash
jira-mgmt q 'list(sprint=current,status=!done){default}'
```

**Full review:**
```bash
jira-mgmt q 'summary(); list(sprint=current,status=done){overview}; list(sprint=current,status=!done){default}'
```

---

### Bug Triage

**Open bugs:**
```bash
jira-mgmt q 'list(type=bug,status=!done){default}'
```

**High priority bugs:**
```bash
jira-mgmt q 'search(jql="issuetype=Bug AND priority in (Highest, High) AND statusCategory!=Done"){default}'
```

**Recent bugs (last 7 days):**
```bash
jira-mgmt q 'search(jql="issuetype=Bug AND created>=-7d"){default}'
```

---

### Epic Progress Tracking

**Epic details with subtasks:**
```bash
jira-mgmt q 'get(PROJ-100){full}'
# The `full` preset includes subtasks — shows each subtask's key, summary, status
```

**All stories under epic:**
```bash
jira-mgmt q 'search(jql="parent=PROJ-100"){default}'
```

**Epic + children:**
```bash
jira-mgmt q 'get(PROJ-100){overview}; search(jql="parent=PROJ-100"){default}'
```

**Story with subtasks — check decomposition:**
```bash
jira-mgmt q 'get(PROJ-123){full}'
# Shows subtasks inline — useful to verify story is properly broken down
```

---

### Team Workload

**All in-progress issues:**
```bash
jira-mgmt q 'list(status=in-progress){default}'
```

**By assignee:**
```bash
jira-mgmt q 'list(assignee=alice@example.com,status=!done){default}; list(assignee=bob@example.com,status=!done){default}'
```

**Unassigned work:**
```bash
jira-mgmt q 'search(jql="assignee is EMPTY AND statusCategory!=Done"){default}'
```

---

## When to Use DSL vs JQL

### Use DSL (Preferred)

**Simple filters:**
```bash
# Good: DSL
jira-mgmt q 'list(sprint=current,assignee=me){default}'

# Avoid: JQL for simple cases
jira-mgmt q 'search(jql="sprint in openSprints() AND assignee=currentUser()"){default}'
```

**Token efficiency:**
- DSL queries are more concise
- Less token overhead for agents
- Faster parsing

---

### Use JQL (When Necessary)

**Complex conditions:**
```bash
# JQL required
jira-mgmt q 'search(jql="statusCategoryChangedDate>=-1d AND statusCategory=Done"){default}'
```

**Historical queries:**
```bash
# JQL required
jira-mgmt q 'search(jql="assignee was currentUser() AND assignee!=currentUser()"){default}'
```

**Advanced filters:**
```bash
# JQL required
jira-mgmt q 'search(jql="due<now() AND due>=-7d AND statusCategory!=Done"){default}'
```

**See `jql-patterns.md` for comprehensive JQL reference.**

---

**Document Version:** 1.0
**Last Updated:** 2026-02-12
