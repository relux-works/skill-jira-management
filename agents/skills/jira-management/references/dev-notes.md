# Development Notes

Architecture notes for agents modifying the `jira-mgmt` CLI codebase.

---

## Duplicate Field Definitions — Keep in Sync

Field names, valid fields, and presets are defined in **two places**:

| File | What | Used by |
|------|------|---------|
| `internal/query/parser.go` | `ValidFields`, `FieldPresets` | DSL parser — decides which field names are valid in queries and what presets expand to |
| `internal/fields/selector.go` | `ValidFields`, `Presets`, `JiraAPIFields()` | Field selector — maps fields to Jira API names and renders them in output |

**When adding a new field:** update BOTH files. If you only update `selector.go`, the DSL parser will reject the field or the preset won't include it — and the field silently won't appear in output.

---

## Description Format (Cloud vs Server/DC)

`IssueFields.Description` is tagged `json:"-"` (not deserialized directly). Raw JSON goes into `DescriptionRaw json.RawMessage`, and `DescriptionText()` handles both formats:
- **Cloud v3:** ADF (JSON object) — parsed and text extracted
- **Server/DC v2:** plain string — used as-is

Never access `Description *ADFDoc` directly from deserialized JSON — use `DescriptionText()`.

---

## Nil Safety in Subtask Fields

Subtask entries from the API may have nil `Status`, `Priority`, etc. Always nil-check before accessing nested fields (e.g. `st.Fields.Status.Name`).

---

## API Path Selection

`client.apiPathFor(parts...)` automatically selects `/rest/api/3/` for Cloud and `/rest/api/2/` for Server/DC based on `instanceType`. All endpoint methods must use `c.apiPathFor()`, never hardcode API version.

---

## Pagination

Two pagination strategies handled by `SearchAll()`:
- **Cloud:** cursor-based (`nextPageToken`) — must pass token from previous response
- **Server/DC:** offset-based (`startAt`) — increment by `maxResults` each page

`ListIssues` delegates to `SearchAll` which handles both transparently.

---

## Key Files

| File | Purpose |
|------|---------|
| `internal/jira/client.go` | HTTP client, auth, retry, API path selection |
| `internal/jira/types.go` | All domain types, description deserialization |
| `internal/jira/issues.go` | Issue CRUD, search, pagination |
| `internal/jira/projects.go` | Project listing (Cloud paginated, Server full array) |
| `internal/query/parser.go` | DSL tokenizer + parser, field presets |
| `internal/query/ops.go` | DSL operation handlers (get, list, search, summary) |
| `internal/fields/selector.go` | Field selection, API field mapping, output projection |
| `internal/config/` | Config + credentials management |
| `cmd/jira-mgmt/` | Cobra commands |

---

**Document Version:** 1.0
**Last Updated:** 2026-02-12
