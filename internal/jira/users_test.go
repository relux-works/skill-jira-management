package jira

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTypedTestClient(t *testing.T, baseURL string, kind InstanceType) *Client {
	t.Helper()
	c, err := NewClient(Config{
		BaseURL:      baseURL,
		Email:        "user@test.com",
		Token:        "test-token",
		InstanceType: kind,
	})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return c
}

// Proves each instance type is addressed by the field it actually accepts:
// Cloud rejects "name" and Server/DC rejects "accountId", so sending the wrong
// one silently leaves the issue unassigned.
func TestAssignIssue_UsesInstanceSpecificField(t *testing.T) {
	cases := []struct {
		kind      InstanceType
		user      User
		wantKey   string
		wantValue interface{}
	}{
		{InstanceCloud, User{AccountID: "acc-1", Name: "ignored"}, "accountId", "acc-1"},
		{InstanceServer, User{Name: "aagrigore1", AccountID: "ignored"}, "name", "aagrigore1"},
	}
	for _, tc := range cases {
		var gotBody map[string]interface{}
		var gotMethod, gotPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod, gotPath = r.Method, r.URL.Path
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &gotBody)
			w.WriteHeader(http.StatusNoContent)
		}))
		c := newTypedTestClient(t, srv.URL, tc.kind)
		if err := c.AssignIssue("PROJ-1", &tc.user); err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.kind, err)
		}
		srv.Close()

		if gotMethod != "PUT" {
			t.Errorf("%s: method = %q, want PUT", tc.kind, gotMethod)
		}
		if !strings.HasSuffix(gotPath, "/issue/PROJ-1/assignee") {
			t.Errorf("%s: path = %q", tc.kind, gotPath)
		}
		if gotBody[tc.wantKey] != tc.wantValue {
			t.Errorf("%s: body[%q] = %v, want %v", tc.kind, tc.wantKey, gotBody[tc.wantKey], tc.wantValue)
		}
		if len(gotBody) != 1 {
			t.Errorf("%s: body carries %d fields, want only %q: %v", tc.kind, len(gotBody), tc.wantKey, gotBody)
		}
	}
}

// Clearing must send an explicit null, not an omitted field: an empty body
// leaves the assignee untouched while the command reports success.
func TestAssignIssue_ClearSendsExplicitNull(t *testing.T) {
	var raw string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		raw = string(data)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTypedTestClient(t, srv.URL, InstanceServer)
	if err := c.AssignIssue("PROJ-1", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(raw, `"name":null`) {
		t.Errorf("clear body = %s, want an explicit null name", raw)
	}
}

// A user resolved without the field the instance needs must fail loudly,
// rather than issue a PUT that quietly does nothing.
func TestAssignIssue_RefusesUserMissingAddressingField(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTypedTestClient(t, srv.URL, InstanceCloud)
	if err := c.AssignIssue("PROJ-1", &User{Name: "only-a-username"}); err == nil {
		t.Fatal("expected an error for a Cloud assignment without accountId")
	}
	if called {
		t.Error("a request was sent although the user could not be addressed")
	}
}

func TestFindUser(t *testing.T) {
	cases := []struct {
		name    string
		kind    InstanceType
		users   []User
		query   string
		wantErr string
		want    string
	}{
		{
			name:  "single match",
			kind:  InstanceServer,
			users: []User{{Name: "aagrigore1", DisplayName: "Alexis"}},
			query: "aagrigore1",
			want:  "aagrigore1",
		},
		{
			name: "ambiguous resolved by exact username",
			kind: InstanceServer,
			users: []User{
				{Name: "alex", DisplayName: "Alex One"},
				{Name: "alexander", DisplayName: "Alex Two"},
			},
			query: "alex",
			want:  "alex",
		},
		{
			name: "ambiguous without an exact match is an error",
			kind: InstanceServer,
			users: []User{
				{Name: "alex1", DisplayName: "Alex One"},
				{Name: "alex2", DisplayName: "Alex Two"},
			},
			query:   "alex",
			wantErr: "matches 2 Jira users",
		},
		{
			name:    "no match is an error",
			kind:    InstanceServer,
			users:   []User{},
			query:   "nobody",
			wantErr: "no Jira user matches",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(tc.users)
			}))
			defer srv.Close()

			c := newTypedTestClient(t, srv.URL, tc.kind)
			u, err := c.FindUser(tc.query)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected an error containing %q, got user %+v", tc.wantErr, u)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("error = %v, want it to contain %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.Name != tc.want {
				t.Errorf("resolved %q, want %q", u.Name, tc.want)
			}
		})
	}
}

// Cloud and Server/DC use different query parameters for user search; sending
// the wrong one returns an unrelated result set.
func TestFindUser_QueryParameterPerInstance(t *testing.T) {
	for _, tc := range []struct {
		kind  InstanceType
		param string
	}{{InstanceCloud, "query"}, {InstanceServer, "username"}} {
		var got string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got = r.URL.Query().Get(tc.param)
			json.NewEncoder(w).Encode([]User{{AccountID: "a", Name: "n"}})
		}))
		c := newTypedTestClient(t, srv.URL, tc.kind)
		if _, err := c.FindUser("someone"); err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.kind, err)
		}
		srv.Close()
		if got != "someone" {
			t.Errorf("%s: parameter %q = %q, want \"someone\"", tc.kind, tc.param, got)
		}
	}
}
