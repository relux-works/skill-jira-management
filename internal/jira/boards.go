package jira

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// ListBoards returns all boards visible to the authenticated user.
// Uses the Agile REST API with offset-based pagination.
func (c *Client) ListBoards(projectKeyOrID string) ([]Board, error) {
	var all []Board
	startAt := 0
	maxResults := 50

	for {
		q := url.Values{}
		q.Set("startAt", strconv.Itoa(startAt))
		q.Set("maxResults", strconv.Itoa(maxResults))
		if projectKeyOrID != "" {
			q.Set("projectKeyOrId", projectKeyOrID)
		}

		data, err := c.Get(agileAPIPath("board"), q)
		if err != nil {
			return nil, fmt.Errorf("ListBoards: %w", err)
		}

		var result BoardSearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("ListBoards: failed to unmarshal: %w", err)
		}

		all = append(all, result.Values...)

		if result.IsLast || len(result.Values) == 0 {
			break
		}
		startAt += len(result.Values)
	}

	return all, nil
}

// GetBoard retrieves a single board by ID.
func (c *Client) GetBoard(boardID int) (*Board, error) {
	data, err := c.Get(agileAPIPath("board", strconv.Itoa(boardID)), nil)
	if err != nil {
		return nil, fmt.Errorf("GetBoard %d: %w", boardID, err)
	}

	var board Board
	if err := json.Unmarshal(data, &board); err != nil {
		return nil, fmt.Errorf("GetBoard %d: failed to unmarshal: %w", boardID, err)
	}
	return &board, nil
}

// ListSprints returns all sprints for a board.
// Uses the Agile REST API with offset-based pagination.
func (c *Client) ListSprints(boardID int) ([]Sprint, error) {
	var all []Sprint
	startAt := 0
	maxResults := 50

	for {
		q := url.Values{}
		q.Set("startAt", strconv.Itoa(startAt))
		q.Set("maxResults", strconv.Itoa(maxResults))

		path := agileAPIPath("board", strconv.Itoa(boardID), "sprint")
		data, err := c.Get(path, q)
		if err != nil {
			return nil, fmt.Errorf("ListSprints board=%d: %w", boardID, err)
		}

		var result SprintSearchResult
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("ListSprints board=%d: failed to unmarshal: %w", boardID, err)
		}

		all = append(all, result.Values...)

		if result.IsLast || len(result.Values) == 0 {
			break
		}
		startAt += len(result.Values)
	}

	return all, nil
}
