# Jira Cloud REST API Research

**Date:** 2026-02-12
**Purpose:** Research for building a Go CLI tool to interact with Jira Cloud

---

## 1. API Versions

### v2 vs v3

**Similarities:**
- Both versions offer the same collection of operations
- Most endpoints have counterparts in both versions

**Key Differences:**
- **v3** provides support for **Atlassian Document Format (ADF)** in rich text fields:
  - Issue descriptions
  - Comments (in issue, issue link, and transition resources)
  - Textarea-type custom fields (multi-line text fields)
- **v2** uses **wiki markdown** format for rich text fields by default

**Current Status:**
- **v3 is the latest version** and is recommended for new integrations
- v2 is still supported but lacks ADF support, which is critical for working with modern Jira Cloud instances

**Why ADF Matters:**
- Rich text fields are among the most important in Jira (e.g., issue descriptions)
- Cannot meaningfully work with issues without proper description field handling
- v3's ADF support makes it preferable for CLI tools that need to read/write formatted content

**Recommendation:** Use **v3** for new tools to ensure compatibility with rich text formatting.

### Base URL Format

```
https://your-domain.atlassian.net/rest/api/3/
```

For Jira Agile/Software endpoints:
```
https://your-domain.atlassian.net/rest/agile/1.0/
```

---

## 2. Authentication

### Recommended Method: Basic Auth with API Token

**API Token Generation:**
- Generate API tokens from your Atlassian account settings
- Each token is specific to your account and can be revoked independently
- **IMPORTANT:** API tokens will expire between **March 14 and May 12, 2026**

**Security Benefits:**
- Not storing primary account password
- Quick revocation on per-use basis
- Works with 2FA and SAML-enabled organizations

**Implementation:**
1. Obtain email address and API token
2. Create Basic Auth string: `email:token`
3. Base64 encode the string
4. Send in `Authorization` header: `Basic <base64-encoded-string>`

**Example Authorization Header:**
```
Authorization: Basic dXNlckBleGFtcGxlLmNvbTphcGlfdG9rZW5faGVyZQ==
```

**Required Headers:**
- `Authorization: Basic <credentials>`
- `Content-Type: application/json` (for POST/PUT requests)
- `Accept: application/json`

**Alternative Methods:**
- OAuth 1.0a (deprecated for Jira Cloud)
- Cookie-based auth (deprecated)
- Session-based auth (not recommended)

**Note:** API token-based traffic is not affected by the new 2026 points-based rate limits and continues to be governed by existing burst rate limits.

---

## 3. Issue Operations

### Get Issue

**Endpoint:** `GET /rest/api/3/issue/{issueIdOrKey}`

**Usage:**
- `issueIdOrKey` can be the numeric ID (e.g., `10000`) or issue key (e.g., `PROJ-123`)
- Returns full issue details including fields, comments, attachments, etc.

**Useful Query Parameters:**
- `fields` - comma-separated list of fields to return (reduces payload size)
- `expand` - expand additional data (e.g., `expand=names` lists all custom field names)
- `properties` - issue properties to return

**Response:**
- JSON object with issue data
- Custom fields appear as `customfield_XXXXX` in the response

### Create Issue

**Endpoint:** `POST /rest/api/3/issue`

**Required Fields:**
- `project.key` or `project.id`
- `issuetype.name` or `issuetype.id`
- `summary` (issue title)

**Common Optional Fields:**
- `description` (ADF format in v3)
- `assignee`
- `priority`
- `labels`
- `components`
- `parent` (for subtasks, links to parent issue)
- Custom fields (e.g., `customfield_10001`)

**Issue Type Support:**
- **Epic** - high-level initiatives
- **Story** - user-centric features
- **Task** - generic work items
- **Subtask** - children of other issues
- **Bug** - defects/problems

**Example Request Body:**
```json
{
  "fields": {
    "project": {
      "key": "PROJ"
    },
    "issuetype": {
      "name": "Task"
    },
    "summary": "Issue summary",
    "description": {
      "type": "doc",
      "version": 1,
      "content": [
        {
          "type": "paragraph",
          "content": [
            {
              "type": "text",
              "text": "Issue description in ADF format"
            }
          ]
        }
      ]
    }
  }
}
```

### Update Issue

**Endpoint:** `PUT /rest/api/3/issue/{issueIdOrKey}`

**Supported Operations:**
- `SET` - for single-value fields (replaces value)
- `ADD` - for multi-value fields (appends value)
- `REMOVE` - for multi-value fields (removes value)

**Update vs Edit:**
- Use `PUT /rest/api/3/issue/{issueIdOrKey}` for field updates
- Use operations (SET/ADD/REMOVE) to modify field values

**Important:** Not all fields support all operations. Single-value fields typically support only SET.

### Search Issues (JQL)

**Endpoint (Current):** `POST /rest/api/3/search/jql`
**Legacy Endpoint:** `GET /rest/api/3/search` (deprecated, removed from Jira Cloud)

**IMPORTANT MIGRATION:**
- The legacy `/rest/api/3/search` endpoint is **deprecated and fully removed** from Jira Cloud
- All implementations must migrate to `/rest/api/3/search/jql`
- New endpoint supports both GET and POST methods
- Use POST for longer JQL strings to avoid URL length limitations

**Request Body (POST):**
```json
{
  "jql": "project = PROJ AND status = 'In Progress'",
  "startAt": 0,
  "maxResults": 50,
  "fields": ["summary", "status", "assignee"]
}
```

**Response Fields:**
- `issues` - array of issue objects
- `total` - total number of matching issues (note: may not be available with nextPageToken pagination)
- `nextPageToken` - token for next page (new pagination method)
- `isLast` - boolean indicating if this is the last page

**Pagination:**
- Old method: `startAt` parameter (deprecated)
- New method: `nextPageToken` (dynamically generated per request)
- Maximum page size: 5,000 issues per request
- See section 12 for detailed pagination patterns

