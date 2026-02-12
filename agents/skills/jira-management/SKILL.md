---
name: jira-management
description: Drive jira-mgmt CLI for Jira operations (Cloud & Server/DC, auto-detected). Translates natural language intent to CLI commands. Uses DSL for reads (token-efficient), CLI for writes. Handles multi-step workflows (create epic with stories, bulk transitions, sprint reviews).
triggers:
  - jira
  - джира
  - ticket
  - тикет
  - issue
  - задача
  - issues
  - задачи
  - epic
  - эпик
  - epics
  - эпики
  - story
  - стори
  - stories
  - сторя
  - board
  - борда
  - boards
  - борды
  - доска
  - sprint
  - спринт
  - sprints
  - спринты
  - create issue
  - создай задачу
  - create ticket
  - заведи тикет
  - move issue
  - двинь задачу
  - transition
  - переведи статус
  - show board
  - покажи борду
  - покажи доску
  - issue status
  - статус задачи
  - comment
  - комментарий
  - коммент
  - add comment
  - добавь коммент
  - search
  - поиск
  - find
  - найди
  - jql
  - dod
  - дод
  - definition of done
---

# Jira Management Skill

**Purpose:** Agent-agnostic skill (Claude Code / Codex CLI) for managing Jira (Cloud & Server/DC) via `jira-mgmt` CLI. Instance type is auto-detected during `auth`.

**Tool:** `jira-mgmt` (installed via `scripts/setup.sh`)

---

## Quick Start

### 1. Setup

**Cloud (email + API token → Basic auth):**
```bash
jira-mgmt auth --instance https://mycompany.atlassian.net --email user@company.com --token API_TOKEN
```

**Server/DC (Personal Access Token → Bearer auth):**
```bash
jira-mgmt auth --instance https://jira.company.com --token PAT_TOKEN
```

Auth type is auto-determined: email provided → Basic (Cloud), no email → Bearer (Server/DC PAT).
Instance type (Cloud vs Server/DC) is auto-detected via `/rest/api/2/serverInfo` during auth.

```bash
# Set active project/board/locale
jira-mgmt config set project YOUR-KEY
jira-mgmt config set board 123
jira-mgmt config set locale en
```

### 2. Basic Operations

**Read (use DSL for token efficiency):**
```bash
# Single issue
jira-mgmt q 'get(PROJ-123)'

# List current sprint
jira-mgmt q 'list(sprint=current){default}'

# Search with JQL
jira-mgmt q 'search(jql="assignee=currentUser() AND statusCategory!=Done"){default}'
```

**Write (use CLI commands):**
```bash
# Create issue
jira-mgmt create --type story --summary "Login UI" --project PROJ

# Transition
jira-mgmt transition PROJ-123 --to "In Progress"

# Comment
jira-mgmt comment PROJ-123 --body "Started work"

# Set DoD
jira-mgmt dod PROJ-123 --set "Tests pass\nCode reviewed"
```

---

## Commands Overview

### Authentication & Config
- `jira-mgmt auth` — interactive setup; or non-interactive with flags:
  - Cloud: `jira-mgmt auth --instance URL --email EMAIL --token API_TOKEN`
  - Server/DC: `jira-mgmt auth --instance URL --token PAT` (no email = Bearer auth)
- `jira-mgmt config set <key> <value>` — set project/board/locale
- `jira-mgmt config show` — display current config (includes instance type, auth type)

### Queries (DSL)
- `jira-mgmt q 'get(KEY){preset}'` — single issue
- `jira-mgmt q 'list(filters){preset}'` — multiple issues
- `jira-mgmt q 'summary()'` — board statistics
- `jira-mgmt q 'search(jql="..."){preset}'` — JQL search

**Presets:** `minimal`, `default`, `overview`, `full` (includes subtasks)

**List Filters:**
- `sprint=current|ID`
- `assignee=me|email`
- `status=!done|in-progress|todo`
- `type=epic|story|task|bug`

**Batch queries:** Use `;` separator

