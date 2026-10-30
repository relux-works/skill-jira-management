package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
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

const uploadTestToken = "FAKE-ATTACHMENT-TOKEN-DO-NOT-PRINT"
const uploadTestBytes = "exact\x00file\r\nbytes"

type uploadFixture struct {
	file     string
	requests atomic.Int32
	posts    atomic.Int32
}

func newUploadFixture(t *testing.T, instance, filename string, getStatus, postStatus int, tls bool) *uploadFixture {
	t.Helper()
	f := &uploadFixture{}
	dir := t.TempDir()
	f.file = filepath.Join(dir, "source.bin")
	if err := os.WriteFile(f.file, []byte(uploadTestBytes), 0o600); err != nil {
		t.Fatal(err)
	}
	version := "3"
	authType, email := "basic", "user@example.com"
	if instance == "server" {
		version, authType, email = "2", "bearer", ""
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.requests.Add(1)
		if r.Header.Get("Authorization") == "" {
			t.Error("missing authentication")
		}
		if instance == "" && r.URL.Path == "/jira/rest/api/2/serverInfo" && r.Method == http.MethodGet {
			io.WriteString(w, `{"deploymentType":"Cloud"}`)
			return
		}
		if r.URL.Path == "/jira/rest/api/"+version+"/issue/PROJ-123" && r.Method == http.MethodGet {
			if getStatus != 200 {
				w.WriteHeader(getStatus)
				fmt.Fprintf(w, `{"errorMessages":["issue refused %s","Authorization: %s","Cookie: session=fake-cookie-secret"]}`, uploadTestToken, r.Header.Get("Authorization"))
				return
			}
			io.WriteString(w, `{"key":"PROJ-123","fields":{}}`)
			return
		}
		if r.Method != http.MethodPost || r.URL.Path != "/jira/rest/api/"+version+"/issue/PROJ-123/attachments" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(400)
			return
		}
		f.posts.Add(1)
		wantRequests := int32(2)
		if instance == "" {
			wantRequests = 3
		}
		if f.requests.Load() != wantRequests {
			t.Error("upload did not follow exactly one pre-check")
		}
		if r.Header.Get("X-Atlassian-Token") != "no-check" {
			t.Error("missing CSRF bypass header")
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		part, err := reader.NextPart()
		if err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		data, err := io.ReadAll(part)
		if err != nil {
			t.Error(err)
		}
		if part.FormName() != "file" || part.FileName() != filename || string(data) != uploadTestBytes {
			t.Errorf("multipart mismatch: field=%q filename=%q bytes=%q", part.FormName(), part.FileName(), data)
		}
		if _, err := reader.NextPart(); err != io.EOF {
			t.Errorf("unexpected extra part: %v", err)
		}
		if postStatus != 200 {
			w.WriteHeader(postStatus)
			fmt.Fprintf(w, `{"errorMessages":["upload refused %s","Authorization: %s","Set-Cookie: session=fake-cookie-secret"]}`, uploadTestToken, r.Header.Get("Authorization"))
			return
		}
		json.NewEncoder(w).Encode([]map[string]any{{"id": "20001", "filename": filename, "size": len(uploadTestBytes)}})
	})
	var server *httptest.Server
	if tls {
		server = httptest.NewTLSServer(handler)
	} else {
		server = httptest.NewServer(handler)
	}
	t.Cleanup(server.Close)
	cfgPath := filepath.Join(dir, "config.yaml")
	cfg := fmt.Sprintf("instance_url: %s/jira\ninstance_type: %s\nauth_type: %s\n", server.URL, instance, authType)
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	oldManager, oldResolver := newConfigManager, getCredentialResolver
	newConfigManager = func() (*config.ConfigManager, error) { return config.NewConfigManagerWithPath(cfgPath), nil }
	getCredentialResolver = func() *config.Resolver {
		return config.NewResolverWithAuthFilePath(config.Runtime{GOOS: "linux", Getenv: func(key string) string {
			return map[string]string{config.EnvAPIToken: uploadTestToken, config.EnvEmail: email, config.EnvAuthType: authType}[key]
		}}, nil, filepath.Join(dir, "unused-auth.json"))
	}
	t.Cleanup(func() { newConfigManager, getCredentialResolver = oldManager, oldResolver })
	return f
}