**JQL Examples:**
```
project = PROJ
status = "In Progress"
assignee = currentUser()
created >= -7d
labels = backend AND status != Done
parent = PROJ-123
issuetype in (Story, Task) AND sprint in openSprints()
```

**Performance Considerations:**
- The new `/rest/api/3/search/jql` endpoint with nextPageToken can be slower than the legacy approach
- Sequential pagination required (cannot parallelize requests)
- Some users report intermittent issues with cursor-based search

**Issues Returned Include Agile Fields:**
- `sprint`
- `closedSprints`
- `flagged`
- `epic`

---

## 4. Issue Types

### Standard Issue Types

Jira Cloud provides five standard issue types:

1. **Epic**
   - High-level initiatives or large pieces of work
   - Can contain Stories, Tasks, and Bugs as children
   - Represents features or major work streams
   - Cannot be a child of another issue

2. **Story** (User Story)
   - User-centric requirement or feature
   - Defined using non-technical language
   - Typically a child of an Epic
   - Can have Subtasks as children

3. **Task**
   - Generic piece of work not directly related to user requirements
   - Examples: server upgrades, code refactoring, administrative work
   - Can be standalone or child of Epic
   - Can have Subtasks as children

4. **Bug**
   - Problem that impairs or prevents product functions
   - Can be standalone or child of Epic
   - Can have Subtasks as children

5. **Subtask**
   - Smaller, actionable piece of a larger issue
   - Must have a parent (Story, Task, or Bug)
   - Cannot have children of its own
   - Lowest level in hierarchy

### Hierarchy Relationships

**Recommended Structure:**
```
Epic → Story/Task/Bug → Subtask
```

**Rules:**
- Epic can have: Story, Task, Bug as children
- Story/Task/Bug can have: Subtask as children
- Subtask cannot have children
- **Important:** Epics can have subtasks under them along with child issues, but this is not the recommended structure

### Custom Issue Types

- Organizations can create custom issue types
- Custom types can be configured with specific fields and workflows
- Access via `issuetype` field in API requests
- Use `issuetype.name` or `issuetype.id` when creating issues

### Issue Type Field

**In API Requests:**
```json
{
  "fields": {
    "issuetype": {
      "name": "Story"
    }
  }
}
```

**Or by ID:**
```json
{
  "fields": {
    "issuetype": {
      "id": "10001"
    }
  }
}
```

**Retrieving Issue Types:**
- Endpoint: `GET /rest/api/3/issuetype`
- Returns all issue types available in the Jira instance
- Each type includes: id, name, description, iconUrl, subtask flag

---

## 5. Projects

### List Projects

**Current Endpoint:** `GET /rest/api/3/project/search`
**Legacy Endpoint:** `GET /rest/api/2/project` (deprecated in Jira Cloud)

**Required Permissions:**
- `read:project:jira` scope

**Response Format:**
```json
{
  "self": "...",
  "nextPage": "...",
  "maxResults": 50,
  "startAt": 0,
  "total": 100,
  "isLast": false,
  "values": [
    {
      "id": "10000",
      "key": "PROJ",
      "name": "Project Name",
      "projectTypeKey": "software",
      "simplified": false,
      "style": "classic"
    }
  ]
}
```

**Pagination:**
- Returns 50 projects per page by default
- For large instances (5000 projects), requires 100 API calls
- Use `startAt` and `maxResults` parameters

**Query Parameters:**
- `startAt` - starting index
- `maxResults` - number of results per page (max 50)
- `orderBy` - sort order (e.g., `name`, `key`)
- `query` - search by project name or key

**Project Types:**
- `software` - Jira Software (Scrum/Kanban)
- `service_desk` - Jira Service Management
- `business` - Jira Work Management

**Get Single Project:**
- Endpoint: `GET /rest/api/3/project/{projectIdOrKey}`
- Returns detailed project information

---

## 6. Boards

### Jira Agile REST API

**Base URL:** `https://your-domain.atlassian.net/rest/agile/1.0/`

### List Boards

**Endpoint:** `GET /rest/agile/1.0/board`

**Response:**
- Returns all boards the user has permission to view
- Includes board type (Scrum/Kanban)

**Response Format:**
```json
{
  "maxResults": 50,
  "startAt": 0,
  "total": 10,
  "isLast": true,
  "values": [
    {
      "id": 1,
      "self": "...",
      "name": "Board Name",
      "type": "scrum"
    }
  ]
}
```

**Board Types:**
- **scrum** - Sprint-based board with backlog
- **kanban** - Continuous flow board

**Query Parameters:**
- `startAt` - pagination offset
- `maxResults` - results per page
- `type` - filter by board type (`scrum` or `kanban`)
- `name` - filter by board name
- `projectKeyOrId` - filter by project

### Get Board

**Endpoint:** `GET /rest/agile/1.0/board/{boardId}`

**Returns:**
- Board details including name, type, location

### Board Issues

**Endpoint:** `GET /rest/agile/1.0/board/{boardId}/issue`

**Returns:**
- Issues on the board
- Includes Agile fields (sprint, epic, etc.)

---

## 7. Sprints

### Get Sprints for Board

**Endpoint:** `GET /rest/agile/1.0/board/{boardId}/sprint`

**Response:**
```json
{
  "maxResults": 50,
  "startAt": 0,
  "isLast": true,
  "values": [
    {
      "id": 1,
      "self": "...",
      "state": "active",
      "name": "Sprint 1",
      "startDate": "2026-01-01T10:00:00.000Z",
      "endDate": "2026-01-14T10:00:00.000Z",
      "originBoardId": 1,
      "goal": "Sprint goal description"
    }
  ]
}
```

**Sprint States:**
- `future` - not started
- `active` - in progress
- `closed` - completed

### Create Sprint

**Endpoint:** `POST /rest/agile/1.0/sprint`

**Required Fields:**
- `name` - sprint name
- `originBoardId` - board ID

**Optional Fields:**
- `startDate` - ISO 8601 format
- `endDate` - ISO 8601 format
- `goal` - sprint goal description

### Get Sprint

