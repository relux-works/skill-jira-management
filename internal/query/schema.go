// Schema setup: registers fields, presets, filterable/sortable fields,
// and operations on an agentquery.Schema[jira.Issue].
// Replaces the hand-rolled parser, field selector, and executor.

package query

import (
	"fmt"
	"strconv"

	"github.com/ivalx1s/skill-agent-facing-api/agentquery"
	"github.com/ivalx1s/skill-jira-management/internal/jira"
)

// JiraAPIFieldMap maps DSL field names to Jira REST API field names.
// "key" is always returned by the API, so it maps to "".
var JiraAPIFieldMap = map[string]string{
	"key":         "",
	"summary":     "summary",
	"status":      "status",
	"assignee":    "assignee",
	"type":        "issuetype",
	"priority":    "priority",
	"parent":      "parent",
	"description": "description",
	"labels":      "labels",
	"reporter":    "reporter",
	"created":     "created",
	"updated":     "updated",
	"project":     "project",
	"subtasks":    "subtasks",
}

// APIFieldsFromSelector returns the Jira REST API field names for the fields
// selected in the given FieldSelector. This optimizes API calls by requesting
// only the fields that will be projected.
func APIFieldsFromSelector(sel *agentquery.FieldSelector[jira.Issue]) []string {
	selected := sel.Fields()
	seen := make(map[string]bool, len(selected))
	var apiFields []string
	for _, f := range selected {
		apiField, ok := JiraAPIFieldMap[f]
		if !ok || apiField == "" {
			continue
		}
		if !seen[apiField] {
			seen[apiField] = true
			apiFields = append(apiFields, apiField)
		}
	}
	return apiFields
}

