package jira

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ListProjects returns all projects visible to the authenticated user.
// Cloud uses paginated /project/search, Server/DC uses /project (returns all at once).
func (c *Client) ListProjects() ([]Project, error) {
	if c.instanceType == InstanceServer {
		return c.listProjectsV2()
	}
	return c.listProjectsCloud()
}

// listProjectsCloud uses the paginated /project/search endpoint (Cloud).
func (c *Client) listProjectsCloud() ([]Project, error) {
	var all []Project
	startAt := 0
	maxResults := 50

	for {
		q := url.Values{}
		q.Set("startAt", strconv.Itoa(startAt))
		q.Set("maxResults", strconv.Itoa(maxResults))

		data, err := c.Get(c.apiPathFor("project", "search"), q)
		if err != nil {
			return nil, fmt.Errorf("ListProjects: %w", err)
		}

		var result ProjectSearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("ListProjects: failed to unmarshal: %w", err)
		}

		all = append(all, result.Values...)

		if result.IsLast || len(result.Values) == 0 {
			break
		}
		startAt += len(result.Values)
	}

	return all, nil
}

// listProjectsV2 uses /project endpoint (Server/DC — returns all projects as array).
func (c *Client) listProjectsV2() ([]Project, error) {
	data, err := c.Get(c.apiPathFor("project"), nil)
	if err != nil {
		return nil, fmt.Errorf("ListProjects: %w", err)
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("ListProjects: failed to unmarshal: %w", err)
	}
	return projects, nil
}

// GetProject retrieves a single project by key or ID.
func (c *Client) GetProject(projectKeyOrID string) (*Project, error) {
	data, err := c.Get(c.apiPathFor("project", projectKeyOrID), nil)
	if err != nil {
		return nil, fmt.Errorf("GetProject %s: %w", projectKeyOrID, err)
	}

	var project Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("GetProject %s: failed to unmarshal: %w", projectKeyOrID, err)
	}
	return &project, nil
}
