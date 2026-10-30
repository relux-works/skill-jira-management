package jira

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const attachmentSecret = "FAKE-UPLOAD-CREDENTIAL-SECRET"

func preparedUpload(t *testing.T, key, name string, limit int64) *AttachmentUpload {
	t.Helper()
	path := filepath.Join(t.TempDir(), "source.bin")
	if err := os.WriteFile(path, []byte("payload"), 0o600); err != nil {
		t.Fatal(err)
	}
	upload, err := PrepareAttachmentUpload(key, path, name, limit)
	if err != nil {
		t.Fatal(err)
	}
	return upload
}

func assertSafeUpload(t *testing.T, c *Client, attachment *Attachment, err error) {
	t.Helper()
	value, _ := json.Marshal(attachment)
	text := string(value)
	if err != nil {
		text += err.Error()
	}
	for _, secret := range []string{attachmentSecret, c.authHeader, "cookie-secret", "user@example.com"} {
		if strings.Contains(text, secret) {
			t.Fatal("upload exposed credentials or cookies")
		}
	}
}

// Proves conventional key/name forms and exact or maximal positive caps work while invalid forms are refused locally.
func TestPrepareAttachmentUploadCompatibility(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.bin")
	if err := os.WriteFile(path, []byte("payload"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"PROJ-123", "A-1", "Proj_2-123", "proj-123"} {
		for _, name := range []string{"", "report.txt", "report with spaces.txt", "отчёт.txt", `quote"file.txt`} {
			for _, limit := range []int64{7, DefaultAttachmentUploadBytes, math.MaxInt64} {
				upload, err := PrepareAttachmentUpload(key, path, name, limit)
				if err != nil || string(upload.data) != "payload" {
					t.Fatalf("key=%q name=%q limit=%d: %v", key, name, limit, err)
				}
			}
		}
	}
	for _, key := range []string{"", "PROJ", "123", "PROJ-", "-123", "PROJ-123/path", "PROJ-123?x=1", " PROJ-123", "PROJ-123\n"} {
		if _, err := PrepareAttachmentUpload(key, path, "", 7); err == nil {
			t.Errorf("accepted key %q", key)
		}
	}
	for _, name := range []string{".", "..", " ", "../file", `dir\file`, "/file", "file\nname", "file\x00name"} {
		if _, err := PrepareAttachmentUpload("PROJ-123", path, name, 7); err == nil {
			t.Errorf("accepted name %q", name)
		}
	}
	if _, err := PrepareAttachmentUpload("PROJ-123", path, "", 6); err == nil {
		t.Fatal("accepted bytes above cap")
	}
}

// Proves preparation snapshots file bytes so later source edits cannot change the upload.
func TestAttachmentUploadPreparedBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.bin")
	if err := os.WriteFile(path, []byte("payload"), 0o600); err != nil {
		t.Fatal(err)
	}
	upload, err := PrepareAttachmentUpload("PROJ-123", path, "", 7)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("replacement"), 0o600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			io.WriteString(w, `{"key":"PROJ-123","fields":{}}`)
			return
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Error(err)
			return
		}
		part, err := reader.NextPart()
		if err != nil {
			t.Error(err)
			return
		}
		body, _ := io.ReadAll(part)
		if string(body) != "payload" {
			t.Error("source edits changed prepared bytes")
		}
		io.WriteString(w, `[{"id":"20001","filename":"source.bin","size":7}]`)
	}))
	defer server.Close()
	client, err := NewClient(Config{BaseURL: server.URL, Token: attachmentSecret, InstanceType: InstanceCloud})
	if err != nil {
		t.Fatal(err)
	}
	attachment, err := client.UploadAttachment(upload)
	assertSafeUpload(t, client, attachment, err)
	if err != nil {
		t.Fatal(err)
	}
}