// NewSchema builds a fully configured agentquery.Schema[jira.Issue].
// The client, defaultProject, and defaultBoard are captured by operation closures.
func NewSchema(client *jira.Client, defaultProject string, defaultBoard int) *agentquery.Schema[jira.Issue] {
	schema := agentquery.NewSchema[jira.Issue]()

	// --- Fields ---
	schema.Field("key", func(i jira.Issue) any { return i.Key })
	schema.Field("summary", func(i jira.Issue) any { return i.Fields.Summary })
	schema.Field("status", func(i jira.Issue) any {
		if i.Fields.Status != nil {
			return i.Fields.Status.Name
		}
		return nil
	})
	schema.Field("assignee", func(i jira.Issue) any {
		if i.Fields.Assignee != nil {
			return i.Fields.Assignee.DisplayName
		}
		return nil
	})
	schema.Field("type", func(i jira.Issue) any { return i.Fields.IssueType.Name })
	schema.Field("priority", func(i jira.Issue) any {
		if i.Fields.Priority != nil {
			return i.Fields.Priority.Name
		}
		return nil
	})
	schema.Field("parent", func(i jira.Issue) any {
		if i.Fields.Parent != nil {
			return i.Fields.Parent.Key
		}
		return nil
	})
	schema.Field("description", func(i jira.Issue) any { return i.Fields.DescriptionText() })
	schema.Field("labels", func(i jira.Issue) any { return i.Fields.Labels })
	schema.Field("reporter", func(i jira.Issue) any {
		if i.Fields.Reporter != nil {
			return i.Fields.Reporter.DisplayName
		}
		return nil
	})
	schema.Field("created", func(i jira.Issue) any { return i.Fields.Created })
	schema.Field("updated", func(i jira.Issue) any { return i.Fields.Updated })
	schema.Field("project", func(i jira.Issue) any { return i.Fields.Project.Key })
	schema.Field("subtasks", func(i jira.Issue) any {
		if len(i.Fields.Subtasks) == 0 {
			return nil
		}
		subs := make([]map[string]any, len(i.Fields.Subtasks))
		for j, st := range i.Fields.Subtasks {
			sub := map[string]any{
				"key":     st.Key,
				"summary": st.Fields.Summary,
			}
			if st.Fields.Status != nil {
				sub["status"] = st.Fields.Status.Name
			}
			subs[j] = sub
		}
		return subs
	})

	// --- Presets ---
	schema.Preset("minimal", "key", "status")
	schema.Preset("default", "key", "summary", "status", "assignee")
	schema.Preset("overview", "key", "summary", "status", "assignee", "type", "priority", "parent")
	schema.Preset("full", "key", "summary", "status", "assignee", "type", "priority", "parent",
		"description", "labels", "reporter", "created", "updated", "project", "subtasks")

	// --- Default fields ---
	schema.DefaultFields("default")

	// --- Filterable fields ---
	agentquery.FilterableField(schema, "status", func(i jira.Issue) string {
		if i.Fields.Status != nil {
			return i.Fields.Status.Name
		}
		return ""
	})
	agentquery.FilterableField(schema, "assignee", func(i jira.Issue) string {
		if i.Fields.Assignee != nil {
			return i.Fields.Assignee.DisplayName
		}
		return ""
	})
	agentquery.FilterableField(schema, "type", func(i jira.Issue) string {
		return i.Fields.IssueType.Name
	})
	agentquery.FilterableField(schema, "priority", func(i jira.Issue) string {
		if i.Fields.Priority != nil {
			return i.Fields.Priority.Name
		}
		return ""
	})
	agentquery.FilterableField(schema, "project", func(i jira.Issue) string {
		return i.Fields.Project.Key
	})

	// --- Sortable fields ---
	agentquery.SortableField(schema, "key", func(i jira.Issue) string { return i.Key })
	agentquery.SortableField(schema, "summary", func(i jira.Issue) string { return i.Fields.Summary })
	agentquery.SortableField(schema, "status", func(i jira.Issue) string {
		if i.Fields.Status != nil {
			return i.Fields.Status.Name
		}
		return ""
	})
	agentquery.SortableField(schema, "assignee", func(i jira.Issue) string {
		if i.Fields.Assignee != nil {
			return i.Fields.Assignee.DisplayName
		}
		return ""
	})
	agentquery.SortableField(schema, "type", func(i jira.Issue) string {
		return i.Fields.IssueType.Name
	})
	agentquery.SortableField(schema, "priority", func(i jira.Issue) string {
		if i.Fields.Priority != nil {
			return i.Fields.Priority.Name
		}
		return ""
	})

	// --- Operations ---
	// Note: operations are closures that capture client, defaultProject, defaultBoard.
	// The schema's SetLoader is NOT used because each operation needs to call the Jira API
	// with different parameters (issue key, JQL, project filters, etc.).

	schema.OperationWithMetadata("get", opGet(client), agentquery.OperationMetadata{
		Description: "Fetch a single Jira issue by key",
		Parameters: []agentquery.ParameterDef{
			{Name: "key", Type: "string", Optional: false, Description: "Issue key (positional), e.g. PROJ-123"},
		},
		Examples: []string{
			"get(PROJ-123) { overview }",
			"get(PROJ-456) { full }",
		},
	})

	schema.OperationWithMetadata("list", opList(client, defaultProject, schema), agentquery.OperationMetadata{
		Description: "List issues with filters, sorting, and pagination",
		Parameters: []agentquery.ParameterDef{
			{Name: "project", Type: "string", Optional: true, Description: "Project key (defaults to configured project)"},
			{Name: "type", Type: "string", Optional: true, Description: "Issue type filter (epic, story, task, bug)"},
			{Name: "status", Type: "string", Optional: true, Description: "Status filter"},
			{Name: "sort_<field>", Type: "asc|desc", Optional: true, Description: "Sort by field (key, summary, status, assignee, type, priority)"},
			{Name: "skip", Type: "int", Optional: true, Default: 0, Description: "Skip first N items"},
			{Name: "take", Type: "int", Optional: true, Description: "Return at most N items"},
		},
		Examples: []string{
			"list() { overview }",
			"list(project=PROJ, type=epic) { minimal }",
			"list(status=open, sort_key=asc) { default }",
			"list(skip=10, take=5) { overview }",
		},
	})

	schema.OperationWithMetadata("count", opCount(client, defaultProject), agentquery.OperationMetadata{
		Description: "Count issues matching filters",
		Parameters: []agentquery.ParameterDef{
			{Name: "project", Type: "string", Optional: true, Description: "Project key (defaults to configured project)"},
			{Name: "type", Type: "string", Optional: true, Description: "Issue type filter"},
			{Name: "status", Type: "string", Optional: true, Description: "Status filter"},
		},
		Examples: []string{
			"count()",
			"count(status=done)",
			"count(project=PROJ, type=bug)",
		},
	})

	schema.OperationWithMetadata("summary", opSummary(client, defaultProject, defaultBoard), agentquery.OperationMetadata{
		Description: "Project/board overview with issue counts by status and type",
		Parameters: []agentquery.ParameterDef{
			{Name: "project", Type: "string", Optional: true, Description: "Project key (defaults to configured project)"},
			{Name: "board", Type: "int", Optional: true, Description: "Board ID (defaults to configured board)"},
		},
		Examples: []string{
			"summary()",
			"summary(project=PROJ)",
		},
	})

	schema.OperationWithMetadata("search", opSearch(client), agentquery.OperationMetadata{
		Description: "Search issues using raw JQL",
		Parameters: []agentquery.ParameterDef{
			{Name: "jql", Type: "string", Optional: false, Description: "JQL query string"},
		},
		Examples: []string{
			`search(jql="assignee = currentUser()") { default }`,
			`search(jql="project = PROJ AND status = Open") { overview }`,
		},
	})

	return schema
}

