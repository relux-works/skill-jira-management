package jira

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

const DefaultAttachmentUploadBytes int64 = 25 << 20
const maxAttachmentResponseBytes int64 = 1 << 20

var uploadIssueKey = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*-[0-9]+$`)
var sensitiveAttachmentHeader = regexp.MustCompile(`(?im)\b(authorization|(?:set-)?cookie)["']?\s*[:=][^\r\n]*`)

// AttachmentUpload holds validated, bounded file bytes, ready before client discovery makes any requests.
type AttachmentUpload struct {
	issueKey string
	filename string
	data     []byte
}

// PrepareAttachmentUpload validates all local inputs and reads at most maxBytes of a regular file.
// An empty name selects the source basename; callers must reject an explicitly empty --name.
func PrepareAttachmentUpload(issueKey, path, name string, maxBytes int64) (*AttachmentUpload, error) {
	if !uploadIssueKey.MatchString(issueKey) {
		return nil, fmt.Errorf("jira: invalid issue key (expected PROJ-123)")
	}
	if maxBytes <= 0 {
		return nil, fmt.Errorf("jira: attachment byte limit must be positive")
	}
	if name == "" {
		name = filepath.Base(path)
	}
	if !validAttachmentName(name) {
		return nil, fmt.Errorf("jira: attachment filename must be a plain base name")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("jira: cannot inspect attachment file")
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("jira: attachment must be a regular file, not a symlink or directory")
	}
	if info.Mode().Perm()&0o444 == 0 {
		return nil, fmt.Errorf("jira: attachment file is not readable")
	}
	if info.Size() == 0 {
		return nil, fmt.Errorf("jira: attachment file is empty")
	}
	if info.Size() > maxBytes {
		return nil, fmt.Errorf("jira: attachment exceeds the configured byte limit")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("jira: cannot read attachment file")
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return nil, fmt.Errorf("jira: attachment file changed during validation")
	}
	data, err := io.ReadAll(io.LimitReader(file, maxBytes))
	if err != nil {
		return nil, fmt.Errorf("jira: cannot read attachment file")
	}
	var extra [1]byte
	n, err := file.Read(extra[:])
	if n != 0 {
		return nil, fmt.Errorf("jira: attachment exceeds the configured byte limit")
	}
	if err != io.EOF {
		return nil, fmt.Errorf("jira: cannot finish reading attachment file")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("jira: attachment file is empty")
	}
	return &AttachmentUpload{issueKey: issueKey, filename: name, data: data}, nil
}

func validAttachmentName(name string) bool {
	if strings.TrimSpace(name) == "" || name == "." || name == ".." {
		return false
	}
	for _, ch := range name {
		if ch == '/' || ch == '\\' || unicode.IsControl(ch) {
			return false
		}
	}
	return true
}

// UploadAttachment verifies the issue with the normal read before one multipart POST.
// The POST is never retried: a lost response does not prove that Jira failed to create the attachment.
func (c *Client) UploadAttachment(upload *AttachmentUpload) (*Attachment, error) {
	if upload == nil || !uploadIssueKey.MatchString(upload.issueKey) || !validAttachmentName(upload.filename) || len(upload.data) == 0 {
		return nil, fmt.Errorf("jira: a prepared attachment upload is required")
	}
	if _, err := c.GetIssue(upload.issueKey, nil); err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			return nil, c.attachmentAPIError("issue pre-check", apiErr.StatusCode, []byte(apiErr.Error()))
		}
		return nil, fmt.Errorf("jira: issue pre-check failed: %s", c.safeAttachmentText(err.Error()))
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", upload.filename)
	if err != nil {
		return nil, fmt.Errorf("jira: cannot prepare attachment multipart form")
	}
	if _, err := part.Write(upload.data); err != nil {
		return nil, fmt.Errorf("jira: cannot prepare attachment bytes")
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("jira: cannot finish attachment multipart form")
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+c.apiPathFor("issue", upload.issueKey, "attachments"), &body)
	if err != nil {
		return nil, fmt.Errorf("jira: cannot create attachment upload request")
	}
	req.Header.Set("Authorization", c.authHeader)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Atlassian-Token", "no-check")

	httpClient := *c.httpClient
	httpClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := httpClient.Do(req)
	if err != nil {
		// Transport errors can contain URLs or reflected headers; do not expose them.
		return nil, fmt.Errorf("jira: attachment upload request failed (check network and TLS configuration)")
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxAttachmentResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("jira: cannot read attachment upload response (HTTP %d)", resp.StatusCode)
	}
	if int64(len(data)) > maxAttachmentResponseBytes {
		return nil, fmt.Errorf("jira: attachment upload response exceeds byte limit (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, c.attachmentAPIError("attachment upload", resp.StatusCode, data)
	}
	var attachments []Attachment
	if err := json.Unmarshal(data, &attachments); err != nil {
		return nil, fmt.Errorf("jira: invalid attachment upload response")
	}
	if len(attachments) != 1 || attachments[0].ID == "" || attachments[0].Filename == "" || attachments[0].Size < 0 {
		return nil, fmt.Errorf("jira: expected one attachment in upload response")
	}
	// Return only printable metadata; Jira may echo authentication material in a response.
	return &Attachment{
		ID:       c.safeAttachmentText(attachments[0].ID),
		Filename: c.safeAttachmentText(attachments[0].Filename),
		Size:     attachments[0].Size,
	}, nil
}

func (c *Client) attachmentAPIError(operation string, status int, body []byte) error {
	message := c.safeAttachmentText(parseAPIError(status, body).Error())
	return fmt.Errorf("jira: %s returned HTTP %d (%s): %s", operation, status, http.StatusText(status), message)
}

// safeAttachmentText removes known credentials and header values before limiting diagnostic output.
func (c *Client) safeAttachmentText(value string) string {
	secrets := []string{c.authHeader}
	if strings.HasPrefix(c.authHeader, "Bearer ") {
		secrets = append(secrets, strings.TrimPrefix(c.authHeader, "Bearer "))
	} else if encoded, ok := strings.CutPrefix(c.authHeader, "Basic "); ok {
		secrets = append(secrets, encoded)
		if decoded, err := base64.StdEncoding.DecodeString(encoded); err == nil {
			secrets = append(secrets, string(decoded))
			if email, token, ok := strings.Cut(string(decoded), ":"); ok {
				secrets = append(secrets, email, token)
			}
		}
	}
	for _, secret := range secrets {
		if secret != "" {
			value = strings.ReplaceAll(value, secret, "[redacted]")
		}
	}
	value = sensitiveAttachmentHeader.ReplaceAllString(value, "$1: [redacted]")
	value = strings.Map(func(ch rune) rune {
		if unicode.IsControl(ch) {
			return ' '
		}
		return ch
	}, value)
	const maxDiagnosticRunes = 1024
	runes := []rune(value)
	if len(runes) > maxDiagnosticRunes {
		value = string(runes[:maxDiagnosticRunes]) + "..."
	}
	return value
}
