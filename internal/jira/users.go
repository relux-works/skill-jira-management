package jira

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Myself returns the authenticated user.
func (c *Client) Myself() (*User, error) {
	data, err := c.Get(c.apiPathFor("myself"), nil)
	if err != nil {
		return nil, fmt.Errorf("Myself: %w", err)
	}
	var u User
	if err := json.Unmarshal(data, &u); err != nil {
		return nil, fmt.Errorf("Myself: failed to unmarshal: %w", err)
	}
	return &u, nil
}

// FindUser resolves a query (username, email or display name) to a single user.
// It is an error for the query to match more than one user: silently picking
// one would assign somebody else's work to the wrong person.
func (c *Client) FindUser(query string) (*User, error) {
	q := url.Values{}
	path := c.apiPathFor("user", "search")
	if c.IsCloud() {
		q.Set("query", query)
	} else {
		q.Set("username", query)
	}

	data, err := c.Get(path, q)
	if err != nil {
		return nil, fmt.Errorf("FindUser %q: %w", query, err)
	}
	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("FindUser %q: failed to unmarshal: %w", query, err)
	}
	switch len(users) {
	case 0:
		return nil, fmt.Errorf("no Jira user matches %q", query)
	case 1:
		return &users[0], nil
	}

	// An exact match on the addressing field or the email resolves the ambiguity.
	var exact []User
	for _, u := range users {
		if u.Name == query || u.AccountID == query || u.EmailAddress == query {
			exact = append(exact, u)
		}
	}
	if len(exact) == 1 {
		return &exact[0], nil
	}

	var names []string
	for _, u := range users {
		label := u.Name
		if label == "" {
			label = u.AccountID
		}
		names = append(names, fmt.Sprintf("%s (%s)", label, u.DisplayName))
	}
	return nil, fmt.Errorf("%q matches %d Jira users, be exact: %v", query, len(users), names)
}

// AssignIssue sets the assignee. A nil user clears the assignment.
//
// The addressing field differs by instance and there is no overlap: Cloud
// rejects "name", Server/DC rejects "accountId". Sending both is not a safe
// fallback, so the field is chosen from the detected instance type.
func (c *Client) AssignIssue(issueKey string, user *User) error {
	body := map[string]interface{}{}
	switch {
	case user == nil:
		if c.IsCloud() {
			body["accountId"] = nil
		} else {
			body["name"] = nil
		}
	case c.IsCloud():
		if user.AccountID == "" {
			return fmt.Errorf("AssignIssue %s: Cloud needs an accountId, resolved user has none", issueKey)
		}
		body["accountId"] = user.AccountID
	default:
		if user.Name == "" {
			return fmt.Errorf("AssignIssue %s: Server/DC needs a username, resolved user has none", issueKey)
		}
		body["name"] = user.Name
	}

	if _, err := c.Put(c.apiPathFor("issue", issueKey, "assignee"), body); err != nil {
		return fmt.Errorf("AssignIssue %s: %w", issueKey, err)
	}
	return nil
}
