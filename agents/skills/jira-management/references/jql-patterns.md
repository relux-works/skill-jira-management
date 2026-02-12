# JQL Query Patterns Reference

**Purpose:** Comprehensive JQL (Jira Query Language) patterns for AI agents building Jira queries.

**Target Audience:** AI agents implementing Jira search functionality, automation tools, and reporting systems.

**Context:** Works with both Jira Cloud (API v3) and Server/DC (API v2). The `jira-mgmt` CLI auto-detects instance type.

---

## IMPORTANT: Server/DC JQL Compatibility

**The `!=` operator causes escaping errors on Jira Server/DC.** The `!` character is interpreted as an escape prefix, producing:
```
Ошибка в запросе JQL: «\!» — некорректная escape-последовательность
```

**Always use `NOT field = value` instead of `field != value` on Server/DC:**

| Cloud (works) | Server/DC (use this instead) |
|----------------|------------------------------|
| `statusCategory != "Done"` | `NOT statusCategory = "Done"` |
| `project != "ARCHIVE"` | `NOT project = "ARCHIVE"` |
| `issuetype != "Subtask"` | `NOT issuetype = "Subtask"` |
| `assignee != currentUser()` | `NOT assignee = currentUser()` |

The `NOT` syntax works on **both** Cloud and Server/DC — prefer it universally.

**Other Server/DC notes:**
- Status names may be localized (e.g. `"Бэклог"`, `"В работе"`, `"Открыто"` instead of `"Backlog"`, `"In Progress"`, `"Open"`)
- Use exact status names as returned by `jira-mgmt q 'get(KEY){full}'`
- Pagination is offset-based (`startAt`), not cursor-based — handled automatically by CLI

---

## Table of Contents