// --- Operation handlers (closures capturing client) ---

// opGet: get(ISSUE-KEY) { fields }
func opGet(client *jira.Client) agentquery.OperationHandler[jira.Issue] {
	return func(ctx agentquery.OperationContext[jira.Issue]) (any, error) {
		if len(ctx.Statement.Args) == 0 {
			return nil, &agentquery.Error{
				Code:    agentquery.ErrValidation,
				Message: "get requires an issue key argument",
			}
		}

		issueKey := ctx.Statement.Args[0].Value
		apiFields := APIFieldsFromSelector(ctx.Selector)

		issue, err := client.GetIssue(issueKey, apiFields)
		if err != nil {
			return nil, err
		}

		return ctx.Selector.Apply(*issue), nil
	}
}

// opList: list(project=X, type=epic, status=open) { fields }
func opList(client *jira.Client, defaultProject string, schema *agentquery.Schema[jira.Issue]) agentquery.OperationHandler[jira.Issue] {
	return func(ctx agentquery.OperationContext[jira.Issue]) (any, error) {
		opts := jira.ListIssuesOptions{}

		for _, arg := range ctx.Statement.Args {
			switch arg.Key {
			case "project":
				opts.ProjectKey = arg.Value
			case "type":
				opts.IssueType = arg.Value
			case "status":
				opts.Status = arg.Value
			case "":
				// Positional arg - treat as project key if none set
				if opts.ProjectKey == "" {
					opts.ProjectKey = arg.Value
				}
			}
		}

		if opts.ProjectKey == "" {
			opts.ProjectKey = defaultProject
		}
		if opts.ProjectKey == "" {
			return nil, &agentquery.Error{
				Code:    agentquery.ErrValidation,
				Message: "list requires a project (via argument or config)",
			}
		}

		apiFields := APIFieldsFromSelector(ctx.Selector)
		opts.Fields = apiFields

		issues, err := client.ListIssues(opts)
		if err != nil {
			return nil, err
		}

		// Sort after fetching, before pagination.
		if err := agentquery.SortSlice(issues, ctx.Statement.Args, schema.SortFields()); err != nil {
			return nil, err
		}

		// Apply skip/take pagination.
		page, err := agentquery.PaginateSlice(issues, ctx.Statement.Args)
		if err != nil {
			return nil, err
		}

		results := make([]map[string]any, 0, len(page))
		for _, issue := range page {
			results = append(results, ctx.Selector.Apply(issue))
		}
		return results, nil
	}
}