**Endpoint:** `GET /rest/agile/1.0/sprint/{sprintId}`

**Returns:**
- Sprint details including state, dates, goal

### Sprint Issues

**Endpoint:** `GET /rest/agile/1.0/sprint/{sprintId}/issue`

**Returns:**
- All issues in the sprint
- Includes standard issue fields plus Agile data

### Finding Sprint/Board IDs

**Pattern:**
1. Get board ID: `GET /rest/agile/1.0/board`
2. Get sprints for board: `GET /rest/agile/1.0/board/{boardId}/sprint`
3. Use sprint ID in further operations

---

## 8. Transitions

### Get Available Transitions

**Endpoint:** `GET /rest/api/3/issue/{issueIdOrKey}/transitions`

**Purpose:**
- Retrieves all available transitions for an issue in its current status
- Transitions depend on workflow configuration and user permissions

**Response:**
```json
{
  "transitions": [
    {
      "id": "11",
      "name": "Start Progress",
      "to": {
        "self": "...",
        "id": "3",
        "name": "In Progress"
      },
      "hasScreen": false,
      "isGlobal": false,
      "isInitial": false
    }
  ]
}
```

**Key Fields:**
- `id` - transition ID (required for executing transition)
- `name` - transition name (e.g., "Start Progress")
- `to` - destination status
- `hasScreen` - whether transition requires user input
- `fields` - required/optional fields for transition

### Execute Transition

**Endpoint:** `POST /rest/api/3/issue/{issueIdOrKey}/transitions`

**Request Body:**
```json
{
  "transition": {
    "id": "11"
  },
  "fields": {
    "resolution": {
      "name": "Done"
    }
  },
  "update": {
    "comment": [
      {
        "add": {
          "body": {
            "type": "doc",
            "version": 1,
            "content": [
              {
                "type": "paragraph",
                "content": [
                  {
                    "type": "text",
                    "text": "Transition comment in ADF format"
                  }
                ]
              }
            ]
          }
        }
      }
    ]
  }
}
```

**Pattern:**
1. Get available transitions for issue
2. Find desired transition by name or ID
3. Execute transition with required fields

### Workflow Context

- Transitions are workflow-specific
- Available transitions vary by:
  - Current issue status
  - User permissions
  - Workflow configuration
  - Project settings

### Transition Properties (IMPORTANT 2026 CHANGE)

**Deprecation Notice:**
Several workflow transition property endpoints will be **removed on June 1, 2026:**

- Fetching transition properties → Use **Bulk get workflows** endpoint instead
- Adding transition properties → Use **Bulk update workflows** endpoint instead
- Deleting transition properties → Use **Bulk update workflows** endpoint instead

**Workflow Editor Change:**
Starting **June 26, 2026**, the old workflow editor will be removed. Workflows will only be editable in the new workflow editor.

**Impact:**
- Transition properties change workflow behavior
- Moving to bulk operations for better efficiency
- Update integrations before June 2026 deadlines

---

## 9. Comments

### Get Comments

**Endpoint:** `GET /rest/api/3/issue/{issueIdOrKey}/comment`

**Response:**
```json
{
  "startAt": 0,
  "maxResults": 50,
  "total": 10,
  "comments": [
    {
      "self": "...",
      "id": "10000",
      "author": {
        "accountId": "...",
        "displayName": "User Name"
      },
      "body": {
        "type": "doc",
        "version": 1,
        "content": [...]
      },
      "created": "2026-01-01T10:00:00.000+0000",
      "updated": "2026-01-01T11:00:00.000+0000"
    }
  ]
}
```

### Add Comment

**Endpoint:** `POST /rest/api/3/issue/{issueIdOrKey}/comment`

**Request Body (ADF Format):**
```json
{
  "body": {
    "type": "doc",
    "version": 1,
    "content": [
      {
        "type": "paragraph",
        "content": [
          {
            "type": "text",
            "text": "This is a comment in ADF format"
          }
        ]
      }
    ]
  }
}
```

### Atlassian Document Format (ADF)

**Overview:**
- JSON-based format for rich text in Jira Cloud
- Used in v3 API for comments, descriptions, and custom text fields
- v2 API uses wiki markdown instead

**Structure:**
- Hierarchy of nodes: block and inline
- **Block nodes:** structural elements (headings, paragraphs, lists)
- **Inline nodes:** content (text, images, links)
- **Marks:** formatting (bold, italics, color, etc.)

**Basic ADF Document:**
```json
{
  "type": "doc",
  "version": 1,
  "content": [
    {
      "type": "paragraph",
      "content": [
        {
          "type": "text",
          "text": "Plain text"
        }
      ]
    }
  ]
}
```

**With Formatting:**
```json
{
  "type": "doc",
  "version": 1,
  "content": [
    {
      "type": "paragraph",
      "content": [
        {
          "type": "text",
          "text": "Bold text",
          "marks": [
            {
              "type": "strong"
            }
          ]
        }
      ]
    }
  ]
}
```

**Common Node Types:**
- `paragraph` - text paragraph
- `heading` - heading (with level attribute)
- `bulletList` / `orderedList` - lists
- `listItem` - list item
- `codeBlock` - code block
- `blockquote` - quote block

**Common Marks:**
- `strong` - bold
- `em` - italic
- `code` - inline code
- `link` - hyperlink
- `textColor` - text color

**Challenges:**
- No official REST API endpoint to convert between ADF and other formats
- Feature request tracked in JRACLOUD-77436
- For plain text comments, use v2 API which returns wiki markdown by default

**Workaround for v2 API:**
- Use v2 counterpart endpoints to get wiki markdown format instead of ADF
- Example: `/rest/api/2/issue/{issueIdOrKey}/comment` returns markdown

---

## 10. Statuses

### Get All Statuses

**Endpoint:** `GET /rest/api/3/status`

**Purpose:**
- Returns list of all statuses associated with workflows
- Provides status details (id, name, category, etc.)

