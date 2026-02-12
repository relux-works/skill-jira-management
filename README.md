# jira-management

CLI tool + Claude Code skill for working with Jira Cloud projects.

## Architecture

```
Skill (SKILL.md) → CLI (jira-mgmt) → Library (internal/) → Jira Cloud REST API v3
```

**Monorepo** with three components:

| Component | Location | Purpose |
|-----------|----------|---------|
| Go library | `internal/jira/` | Jira Cloud REST API v3 client |
| Config | `internal/config/` | Keychain auth storage + YAML config |
| CLI tool | `cmd/jira-mgmt/` | Agents-facing CLI with DSL query layer |
| Skill | `agents/skills/jira-management/` | Claude Code / Codex skill |

## Setup

```bash
# Build binary, create symlinks
./scripts/setup.sh

# Authenticate with Jira Cloud
jira-mgmt auth

# Set active project
jira-mgmt config set project YOUR-KEY
```

### Deinit

```bash
# Remove symlinks (keep config)
./scripts/deinit.sh

# Remove symlinks + config
./scripts/deinit.sh --purge
```

## CLI Reference

### Auth & Config

```bash
jira-mgmt auth                        # Interactive setup: instance URL, email, API token
jira-mgmt config set project KEY       # Set active project
jira-mgmt config set board 42          # Set active board
jira-mgmt config set locale en         # Set locale (en/ru)
jira-mgmt config show                  # Show current config
```

### DSL Queries (agents-facing reads)

```bash
jira-mgmt q 'get(PROJ-123) { default }'                           # Single issue
jira-mgmt q 'list(project=PROJ, type=epic) { overview }'          # List epics
jira-mgmt q 'summary()'                                            # Project overview
jira-mgmt q 'search(jql="assignee=currentUser()") { minimal }'    # JQL search
jira-mgmt q 'get(A-1) { status }; get(A-2) { status }'            # Batch queries
```

Field presets: `minimal` (key, status), `default` (+summary, assignee), `overview` (+type, priority, parent), `full` (all fields).

### Grep (text search)

```bash
jira-mgmt grep "pattern"                    # Search all
jira-mgmt grep "bug" --scope issues -i      # Search issues, case-insensitive
jira-mgmt grep "deploy" --scope comments    # Search comments
```

### Write Operations

```bash
jira-mgmt create --type epic --summary "New feature" --project KEY
jira-mgmt create --type story --summary "Auth flow" --parent PROJ-10
jira-mgmt create --type task --summary "Implement login" --parent PROJ-11
jira-mgmt transition PROJ-123 --to "In Progress"
jira-mgmt comment PROJ-123 --body "Done, ready for review"
jira-mgmt dod PROJ-123 --set "- Unit tests\n- Code review\n- Docs updated"
```

### Global Flags

```bash
--project KEY    # Override active project
--board ID       # Override active board
--format json    # Output format (json/text)
```

## Testing

```bash
go test ./... -v                                   # Run all tests
go test ./internal/jira/... -v                     # Library tests only
go test ./internal/config/... -v                   # Config tests only
go test ./cmd/jira-mgmt/... -v                     # CLI tests only
go test ./... -cover                               # Coverage report
UPDATE_SNAPSHOTS=1 go test ./... -v                # Update golden files
```

## Project Structure

```
.
├── cmd/jira-mgmt/          # CLI entry point + commands
├── internal/
│   ├── jira/               # Jira Cloud API client library
│   ├── config/             # Auth (keychain) + config (YAML)
│   ├── query/              # DSL parser & executor
│   ├── fields/             # Field selection & projection
│   └── search/             # Scoped grep
├── agents/skills/jira-management/   # Skill
│   ├── SKILL.md
│   └── references/
│       ├── cli-commands.md
│       ├── dsl-examples.md
│       ├── jql-patterns.md
│       └── workflows.md
├── scripts/
│   ├── setup.sh            # Build + install
│   └── deinit.sh           # Uninstall
├── .spec/                  # Project specifications
├── .task-board/            # Task board (project management)
├── .research/              # Research artifacts
└── .planning/              # Planning artifacts
```

## Tools

| Tool | Purpose | How to run |
|------|---------|-----------|
| `jira-mgmt` | CLI for Jira Cloud | `jira-mgmt --help` |
| `go test` | Run tests | `go test ./... -v` |
| `task-board` | Project management board | `task-board summary` |
| `scripts/setup.sh` | Build and install | `./scripts/setup.sh` |
| `scripts/deinit.sh` | Uninstall | `./scripts/deinit.sh [--purge]` |

## Config Location

- **Config:** `~/.config/jira-mgmt/config.yaml` (global, per-user)
- **Credentials:** OS keychain (macOS Keychain, Linux Secret Service)
- **Binary:** `~/.local/bin/jira-mgmt` (symlink)
- **Skill:** `~/.claude/skills/jira-management` + `~/.codex/skills/jira-management` (symlinks)

## License

MIT
