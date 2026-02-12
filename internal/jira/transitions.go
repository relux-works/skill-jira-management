package jira

import (
	"encoding/json"
	"fmt"
)

// GetTransitions returns the available workflow transitions for an issue.
func (c *Client) GetTransitions(issueKey string) ([]Transition, error) {
	data, err := c.Get(c.apiPathFor("issue", issueKey, "transitions"), nil)
	if err != nil {
		return nil, fmt.Errorf("GetTransitions %s: %w", issueKey, err)
	}

	var resp TransitionsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("GetTransitions %s: failed to unmarshal: %w", issueKey, err)
	}
	return resp.Transitions, nil
}

// DoTransition executes a workflow transition on an issue.
// transitionID is the ID of the transition to execute.
// fields is an optional map of fields required by the transition (e.g. resolution).
func (c *Client) DoTransition(issueKey string, transitionID string, fields map[string]interface{}) error {
	req := DoTransitionRequest{
		Transition: TransitionRef{ID: transitionID},
		Fields:     fields,
	}

	_, err := c.Post(c.apiPathFor("issue", issueKey, "transitions"), &req)
	if err != nil {
		return fmt.Errorf("DoTransition %s (transition=%s): %w", issueKey, transitionID, err)
	}
	return nil
}

// ListStatuses returns all statuses in the Jira instance.
func (c *Client) ListStatuses() ([]Status, error) {
	data, err := c.Get(c.apiPathFor("status"), nil)
	if err != nil {
		return nil, fmt.Errorf("ListStatuses: %w", err)
	}

	var statuses []Status
	if err := json.Unmarshal(data, &statuses); err != nil {
		return nil, fmt.Errorf("ListStatuses: failed to unmarshal: %w", err)
	}
	return statuses, nil
}

// ListProjectStatuses returns statuses grouped by issue type for a specific project.
func (c *Client) ListProjectStatuses(projectKeyOrID string) ([]ProjectIssueTypeStatuses, error) {
	data, err := c.Get(c.apiPathFor("project", projectKeyOrID, "statuses"), nil)
	if err != nil {
		return nil, fmt.Errorf("ListProjectStatuses %s: %w", projectKeyOrID, err)
	}

	var result []ProjectIssueTypeStatuses
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("ListProjectStatuses %s: failed to unmarshal: %w", projectKeyOrID, err)
	}
	return result, nil
}

// ProjectIssueTypeStatuses groups statuses by issue type within a project.
type ProjectIssueTypeStatuses struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Subtask  bool     `json:"subtask"`
	Statuses []Status `json:"statuses"`
}