**Response:**
```json
[
  {
    "self": "...",
    "id": "1",
    "name": "Open",
    "statusCategory": {
      "self": "...",
      "id": 2,
      "key": "new",
      "colorName": "blue-gray",
      "name": "To Do"
    }
  },
  {
    "self": "...",
    "id": "3",
    "name": "In Progress",
    "statusCategory": {
      "self": "...",
      "id": 4,
      "key": "indeterminate",
      "colorName": "yellow",
      "name": "In Progress"
    }
  }
]
```

### Get Statuses for Project

**Endpoint:** `GET /rest/api/3/project/{projectIdOrKey}/statuses`
**Alternative:** `GET /rest/api/2/project/{projectIdOrKey}/statuses`

**Purpose:**
- Returns statuses specific to a project
- Shows statuses used by issue types in the project

**Response:**
- Array of issue types with their associated statuses
- Each issue type lists its workflow statuses

### Status Categories

All statuses map to one of three categories:

1. **To Do** (new)
   - Key: `new`
   - Color: blue-gray
   - Indicates work not started

2. **In Progress** (indeterminate)
   - Key: `indeterminate`
   - Color: yellow
   - Indicates work in progress

3. **Done** (done)
   - Key: `done`
   - Color: green
   - Indicates completed work

### Important Limitations

**Known Issues:**
- Status must be associated with an **active workflow** to be returned
- Does **not** include statuses unused in any workflow
- Does **not** include statuses from Next Gen (team-managed) projects
- No REST call exists for "all statuses from a specific workflow"

**Workarounds:**
- Use project-specific status endpoint to see statuses by project
- Get transitions for an issue to see available destination statuses
- Query workflow configuration (if needed)

### Status vs Transition

- **Status:** Current state of an issue (e.g., "Open", "In Progress", "Done")
- **Transition:** Action that moves issue from one status to another (e.g., "Start Progress")

To change status, execute a transition (see section 8).

---

## 11. Custom Fields

### Understanding Custom Fields

Custom fields in Jira extend the standard field set with project-specific or organization-specific data.

**Field ID Format:** `customfield_XXXXX` (e.g., `customfield_10001`)

**Common Use Cases:**
- Definition of Done (DoD)
- Story points
- Sprint fields
- Custom dropdowns, text fields, dates, etc.

### Finding Custom Field IDs

**Method 1: Field Search API**

**Endpoint:** `GET /rest/api/3/field/search`

**Query Parameters:**
- `type=custom` - show only custom fields
- `query` - search by field name

**Response:**
```json
{
  "maxResults": 50,
  "startAt": 0,
  "total": 10,
  "values": [
    {
      "id": "customfield_10001",
      "name": "Story Points",
      "schema": {
        "type": "number",
        "custom": "com.atlassian.jira.plugin.system.customfieldtypes:float",
        "customId": 10001
      }
    }
  ]
}
```

**Method 2: Get Issue with Names Expansion**

**Endpoint:** `GET /rest/api/3/issue/{issueIdOrKey}?expand=names`

**Purpose:**
- Lists all field names in scope for the issue
- Shows mapping of `customfield_XXXXX` to field names

**Method 3: Browser Developer Tools**

- Open issue in browser
- Open Developer Tools (F12)
- Inspect field element
- Field ID appears in HTML attributes

**Method 4: List All Fields**

**Endpoint:** `GET /rest/api/3/field`

**Returns:**
- All fields (system + custom)
- Each field includes: id, name, schema, custom flag

### Working with Custom Fields

**In Issue Creation:**
```json
{
  "fields": {
    "customfield_10001": 8,
    "customfield_10002": "Custom text value"
  }
}
```

**In Issue Updates:**
```json
{
  "fields": {
    "customfield_10001": 13
  }
}
```

**In Search Results:**
- Custom fields appear alongside standard fields
- Specify in `fields` parameter to include in response

### Custom Field Types

- **Text fields:** Single/multi-line text
- **Number fields:** Numeric values
- **Date fields:** Date/datetime pickers
- **Select fields:** Dropdown (single/multi-select)
- **User picker:** Atlassian user
- **Checkboxes:** Boolean or multi-value
- **URL fields:** Web links

### Custom Field Contexts

Custom fields can have different configurations (options, defaults) per project or issue type, called **contexts**.

**Endpoint:** `GET /rest/api/3/field/{fieldId}/context`

**Use Cases:**
- Different dropdown options per project
- Different default values per issue type

**Get Context Options:**

**Endpoint:** `GET /rest/api/3/field/{fieldId}/context/{contextId}/option`

**Purpose:**
- Retrieve available options for select-type fields
- Each option has an ID and value

### Definition of Done (DoD) Field

DoD is typically implemented as:
- Custom text field (single or multi-line)
- Custom checklist (via app/plugin)
- Standard description with convention

**Pattern:**
1. Find DoD field ID via field search
2. Use field ID in issue create/update operations
3. Store as text or ADF (depending on field type)

---

## 12. Pagination

### Overview

Jira Cloud REST API uses pagination for endpoints returning large result sets. Two systems exist:

1. **Legacy offset-based pagination** (`startAt`, `maxResults`)
2. **New cursor-based pagination** (`nextPageToken`)

### Legacy Pagination (Deprecated for Search)

**Parameters:**
- `startAt` - starting index (0-based)
- `maxResults` - number of results per page
- `total` - total number of results

**Example:**
```
GET /rest/api/3/project/search?startAt=0&maxResults=50
```

**Response:**
```json
{
  "startAt": 0,
  "maxResults": 50,
  "total": 150,
  "isLast": false,
  "values": [...]
}
```

**Next Page:**
```
GET /rest/api/3/project/search?startAt=50&maxResults=50
```

**Advantages:**
- Simple offset-based navigation
- Can jump to any page directly
- Parallel requests possible
- Total count known upfront

**Disadvantages:**
- Performance degrades with large offsets
- Inconsistent results if data changes during pagination

**Still Used By:**
- Projects API (`/rest/api/3/project/search`)
- Some other non-search endpoints

