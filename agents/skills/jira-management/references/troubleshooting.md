# Troubleshooting

Common issues and solutions when working with `jira-mgmt` CLI.

---

## Authentication Issues

**Problem:** `401 Unauthorized`
**Solution:**
```bash
jira-mgmt auth
jira-mgmt config show
```

**Problem:** Redirects to login page (HTML instead of JSON)
**Cause:** Corporate Jira behind SSO/OAuth proxy (e.g. Keycloak). API calls intercepted before reaching Jira.
**Solution:**
- Connect to corporate VPN — SSO proxy may only intercept external traffic
- Use a Personal Access Token (PAT) with Bearer auth (no email)
- Ask admin if there's an internal API URL that bypasses SSO

---

## Invalid JQL

**Problem:** `400 Bad Request` with JQL
**Solution:**
- Test JQL in Jira UI Advanced Search first
- On Server/DC: `!=` may cause escaping issues; use `NOT status = "Done"` instead of `status != "Done"`
- Status names must match exactly (case and language)
- Check field names (quotes for custom fields)
- See `jql-patterns.md` for examples

---

## Transition Failed

**Problem:** Cannot transition to status
**Solution:**
```bash
# Check available transitions
jira-mgmt q 'get(PROJ-123){full}'

# Use exact status name (may be localized on Server/DC)
jira-mgmt transition PROJ-123 --to "In Progress"
```

---

## Description Not Showing

**Problem:** Description field is empty in `get` output
**Cause:** Server/DC v2 returns description as plain string, Cloud v3 as ADF. If deserialization fails, description may show as empty.
**Solution:** The CLI handles both formats automatically via `DescriptionText()`. If still empty — the issue genuinely has no description.

---

## Subtasks Not Showing

**Problem:** `get(KEY){full}` doesn't show subtasks
**Possible causes:**
1. Issue has no subtasks
2. Field not in preset — verify `full` preset includes `subtasks` in both `parser.go` and `selector.go` (see `dev-notes.md`)
3. Binary not rebuilt after code change — rebuild with `go build` and redeploy

---

**Document Version:** 1.0
**Last Updated:** 2026-02-12
