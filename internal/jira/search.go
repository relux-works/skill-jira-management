package jira

import (
	"encoding/json"
	"fmt"
)

// SearchJQL executes a JQL search. Uses the appropriate endpoint based on instance type:
// - Cloud: POST /rest/api/3/search/jql (cursor-based pagination)
// - Server/DC: POST /rest/api/2/search (offset-based pagination)
func (c *Client) SearchJQL(req *SearchRequest) (*SearchResponse, error) {
	if req.MaxResults <= 0 {
		req.MaxResults = 100
	}

	if c.instanceType == InstanceServer {
		return c.searchV2(req)
	}
	return c.searchV3(req)
}

// searchV3 uses Cloud endpoint: POST /rest/api/3/search/jql with cursor pagination.
func (c *Client) searchV3(req *SearchRequest) (*SearchResponse, error) {
	data, err := c.Post(apiPath("search", "jql"), req)
	if err != nil {
		return nil, fmt.Errorf("SearchJQL: %w", err)
	}

	var resp SearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("SearchJQL: failed to unmarshal: %w", err)
	}
	return &resp, nil
}

// searchV2 uses Server/DC endpoint: POST /rest/api/2/search with offset pagination.
func (c *Client) searchV2(req *SearchRequest) (*SearchResponse, error) {
	v2Req := map[string]interface{}{
		"jql":        req.JQL,
		"maxResults": req.MaxResults,
	}
	if len(req.Fields) > 0 {
		v2Req["fields"] = req.Fields
	}
	if req.StartAt > 0 {
		v2Req["startAt"] = req.StartAt
	}

	data, err := c.Post(c.apiPathFor("search"), v2Req)
	if err != nil {
		return nil, fmt.Errorf("SearchJQL: %w", err)
	}

	var v2Resp struct {
		StartAt    int     `json:"startAt"`
		MaxResults int     `json:"maxResults"`
		Total      int     `json:"total"`
		Issues     []Issue `json:"issues"`
	}
	if err := json.Unmarshal(data, &v2Resp); err != nil {
		return nil, fmt.Errorf("SearchJQL: failed to unmarshal: %w", err)
	}

	isLast := v2Resp.StartAt+len(v2Resp.Issues) >= v2Resp.Total
	return &SearchResponse{
		MaxResults: v2Resp.MaxResults,
		Issues:     v2Resp.Issues,
		Total:      v2Resp.Total,
		IsLast:     isLast,
	}, nil
}

// SearchAll executes a JQL search and fetches all pages.
func (c *Client) SearchAll(jql string, fields []string) ([]Issue, error) {
	var allIssues []Issue

	if c.instanceType == InstanceServer {
		return c.searchAllV2(jql, fields)
	}
	return c.searchAllV3(jql, fields, allIssues)
}

func (c *Client) searchAllV3(jql string, fields []string, allIssues []Issue) ([]Issue, error) {
	nextPageToken := ""
	for {
		req := &SearchRequest{
			JQL:           jql,
			MaxResults:    100,
			Fields:        fields,
			NextPageToken: nextPageToken,
		}

		resp, err := c.searchV3(req)
		if err != nil {
			return nil, err
		}

		allIssues = append(allIssues, resp.Issues...)

		if resp.IsLast || resp.NextPageToken == "" {
			break
		}
		nextPageToken = resp.NextPageToken
	}
	return allIssues, nil
}

func (c *Client) searchAllV2(jql string, fields []string) ([]Issue, error) {
	var allIssues []Issue
	startAt := 0
	for {
		req := &SearchRequest{
			JQL:        jql,
			MaxResults: 100,
			Fields:     fields,
			StartAt:    startAt,
		}

		resp, err := c.searchV2(req)
		if err != nil {
			return nil, err
		}

		allIssues = append(allIssues, resp.Issues...)

		if resp.IsLast || len(resp.Issues) == 0 {
			break
		}
		startAt += len(resp.Issues)
	}
	return allIssues, nil
}