### New Cursor-Based Pagination (Search API)

**IMPORTANT:** The `/rest/api/3/search/jql` endpoint now uses **cursor-based pagination** exclusively.

**Parameters:**
- `nextPageToken` - opaque token for next page (dynamically generated)
- `maxResults` - results per page (max 5,000 for search)

**First Request:**
```json
POST /rest/api/3/search/jql
{
  "jql": "project = PROJ",
  "maxResults": 100
}
```

**Response:**
```json
{
  "maxResults": 100,
  "issues": [...],
  "nextPageToken": "abc123xyz",
  "isLast": false
}
```

**Subsequent Request:**
```json
POST /rest/api/3/search/jql
{
  "jql": "project = PROJ",
  "maxResults": 100,
  "nextPageToken": "abc123xyz"
}
```

**Advantages:**
- Consistent results even if data changes
- Better performance for large datasets
- Prevents aggressive API usage

**Disadvantages:**
- **Sequential only** - cannot parallelize requests
- Must wait for previous response to get next token
- **No total count** available upfront
- Cannot jump to arbitrary page
- **Slower** for bulk retrieval compared to legacy method

**Implementation Pattern:**
```go
func fetchAllIssues(jql string) ([]Issue, error) {
    var allIssues []Issue
    nextPageToken := ""

    for {
        req := SearchRequest{
            JQL: jql,
            MaxResults: 100,
        }
        if nextPageToken != "" {
            req.NextPageToken = nextPageToken
        }

        resp, err := postSearchJQL(req)
        if err != nil {
            return nil, err
        }

        allIssues = append(allIssues, resp.Issues...)

        if resp.IsLast {
            break
        }
        nextPageToken = resp.NextPageToken
    }

    return allIssues, nil
}
```

### Pagination Patterns by Endpoint

| Endpoint | Method | Max Results | Total Count |
|----------|--------|-------------|-------------|
| `/rest/api/3/search/jql` | nextPageToken | 5,000 | No |
| `/rest/api/3/project/search` | startAt | 50 | Yes |
| `/rest/api/3/issue/{key}/comment` | startAt | 50 | Yes |
| `/rest/agile/1.0/board` | startAt | 50 | Yes |
| `/rest/agile/1.0/board/{id}/sprint` | startAt | 50 | Yes |
| `/rest/agile/1.0/sprint/{id}/issue` | startAt | 50 | Yes |

### Performance Considerations

**For Large Datasets (Search):**
- New pagination can be significantly slower
- Cannot parallelize requests
- Some users report intermittent issues
- Consider caching results if frequent access needed

**For Project/Board Listing:**
- Legacy pagination still works
- Can parallelize if needed
- More predictable performance

### Migration Guide

**If Using Legacy Search API:**
1. Replace `/rest/api/3/search` with `/rest/api/3/search/jql`
2. Switch from GET to POST (or continue using GET)
3. Replace `startAt` logic with `nextPageToken` handling
4. Remove reliance on `total` count
5. Update pagination loop to use `isLast` flag

---

## 13. Error Handling

### Error Response Format

**Standard Error Response:**
```json
{
  "errorMessages": [
    "High-level error message"
  ],
  "errors": {
    "field1": "Field-specific error message",
    "field2": "Another field-specific error"
  },
  "status": 400
}
```

**Fields:**
- `errorMessages` - array of general error messages
- `errors` - object mapping field names to error messages
- `status` - HTTP status code

### Common HTTP Status Codes

#### 2xx Success
- **200 OK** - request succeeded
- **201 Created** - resource created successfully
- **204 No Content** - request succeeded, no response body

#### 4xx Client Errors
- **400 Bad Request**
  - Catch-all for malformed requests
  - Invalid JQL syntax (misspelled keywords, search terms)
  - References to non-existent entities (issue keys, projects)
  - Missing required fields
  - Invalid field values

- **401 Unauthorized**
  - Missing authentication credentials
  - Invalid API token
  - Credentials passed incorrectly

- **403 Forbidden**
  - User authenticated but lacks permissions
  - Project/issue not accessible to user
  - Operation not allowed by workflow

- **404 Not Found**
  - Resource doesn't exist
  - Issue key, project, or endpoint not found

- **410 Gone**
  - API endpoint removed (planned change)
  - Example: JVIS API removed (announced August 2024)
  - Permanent removal, cannot be recovered

- **429 Too Many Requests**
  - Rate limit exceeded
  - Response includes `Retry-After` header
  - Implement exponential backoff

#### 5xx Server Errors
- **500 Internal Server Error**
  - Server-side problem
  - Request might be valid but server failed to process
  - May be transient, retry with backoff

- **502 Bad Gateway** - upstream server error
- **503 Service Unavailable** - temporary unavailability, retry later

### Error Handling Best Practices

**1. Check HTTP Status Code First**
```go
if resp.StatusCode != http.StatusOK {
    // Handle error
}
```

**2. Parse Error Response**
```go
var errResp ErrorResponse
if err := json.Unmarshal(body, &errResp); err != nil {
    // Handle JSON parsing error
}
```

**3. Handle Specific Status Codes**
```go
switch resp.StatusCode {
case http.StatusUnauthorized:
    return fmt.Errorf("authentication failed: check API token")
case http.StatusForbidden:
    return fmt.Errorf("permission denied: %v", errResp.ErrorMessages)
case http.StatusNotFound:
    return fmt.Errorf("resource not found")
case http.StatusTooManyRequests:
    retryAfter := resp.Header.Get("Retry-After")
    return fmt.Errorf("rate limited, retry after %s seconds", retryAfter)
case http.StatusBadRequest:
    return fmt.Errorf("bad request: %v", errResp.ErrorMessages)
default:
    return fmt.Errorf("request failed: %v", errResp.ErrorMessages)
}
```

**4. Implement Retry Logic**
- Retry on 429, 500, 502, 503
- Use exponential backoff
- Respect `Retry-After` header for 429
- Maximum retry attempts (e.g., 3-5)

