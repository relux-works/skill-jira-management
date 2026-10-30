package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/relux-works/skill-jira-management/internal/config"
)

func configureUploadPolicyTest(t *testing.T, url, instance string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "source.bin")
	if err := os.WriteFile(path, []byte(uploadTestBytes), 0o600); err != nil {
		t.Fatal(err)
	}
	cfgPath := filepath.Join(dir, "config.yaml")
	data := fmt.Sprintf("instance_url: %s/jira\ninstance_type: %s\nauth_type: basic\n", url, instance)
	if err := os.WriteFile(cfgPath, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	oldManager, oldResolver := newConfigManager, getCredentialResolver
	newConfigManager = func() (*config.ConfigManager, error) { return config.NewConfigManagerWithPath(cfgPath), nil }
	getCredentialResolver = func() *config.Resolver {
		return config.NewResolverWithAuthFilePath(config.Runtime{GOOS: "linux", Getenv: func(key string) string {
			return map[string]string{config.EnvAPIToken: uploadTestToken, config.EnvEmail: "user@example.com", config.EnvAuthType: "basic"}[key]
		}}, nil, filepath.Join(dir, "unused-auth.json"))
	}
	t.Cleanup(func() { newConfigManager, getCredentialResolver = oldManager, oldResolver })
	return path
}

func paddedUploadResponse(body string, size int) string {
	return body + strings.Repeat(" ", size-len(body))
}

// Proves redirects and oversized success/error responses at discovery or pre-check stop the CLI before POST, while boundary and direct controls pass.
func TestAttachmentUploadReadPolicy(t *testing.T) {
	const capBytes = 1 << 20
	for _, phase := range []string{"discovery", "pre-check"} {
		for _, tc := range []struct {
			name            string
			status, size    int
			chunked, refuse bool
			want            string
		}{
			{"small direct control", 200, 128, false, false, ""},
			{"exact boundary", 200, capBytes, false, false, ""},
			{"exact chunked boundary", 200, capBytes, true, false, ""},
			{"redirect", 302, 128, false, true, "HTTP 302"},
			{"redirect with truncated body", 302, 128, false, true, "HTTP 302"},
			{"over boundary", 200, capBytes + 1, false, true, "exceeds byte limit"},
			{"over boundary truncated body", 200, capBytes + 1, false, true, "exceeds byte limit"},
			{"over chunked boundary", 200, capBytes + 1, true, true, "exceeds byte limit"},
			{"oversized error", 403, capBytes + 1, false, true, "exceeds byte limit"},
			{"oversized chunked error", 403, capBytes + 1, true, true, "exceeds byte limit"},
		} {
			t.Run(phase+"/"+tc.name, func(t *testing.T) {
				var requests, posts, followed atomic.Int32
				issueBody := `{"key":"PROJ-123","fields":{}}`
				discoveryBody := `{"deploymentType":"Cloud"}`
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests.Add(1)
					if r.Header.Get("Authorization") == "" {
						t.Error("missing Authorization")
					}
					if r.Method == http.MethodPost {
						posts.Add(1)
						io.WriteString(w, `[{"id":"20001","filename":"source.bin","size":17}]`)
						return
					}
					if r.URL.Path == "/jira/redirected" {
						followed.Add(1)
						if phase == "discovery" {
							io.WriteString(w, discoveryBody)
						} else {
							io.WriteString(w, issueBody)
						}
						return
					}
					target := "/jira/rest/api/3/issue/PROJ-123"
					body := issueBody
					if phase == "discovery" {
						target = "/jira/rest/api/2/serverInfo"
						body = discoveryBody
					}
					if r.URL.Path != target {
						if r.URL.Path != "/jira/rest/api/3/issue/PROJ-123" {
							t.Errorf("unexpected path: %s", r.URL.Path)
						}
						io.WriteString(w, issueBody)
						return
					}
					if tc.status == 302 {
						w.Header().Set("Location", "/jira/redirected")
					}
					if tc.status == 403 {
						body = `{"errorMessages":["` + uploadTestToken + ` Cookie: fake-cookie-secret"]}`
					}
					body = paddedUploadResponse(body, tc.size)
					if !tc.chunked {
						length := len(body)
						if strings.Contains(tc.name, "truncated body") {
							length++
						}
						w.Header().Set("Content-Length", fmt.Sprint(length))
					}
					w.WriteHeader(tc.status)
					if tc.chunked {
						w.(http.Flusher).Flush()
					}
					io.WriteString(w, body)
				}))
				defer server.Close()
				instance := "cloud"
				if phase == "discovery" {
					instance = ""
				}
				path := configureUploadPolicyTest(t, server.URL, instance)
				out, err := executeUpload(t, "PROJ-123", path)
				if tc.refuse {
					if err == nil || !strings.Contains(err.Error(), tc.want) {
						t.Errorf("error=%v, want %q", err, tc.want)
					}
					if out != "" || posts.Load() != 0 || followed.Load() != 0 || requests.Load() != 1 {
						t.Errorf("output=%q requests=%d posts=%d redirects=%d", out, requests.Load(), posts.Load(), followed.Load())
					}
				} else {
					wantRequests := int32(2)
					if phase == "discovery" {
						wantRequests = 3
					}
					if err != nil || posts.Load() != 1 || followed.Load() != 0 || requests.Load() != wantRequests {
						t.Fatalf("control error=%v requests=%d posts=%d redirects=%d", err, requests.Load(), posts.Load(), followed.Load())
					}
				}
			})
		}
	}
}
