package query

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/relux-works/skill-agent-facing-api/agentquery"
	"github.com/relux-works/skill-jira-management/internal/jira"
)

func TestGetProjectsTargetedCommentsAndAttachmentMetadata(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/rest/api/2/issue/PROJ-1":
			if fields := r.URL.Query().Get("fields"); !strings.Contains(fields, "attachment") {
				t.Fatalf("issue fields = %q, want attachment", fields)
			}
			w.Write([]byte(`{"id":"1","key":"PROJ-1","fields":{"attachment":[{"id":"20001","filename":"review.docx","mimeType":"application/vnd.openxmlformats-officedocument.wordprocessingml.document","size":18432,"content":"https://jira.example/secure/attachment/20001/review.docx"}]}}`))
		case "/rest/api/2/issue/PROJ-1/comment":
			w.Write([]byte(`{"startAt":0,"maxResults":50,"total":1,"comments":[{"id":"10001","author":{"displayName":"Patent Counsel"},"body":"Please clarify the clock source.","created":"2026-08-26T10:00:00.000+0300"}]}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client, err := jira.NewClient(jira.Config{
		BaseURL:      srv.URL,
		Token:        "test-token",
		AuthType:     jira.AuthBearer,
		InstanceType: jira.InstanceServer,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	client.SetHTTPClient(srv.Client())

	schema := NewSchema(client, "", 0)
	data, err := schema.QueryJSONWithMode("get(PROJ-1) { comments attachments }", agentquery.HumanReadable)
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	comments, ok := result["comments"].([]any)
	if !ok || len(comments) != 1 {
		t.Fatalf("comments = %#v, want one comment", result["comments"])
	}
	comment := comments[0].(map[string]any)
	if comment["body"] != "Please clarify the clock source." {
		t.Fatalf("comment body = %#v", comment["body"])
	}
	attachments, ok := result["attachments"].([]any)
	if !ok || len(attachments) != 1 {
		t.Fatalf("attachments = %#v, want one attachment", result["attachments"])
	}
	attachment := attachments[0].(map[string]any)
	if attachment["filename"] != "review.docx" {
		t.Fatalf("attachment filename = %#v", attachment["filename"])
	}
	if _, leaked := attachment["content"]; leaked {
		t.Fatal("attachment projection leaked authenticated content URL")
	}
}