func executeUpload(t *testing.T, args ...string) (string, error) {
	t.Helper()
	// Cobra commands retain flag values between executions; restore defaults for every case.
	for _, name := range []string{"name", "max-bytes"} {
		flag := attachmentUploadCmd.Flags().Lookup(name)
		if err := flag.Value.Set(flag.DefValue); err != nil {
			t.Fatal(err)
		}
		flag.Changed = false
	}
	var stdout, stderr bytes.Buffer
	oldOut, oldErr := rootCmd.OutOrStdout(), rootCmd.ErrOrStderr()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	t.Cleanup(func() { rootCmd.SetOut(oldOut); rootCmd.SetErr(oldErr) })
	rootCmd.SetArgs(append([]string{"attachment", "upload"}, args...))
	err := rootCmd.Execute()
	errText := ""
	if err != nil {
		errText = err.Error()
	}
	for _, output := range []string{stdout.String(), stderr.String(), errText} {
		for _, secret := range []string{uploadTestToken, "Bearer " + uploadTestToken,
			"Basic " + base64.StdEncoding.EncodeToString([]byte("user@example.com:"+uploadTestToken)), "fake-cookie-secret"} {
			if strings.Contains(output, secret) {
				t.Fatal("credential leaked in command output or error")
			}
		}
	}
	return stdout.String(), err
}

// Proves Cloud and Server/DC uploads preserve exact multipart bytes, base paths, auth, names and both output formats.
func TestAttachmentUploadSuccess(t *testing.T) {
	for _, tc := range []struct {
		name, instance, filename, format string
		override, tls                    bool
	}{
		{"cloud default", "cloud", "source.bin", "json", false, false},
		{"cloud discovery", "", "source.bin", "json", false, false},
		{"server override", "server", "renamed.bin", "json", true, false},
		{"text TLS insecure", "cloud", "source.bin", "text", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newUploadFixture(t, tc.instance, tc.filename, 200, 200, tc.tls)
			args := []string{"PROJ-123", f.file}
			if tc.override {
				args = append(args, "--name", tc.filename)
			}
			if tc.format != "json" {
				args = append(args, "--format", tc.format)
			}
			if tc.tls {
				args = append(args, "--insecure")
			}
			defer func() { flagFormat = "json"; flagInsecure = false }()
			out, err := executeUpload(t, args...)
			if err != nil {
				t.Fatal(err)
			}
			wantRequests := int32(2)
			if tc.instance == "" {
				wantRequests = 3
			}
			if f.requests.Load() != wantRequests || f.posts.Load() != 1 {
				t.Fatalf("requests=%d posts=%d", f.requests.Load(), f.posts.Load())
			}
			if tc.format == "json" {
				var got map[string]any
				if err := json.Unmarshal([]byte(out), &got); err != nil {
					t.Fatal(err)
				}
				if len(got) != 4 || got["issue"] != "PROJ-123" || got["id"] != "20001" || got["filename"] != tc.filename || got["size"] != float64(len(uploadTestBytes)) {
					t.Fatalf("unexpected result: %s", out)
				}
			} else if strings.Count(out, "\n") != 1 || !strings.Contains(out, "PROJ-123") || !strings.Contains(out, "20001") || !strings.Contains(out, tc.filename) {
				t.Fatalf("unexpected text: %q", out)
			}
		})
	}
}

