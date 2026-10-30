//go:build darwin || linux

package jira

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// Proves the actual post-Lstat open boundary promptly rejects FIFO/same-inode symlink substitutions before network activity and accepts regular files.
func TestAttachmentUploadFileSubstitution(t *testing.T) {
	for _, kind := range []string{"regular control", "FIFO", "same-inode symlink", "different regular inode"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "source.bin")
			if err := os.WriteFile(path, []byte("payload"), 0o600); err != nil {
				t.Fatal(err)
			}
			expected, err := os.Lstat(path)
			if err != nil {
				t.Fatal(err)
			}
			if kind != "regular control" {
				backup := path + ".original"
				if err := os.Rename(path, backup); err != nil {
					t.Fatal(err)
				}
				switch kind {
				case "FIFO":
					if err := syscall.Mkfifo(path, 0o600); err != nil {
						t.Fatal(err)
					}
				case "same-inode symlink":
					if err := os.Symlink(backup, path); err != nil {
						t.Fatal(err)
					}
				case "different regular inode":
					if err := os.WriteFile(path, []byte("replacement"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.Method == http.MethodGet {
					io.WriteString(w, `{"key":"PROJ-123","fields":{}}`)
					return
				}
				io.WriteString(w, `[{"id":"20001","filename":"source.bin","size":7}]`)
			}))
			defer server.Close()
			type result struct {
				file *os.File
				err  error
			}
			done := make(chan result, 1)
			// Supply the saved Lstat identity to the same helper used by production, after deterministic substitution.
			go func() { file, err := openAttachmentFile(path, expected); done <- result{file, err} }()
			var got result
			select {
			case got = <-done:
			case <-time.After(time.Second):
				// A blocking-open regression is unblocked and joined without leaving a goroutine behind.
				if kind == "FIFO" {
					fd, err := syscall.Open(path, syscall.O_WRONLY|syscall.O_NONBLOCK, 0)
					if err != nil {
						t.Fatalf("blocked FIFO could not be released: %v", err)
					}
					syscall.Close(fd)
				}
				got = <-done
				if got.file != nil {
					got.file.Close()
				}
				t.Fatal("opening substituted file blocked for one second")
			}
			if kind != "regular control" {
				if got.file != nil {
					got.file.Close()
				}
				if got.err == nil || requests.Load() != 0 {
					t.Fatalf("substitution accepted: error=%v requests=%d", got.err, requests.Load())
				}
				return
			}
			if got.err != nil {
				t.Fatal(got.err)
			}
			body, err := io.ReadAll(got.file)
			got.file.Close()
			if err != nil || string(body) != "payload" {
				t.Fatalf("regular descriptor: bytes=%q error=%v", body, err)
			}
			upload, err := PrepareAttachmentUpload("PROJ-123", path, "", 7)
			if err != nil {
				t.Fatal(err)
			}
			client, err := NewClient(Config{BaseURL: server.URL, Token: attachmentSecret, InstanceType: InstanceCloud})
			if err != nil {
				t.Fatal(err)
			}
			attachment, err := client.UploadAttachment(upload)
			assertSafeUpload(t, client, attachment, err)
			if err != nil || requests.Load() != 2 {
				t.Fatalf("regular upload error=%v requests=%d", err, requests.Load())
			}
		})
	}
}