**5. Log Error Details**
- Include HTTP status code
- Include error messages from response
- Include request context (endpoint, method, etc.)

### Field Validation Errors

For **400** errors with field-specific problems:

```go
if len(errResp.Errors) > 0 {
    for field, msg := range errResp.Errors {
        fmt.Printf("Field '%s': %s\n", field, msg)
    }
}
```

### JQL Validation

**Invalid JQL returns 400:**
- Misspelled JQL keywords
- Invalid field names
- Syntax errors
- References to non-existent projects/issues

**Example Error:**
```json
{
  "errorMessages": [
    "Error in the JQL Query: The character '=' is a reserved JQL character. You must enclose it in a string or use the escape '\\u0027' instead."
  ]
}
```

### Deprecation Warnings

Watch for endpoints returning **410 Gone** - indicates permanent removal.

**Example:**
- Legacy search endpoint `/rest/api/3/search` → 410 after deprecation
- Transition property endpoints → 410 after June 1, 2026

**Mitigation:**
- Monitor Atlassian changelog
- Test against Jira Cloud regularly
- Implement endpoint fallbacks if needed

---

## 14. Go Libraries

### andygrunwald/go-jira

**Repository:** https://github.com/andygrunwald/go-jira
**Go Package:** https://pkg.go.dev/github.com/andygrunwald/go-jira

**Overview:**
- Most popular Go client library for Jira
- Supports both Jira Cloud and On-Premise (Server/Data Center)
- Code structure inspired by `google/go-github`

**Current Status:**
- **v1.x** - stable, production-ready
- **v2.x** - in development, contains breaking changes

**Architecture:**
- One main client
- Services extracted from client (Issues, Authentication, etc.)
- Service-oriented design for endpoint organization

**Supported Authentication:**
- HTTP Basic Auth
- OAuth
- Session Cookie (deprecated)

**Key Features:**
- Not API-complete but extensible
- Can call any API endpoint you want via raw requests
- Comprehensive coverage of common operations
- Active maintenance and community

**Version 2 Changes (In Development):**
- Split into two separate clients:
  - One for **Jira Cloud**
  - One for **Jira On-Premise** (Server/Data Center)
- Breaking API changes
- Improved type safety and error handling

**Latest Stable Release:**
- v1.16.0 (check repository for current version)

**Example Usage:**
```go
import "github.com/andygrunwald/go-jira"

tp := jira.BasicAuthTransport{
    Username: "user@example.com",
    Password: "api-token-here",
}

client, err := jira.NewClient(tp.Client(), "https://your-domain.atlassian.net")
if err != nil {
    panic(err)
}

issue, _, err := client.Issue.Get("PROJ-123", nil)
if err != nil {
    panic(err)
}

fmt.Printf("Issue: %s - %s\n", issue.Key, issue.Fields.Summary)
```

**Pros:**
- Battle-tested in production
- Good documentation
- Active community
- Handles auth, retries, etc.
- Type-safe API

**Cons:**
- Not 100% API coverage (though extensible)
- v2 still in development (breaking changes coming)
- Some endpoints might require manual implementation

### Other Libraries

**ctreminiom/go-atlassian:**
- Newer library
- Supports multiple Atlassian products (Jira, Confluence, etc.)
- More comprehensive API coverage
- Less mature than andygrunwald/go-jira

**dghubble/sling + manual implementation:**
- Sling is a Go HTTP client library
- Build custom Jira client on top
- Full control over implementation
- More work upfront

### Recommendation

**Use andygrunwald/go-jira if:**
- You need a stable, production-ready library
- Common operations are sufficient
- You want quick setup and good defaults

**Build from scratch if:**
- You need specific API features not in go-jira
- You want full control over implementation
- You need to minimize dependencies
- CLI tool is relatively simple

**Hybrid Approach:**
- Use go-jira for common operations (issue CRUD, search)
- Implement custom endpoints manually when needed
- Leverage go-jira's auth and HTTP client setup

### CLI Tool Considerations

For a CLI tool, consider:

1. **Minimal dependencies** - smaller binary size
2. **Custom requirements** - DoD field, specific workflows
3. **Error handling** - CLI-specific error formatting
4. **Configuration** - API token storage, domain config

**Verdict:**
- If building a **general-purpose CLI**, use `andygrunwald/go-jira`
- If building a **specialized CLI** with custom workflows, consider manual implementation with `net/http` or minimal HTTP client library

---

## 15. Rate Limiting

### Overview

Jira Cloud enforces **three independent rate limiting systems** simultaneously to protect platform stability:

1. **Burst rate limits** (traditional)
2. **Points-based rate limits** (new, 2026)
3. **Tiered quota rate limits** (new, 2026)

### Important 2026 Changes

**Enforcement Date:** March 2, 2026

**Scope:**
- Applies to **Forge, Connect, and OAuth 2.0 (3LO) apps**
- **API token-based traffic is NOT affected** and continues using burst rate limits only

**Impact:**
- Vast majority of apps won't be affected
- No action required for most apps
- CLI tools using API tokens unaffected by new limits

### Points-Based Model

**How It Works:**
- Each API call consumes **points** based on work performed
- Factors: data volume returned, operation complexity
- Different endpoints have different point costs

**Example:**
- Simple GET for single issue: 1-2 points
- Complex search returning 1000 issues: 50+ points

**Quota:**
- Points accumulated over a time window
- When quota exceeded, rate limit applies

### Burst Rate Limits (Traditional)

**Still Active:**
- Primary rate limiting system for API token auth
- Short time windows (e.g., per second, per minute)
- Limits number of requests, not points

**Typical Limits:**
- ~100-300 requests per minute (varies by plan)
- ~10 requests per second

**Note:** Exact limits not publicly documented and may vary by:
- Atlassian plan (Free, Standard, Premium, Enterprise)
- Endpoint
- System load

### Rate Limit Headers

