# CLI Commands Reference

Complete reference for `jira-mgmt` CLI commands with all options and examples.

---

## Authentication & Configuration

### jira-mgmt auth

Authentication setup. Supports both interactive and non-interactive (agent-friendly) modes.

**Non-interactive (recommended for agents):**
```bash
# Cloud (Basic auth — email + API token)
jira-mgmt auth --instance https://mycompany.atlassian.net --email user@company.com --token API_TOKEN

# Server/DC (Bearer auth — Personal Access Token, no email)
jira-mgmt auth --instance https://jira.company.com --token PAT_TOKEN
```

**Interactive (prompts for input):**
```bash
jira-mgmt auth
```

**Flags:**
- `--instance URL` — Jira instance URL
- `--email EMAIL` — Account email (omit for Server/DC PAT auth)
- `--token TOKEN` — API token (Cloud) or PAT (Server/DC)

**Auth type detection:**
- Email provided → Basic auth (Cloud: `base64(email:token)`)
- No email → Bearer auth (Server/DC: `Bearer <token>`)

**Instance type detection:**
During auth, the CLI probes `/rest/api/2/serverInfo` to auto-detect Cloud vs Server/DC. This determines which API version (v2/v3) and pagination style to use.

**Storage:**
- Credentials: OS keychain (macOS Keychain / Linux Secret Service)
- Config: `~/.config/jira-mgmt/config.yaml`

**What gets saved:**
- `instance_url` — Jira base URL
- `instance_type` — `cloud` or `server` (auto-detected)
- `auth_type` — `basic` or `bearer` (auto-determined)

**Notes:**
- Cloud: create API token at https://id.atlassian.com/manage-profile/security/api-tokens
- Server/DC: create PAT in Jira → Profile → Personal Access Tokens
- Corporate Jira behind SSO/OAuth proxy may require VPN for API access
- One-time setup per environment

---

### jira-mgmt config set

Set configuration values.

**Syntax:**
```bash
jira-mgmt config set <key> <value>
```

**Keys:**
- `project` — default project key (e.g., `PROJ`, `ACME`)
- `board` — default board ID (numeric, e.g., `123`)
- `locale` — locale for content creation (e.g., `en-US`, `ru-RU`, `hy-AM`)

**Examples:**
```bash
# Set default project
jira-mgmt config set project ACME

# Set default board
jira-mgmt config set board 456

# Set locale to Russian
jira-mgmt config set locale ru-RU

# Set locale to English
jira-mgmt config set locale en-US
```

**Notes:**
- Config persists across sessions
- Can override per-command with flags
- Locale affects summary/description language for created issues

---

### jira-mgmt config show

Display current configuration.

**Syntax:**
```bash
jira-mgmt config show
```

**Example Output:**
```
Jira Configuration
------------------
Instance: https://acme.atlassian.net
Email: user@example.com
Project: ACME
Board: 456
Locale: en-US
```

---

## Query Commands (DSL)

### jira-mgmt q

Execute DSL queries. Token-efficient for reads.

**Syntax:**
```bash
jira-mgmt q '<query>'
```

**Query Types:**

#### 1. get(ISSUE-KEY){preset}

Get single issue.

**Presets:**
- `minimal` — key, status
- `default` — key, summary, status, assignee
- `overview` — + type, priority, parent
- `full` — all fields including subtasks (key, summary, status per subtask)

**Examples:**
```bash
# Minimal (default)
jira-mgmt q 'get(PROJ-123)'

# With full details
jira-mgmt q 'get(PROJ-123){full}'
```

**Output (minimal):**
```
PROJ-123: User Authentication
Status: In Progress
```

---

#### 2. list(filters){preset}

List multiple issues.

**Filters (comma-separated):**
- `sprint=current|ID` — filter by sprint
- `assignee=me|email` — filter by assignee
- `status=!done|in-progress|todo` — filter by status (use `!` for negation)
- `type=epic|story|task|bug|subtask` — filter by issue type

**Examples:**
```bash
# Current sprint
jira-mgmt q 'list(sprint=current){default}'

# My open issues
jira-mgmt q 'list(assignee=me,status=!done){default}'

# Epics only
jira-mgmt q 'list(type=epic){overview}'
```

---

#### 3. summary()

Board statistics and summary.

**Example:**
```bash
jira-mgmt q 'summary()'
```

---

#### 4. search(jql="..."){preset}

JQL search.

**Examples:**
```bash
# All epics
jira-mgmt q 'search(jql="project=PROJ AND issuetype=Epic"){overview}'

# Overdue issues
jira-mgmt q 'search(jql="due<now() AND statusCategory!=Done"){default}'
```

