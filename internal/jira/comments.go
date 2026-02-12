package jira

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// --- ADF body builders ---

// NewADFText creates an ADF document with a single paragraph of plain text.
func NewADFText(text string) *ADFDoc {
	return &ADFDoc{
		Type:    "doc",
		Version: 1,
		Content: []ADFNode{
			{
				Type: "paragraph",
				Content: []ADFNode{
					{
						Type: "text",
						Text: text,
					},
				},
			},
		},
	}
}

// NewADFParagraphs creates an ADF document with multiple paragraphs.
func NewADFParagraphs(paragraphs []string) *ADFDoc {
	nodes := make([]ADFNode, 0, len(paragraphs))
	for _, p := range paragraphs {
		nodes = append(nodes, ADFNode{
			Type: "paragraph",
			Content: []ADFNode{
				{
					Type: "text",
					Text: p,
				},
			},
		})
	}
	return &ADFDoc{
		Type:    "doc",
		Version: 1,
		Content: nodes,
	}
}

// NewADFWithHeading creates an ADF document with a heading followed by a text paragraph.
func NewADFWithHeading(level int, heading string, bodyText string) *ADFDoc {
	if level < 1 || level > 6 {
		level = 1
	}

	attrs, _ := json.Marshal(map[string]int{"level": level})

	nodes := []ADFNode{
		{
			Type:  "heading",
			Attrs: attrs,
			Content: []ADFNode{
				{
					Type: "text",
					Text: heading,
				},
			},
		},
	}

	if bodyText != "" {
		nodes = append(nodes, ADFNode{
			Type: "paragraph",
			Content: []ADFNode{
				{
					Type: "text",
					Text: bodyText,
				},
			},
		})
	}

	return &ADFDoc{
		Type:    "doc",
		Version: 1,
		Content: nodes,
	}
}

// NewADFBulletList creates an ADF document with a bullet list.
func NewADFBulletList(items []string) *ADFDoc {
	listItems := make([]ADFNode, 0, len(items))
	for _, item := range items {
		listItems = append(listItems, ADFNode{
			Type: "listItem",
			Content: []ADFNode{
				{
					Type: "paragraph",
					Content: []ADFNode{
						{
							Type: "text",
							Text: item,
						},
					},
				},
			},
		})
	}

	return &ADFDoc{
		Type:    "doc",
		Version: 1,
		Content: []ADFNode{
			{
				Type:    "bulletList",
				Content: listItems,
			},
		},
	}
}

// NewADFCodeBlock creates an ADF document with a code block.
func NewADFCodeBlock(code string, language string) *ADFDoc {
	var attrs json.RawMessage
	if language != "" {
		attrs, _ = json.Marshal(map[string]string{"language": language})
	}

	return &ADFDoc{
		Type:    "doc",
		Version: 1,
		Content: []ADFNode{
			{
				Type:  "codeBlock",
				Attrs: attrs,
				Content: []ADFNode{
					{
						Type: "text",
						Text: code,
					},
				},
			},
		},
	}
}

// --- Comment operations ---

// AddComment adds a comment to an issue.
func (c *Client) AddComment(issueKey string, body *ADFDoc) (*Comment, error) {
	req := AddCommentRequest{Body: body}

	data, err := c.Post(c.apiPathFor("issue", issueKey, "comment"), &req)
	if err != nil {
		return nil, fmt.Errorf("AddComment %s: %w", issueKey, err)
	}

	var comment Comment
	if err := json.Unmarshal(data, &comment); err != nil {
		return nil, fmt.Errorf("AddComment %s: failed to unmarshal: %w", issueKey, err)
	}
	return &comment, nil
}

// ListComments returns comments on an issue with offset-based pagination.
func (c *Client) ListComments(issueKey string, startAt, maxResults int) (*CommentsResponse, error) {
	if maxResults <= 0 {
		maxResults = 50
	}

	q := url.Values{}
	q.Set("startAt", strconv.Itoa(startAt))
	q.Set("maxResults", strconv.Itoa(maxResults))

	data, err := c.Get(c.apiPathFor("issue", issueKey, "comment"), q)
	if err != nil {
		return nil, fmt.Errorf("ListComments %s: %w", issueKey, err)
	}

	var resp CommentsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ListComments %s: failed to unmarshal: %w", issueKey, err)
	}
	return &resp, nil
}

// ListAllComments fetches all comments for an issue, handling pagination.
func (c *Client) ListAllComments(issueKey string) ([]Comment, error) {
	var all []Comment
	startAt := 0
	maxResults := 50

	for {
		resp, err := c.ListComments(issueKey, startAt, maxResults)
		if err != nil {
			return nil, err
		}

		all = append(all, resp.Comments...)

		if startAt+len(resp.Comments) >= resp.Total {
			break
		}
		startAt += len(resp.Comments)
	}

	return all, nil
}