// opCount: count(project=X, type=epic, status=open)
func opCount(client *jira.Client, defaultProject string) agentquery.OperationHandler[jira.Issue] {
	return func(ctx agentquery.OperationContext[jira.Issue]) (any, error) {
		opts := jira.ListIssuesOptions{}

		for _, arg := range ctx.Statement.Args {
			switch arg.Key {
			case "project":
				opts.ProjectKey = arg.Value
			case "type":
				opts.IssueType = arg.Value
			case "status":
				opts.Status = arg.Value
			case "":
				if opts.ProjectKey == "" {
					opts.ProjectKey = arg.Value
				}
			}
		}

		if opts.ProjectKey == "" {
			opts.ProjectKey = defaultProject
		}
		if opts.ProjectKey == "" {
			return nil, &agentquery.Error{
				Code:    agentquery.ErrValidation,
				Message: "count requires a project (via argument or config)",
			}
		}

		// Only need status and issuetype for counting.
		opts.Fields = []string{"status", "issuetype"}

		issues, err := client.ListIssues(opts)
		if err != nil {
			return nil, err
		}

		return map[string]any{"count": len(issues)}, nil
	}
}

// opSummary: summary() or summary(project=X, board=42)
func opSummary(client *jira.Client, defaultProject string, defaultBoard int) agentquery.OperationHandler[jira.Issue] {
	return func(ctx agentquery.OperationContext[jira.Issue]) (any, error) {
		projectKey := defaultProject
		boardID := defaultBoard

		for _, arg := range ctx.Statement.Args {
			switch arg.Key {
			case "project":
				projectKey = arg.Value
			case "board":
				id, err := strconv.Atoi(arg.Value)
				if err != nil {
					return nil, &agentquery.Error{
						Code:    agentquery.ErrValidation,
						Message: fmt.Sprintf("invalid board ID: %s", arg.Value),
					}
				}
				boardID = id
			case "":
				if projectKey == "" {
					projectKey = arg.Value
				}
			}
		}

		result := map[string]any{}

		// Get project info.
		if projectKey != "" {
			proj, err := client.GetProject(projectKey)
			if err != nil {
				return nil, fmt.Errorf("getting project: %w", err)
			}
			result["project"] = map[string]any{
				"key":  proj.Key,
				"name": proj.Name,
				"type": proj.ProjectTypeKey,
			}

			// Count issues by status and type.
			issues, err := client.ListIssues(jira.ListIssuesOptions{
				ProjectKey: projectKey,
				Fields:     []string{"status", "issuetype"},
			})
			if err != nil {
				return nil, fmt.Errorf("listing issues: %w", err)
			}

			statusCounts := map[string]int{}
			typeCounts := map[string]int{}
			for _, issue := range issues {
				if issue.Fields.Status != nil {
					statusCounts[issue.Fields.Status.Name]++
				}
				typeCounts[issue.Fields.IssueType.Name]++
			}
			result["total_issues"] = len(issues)
			result["by_status"] = statusCounts
			result["by_type"] = typeCounts
		}

		// Get board info.
		if boardID != 0 {
			board, err := client.GetBoard(boardID)
			if err != nil {
				return nil, fmt.Errorf("getting board: %w", err)
			}
			result["board"] = map[string]any{
				"id":   board.ID,
				"name": board.Name,
				"type": board.Type,
			}
		}

		if len(result) == 0 {
			return nil, &agentquery.Error{
				Code:    agentquery.ErrValidation,
				Message: "summary requires a project or board (via argument or config)",
			}
		}

		return result, nil
	}
}

// opSearch: search(jql="...") { fields }
func opSearch(client *jira.Client) agentquery.OperationHandler[jira.Issue] {
	return func(ctx agentquery.OperationContext[jira.Issue]) (any, error) {
		var jql string
		for _, arg := range ctx.Statement.Args {
			switch arg.Key {
			case "jql":
				jql = arg.Value
			case "":
				if jql == "" {
					jql = arg.Value
				}
			}
		}
		if jql == "" {
			return nil, &agentquery.Error{
				Code:    agentquery.ErrValidation,
				Message: "search requires a jql argument",
			}
		}

		apiFields := APIFieldsFromSelector(ctx.Selector)

		issues, err := client.SearchAll(jql, apiFields)
		if err != nil {
			return nil, err
		}

		results := make([]map[string]any, 0, len(issues))
		for _, issue := range issues {
			results = append(results, ctx.Selector.Apply(issue))
		}
		return results, nil
	}
}