### Search
- `jira-mgmt grep <pattern>` — search issues/comments
  - Flags: `--scope issues|comments|all`, `-i`, `-C <num>`

### Create
- `jira-mgmt create --type <type> --summary "..." --project KEY`
  - Types: `epic`, `story`, `task`, `subtask`, `bug`
  - Optional: `--description`, `--parent`, `--assignee`, `--priority`, `--labels`

### Update
- `jira-mgmt update ISSUE-KEY --summary "..." --description "..."` — update issue fields
- `jira-mgmt transition ISSUE-KEY --to "Status Name"` — move to status
- `jira-mgmt comment ISSUE-KEY --body "text"` — add comment
- `jira-mgmt dod ISSUE-KEY --set "criteria"` — set Definition of Done

### Global Flags
- `--project KEY` — override default project
- `--board ID` — override default board
- `--format json|text` — output format

---

## Agent Usage Patterns

### Natural Language → CLI

**User:** "Show me all epics in the current sprint"
```bash
jira-mgmt q 'list(sprint=current,type=epic){overview}'
```

**User:** "Create a story for login UI under epic PROJ-100"
```bash
jira-mgmt create --type story --summary "Login UI implementation" --parent PROJ-100 --project PROJ
```

**User:** "Find all issues mentioning 'performance'"
```bash
jira-mgmt grep -i "performance"
```

**User:** "Move PROJ-123 to done"
```bash
jira-mgmt transition PROJ-123 --to "Done"
```

### Read Operations: Use DSL

Prefer DSL over JQL for token efficiency.

**Good (DSL):**
```bash
jira-mgmt q 'list(sprint=current,assignee=me){default}'
```

**Use JQL only when:**
- DSL filters insufficient
- Complex multi-condition queries needed
- Historical/advanced queries required

### Write Operations: Use CLI

Always use explicit CLI commands for modifications:
```bash
jira-mgmt create --type story --summary "..." --project PROJ
jira-mgmt transition PROJ-123 --to "In Progress"
jira-mgmt comment PROJ-123 --body "..."
```

### Locale-Aware Content

Respect `locale` config for all content creation:
```bash
# Check locale
LOCALE=$(jira-mgmt config show | grep Locale | awk '{print $2}')

# Create in configured locale
if [[ "$LOCALE" == "ru-RU" ]]; then
  jira-mgmt create --type task --summary "Добавить тесты" --project PROJ
else
  jira-mgmt create --type task --summary "Add tests" --project PROJ
fi
```

---

## Cloud vs Server/DC Differences

The CLI auto-detects instance type and adapts. Key differences to be aware of:

| Aspect | Cloud | Server/DC |
|--------|-------|-----------|
| Auth | Basic (email + API token) | Bearer (Personal Access Token) |
| API version | v3 (`/rest/api/3/`) | v2 (`/rest/api/2/`) |
| Search pagination | Cursor-based (`nextPageToken`) | Offset-based (`startAt`) |
| Descriptions/comments | ADF (Atlassian Document Format) | ADF (Jira 8.x+) or wiki markup |
| Project listing | Paginated `/project/search` | `/project` returns full array |
| Status names | English by default | May be localized (e.g. Russian) |

**Status names on Server/DC** may be in the instance's language. Use exact names as returned by Jira:
```bash
# Check actual status names for an issue
jira-mgmt q 'get(PROJ-123){full}'
# Transition using the exact name
jira-mgmt transition PROJ-123 --to "В работе"
```

---

## References

- `references/cli-commands.md` — Complete CLI command reference with all options
- `references/dsl-examples.md` — Comprehensive DSL query patterns and examples
- `references/workflows.md` — Multi-step workflow patterns (sprint review, daily standup, bulk ops)
- `references/jql-patterns.md` — JQL query patterns for advanced searches
- `references/troubleshooting.md` — Common issues and solutions (auth, JQL, transitions)
- `references/dev-notes.md` — Architecture notes for agents modifying the CLI codebase

---

**Skill Version:** 1.2
**Last Updated:** 2026-02-12
**Tool:** jira-mgmt CLI
