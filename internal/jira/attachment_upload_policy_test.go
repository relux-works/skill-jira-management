package jira

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// Proves direct library uploads enforce the pre-check policy without changing the caller's redirect handling or ordinary read size.
func TestAttachmentUploadPolicyIsolation(t *testing.T) {
	for _, redirect := range []bool{false, true} {
		name := "oversized read"
		if redirect {
			name = "redirected read"
		}
		t.Run(name, func(t *testing.T) {
			var posts, followed, redirects atomic.Int32
			body := `{"key":"PROJ-123","fields":{},"ignored":"` + strings.Repeat("x", int(maxAttachmentResponseBytes)) + `"}`
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					posts.Add(1)
					io.WriteString(w, `[{"id":"20001","filename":"report.txt","size":7}]`)
					return
				}
				if redirect && r.URL.Path != "/redirected" {
					http.Redirect(w, r, "/redirected", http.StatusFound)
					return
				}
				if r.URL.Path == "/redirected" {
					followed.Add(1)
				}
				io.WriteString(w, body)
			}))
			defer server.Close()
			client, err := NewClient(Config{BaseURL: server.URL, Token: attachmentSecret, InstanceType: InstanceCloud})
			if err != nil {
				t.Fatal(err)
			}
			client.SetHTTPClient(&http.Client{CheckRedirect: func(_ *http.Request, _ []*http.Request) error { redirects.Add(1); return nil }})
			attachment, err := client.UploadAttachment(preparedUpload(t, "PROJ-123", "report.txt", 7))
			assertSafeUpload(t, client, attachment, err)
			if err == nil || posts.Load() != 0 || followed.Load() != 0 || redirects.Load() != 0 {
				t.Fatalf("upload error=%v posts=%d followed=%d redirects=%d", err, posts.Load(), followed.Load(), redirects.Load())
			}
			issue, err := client.GetIssue("PROJ-123", nil)
			if err != nil || issue.Key != "PROJ-123" {
				t.Fatalf("ordinary read control: %v", err)
			}
			want := int32(0)
			if redirect {
				want = 1
			}
			if followed.Load() != want || redirects.Load() != want {
				t.Fatal("upload mutated the caller's redirect behavior")
			}
		})
	}
}