See `jql-patterns.md` for comprehensive JQL examples.

---

#### Batch Queries

Execute multiple queries with `;` separator.

**Examples:**
```bash
# Summary + current sprint
jira-mgmt q 'summary(); list(sprint=current){default}'
```

---

## Search Commands

### jira-mgmt grep

Search issues and comments with regex.

**Syntax:**
```bash
jira-mgmt grep <pattern> [flags]
```

**Flags:**
- `--scope <issues|comments|all>` — search scope (default: `all`)
- `-i` — case-insensitive
- `-C <num>` — context lines around match

**Examples:**
```bash
# Search everywhere
jira-mgmt grep "authentication"

# Case-insensitive, issues only
jira-mgmt grep -i "AUTH" --scope issues

# With context
jira-mgmt grep -C 2 "performance"
```

---

## Create Commands

### jira-mgmt create

Create new issue.

**Syntax:**
```bash
jira-mgmt create --type <type> --summary "..." [options]
```

**Required:**
- `--type <epic|story|task|subtask|bug>` — issue type
- `--summary "..."` — issue title
- `--project KEY` — project key (or use default from config)

**Optional:**
- `--description "..."` — issue description
- `--parent ISSUE-KEY` — parent issue (required for `subtask`)
- `--assignee email@example.com` — assignee email
- `--priority <highest|high|medium|low|lowest>` — priority level
- `--labels "label1,label2"` — comma-separated labels

**Examples:**

#### Epic
```bash
jira-mgmt create --type epic --summary "User Authentication System" --description "Implement OAuth2 and local authentication" --project PROJ
```

#### Story
```bash
jira-mgmt create --type story --summary "Login page UI" --parent PROJ-123 --project PROJ
```

#### Task
```bash
jira-mgmt create --type task --summary "Write unit tests" --project PROJ
```

#### Subtask
```bash
jira-mgmt create --type subtask --summary "Add password validation" --parent PROJ-125 --project PROJ
```

#### Bug
```bash
jira-mgmt create --type bug --summary "Login fails with special characters" --priority high --labels "security,authentication" --project PROJ
```

---

## Update Commands

### jira-mgmt update

Update issue fields (summary, description).

**Syntax:**
```bash
jira-mgmt update ISSUE-KEY [flags]
```

**Flags:**
- `--summary "..."` — new issue summary/title
- `--description "..."` — new issue description

At least one flag is required.

**Examples:**
```bash
# Update summary only
jira-mgmt update PROJ-123 --summary "New title"

# Update description only
jira-mgmt update PROJ-123 --description "New description text"

# Update both
jira-mgmt update PROJ-123 --summary "New title" --description "Updated description"
```

**Notes:**
- Description format is handled automatically: plain string for Server/DC, ADF for Cloud
- Useful for bulk cleanup of subtask titles/descriptions

---

### jira-mgmt transition

Move issue to different status.

**Syntax:**
```bash
jira-mgmt transition ISSUE-KEY --to "Status Name"
```

**Examples:**
```bash
# Start work
jira-mgmt transition PROJ-123 --to "In Progress"

# Mark done
jira-mgmt transition PROJ-123 --to "Done"
```

**Notes:**
- Status name must match workflow exactly (case-sensitive)
- Check available transitions: `jira-mgmt q 'get(ISSUE-KEY){full}'`

---

### jira-mgmt comment

Add comment to issue.

**Syntax:**
```bash
jira-mgmt comment ISSUE-KEY --body "text"
```

**Examples:**
```bash
# Simple comment
jira-mgmt comment PROJ-123 --body "Started implementation"

# Multi-line comment
jira-mgmt comment PROJ-123 --body "Implementation notes:
- Added OAuth2 provider
- Configured redirect URIs
- Updated tests"
```

---

### jira-mgmt dod

Set Definition of Done criteria.

**Syntax:**
```bash
jira-mgmt dod ISSUE-KEY --set "criteria"
```

**Example:**
```bash
jira-mgmt dod PROJ-123 --set "Unit tests pass
E2E tests pass
Code reviewed
Documentation updated"
```

---

## Global Flags

All commands support:
- `--project KEY` — override default project
- `--board ID` — override default board
- `--format <json|text>` — output format (default: `text`)

**Examples:**
```bash
# JSON output for scripting
jira-mgmt q 'get(PROJ-123)' --format json

# Override project
jira-mgmt create --type task --summary "Test" --project TEMP
```

---

## Version

### jira-mgmt version

Display CLI version.

**Syntax:**
```bash
jira-mgmt version
```

---

**Document Version:** 1.0
**Last Updated:** 2026-02-12