1. [Project Filters](#1-project-filters)
2. [Issue Type Filters](#2-issue-type-filters)
3. [Status Filters](#3-status-filters)
4. [Assignee Filters](#4-assignee-filters)
5. [Sprint Filters](#5-sprint-filters)
6. [Date Range Filters](#6-date-range-filters)
7. [Combined Query Patterns](#7-combined-query-patterns)

---

## 1. Project Filters

### 1.1 Single Project

**Pattern:**
```jql
project = "PROJECT_KEY"
```

**Example:**
```jql
project = "ACME"
```

**Notes:**
- Project key is case-insensitive
- Use exact key (not project name)
- To find project key: `/rest/api/3/project` endpoint

---

### 1.2 Multiple Projects

**Pattern:**
```jql
project in ("KEY1", "KEY2", "KEY3")
```

**Example:**
```jql
project in ("ACME", "PLATFORM", "MOBILE")
```

**Notes:**
- Use `in` operator for OR logic across projects
- Quotes required for each key
- Comma-separated list

---

### 1.3 Exclude Project

**Pattern:**
```jql
project != "PROJECT_KEY"
```

**Example:**
```jql
project != "ARCHIVE"
```

**Alternative (exclude multiple):**
```jql
project not in ("ARCHIVE", "DEPRECATED")
```

---

## 2. Issue Type Filters

### 2.1 Single Issue Type

**Pattern:**
```jql
issuetype = "TYPE_NAME"
```

**Standard Types:**
```jql
issuetype = "Epic"
issuetype = "Story"
issuetype = "Task"
issuetype = "Bug"
issuetype = "Subtask"
```

**Notes:**
- Type names are case-insensitive
- Use exact type name as configured in Jira
- Custom types: use exact name from instance

---

### 2.2 Multiple Issue Types

**Pattern:**
```jql
issuetype in ("TYPE1", "TYPE2", "TYPE3")
```

**Example:**
```jql
issuetype in ("Story", "Task", "Bug")
```

**Common Patterns:**

**Work items (no epics):**
```jql
issuetype in ("Story", "Task", "Bug")
```

**High-level planning:**
```jql
issuetype = "Epic"
```

**Development work:**
```jql
issuetype in ("Story", "Bug", "Task")
```

---

### 2.3 Exclude Issue Types

**Pattern:**
```jql
issuetype != "TYPE_NAME"
```

**Example (exclude subtasks):**
```jql
issuetype != "Subtask"
```

**Alternative:**
```jql
issuetype not in ("Subtask", "Epic")
```

**Notes:**
- Useful for excluding subtasks from counts
- Subtasks often clutter sprint views

---

## 3. Status Filters

### 3.1 Single Status

**Pattern:**
```jql
status = "STATUS_NAME"
```

**Standard Statuses:**
```jql
status = "To Do"
status = "In Progress"
status = "Done"
status = "Blocked"
status = "In Review"
```

**Notes:**
- Status names are case-insensitive
- Use exact status name from workflow
- Quotes required if status has spaces

---

### 3.2 Multiple Statuses

**Pattern:**
```jql
status in ("STATUS1", "STATUS2", "STATUS3")
```

**Example (active work):**
```jql
status in ("In Progress", "In Review", "Testing")
```

**Example (not started):**
```jql
status in ("To Do", "Backlog")
```

---

### 3.3 Status Category

**Pattern:**
```jql
statusCategory = "CATEGORY"
```

**Standard Categories:**
```jql
statusCategory = "To Do"       # Not started
statusCategory = "In Progress" # Active work
statusCategory = "Done"        # Completed
```

**Why Use Categories:**
- Workflow-agnostic (works across different projects)
- Simpler than listing all custom status names
- More maintainable for tools

**Example (all open issues):**
```jql
statusCategory != "Done"
```

**Example (work in flight):**
```jql
statusCategory = "In Progress"
```

---

### 3.4 Status Category vs Status

**Use Status When:**
- Need precise workflow control
- Building status-specific reports
- Working with known workflow

**Use Status Category When:**
- Building cross-project tools
- Don't know exact status names
- Want workflow-independent queries

---

## 4. Assignee Filters

### 4.1 Current User

**Pattern:**
```jql
assignee = currentUser()
```

**Example (my open tasks):**
```jql
assignee = currentUser() AND statusCategory != "Done"
```

**Notes:**
- `currentUser()` is a function, not a string
- No quotes around `currentUser()`
- Resolves to authenticated user's account ID

---

### 4.2 Specific User

**Pattern (by email):**
```jql
assignee = "user@example.com"
```

**Pattern (by account ID):**
```jql
assignee = "5b10a2844c20165700ede21g"
```

**Notes:**
- Email-based queries may not work on all instances
- Account ID is more reliable
- Get account ID from `/rest/api/3/user/search` endpoint

---

### 4.3 Unassigned Issues

**Pattern:**
```jql
assignee is EMPTY
```

**Alternative:**
```jql
assignee = null
```

**Example (unassigned bugs):**
```jql
issuetype = "Bug" AND assignee is EMPTY
```

**Notes:**
- `is EMPTY` and `= null` are equivalent
- Use `is EMPTY` for readability

---

### 4.4 Multiple Assignees

**Pattern:**
```jql
assignee in ("user1@example.com", "user2@example.com")
```

**Example (team members):**
```jql
assignee in ("alice@example.com", "bob@example.com", "charlie@example.com")
```

---

### 4.5 Was Assigned (Ever)

**Pattern:**
```jql
assignee was "user@example.com"
```

**Example:**
```jql
assignee was currentUser() AND assignee != currentUser()
```

**Notes:**
- `was` operator checks historical values
- Useful for finding reassigned issues

---

## 5. Sprint Filters

### 5.1 Specific Sprint

**Pattern (by name):**
```jql
sprint = "Sprint Name"
```

**Pattern (by ID):**
```jql
sprint = 123
```

**Example:**
```jql
sprint = "Sprint 24"
```

**Notes:**
- Sprint names must match exactly
- Sprint IDs are more reliable
- Get sprint ID from `/rest/agile/1.0/board/{boardId}/sprint` endpoint

---

### 5.2 Open Sprints

**Pattern:**
```jql
sprint in openSprints()
```

**Example (current sprint work):**
```jql
sprint in openSprints() AND statusCategory != "Done"
```

**Notes:**
- `openSprints()` includes all active sprints
- Does not include future or closed sprints
- Most common sprint query for active work

---

### 5.3 Future Sprints

**Pattern:**
```jql
sprint in futureSprints()
```

**Example (planned work):**
```jql
sprint in futureSprints() AND issuetype = "Story"
```

**Notes:**
- Sprints that haven't started yet
- Useful for backlog planning

---

### 5.4 Closed Sprints

**Pattern:**
```jql
sprint in closedSprints()
```

**Example (sprint retrospective data):**
```jql
sprint in closedSprints() AND updated >= -14d
```

**Notes:**
- Completed sprints
- Useful for historical reporting

---

### 5.5 Backlog (No Sprint)

**Pattern:**
```jql
sprint is EMPTY
```

**Alternative:**
```jql
sprint = null
```

**Example (unplanned stories):**
```jql
issuetype = "Story" AND sprint is EMPTY AND statusCategory = "To Do"
```

**Notes:**
- Issues not assigned to any sprint
- Backlog items waiting for sprint assignment

---

### 5.6 Multiple Sprints

**Pattern:**
```jql
sprint in ("Sprint 1", "Sprint 2", "Sprint 3")
```

**Example (last 3 sprints):**
```jql
sprint in ("Sprint 22", "Sprint 23", "Sprint 24")
```

---

## 6. Date Range Filters

### 6.1 Relative Date Ranges

**Pattern:**
```jql
FIELD >= -Nd
FIELD <= -Nd
```

**Examples:**

**Created in last 7 days:**
```jql
created >= -7d
```

**Updated in last 30 days:**
```jql
updated >= -30d
```

**Due in next 14 days:**
```jql
due <= 14d
```

**Resolved in last week:**
```jql
resolved >= -7d
```

**Notes:**
- `d` = days, `w` = weeks, `m` = months, `y` = years
- Negative value = past (e.g., `-7d` = 7 days ago)
- Positive value = future (e.g., `14d` = 14 days from now)
- Time units: `h` (hours), `d` (days), `w` (weeks), `m` (months), `y` (years)

---

### 6.2 Start/End of Period Functions

**Start Functions:**
```jql
startOfDay()
startOfWeek()
startOfMonth()
startOfYear()
```

**End Functions:**
```jql
endOfDay()
endOfWeek()
endOfMonth()
endOfYear()
```

**Examples:**

**Updated this week:**
```jql
updated >= startOfWeek()
```

**Created this month:**
```jql
created >= startOfMonth()
```

**Due by end of month:**
```jql
due <= endOfMonth()
```

**Resolved this year:**
```jql
resolved >= startOfYear()
```

---

### 6.3 Specific Date Ranges

**Pattern:**
```jql
FIELD >= "YYYY-MM-DD" AND FIELD <= "YYYY-MM-DD"
```

**Example:**
```jql
created >= "2026-01-01" AND created <= "2026-01-31"
```

**Notes:**
- Date format: `YYYY-MM-DD`
- Quotes required
- Can include time: `"YYYY-MM-DD HH:mm"`

---

### 6.4 Overdue Issues

**Pattern:**
```jql
due < now() AND statusCategory != "Done"
```

**Alternative (overdue by more than 3 days):**
```jql
due < -3d AND statusCategory != "Done"
```

**Notes:**
- `now()` returns current date/time
- Common pattern for reports

---

### 6.5 No Due Date

**Pattern:**
```jql
due is EMPTY
```

**Example (stories without due date):**
```jql
issuetype = "Story" AND due is EMPTY
```

---

### 6.6 Common Date Field Names

- `created` — when issue was created
- `updated` — last modification time
- `resolved` — when issue was resolved
- `due` — due date (if set)
- `resolutiondate` — resolution timestamp

---

## 7. Combined Query Patterns

### 7.1 My Open Tasks

**Query:**
```jql
assignee = currentUser() AND statusCategory != "Done"
```

**With ordering:**
```jql
assignee = currentUser() AND statusCategory != "Done" ORDER BY updated DESC
```

**Notes:**
- Most common personal query
- Add `ORDER BY priority DESC, updated DESC` for prioritized view

---

### 7.2 Current Sprint Backlog

**Query:**
```jql
sprint in openSprints() AND statusCategory = "To Do"
```

**With project filter:**
```jql
project = "ACME" AND sprint in openSprints() AND statusCategory = "To Do" ORDER BY rank
```

**Notes:**
- `ORDER BY rank` maintains backlog order
- `rank` is Jira's internal priority field

---

### 7.3 Active Sprint Work

**Query:**
```jql
sprint in openSprints() AND statusCategory = "In Progress"
```

**With assignee:**
```jql
sprint in openSprints() AND statusCategory = "In Progress" AND assignee = currentUser()
```

---

### 7.4 Overdue Issues in Project

**Query:**
```jql
project = "ACME" AND due < now() AND statusCategory != "Done" ORDER BY due ASC
```

**Notes:**
- `ORDER BY due ASC` shows most overdue first
- Filter out completed issues

---

### 7.5 Recent Bugs

**Query:**
```jql
issuetype = "Bug" AND created >= -7d ORDER BY created DESC
```

**With status filter:**
```jql
issuetype = "Bug" AND created >= -7d AND statusCategory != "Done" ORDER BY priority DESC
```

---

### 7.6 Unassigned Backlog Items

**Query:**
```jql
assignee is EMPTY AND sprint is EMPTY AND statusCategory = "To Do"
```

**With project and type:**
```jql
project = "ACME" AND issuetype in ("Story", "Bug") AND assignee is EMPTY AND sprint is EMPTY ORDER BY created DESC
```

---

### 7.7 Team Sprint Progress

**Query:**
```jql
sprint in openSprints() AND assignee in ("alice@example.com", "bob@example.com")
```

**With status breakdown:**
```jql
sprint in openSprints() AND assignee in ("alice@example.com", "bob@example.com") ORDER BY status, assignee
```

---

### 7.8 Epic Progress

**Query:**
```jql
"Epic Link" = "ACME-123" ORDER BY status, rank
```

**Notes:**
- `"Epic Link"` is a standard Jira field
- Requires quotes because of space
- Shows all issues under an epic

---

### 7.9 Recently Updated Issues

**Query:**
```jql
project = "ACME" AND updated >= -7d ORDER BY updated DESC
```

**With status filter:**
```jql
project = "ACME" AND updated >= -7d AND statusCategory != "Done" ORDER BY updated DESC
```

---

### 7.10 Cross-Project Search

**Query:**
```jql
project in ("ACME", "PLATFORM", "MOBILE") AND assignee = currentUser() AND statusCategory = "In Progress"
```

**Notes:**
- Use `in` for multiple projects
- Useful for org-wide views

---

### 7.11 Stories Ready for Development

**Query:**
```jql
issuetype = "Story" AND status = "Ready for Dev" AND sprint is EMPTY
```

**With refinement criteria:**
```jql
issuetype = "Story" AND status = "Ready for Dev" AND sprint is EMPTY AND "Story Points" is not EMPTY
```

**Notes:**
- Custom field `"Story Points"` requires quotes
- Filters for estimated stories

---

### 7.12 Blocked Issues

**Query:**
```jql
status = "Blocked" ORDER BY created ASC
```

**With project:**
```jql
project = "ACME" AND status = "Blocked" ORDER BY created ASC
```

**Notes:**
- Shows oldest blocked issues first
- Useful for stand-ups

---

### 7.13 Issues in Review

**Query:**
```jql
status in ("In Review", "Code Review", "QA") AND assignee = currentUser()
```

**Notes:**
- Adapt status names to your workflow
- Shows items awaiting feedback

---

### 7.14 Sprint Completion Report

**Query:**
```jql
sprint = "Sprint 24" ORDER BY statusCategory, issuetype
```

**With time filter:**
```jql
sprint = "Sprint 24" AND resolved >= startOfDay(-14d)
```

**Notes:**
- Shows sprint scope
- Orders by completion status

---

### 7.15 This Month's Completed Work

**Query:**
```jql
resolved >= startOfMonth() AND assignee = currentUser() ORDER BY resolved DESC
```

**With project:**
```jql
project = "ACME" AND resolved >= startOfMonth() ORDER BY resolved DESC
```

---

## Query Operators Reference

### Comparison Operators
- `=` — equals
- `!=` — not equals
- `>` — greater than
- `>=` — greater than or equal
- `<` — less than
- `<=` — less than or equal

### List Operators
- `in` — matches any value in list
- `not in` — matches no value in list

### Null Operators
- `is EMPTY` — field has no value
- `is not EMPTY` — field has a value
- `= null` — equivalent to `is EMPTY`
- `!= null` — equivalent to `is not EMPTY`

### Text Operators
- `~` — contains text (case-insensitive)
- `!~` — does not contain text
- `~` with wildcards: `summary ~ "bug*"`

### Historical Operators
- `was` — field had this value at any time
- `was in` — field had any of these values
- `was not` — field never had this value
- `changed` — field value changed

### Logical Operators
- `AND` — both conditions must be true
- `OR` — either condition must be true
- `NOT` — negates condition

---

## Query Best Practices

### 1. Use Status Category for Portability
```jql
# Portable (works across workflows)
statusCategory = "In Progress"

# Brittle (workflow-specific)
status in ("In Dev", "Code Review", "Testing", "QA")
```

### 2. Order Results
Always add `ORDER BY` for predictable results:
```jql
project = "ACME" ORDER BY created DESC
```

### 3. Limit Scope
Start narrow, expand as needed:
```jql
# Good: specific scope
project = "ACME" AND sprint in openSprints()

# Avoid: too broad
sprint in openSprints()
```

### 4. Quote Field Names with Spaces
```jql
"Epic Link" = "ACME-123"
"Story Points" > 5
```

### 5. Use Functions for Dates
```jql
# Readable
created >= startOfWeek()

# Less readable
created >= "2026-02-10"
```

### 6. Test Queries in Jira UI First
- Use Jira's "Advanced Search" to validate JQL
- Copy working queries into code
- Jira UI provides autocomplete and validation

---

## Common Gotchas

### 1. Field Name Case Sensitivity
- Standard fields are case-insensitive: `PROJECT`, `project`, `Project` all work
- Custom fields may be case-sensitive: check exact name

### 2. Sprint Functions vs Sprint Names
```jql
# Function (no quotes)
sprint in openSprints()

# Name (quotes required)
sprint = "Sprint 24"
```

### 3. Status vs Status Category
```jql
# Different meanings
status = "In Progress"          # Exact workflow status
statusCategory = "In Progress"  # Any status in "In Progress" category
```

### 4. Empty vs Null
```jql
# Both work, prefer "is EMPTY" for readability
assignee is EMPTY
assignee = null
```

### 5. Date Format
```jql
# Correct
created >= "2026-01-01"

# Wrong
created >= "01/01/2026"
```

### 6. currentUser() is a Function
```jql
# Correct
assignee = currentUser()

# Wrong
assignee = "currentUser()"
```

---

## Integration with REST API

### Search Endpoint
```
POST /rest/api/3/search/jql
```

### Request Body
```json
{
  "jql": "project = ACME AND sprint in openSprints()",
  "startAt": 0,
  "maxResults": 50,
  "fields": ["summary", "status", "assignee"]
}
```

### Common Fields to Request
- `summary` — issue title
- `status` — current status
- `assignee` — assigned user
- `priority` — priority level
- `issuetype` — issue type
- `created` — creation date
- `updated` — last update
- `description` — issue description (ADF format in v3)
- `customfield_XXXXX` — custom fields (get IDs from `/rest/api/3/field`)

### Pagination
- `startAt` — offset for results (0-based)
- `maxResults` — max results per page (default: 50, max: 100)
- Response includes `total` count for pagination logic

---

## Custom Fields

### Finding Custom Field IDs
```
GET /rest/api/3/field
```

Returns all fields with their IDs.

### Using Custom Fields in JQL
```jql
# By field name (if unique)
"Story Points" > 5

# By field ID (always works)
cf[10024] > 5
```

**Notes:**
- Custom field names require quotes
- Field IDs are more reliable for automation
- Get field ID from `/rest/api/3/field` endpoint

---

## Performance Considerations

### 1. Index Optimization
Jira indexes these fields for fast queries:
- `project`
- `issuetype`
- `status`
- `assignee`
- `created`, `updated`, `resolved`

Use indexed fields in `WHERE` clause for better performance.

### 2. Avoid Broad Text Searches
```jql
# Slow (full-text search)
summary ~ "performance"

# Fast (indexed field)
project = "ACME" AND summary ~ "performance"
```

### 3. Limit Result Sets
- Use `maxResults` in API calls
- Add specific filters to reduce scope
- Paginate large result sets

---

## Testing Queries

### 1. Jira UI Advanced Search
- Navigate to Filters → Advanced Search
- Enter JQL query
- Validate results
- Copy working query into code

### 2. API Testing
```bash
curl -X POST \
  -H "Authorization: Basic $(echo -n 'email:token' | base64)" \
  -H "Content-Type: application/json" \
  -d '{"jql": "project = ACME", "maxResults": 1}' \
  https://your-domain.atlassian.net/rest/api/3/search
```

### 3. Validation Tools
- Use Jira's query autocomplete
- Check `/rest/api/3/jql/autocompletedata` for valid values
- Validate field names with `/rest/api/3/field`

---

## Additional Resources

- **Official JQL Docs:** https://support.atlassian.com/jira-service-management-cloud/docs/use-advanced-search-with-jira-query-language-jql/
- **JQL Functions Reference:** https://support.atlassian.com/jira-software-cloud/docs/jql-functions/
- **JQL Operators Reference:** https://support.atlassian.com/jira-software-cloud/docs/jql-operators/
- **Search API Docs:** https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/

---

**Document Version:** 1.0
**Last Updated:** 2026-02-12
**Maintained By:** agent-jql-ref
