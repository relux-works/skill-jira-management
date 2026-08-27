package jira

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const MaxAttachmentDownloadBytes int64 = 100 << 20

// DownloadAttachment resolves an attachment through its owning issue and
// downloads it with the configured Jira credentials under a strict byte cap.
func (c *Client) DownloadAttachment(issueKey, attachmentID string, maxBytes int64) ([]byte, *Attachment, error) {
	if issueKey == "" || attachmentID == "" {
		return nil, nil, fmt.Errorf("jira: issue key and attachment ID are required")
	}
	if maxBytes <= 0 || maxBytes > MaxAttachmentDownloadBytes {
		return nil, nil, fmt.Errorf("jira: attachment byte limit must be between 1 and %d", MaxAttachmentDownloadBytes)
	}

	issue, err := c.GetIssue(issueKey, []string{"attachment"})
	if err != nil {
		return nil, nil, err
	}

	var selected *Attachment
	for i := range issue.Fields.Attachments {
		if issue.Fields.Attachments[i].ID == attachmentID {
			selected = &issue.Fields.Attachments[i]
			break
		}
	}
	if selected == nil {
		return nil, nil, fmt.Errorf("jira: attachment %s does not belong to %s", attachmentID, issueKey)
	}
	if selected.Size > maxBytes {
		return nil, nil, fmt.Errorf("jira: attachment exceeds the configured byte limit")
	}

	contentURL, err := c.validatedAttachmentURL(selected.Content)
	if err != nil {
		return nil, nil, err
	}
	data, err := c.getBounded(contentURL, maxBytes)
	if err != nil {
		return nil, nil, err
	}
	return data, selected, nil
}

func (c *Client) validatedAttachmentURL(raw string) (*url.URL, error) {
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("jira: configured base URL is invalid")
	}
	content, err := url.Parse(raw)
	if err != nil || !content.IsAbs() || content.User != nil {
		return nil, fmt.Errorf("jira: attachment URL could not be verified")
	}
	if content.Scheme != base.Scheme || content.Host != base.Host || content.Fragment != "" {
		return nil, fmt.Errorf("jira: attachment URL is outside the configured Jira origin")
	}
	return content, nil
}

func (c *Client) getBounded(target *url.URL, maxBytes int64) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("jira: failed to create attachment request")
	}
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Accept", "application/octet-stream")

	httpClient := *c.httpClient
	httpClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jira: attachment download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jira: attachment download returned HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > maxBytes {
		return nil, fmt.Errorf("jira: attachment exceeds the configured byte limit")
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("jira: failed to read attachment: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("jira: attachment exceeds the configured byte limit")
	}
	return data, nil
}
