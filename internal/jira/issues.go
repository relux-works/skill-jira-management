package jira

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// GetIssue retrieves a single issue by key (e.g. "PROJ-123").
// fields is an optional list of field names to return; nil returns defaults.
func (c *Client) GetIssue(issueKey string, fields []string) (*Issue, error) {
	q := url.Values{}
	if len(fields) > 0 {
		q.Set("fields", strings.Join(fields, ","))
	}

	data, err := c.Get(c.apiPathFor("issue", issueKey), q)
	if err != nil {
		return nil, fmt.Errorf("GetIssue %s: %w", issueKey, err)
	}

	var issue Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("GetIssue %s: failed to unmarshal: %w", issueKey, err)
	}
	return &issue, nil
}

// CreateIssue creates a new issue in Jira.
// Supports Epic, Story, Task, Subtask, Bug.
func (c *Client) CreateIssue(req *CreateIssueRequest) (*CreateIssueResponse, error) {
	// Build the payload; merge Extra custom fields into the fields map.
	payload := buildCreatePayload(req)

	data, err := c.Post(c.apiPathFor("issue"), payload)
	if err != nil {
		return nil, fmt.Errorf("CreateIssue: %w", err)
	}

	var resp CreateIssueResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("CreateIssue: failed to unmarshal response: %w", err)
	}
	return &resp, nil
}

// buildCreatePayload merges the typed CreateIssueFields with any Extra custom fields
// into a single map suitable for JSON marshalling.
func buildCreatePayload(req *CreateIssueRequest) map[string]interface{} {
	fields := map[string]interface{}{
		"project":   req.Fields.Project,
		"issuetype": req.Fields.IssueType,
		"summary":   req.Fields.Summary,
	}

	if req.Fields.Description != nil {
		fields["description"] = req.Fields.Description
	}
	if req.Fields.Assignee != nil {
		fields["assignee"] = req.Fields.Assignee
	}
	if req.Fields.Priority != nil {
		fields["priority"] = req.Fields.Priority
	}
	if len(req.Fields.Labels) > 0 {
		fields["labels"] = req.Fields.Labels
	}
	if req.Fields.Parent != nil {
		fields["parent"] = req.Fields.Parent
	}

	// Merge custom fields.
	for k, v := range req.Fields.Extra {
		fields[k] = v
	}

	return map[string]interface{}{
		"fields": fields,
	}
}

// UpdateIssue updates fields on an existing issue.
func (c *Client) UpdateIssue(issueKey string, req *UpdateIssueRequest) error {
	_, err := c.Put(c.apiPathFor("issue", issueKey), req)
	if err != nil {
		return fmt.Errorf("UpdateIssue %s: %w", issueKey, err)
	}
	return nil
}

// ListIssues lists issues for a project using JQL search.
// Returns all matching issues (handles pagination internally for both Cloud and Server/DC).
func (c *Client) ListIssues(opts ListIssuesOptions) ([]Issue, error) {
	// Build JQL.
	var clauses []string
	if opts.ProjectKey != "" {
		clauses = append(clauses, fmt.Sprintf("project = %s", opts.ProjectKey))
	}
	if opts.IssueType != "" {
		clauses = append(clauses, fmt.Sprintf("issuetype = %q", opts.IssueType))
	}
	if opts.Status != "" {
		clauses = append(clauses, fmt.Sprintf("status = %q", opts.Status))
	}
	jql := strings.Join(clauses, " AND ")

	return c.SearchAll(jql, opts.Fields)
}
