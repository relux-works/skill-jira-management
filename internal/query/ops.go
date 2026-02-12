// Operation handlers for DSL queries.
// Each operation maps to Jira API calls via the jira client library.

package query

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ivalx1s/skill-jira-management/internal/fields"
	"github.com/ivalx1s/skill-jira-management/internal/jira"
)

// Executor runs parsed DSL queries against a Jira client.
type Executor struct {
	client         *jira.Client
	defaultProject string
	defaultBoard   int
}

// NewExecutor creates a new query executor.
func NewExecutor(client *jira.Client, defaultProject string, defaultBoard int) *Executor {
	return &Executor{
		client:         client,
		defaultProject: defaultProject,
		defaultBoard:   defaultBoard,
	}
}

// Execute runs a parsed Query and returns JSON-encoded results.
func (e *Executor) Execute(q *Query) ([]json.RawMessage, error) {
	results := make([]json.RawMessage, 0, len(q.Statements))
	for _, stmt := range q.Statements {
		result, err := e.executeStatement(&stmt)
		if err != nil {
			return nil, fmt.Errorf("operation %s: %w", stmt.Operation, err)
		}
		results = append(results, result)
	}
	return results, nil
}

func (e *Executor) executeStatement(stmt *Statement) (json.RawMessage, error) {
	switch stmt.Operation {
	case "get":
		return e.opGet(stmt)
	case "list":
		return e.opList(stmt)
	case "summary":
		return e.opSummary(stmt)
	case "search":
		return e.opSearch(stmt)
	default:
		return nil, fmt.Errorf("unknown operation: %s", stmt.Operation)
	}
}

// opGet: get(ISSUE-KEY) { fields }
func (e *Executor) opGet(stmt *Statement) (json.RawMessage, error) {
	if len(stmt.Args) < 1 {
		return nil, fmt.Errorf("get requires an issue key argument")
	}
	issueKey := stmt.Args[0].Value

	sel, err := fields.NewSelector(stmt.Fields)
	if err != nil {
		return nil, err
	}

	issue, err := e.client.GetIssue(issueKey, sel.JiraAPIFields())
	if err != nil {
		return nil, err
	}

	projected := sel.Apply(issue)
	return json.Marshal(projected)
}

// opList: list(project=X, type=epic, status=open) { fields }
func (e *Executor) opList(stmt *Statement) (json.RawMessage, error) {
	opts := jira.ListIssuesOptions{}

	for _, arg := range stmt.Args {
		switch arg.Key {
		case "project":
			opts.ProjectKey = arg.Value
		case "type":
			opts.IssueType = arg.Value
		case "status":
			opts.Status = arg.Value
		case "":
			// Positional arg — treat as project key if none set
			if opts.ProjectKey == "" {
				opts.ProjectKey = arg.Value
			}
		}
	}

	// Default to configured project
	if opts.ProjectKey == "" {
		opts.ProjectKey = e.defaultProject
	}
	if opts.ProjectKey == "" {
		return nil, fmt.Errorf("list requires a project (via argument or config)")
	}

	sel, err := fields.NewSelector(stmt.Fields)
	if err != nil {
		return nil, err
	}
	opts.Fields = sel.JiraAPIFields()

	issues, err := e.client.ListIssues(opts)
	if err != nil {
		return nil, err
	}

	projected := sel.ApplyMany(issues)
	return json.Marshal(projected)
}

// opSummary: summary() or summary(project=X)
// Returns project overview: project info, board info, issue counts by status.
func (e *Executor) opSummary(stmt *Statement) (json.RawMessage, error) {
	projectKey := e.defaultProject
	boardID := e.defaultBoard

	for _, arg := range stmt.Args {
		switch arg.Key {
		case "project":
			projectKey = arg.Value
		case "board":
			id, err := strconv.Atoi(arg.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid board ID: %s", arg.Value)
			}
			boardID = id
		case "":
			if projectKey == "" {
				projectKey = arg.Value
			}
		}
	}

	result := map[string]interface{}{}

	// Get project info
	if projectKey != "" {
		proj, err := e.client.GetProject(projectKey)
		if err != nil {
			return nil, fmt.Errorf("getting project: %w", err)
		}
		result["project"] = map[string]interface{}{
			"key":  proj.Key,
			"name": proj.Name,
			"type": proj.ProjectTypeKey,
		}

		// Count issues by status
		issues, err := e.client.ListIssues(jira.ListIssuesOptions{
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

	// Get board info
	if boardID != 0 {
		board, err := e.client.GetBoard(boardID)
		if err != nil {
			return nil, fmt.Errorf("getting board: %w", err)
		}
		result["board"] = map[string]interface{}{
			"id":   board.ID,
			"name": board.Name,
			"type": board.Type,
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("summary requires a project or board (via argument or config)")
	}

	return json.Marshal(result)
}

// opSearch: search(jql="...") { fields }
func (e *Executor) opSearch(stmt *Statement) (json.RawMessage, error) {
	var jql string
	for _, arg := range stmt.Args {
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
		return nil, fmt.Errorf("search requires a jql argument")
	}

	sel, err := fields.NewSelector(stmt.Fields)
	if err != nil {
		return nil, err
	}

	issues, err := e.client.SearchAll(jql, sel.JiraAPIFields())
	if err != nil {
		return nil, err
	}

	projected := sel.ApplyMany(issues)
	return json.Marshal(projected)
}