// Proves invalid local input sends zero requests, paired with the same configuration's successful regular-file control.
func TestAttachmentUploadLocalRefusals(t *testing.T) {
	for _, name := range []string{"missing", "directory", "symlink", "empty", "unreadable", "too large", "before discovery", "default cap", "invalid key", "zero cap", "negative cap", "name path", "name backslash", "empty name", "dot name", "control name"} {
		t.Run(name, func(t *testing.T) {
			instance := "cloud"
			if name == "before discovery" {
				instance = ""
			}
			f := newUploadFixture(t, instance, "source.bin", 200, 200, false)
			args := []string{"PROJ-123", f.file}
			want := ""
			switch name {
			case "missing":
				args[1] += ".missing"
				want = "inspect"
			case "directory":
				args[1] = filepath.Dir(f.file)
				want = "regular"
			case "symlink":
				link := f.file + ".link"
				if err := os.Symlink(f.file, link); err != nil {
					t.Fatal(err)
				}
				args[1] = link
				want = "regular"
			case "empty":
				args[1] = f.file + ".empty"
				if err := os.WriteFile(args[1], nil, 0o600); err != nil {
					t.Fatal(err)
				}
				want = "empty"
			case "unreadable":
				args[1] = f.file + ".unreadable"
				if err := os.WriteFile(args[1], []byte("data"), 0); err != nil {
					t.Fatal(err)
				}
				want = "read"
			case "too large", "before discovery":
				args = append(args, "--max-bytes", "1")
				want = "byte limit"
			case "default cap":
				args[1] = f.file + ".large"
				file, err := os.Create(args[1])
				if err != nil {
					t.Fatal(err)
				}
				err = file.Truncate(26214401)
				file.Close()
				if err != nil {
					t.Fatal(err)
				}
				want = "byte limit"
			case "invalid key":
				args[0] = "PROJ-123/../../other"
				want = "issue key"
			case "zero cap":
				args = append(args, "--max-bytes", "0")
				want = "positive"
			case "negative cap":
				args = append(args, "--max-bytes", "-1")
				want = "positive"
			case "name path":
				args = append(args, "--name", "../file")
				want = "base name"
			case "name backslash":
				args = append(args, "--name", "dir\\file")
				want = "base name"
			case "empty name":
				args = append(args, "--name", "")
				want = "base name"
			case "dot name":
				args = append(args, "--name", "..")
				want = "base name"
			case "control name":
				args = append(args, "--name", "bad\r\nname")
				want = "base name"
			}
			out, err := executeUpload(t, args...)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("error=%v, want %q", err, want)
			}
			if out != "" || f.requests.Load() != 0 {
				t.Fatalf("output=%q requests=%d", out, f.requests.Load())
			}
			if _, err := executeUpload(t, "PROJ-123", f.file); err != nil {
				t.Fatalf("positive control: %v", err)
			}
			wantRequests := int32(2)
			if instance == "" {
				wantRequests = 3
			}
			if f.requests.Load() != wantRequests || f.posts.Load() != 1 {
				t.Fatal("positive control did not upload")
			}
		})
	}
}

// Proves a 404 pre-check prevents POST, and 403/413 uploads return sanitized status errors without retrying.
func TestAttachmentUploadServerRefusals(t *testing.T) {
	for _, tc := range []struct {
		name                       string
		get, post, requests, posts int
	}{
		{"unknown issue", 404, 200, 1, 0}, {"forbidden", 200, 403, 2, 1}, {"too large", 200, 413, 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newUploadFixture(t, "cloud", "source.bin", tc.get, tc.post, false)
			out, err := executeUpload(t, "PROJ-123", f.file)
			status := tc.post
			if tc.get != 200 {
				status = tc.get
			}
			if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("HTTP %d", status)) {
				t.Fatalf("unclear status: %v", err)
			}
			if out != "" || int(f.requests.Load()) != tc.requests || int(f.posts.Load()) != tc.posts {
				t.Fatalf("output=%q requests=%d posts=%d", out, f.requests.Load(), f.posts.Load())
			}
			control := newUploadFixture(t, "cloud", "source.bin", 200, 200, false)
			if _, err := executeUpload(t, "PROJ-123", control.file); err != nil {
				t.Fatalf("positive control: %v", err)
			}
		})
	}
}
