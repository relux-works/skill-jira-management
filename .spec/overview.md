# Jira Management — Overview

CLI tool + Claude Code skill for working with Jira projects.

## Components

1. **Go library** — core Jira API client, shared logic
2. **CLI tool** — agents-facing command-line interface, built on top of the library
3. **Skill** — Claude Code / Codex skill that drives the CLI tool

Architecture: `Skill → CLI → Library → Jira REST API`

## Feature Directions

### 1. Auth & Token Storage

- **Jira Cloud only** (REST API v3 / v2)
- Auth: email + API token (Atlassian API key)
- Secure local storage for the token (keychain / encrypted file — NOT plaintext)

### 2. Local Config & Persistence

- Persistent local config (project focus, board focus, locale, etc.)
- Set "active project" — all commands default to it
- Set "active board" — same idea
- Locale setting (ru / en) — all content created in Jira uses this locale

### 3. Read Operations

- List projects accessible to the user
- List epics, stories, tasks for a specific project
- View different boards for a project
- See available statuses/transitions for a task

### 4. Write Operations

- Create issues: epics, stories, subtasks (various types under stories)
- Move tasks through statuses (transitions)
- Add comments to issues
- Write/update Definition of Done (DoD) on tasks

### 5. Search

- JQL search support via CLI
- Agent can construct JQL queries from natural language intent

### 6. Board Focus

- List boards for a project
- Set active board in persistent config
- Commands operate within focused board context

## Technical Stack

### CLI Tool

- **Language:** Go
- **Testing:** `go-testing-tools` skill + its library (tuitestkit)
- **API layer:** agents-facing design (ref: `agent-facing-api` skill)
- **Dependency:** Go library (monorepo — library + CLI + skill in one repo)

### Skill

- Knows CLI commands and their arguments
- Translates user intent into CLI calls
- Handles multi-step workflows (e.g., create epic → create stories under it)
- **Triggers (ru + en):** jira, тикет, задача, эпик, стори, борда, спринт, issue, epic, story, board, sprint, создай задачу, двинь задачу, покажи борду, статус задачи, etc.

## Decisions

- **CLI binary:** `jira-mgmt`
- **Repo structure:** monorepo (library + CLI + skill)
- **JQL search:** yes, in scope