**Response Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1609459200
```

**Fields:**
- `X-RateLimit-Limit` - total requests allowed in window
- `X-RateLimit-Remaining` - remaining requests
- `X-RateLimit-Reset` - timestamp when limit resets (Unix epoch)

**Note:** Not all endpoints return these headers consistently.

### HTTP 429 Response

**When Rate Limited:**
```
HTTP/1.1 429 Too Many Requests
Retry-After: 60
Content-Type: application/json

{
  "errorMessages": ["Rate limit exceeded"]
}
```

**Headers:**
- `Retry-After` - seconds to wait before retrying (integer or HTTP date)

### Handling Rate Limits

**1. Detect Rate Limit:**
```go
if resp.StatusCode == http.StatusTooManyRequests {
    // Rate limited
}
```

**2. Extract Retry-After:**
```go
retryAfter := resp.Header.Get("Retry-After")
if retryAfter != "" {
    seconds, _ := strconv.Atoi(retryAfter)
    time.Sleep(time.Duration(seconds) * time.Second)
}
```

**3. Implement Exponential Backoff:**
```go
func retryWithBackoff(fn func() error, maxRetries int) error {
    backoff := 1 * time.Second

    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }

        if isRateLimitError(err) {
            time.Sleep(backoff)
            backoff *= 2 // exponential
            if backoff > 60*time.Second {
                backoff = 60 * time.Second // cap at 60s
            }
            continue
        }

        return err // non-rate-limit error
    }

    return fmt.Errorf("max retries exceeded")
}
```

**4. Respect Retry-After Header:**
```go
func handleRateLimit(resp *http.Response) time.Duration {
    retryAfter := resp.Header.Get("Retry-After")
    if retryAfter == "" {
        return 60 * time.Second // default
    }

    seconds, err := strconv.Atoi(retryAfter)
    if err != nil {
        return 60 * time.Second
    }

    return time.Duration(seconds) * time.Second
}
```

### Best Practices

**For CLI Tools:**

1. **Implement retry logic** with exponential backoff
2. **Respect Retry-After header** (mandatory)
3. **Show progress** during long operations
4. **Batch operations** where possible
5. **Cache results** for repeated queries
6. **Use pagination wisely** - don't fetch more than needed
7. **Parallel requests** - limit concurrency to avoid burst limits
8. **User feedback** - inform user when rate limited

**Example:**
```go
const maxConcurrent = 5

limiter := make(chan struct{}, maxConcurrent)

for _, issue := range issues {
    limiter <- struct{}{} // acquire

    go func(iss Issue) {
        defer func() { <-limiter }() // release
        processIssue(iss)
    }(issue)
}
```

**Optimization:**
- Minimize API calls (use JQL search instead of multiple GETs)
- Use `fields` parameter to reduce payload size
- Cache frequently accessed data (projects, issue types, etc.)
- Implement local rate limiting (e.g., max 50 req/min)

**Monitoring:**
- Track API usage in CLI
- Log rate limit responses
- Warn user if approaching limits

### API Token Auth Benefits

**For CLI Tools:**
- **Not affected by new 2026 points-based limits**
- Continues using traditional burst limits
- Simpler to implement
- Recommended for CLI use cases

**Verdict:** CLI tools using API token authentication are in a good position and should not be significantly impacted by 2026 rate limit changes.

---

## Summary & Recommendations

### For Go CLI Tool Development

**1. API Version:**
- Use **v3** for full ADF support (descriptions, comments)
- Fall back to v2 if plain text is sufficient

**2. Authentication:**
- Implement **Basic Auth with API token**
- Store token securely (keychain, config file with appropriate permissions)
- Remember: tokens expire **March-May 2026** (prompt user to refresh)

**3. Issue Operations:**
- Use `POST /rest/api/3/search/jql` for search (not legacy endpoint)
- Implement **cursor-based pagination** with `nextPageToken`
- Handle issue types: Epic, Story, Task, Subtask, Bug

**4. Custom Fields:**
- Implement field discovery (`GET /rest/api/3/field/search`)
- Store field ID mappings (e.g., DoD field)
- Handle field types appropriately (text, number, date, select)

**5. Comments:**
- Support **ADF format** for rich text
- Provide plain text wrapper for user convenience
- Consider v2 API for markdown if simpler

**6. Pagination:**
- Use `nextPageToken` for search operations
- Use `startAt`/`maxResults` for other endpoints
- Implement sequential pagination loop

**7. Error Handling:**
- Check HTTP status codes
- Parse error responses
- Implement retry logic with exponential backoff
- Respect `Retry-After` for 429 responses

**8. Rate Limiting:**
- API token auth unaffected by 2026 changes
- Implement backoff and retry
- Limit concurrent requests (e.g., 5-10)

**9. Go Library:**
- Consider **andygrunwald/go-jira** for quick setup
- Or build custom client for full control
- Hybrid approach: use library + custom endpoints

**10. Testing:**
- Test against real Jira Cloud instance
- Handle pagination edge cases
- Test rate limiting behavior
- Validate ADF generation

### Key Endpoints Summary

```
# Authentication
Authorization: Basic base64(email:token)

# Issues
GET    /rest/api/3/issue/{issueIdOrKey}
POST   /rest/api/3/issue
PUT    /rest/api/3/issue/{issueIdOrKey}
POST   /rest/api/3/search/jql

# Projects
GET    /rest/api/3/project/search
GET    /rest/api/3/project/{projectIdOrKey}

# Boards & Sprints
GET    /rest/agile/1.0/board
GET    /rest/agile/1.0/board/{boardId}/sprint
GET    /rest/agile/1.0/sprint/{sprintId}

# Transitions
GET    /rest/api/3/issue/{issueIdOrKey}/transitions
POST   /rest/api/3/issue/{issueIdOrKey}/transitions

# Comments
GET    /rest/api/3/issue/{issueIdOrKey}/comment
POST   /rest/api/3/issue/{issueIdOrKey}/comment

