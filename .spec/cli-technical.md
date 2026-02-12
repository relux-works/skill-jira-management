# CLI Technical Design

## Architecture

```
Skill (SKILL.md) → CLI (jira-mgmt) → Library (internal/) → Jira Cloud REST API v3
```

Monorepo: library + CLI + skill in one repo.

## Agents-Facing API (ref: agent-facing-api skill)

Three-layer read/write model:

### Layer 1: Mini-Query DSL (primary agent reads)

```bash
jira-mgmt q '<query>'
```

Operations:
- `get(ISSUE-KEY) { fields }` — single issue lookup
- `list(project=X, type=epic, status=open) { overview }` — filtered listing
- `summary()` — project/board overview
- `search(jql="...") { fields }` — JQL search

Field presets: `minimal`, `default`, `overview`, `full`
Batching: semicolons between queries
Output: always JSON, field projection

### Layer 2: Scoped Grep

```bash
jira-mgmt grep "pattern" --scope issues
```

For unstructured text search across cached/local data.

### Layer 3: CLI Commands (writes)

```bash
jira-mgmt create epic --summary "..." --project KEY
jira-mgmt transition ISSUE-KEY --to "In Progress"
jira-mgmt comment ISSUE-KEY --body "..."
```

Human-facing output for writes, JSON for reads.

## Go Project Structure

```
.
├── cmd/
│   └── jira-mgmt/
│       └── main.go              # CLI entry point
├── internal/
│   ├── jira/                    # Jira API client library
│   │   ├── client.go            # HTTP client, auth, base URL
│   │   ├── issues.go            # CRUD operations on issues
│   │   ├── projects.go          # Project listing
│   │   ├── boards.go            # Board operations
│   │   ├── transitions.go       # Status transitions
│   │   ├── search.go            # JQL search
│   │   └── types.go             # Jira domain types
│   ├── config/                  # Local config & persistence
│   │   ├── config.go            # Config read/write
│   │   └── auth.go              # Secure token storage
│   ├── fields/                  # Field selection & projection (agents-facing)
│   │   ├── selector.go
│   │   └── presets.go
│   ├── query/                   # DSL parser & executor
│   │   ├── parser.go
│   │   ├── ops.go
│   │   └── batch.go
│   └── search/                  # Scoped grep
│       └── grep.go
├── agents/skills/jira-management/  # Skill
│   └── SKILL.md
├── .spec/                       # Specs
├── .task-board/                 # Project board
├── go.mod
└── go.sum
```

## Testing (ref: go-testing-tools skill)

- Library: `tuitestkit` from `github.com/ivalx1s/skill-go-testing-tools/tuitestkit`
- Reducer tests for any state machines
- Snapshot tests for CLI output (golden files in `testdata/snapshots/`)
- Mock patterns for Jira API client (interface extraction → mock impl)
- Update snapshots: `UPDATE_SNAPSHOTS=1 go test ./...`
- Coverage target: 80%+

## Config & Auth

### Config file (GLOBAL, per-user): `~/.config/jira-mgmt/config.yaml`

```yaml
active_project: "PROJ"
active_board: 42
locale: "en"           # en | ru — locale for content created in Jira
```

### Auth: `~/.config/jira-mgmt/credentials` (encrypted or keychain)

```
instance: https://mycompany.atlassian.net
email: user@company.com
token: <encrypted-api-token>
```

## DSL Reference Templates

Copy from `agent-facing-api` skill assets:
- `assets/dsl-parser.go` → adapt for Jira operations
- `assets/field-selector.go` → adapt for Jira issue fields
- `assets/scoped-grep.go` → adapt for local issue cache
- `assets/query-patterns.md` → adapt examples for Jira

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- `github.com/ivalx1s/skill-go-testing-tools/tuitestkit` — testing
- `github.com/charmbracelet/bubbletea` — if TUI needed later
- Standard library for HTTP, JSON, crypto