// Proves response validation, diagnostic truncation and credential redaction for Basic and Bearer without repeating POST.
func TestAttachmentUploadResponses(t *testing.T) {
	for _, auth := range []AuthType{AuthBasic, AuthBearer} {
		for _, tc := range []struct {
			name       string
			status     int
			body, want string
		}{
			{"valid", 200, `[{"id":"20001","filename":"report.txt","size":7}]`, ""},
			{"reflected success", 200, `[{"id":"` + attachmentSecret + `","filename":"Authorization: credentials\nCookie: cookie-secret","size":7}]`, ""},
			{"invalid JSON", 200, `{invalid ` + attachmentSecret, "invalid"},
			{"null", 200, "null", "expected one"},
			{"empty", 200, "[]", "expected one"},
			{"many", 200, `[{"id":"1"},{"id":"2"}]`, "expected one"},
			{"missing id", 200, `[{"filename":"report.txt","size":7}]`, "expected one"},
			{"negative size", 200, `[{"id":"1","filename":"report.txt","size":-1}]`, "expected one"},
			{"raw forbidden", 403, "Authorization: credentials\nCookie: cookie-secret\n" + attachmentSecret, "HTTP 403"},
			{"malformed JSON headers", 403, `{"Cookie":"cookie-secret","Authorization":"credentials"`, "HTTP 403"},
			{"JSON too large", 413, `{"errorMessages":["refused ` + attachmentSecret + strings.Repeat("x", 3000) + `"]}`, "HTTP 413"},
			{"server error", 500, `{"errorMessages":["` + attachmentSecret + `"]}`, "HTTP 500"},
			{"oversize response", 200, strings.Repeat("x", int(maxAttachmentResponseBytes)+1), "exceeds byte limit"},
			{"redirect", 307, "redirect", "HTTP 307"},
		} {
			t.Run(string(auth)+"/"+tc.name, func(t *testing.T) {
				var requests, posts atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests.Add(1)
					if r.Method == http.MethodGet {
						io.WriteString(w, `{"key":"PROJ-123","fields":{}}`)
						return
					}
					posts.Add(1)
					if tc.status == 307 {
						w.Header().Set("Location", "/must-not-follow")
					}
					w.WriteHeader(tc.status)
					io.WriteString(w, tc.body)
				}))
				defer server.Close()
				email := ""
				if auth == AuthBasic {
					email = "user@example.com"
				}
				client, err := NewClient(Config{BaseURL: server.URL, Email: email, Token: attachmentSecret, AuthType: auth, InstanceType: InstanceCloud})
				if err != nil {
					t.Fatal(err)
				}
				attachment, err := client.UploadAttachment(preparedUpload(t, "PROJ-123", "report.txt", 7))
				assertSafeUpload(t, client, attachment, err)
				if tc.want == "" {
					if err != nil {
						t.Fatal(err)
					}
				} else if err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("error=%v, want %q", err, tc.want)
				}
				if err != nil && len(err.Error()) > 1300 {
					t.Fatal("diagnostic was not truncated")
				}
				if requests.Load() != 2 || posts.Load() != 1 {
					t.Fatalf("requests=%d posts=%d", requests.Load(), posts.Load())
				}
			})
		}
	}
}

// Proves a reflected Authorization header is removed from the normal issue-read error before the upload is refused.
func TestAttachmentUploadPrecheckRedaction(t *testing.T) {
	for _, auth := range []AuthType{AuthBasic, AuthBearer} {
		t.Run(string(auth), func(t *testing.T) {
			var posts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					posts.Add(1)
				}
				w.WriteHeader(404)
				json.NewEncoder(w).Encode(map[string]any{"errorMessages": []string{fmt.Sprintf("refused %s\nAuthorization: %s\nCookie: cookie-secret", attachmentSecret, r.Header.Get("Authorization"))}})
			}))
			defer server.Close()
			client, err := NewClient(Config{BaseURL: server.URL, Email: "user@example.com", Token: attachmentSecret, AuthType: auth, InstanceType: InstanceCloud})
			if err != nil {
				t.Fatal(err)
			}
			attachment, err := client.UploadAttachment(preparedUpload(t, "PROJ-123", "report.txt", 7))
			assertSafeUpload(t, client, attachment, err)
			if err == nil || !strings.Contains(err.Error(), "HTTP 404") || posts.Load() != 0 {
				t.Fatalf("error=%v posts=%d", err, posts.Load())
			}
		})
	}
}

// Proves transport failure and timeout after POST produce safe errors and no second upload with possible duplicate effects.
func TestAttachmentUploadNoRetryAfterPartialEffect(t *testing.T) {
	for _, timeout := range []bool{false, true} {
		t.Run(fmt.Sprintf("timeout=%t", timeout), func(t *testing.T) {
			var posts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					io.WriteString(w, `{"key":"PROJ-123","fields":{}}`)
					return
				}
				posts.Add(1)
				if timeout {
					io.Copy(io.Discard, r.Body)
					<-r.Context().Done()
					return
				}
				connection, _, err := w.(http.Hijacker).Hijack()
				if err != nil {
					t.Error(err)
					return
				}
				connection.Close()
			}))
			defer server.Close()
			client, err := NewClient(Config{BaseURL: server.URL, Token: attachmentSecret, InstanceType: InstanceCloud})
			if err != nil {
				t.Fatal(err)
			}
			if timeout {
				client.SetHTTPClient(&http.Client{Timeout: 200 * time.Millisecond})
			}
			attachment, err := client.UploadAttachment(preparedUpload(t, "PROJ-123", "report.txt", 7))
			assertSafeUpload(t, client, attachment, err)
			if err == nil || posts.Load() != 1 {
				t.Fatalf("error=%v posts=%d", err, posts.Load())
			}
		})
	}
}

// Proves exact and encoded credentials are removed before truncation even if surrounded by a long diagnostic.
func TestAttachmentUploadSafeDiagnostics(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte("user@example.com:" + attachmentSecret))
	client, err := NewClient(Config{BaseURL: "https://jira.example", Email: "user@example.com", Token: attachmentSecret})
	if err != nil {
		t.Fatal(err)
	}
	message := client.safeAttachmentText(attachmentSecret + " " + encoded + " Authorization: " + client.authHeader + "\nSet-Cookie: cookie-secret\n" + strings.Repeat("x", 3000))
	assertSafeUpload(t, client, nil, fmt.Errorf("%s", message))
	if len(message) > 1100 || !strings.HasSuffix(message, "...") {
		t.Fatal("long message not truncated")
	}
}