# Fields & Statuses
GET    /rest/api/3/field/search
GET    /rest/api/3/status
GET    /rest/api/3/project/{projectIdOrKey}/statuses
```

### 2026 Timeline Awareness

**March 2, 2026:**
- New points-based and tiered quota rate limits enforced (Forge/Connect/OAuth apps)
- **API token auth unaffected**

**March 14 - May 12, 2026:**
- API tokens expire (prompt users to regenerate)

**June 1, 2026:**
- Transition property endpoints removed
- Use bulk workflow endpoints instead

**June 26, 2026:**
- Old workflow editor removed
- New workflow editor only

---

## Sources

### API Versions & Authentication
- [The Jira Cloud platform REST API v3](https://developer.atlassian.com/cloud/jira/platform/rest/v3/intro/)
- [Jira Cloud Platform REST API v2](https://developer.atlassian.com/cloud/jira/platform/rest/v2/intro/)
- [Basic auth for REST APIs](https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/)
- [Manage API tokens for your Atlassian account](https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/)

### Issue Operations & Search
- [The Jira Cloud platform REST API - Issue Search](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-search/)
- [Run JQL search query using Jira Cloud REST API](https://confluence.atlassian.com/jirakb/run-jql-search-query-using-jira-cloud-rest-api-1289424308.html)
- [Jira REST API examples](https://developer.atlassian.com/server/jira/platform/jira-rest-api-examples/)
- [Atlassian REST API Search Endpoints Deprecation](https://docs.adaptavist.com/sr4jc/latest/release-notes/breaking-changes/atlassian-rest-api-search-endpoints-deprecation)

### Issue Types & Hierarchy
- [What are work types? | Atlassian Support](https://support.atlassian.com/jira-cloud-administration/docs/what-are-issue-types/)
- [Jira Issue Types: A Complete Guide for 2026](https://community.atlassian.com/forums/App-Central-articles/Jira-Issue-Types-A-Complete-Guide-for-2026/ba-p/2928042)
- [Mastering Jira: Understanding Issue Types & Hierarchies](https://www.salto.io/blog-posts/jira-issue-type-understanding)
- [Jira Epic issues have subtasks under them along with child issues](https://support.atlassian.com/jira/kb/jira-epic-issues-have-subtasks-under-them-along-with-child-issues/)

### Projects, Boards & Sprints
- [The Jira Cloud platform REST API - Projects](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-projects/)
- [The Jira Software Cloud REST API - Sprint](https://developer.atlassian.com/cloud/jira/software/rest/api-group-sprint/)
- [Creating An Issue In A Sprint Using The JIRA REST API](https://support.atlassian.com/jira/kb/creating-an-issue-in-a-sprint-using-the-jira-rest-api/)

### Transitions & Workflows
- [The Jira Cloud platform REST API - Workflow Transition Properties](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-workflow-transition-properties/)
- [Jira Cloud - REST API - Transitions](https://community.developer.atlassian.com/t/jira-cloud-rest-api-transitions/73400)

### Comments & ADF
- [Atlassian Document Format](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/)
- [Jira Cloud REST API: Unable to add comment via ADF](https://community.atlassian.com/forums/Jira-questions/Jira-Cloud-REST-API-Unable-to-add-comment-via-ADF-receiving-quot/qaq-p/2808955)
- [Post HTML Issue Description with JIRA REST API v3](https://community.developer.atlassian.com/t/post-html-issue-description-with-jira-rest-api-v3/38482)

### Custom Fields
- [Get custom field IDs for Jira and Jira Service Management](https://confluence.atlassian.com/jirakb/get-custom-field-ids-for-jira-and-jira-service-management-744522503.html)
- [The Jira Cloud platform REST API - Issue Fields](https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-issue-fields/)

### Pagination
- [How to use the new Jira cloud issue search API](https://community.atlassian.com/forums/Jira-articles/How-to-use-the-new-Jira-cloud-issue-search-API/ba-p/3006109)
- [Jira Cloud REST API v3 /search/jql: Slower Fetching with nextPageToken](https://community.developer.atlassian.com/t/jira-cloud-rest-api-v3-search-jql-slower-fetching-with-nextpagetoken-no-totalissues-any-workarounds/90176)
- [Optimizing Data Retrieval with /rest/api/3/search/jql Pagination](https://community.atlassian.com/forums/Jira-questions/Optimizing-Data-Retrieval-with-rest-api-3-search-jql-Pagination/qaq-p/2957628)

### Error Handling
- [REST API Error responses](https://community.developer.atlassian.com/t/rest-api-error-responses/1668)
- [WebHooks or Web Requests fail with HTTP Status Code 400, 401 or 403](https://support.atlassian.com/jira/kb/webhooks-or-web-requests-fail-with-http-status-code-400-401-or-403-in-jira/)
- [How to handle HTTP 400 Bad Request errors on Jira search REST API endpoint](https://support.atlassian.com/jira/kb/how-to-handle-http-400-bad-request-errors-on-jira-search-rest-api-endpoint/)

### Go Libraries
- [GitHub - andygrunwald/go-jira](https://github.com/andygrunwald/go-jira)
- [jira package - github.com/andygrunwald/go-jira](https://pkg.go.dev/github.com/andygrunwald/go-jira)

### Rate Limiting
- [Rate limiting - Jira Cloud platform](https://developer.atlassian.com/cloud/jira/platform/rate-limiting/)
- [Scaling responsibly: evolving our API rate limits](https://www.atlassian.com/blog/platform/evolving-api-rate-limits)
- [Action Required: Update your Apps to comply with Jira Cloud Burst API rate limits](https://community.developer.atlassian.com/t/action-required-update-your-apps-to-comply-with-jira-cloud-burst-api-rate-limits/97202)

### Statuses
- [The Jira Cloud platform REST API - Workflow Statuses](https://developer.atlassian.com/cloud/jira/platform/rest/v2/api-group-workflow-statuses/)
- [Jira API: Statuses](https://agiletechnicalexcellence.com/2024/04/12/jira-api-statuses.html)
- [Get All Statuses For Project](https://www.postman.com/api-evangelist/atlassian-jira/request/ap31dth/get-all-statuses-for-project)
