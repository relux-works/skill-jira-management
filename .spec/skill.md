# Skill Specification

## Purpose

Claude Code / Codex skill that drives `jira-mgmt` CLI for Jira Cloud operations.

## Triggers (ru + en)

| Trigger patterns (ru) | Trigger patterns (en) |
|------------------------|------------------------|
| jira, джира | jira |
| тикет, задача, тикеты, задачи | ticket, issue, issues |
| эпик, эпики | epic, epics |
| стори, сторя, стори | story, stories |
| борда, борды, доска | board, boards |
| спринт, спринты | sprint, sprints |
| создай задачу, заведи тикет | create issue, create ticket |
| двинь задачу, переведи статус | move issue, transition |
| покажи борду, покажи доску | show board |
| статус задачи | issue status |
| комментарий, коммент, добавь коммент | comment, add comment |
| поиск, найди, jql | search, find, jql |
| дод, definition of done | dod, definition of done |

## Skill Structure

```
agents/skills/jira-management/
├── SKILL.md              # Main skill file with CLI reference
├── references/
│   └── jql-patterns.md   # Common JQL query patterns
└── assets/               # If needed
```

Symlinks:
- `.claude/skills/jira-management` → `agents/skills/jira-management`
- `.codex/skills/jira-management` → `agents/skills/jira-management`

## Skill Behavior

- Translates natural language intent → CLI commands
- Uses DSL layer for reads (token-efficient)
- Uses CLI commands for writes
- Handles multi-step workflows (create epic → create stories → link)
- Respects locale from config for all content creation
