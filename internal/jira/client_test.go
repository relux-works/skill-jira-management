package jira

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- Story 1: HTTP Client & Auth ---

func TestNewClient_Valid(t *testing.T) {
	c, err := NewClient(Config{
		BaseURL: "https://test.atlassian.net",
		Email:   "user@test.com",
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("client is nil")
	}
	if c.baseURL != "https://test.atlassian.net" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "https://test.atlassian.net")
	}
	// Auth header should be Basic base64("user@test.com:test-token")
	wantAuth := "Basic dXNlckB0ZXN0LmNvbTp0ZXN0LXRva2Vu"
	if c.authHeader != wantAuth {
		t.Errorf("authHeader = %q, want %q", c.authHeader, wantAuth)
	}
}

func TestNewClient_TrailingSlash(t *testing.T) {
	c, err := NewClient(Config{
		BaseURL: "https://test.atlassian.net/",
		Email:   "user@test.com",
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.baseURL != "https://test.atlassian.net" {
		t.Errorf("trailing slash not trimmed: %q", c.baseURL)
	}
}

func TestNewClient_MissingBaseURL(t *testing.T) {
	_, err := NewClient(Config{Email: "x", Token: "y"})
	if err == nil {
		t.Fatal("expected error for missing base URL")
	}
}

func TestNewClient_MissingCredentials(t *testing.T) {
	_, err := NewClient(Config{BaseURL: "https://x.atlassian.net"})
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}
}

func TestClient_AuthHeaderSent(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, _ = c.Get("/test", nil)

	wantAuth := "Basic dXNlckB0ZXN0LmNvbTp0ZXN0LXRva2Vu"
	if gotAuth != wantAuth {
		t.Errorf("Authorization header = %q, want %q", gotAuth, wantAuth)
	}
}

func TestClient_ContentTypeJSON(t *testing.T) {
	var gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, _ = c.Post("/test", map[string]string{"foo": "bar"})

	if gotCT != "application/json" {
		t.Errorf("Content-Type = %q, want %q", gotCT, "application/json")
	}
}

func TestClient_GetPostPutDelete(t *testing.T) {
	var methods []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)

	c.Get("/a", nil)
	c.Post("/b", map[string]string{})
	c.Put("/c", map[string]string{})
	c.Delete("/d")

	want := []string{"GET", "POST", "PUT", "DELETE"}
	if len(methods) != len(want) {
		t.Fatalf("methods = %v, want %v", methods, want)
	}
	for i := range want {
		if methods[i] != want[i] {
			t.Errorf("methods[%d] = %q, want %q", i, methods[i], want[i])
		}
	}
}

func TestClient_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Bad request: field is required"},
			"errors":        map[string]string{},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.Get("/fail", nil)
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
	if !strings.Contains(apiErr.Error(), "Bad request") {
		t.Errorf("error message = %q, want to contain 'Bad request'", apiErr.Error())
	}
}

func TestClient_RetryOn429(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"errorMessages":["Rate limit"]}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	data, err := c.Get("/rate-limit", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts < 2 {
		t.Errorf("expected retry, got %d attempts", attempts)
	}
	if !strings.Contains(string(data), "ok") {
		t.Errorf("unexpected response: %s", string(data))
	}
}

func TestClient_RetryOn500(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"errorMessages":["Server error"]}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	data, err := c.Get("/server-error", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts < 2 {
		t.Errorf("expected retry, got %d attempts", attempts)
	}
	if !strings.Contains(string(data), "ok") {
		t.Errorf("unexpected response: %s", string(data))
	}
}

// --- Story 2: Issues CRUD ---

func TestGetIssue(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/rest/api/3/issue/PROJ-1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		json.NewEncoder(w).Encode(Issue{
			ID:  "10001",
			Key: "PROJ-1",
			Fields: IssueFields{
				Summary: "Test issue",
				IssueType: IssueType{
					ID:   "10000",
					Name: "Task",
				},
				Status: &Status{
					Name: "Open",
				},
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	issue, err := c.GetIssue("PROJ-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if issue.Key != "PROJ-1" {
		t.Errorf("Key = %q, want PROJ-1", issue.Key)
	}
	if issue.Fields.Summary != "Test issue" {
		t.Errorf("Summary = %q, want 'Test issue'", issue.Fields.Summary)
	}
	if issue.Fields.Status.Name != "Open" {
		t.Errorf("Status = %q, want 'Open'", issue.Fields.Status.Name)
	}
}

func TestGetIssue_WithFields(t *testing.T) {
	var gotFields string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFields = r.URL.Query().Get("fields")
		json.NewEncoder(w).Encode(Issue{Key: "X-1"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, _ = c.GetIssue("X-1", []string{"summary", "status"})

	if gotFields != "summary,status" {
		t.Errorf("fields param = %q, want 'summary,status'", gotFields)
	}
}

func TestGetIssue_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Issue does not exist"},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.GetIssue("NOPE-999", nil)
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

func TestCreateIssue(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/rest/api/3/issue" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &gotBody)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(CreateIssueResponse{
			ID:  "10002",
			Key: "PROJ-2",
			Self: "https://test.atlassian.net/rest/api/3/issue/10002",
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.CreateIssue(&CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   ProjectRef{Key: "PROJ"},
			IssueType: IssueTypeRef{Name: "Story"},
			Summary:   "New story",
			Description: NewADFText("Story description"),
			Labels:    []string{"backend"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Key != "PROJ-2" {
		t.Errorf("Key = %q, want PROJ-2", resp.Key)
	}

	// Verify the request body structure.
	fields, ok := gotBody["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("request body missing 'fields'")
	}
	if fields["summary"] != "New story" {
		t.Errorf("summary = %v, want 'New story'", fields["summary"])
	}
}

func TestCreateIssue_Subtask(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &gotBody)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(CreateIssueResponse{Key: "PROJ-3"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.CreateIssue(&CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   ProjectRef{Key: "PROJ"},
			IssueType: IssueTypeRef{Name: "Subtask"},
			Summary:   "Sub task",
			Parent:    &IssueRef{Key: "PROJ-1"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fields := gotBody["fields"].(map[string]interface{})
	parent, ok := fields["parent"].(map[string]interface{})
	if !ok {
		t.Fatal("request body missing 'parent'")
	}
	if parent["key"] != "PROJ-1" {
		t.Errorf("parent.key = %v, want PROJ-1", parent["key"])
	}
}

func TestCreateIssue_CustomFields(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &gotBody)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(CreateIssueResponse{Key: "PROJ-4"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.CreateIssue(&CreateIssueRequest{
		Fields: CreateIssueFields{
			Project:   ProjectRef{Key: "PROJ"},
			IssueType: IssueTypeRef{Name: "Task"},
			Summary:   "Task with custom fields",
			Extra: map[string]interface{}{
				"customfield_10001": 8,
				"customfield_10002": "DoD text",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fields := gotBody["fields"].(map[string]interface{})
	if fields["customfield_10001"] != float64(8) {
		t.Errorf("customfield_10001 = %v, want 8", fields["customfield_10001"])
	}
	if fields["customfield_10002"] != "DoD text" {
		t.Errorf("customfield_10002 = %v, want 'DoD text'", fields["customfield_10002"])
	}
}

func TestUpdateIssue(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	err := c.UpdateIssue("PROJ-1", &UpdateIssueRequest{
		Fields: map[string]interface{}{
			"summary":     "Updated summary",
			"description": NewADFText("Updated description"),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != "PUT" {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	if gotPath != "/rest/api/3/issue/PROJ-1" {
		t.Errorf("path = %q, want /rest/api/3/issue/PROJ-1", gotPath)
	}
}

func TestListIssues(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/search/jql" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		callCount++

		var req SearchRequest
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &req)

		// First page.
		if req.NextPageToken == "" {
			json.NewEncoder(w).Encode(SearchResponse{
				Issues: []Issue{
					{Key: "PROJ-1", Fields: IssueFields{Summary: "Issue 1"}},
					{Key: "PROJ-2", Fields: IssueFields{Summary: "Issue 2"}},
				},
				NextPageToken: "page2token",
				IsLast:        false,
			})
			return
		}

		// Second (last) page.
		json.NewEncoder(w).Encode(SearchResponse{
			Issues: []Issue{
				{Key: "PROJ-3", Fields: IssueFields{Summary: "Issue 3"}},
			},
			NextPageToken: "",
			IsLast:        true,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	issues, err := c.ListIssues(ListIssuesOptions{
		ProjectKey: "PROJ",
		MaxResults: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 3 {
		t.Fatalf("got %d issues, want 3", len(issues))
	}
	if issues[2].Key != "PROJ-3" {
		t.Errorf("issues[2].Key = %q, want PROJ-3", issues[2].Key)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls (pagination), got %d", callCount)
	}
}

// --- Story 3: Projects & Boards ---

func TestListProjects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/project/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(ProjectSearchResult{
			Values: []Project{
				{ID: "1", Key: "PROJ", Name: "Project One"},
				{ID: "2", Key: "TEST", Name: "Test Project"},
			},
			IsLast: true,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	projects, err := c.ListProjects()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("got %d projects, want 2", len(projects))
	}
	if projects[0].Key != "PROJ" {
		t.Errorf("projects[0].Key = %q, want PROJ", projects[0].Key)
	}
}

func TestListProjects_Paginated(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		startAt := r.URL.Query().Get("startAt")

		if startAt == "0" || startAt == "" {
			json.NewEncoder(w).Encode(ProjectSearchResult{
				Values: []Project{{Key: "A"}, {Key: "B"}},
				IsLast: false,
			})
		} else {
			json.NewEncoder(w).Encode(ProjectSearchResult{
				Values: []Project{{Key: "C"}},
				IsLast: true,
			})
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	projects, err := c.ListProjects()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 3 {
		t.Fatalf("got %d projects, want 3", len(projects))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestListBoards(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/agile/1.0/board" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if proj := r.URL.Query().Get("projectKeyOrId"); proj != "PROJ" {
			t.Errorf("projectKeyOrId = %q, want PROJ", proj)
		}
		json.NewEncoder(w).Encode(BoardSearchResult{
			Values: []Board{
				{ID: 1, Name: "Scrum Board", Type: "scrum"},
				{ID: 2, Name: "Kanban Board", Type: "kanban"},
			},
			IsLast: true,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	boards, err := c.ListBoards("PROJ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(boards) != 2 {
		t.Fatalf("got %d boards, want 2", len(boards))
	}
	if boards[0].Type != "scrum" {
		t.Errorf("boards[0].Type = %q, want 'scrum'", boards[0].Type)
	}
}

func TestGetBoard(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/agile/1.0/board/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(Board{
			ID:   42,
			Name: "My Board",
			Type: "scrum",
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	board, err := c.GetBoard(42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if board.ID != 42 {
		t.Errorf("ID = %d, want 42", board.ID)
	}
	if board.Name != "My Board" {
		t.Errorf("Name = %q, want 'My Board'", board.Name)
	}
}

func TestListSprints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/agile/1.0/board/1/sprint" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(SprintSearchResult{
			Values: []Sprint{
				{ID: 1, Name: "Sprint 1", State: "closed"},
				{ID: 2, Name: "Sprint 2", State: "active"},
				{ID: 3, Name: "Sprint 3", State: "future"},
			},
			IsLast: true,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	sprints, err := c.ListSprints(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sprints) != 3 {
		t.Fatalf("got %d sprints, want 3", len(sprints))
	}
	if sprints[1].State != "active" {
		t.Errorf("sprints[1].State = %q, want 'active'", sprints[1].State)
	}
}

// --- Story 4: Transitions & Statuses ---

func TestGetTransitions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/issue/PROJ-1/transitions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(TransitionsResponse{
			Transitions: []Transition{
				{
					ID:   "11",
					Name: "Start Progress",
					To:   Status{ID: "3", Name: "In Progress"},
				},
				{
					ID:   "21",
					Name: "Done",
					To:   Status{ID: "5", Name: "Done"},
				},
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	transitions, err := c.GetTransitions("PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(transitions) != 2 {
		t.Fatalf("got %d transitions, want 2", len(transitions))
	}
	if transitions[0].Name != "Start Progress" {
		t.Errorf("transitions[0].Name = %q, want 'Start Progress'", transitions[0].Name)
	}
	if transitions[0].To.Name != "In Progress" {
		t.Errorf("transitions[0].To.Name = %q, want 'In Progress'", transitions[0].To.Name)
	}
}

func TestDoTransition(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/rest/api/3/issue/PROJ-1/transitions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	err := c.DoTransition("PROJ-1", "11", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	transition, ok := gotBody["transition"].(map[string]interface{})
	if !ok {
		t.Fatal("request body missing 'transition'")
	}
	if transition["id"] != "11" {
		t.Errorf("transition.id = %v, want '11'", transition["id"])
	}
}

func TestDoTransition_WithFields(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	err := c.DoTransition("PROJ-1", "21", map[string]interface{}{
		"resolution": map[string]string{"name": "Done"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fields, ok := gotBody["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("request body missing 'fields'")
	}
	res := fields["resolution"].(map[string]interface{})
	if res["name"] != "Done" {
		t.Errorf("resolution.name = %v, want 'Done'", res["name"])
	}
}

func TestListStatuses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/status" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]Status{
			{ID: "1", Name: "Open", StatusCategory: &StatusCategory{Key: "new", Name: "To Do"}},
			{ID: "3", Name: "In Progress", StatusCategory: &StatusCategory{Key: "indeterminate", Name: "In Progress"}},
			{ID: "5", Name: "Done", StatusCategory: &StatusCategory{Key: "done", Name: "Done"}},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	statuses, err := c.ListStatuses()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(statuses) != 3 {
		t.Fatalf("got %d statuses, want 3", len(statuses))
	}
	if statuses[2].StatusCategory.Key != "done" {
		t.Errorf("statuses[2].StatusCategory.Key = %q, want 'done'", statuses[2].StatusCategory.Key)
	}
}

// --- Story 5: JQL Search ---

func TestSearchJQL(t *testing.T) {
	var gotBody SearchRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/rest/api/3/search/jql" {
			t.Errorf("path = %q, want /rest/api/3/search/jql", r.URL.Path)
		}

		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &gotBody)

		json.NewEncoder(w).Encode(SearchResponse{
			Issues: []Issue{
				{Key: "PROJ-1", Fields: IssueFields{Summary: "Found"}},
			},
			IsLast: true,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.SearchJQL(&SearchRequest{
		JQL:        "project = PROJ AND status = 'Open'",
		MaxResults: 50,
		Fields:     []string{"summary", "status"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Issues) != 1 {
		t.Fatalf("got %d issues, want 1", len(resp.Issues))
	}
	if gotBody.JQL != "project = PROJ AND status = 'Open'" {
		t.Errorf("JQL = %q", gotBody.JQL)
	}
	if !resp.IsLast {
		t.Error("expected IsLast = true")
	}
}

func TestSearchAll_Pagination(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var req SearchRequest
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &req)

		if req.NextPageToken == "" {
			json.NewEncoder(w).Encode(SearchResponse{
				Issues:        []Issue{{Key: "A-1"}, {Key: "A-2"}},
				NextPageToken: "tok2",
				IsLast:        false,
			})
		} else if req.NextPageToken == "tok2" {
			json.NewEncoder(w).Encode(SearchResponse{
				Issues:        []Issue{{Key: "A-3"}},
				NextPageToken: "",
				IsLast:        true,
			})
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	issues, err := c.SearchAll("project = A", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 3 {
		t.Fatalf("got %d issues, want 3", len(issues))
	}
	if callCount != 2 {
		t.Errorf("expected 2 paginated calls, got %d", callCount)
	}
}

// --- Story 6: Comments ---

func TestNewADFText(t *testing.T) {
	doc := NewADFText("Hello world")
	if doc.Type != "doc" {
		t.Errorf("Type = %q, want 'doc'", doc.Type)
	}
	if doc.Version != 1 {
		t.Errorf("Version = %d, want 1", doc.Version)
	}
	if len(doc.Content) != 1 {
		t.Fatalf("Content length = %d, want 1", len(doc.Content))
	}
	para := doc.Content[0]
	if para.Type != "paragraph" {
		t.Errorf("Content[0].Type = %q, want 'paragraph'", para.Type)
	}
	if len(para.Content) != 1 || para.Content[0].Text != "Hello world" {
		t.Errorf("text = %q, want 'Hello world'", para.Content[0].Text)
	}
}

func TestNewADFParagraphs(t *testing.T) {
	doc := NewADFParagraphs([]string{"First", "Second", "Third"})
	if len(doc.Content) != 3 {
		t.Fatalf("Content length = %d, want 3", len(doc.Content))
	}
	for i, text := range []string{"First", "Second", "Third"} {
		if doc.Content[i].Content[0].Text != text {
			t.Errorf("paragraph %d text = %q, want %q", i, doc.Content[i].Content[0].Text, text)
		}
	}
}

func TestNewADFBulletList(t *testing.T) {
	doc := NewADFBulletList([]string{"item1", "item2"})
	if len(doc.Content) != 1 {
		t.Fatalf("Content length = %d, want 1 (bulletList)", len(doc.Content))
	}
	list := doc.Content[0]
	if list.Type != "bulletList" {
		t.Errorf("Type = %q, want 'bulletList'", list.Type)
	}
	if len(list.Content) != 2 {
		t.Fatalf("list items = %d, want 2", len(list.Content))
	}
}

func TestNewADFCodeBlock(t *testing.T) {
	doc := NewADFCodeBlock("fmt.Println(\"hello\")", "go")
	if len(doc.Content) != 1 || doc.Content[0].Type != "codeBlock" {
		t.Fatal("expected codeBlock node")
	}
	// Check attrs contain language.
	var attrs map[string]string
	json.Unmarshal(doc.Content[0].Attrs, &attrs)
	if attrs["language"] != "go" {
		t.Errorf("language = %q, want 'go'", attrs["language"])
	}
	if doc.Content[0].Content[0].Text != "fmt.Println(\"hello\")" {
		t.Errorf("code text mismatch")
	}
}

func TestNewADFWithHeading(t *testing.T) {
	doc := NewADFWithHeading(2, "Title", "Body text")
	if len(doc.Content) != 2 {
		t.Fatalf("Content length = %d, want 2", len(doc.Content))
	}
	heading := doc.Content[0]
	if heading.Type != "heading" {
		t.Errorf("Type = %q, want 'heading'", heading.Type)
	}
	var attrs map[string]int
	json.Unmarshal(heading.Attrs, &attrs)
	if attrs["level"] != 2 {
		t.Errorf("heading level = %d, want 2", attrs["level"])
	}
}

func TestAddComment(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/rest/api/3/issue/PROJ-1/comment" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &gotBody)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Comment{
			ID:      "10001",
			Created: "2026-01-01T10:00:00.000+0000",
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	comment, err := c.AddComment("PROJ-1", NewADFText("Test comment"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comment.ID != "10001" {
		t.Errorf("comment.ID = %q, want '10001'", comment.ID)
	}

	// Verify body is ADF.
	body, ok := gotBody["body"].(map[string]interface{})
	if !ok {
		t.Fatal("missing body in request")
	}
	if body["type"] != "doc" {
		t.Errorf("body.type = %v, want 'doc'", body["type"])
	}
}

func TestListComments(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/issue/PROJ-1/comment" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(CommentsResponse{
			StartAt:    0,
			MaxResults: 50,
			Total:      2,
			Comments: []Comment{
				{ID: "1", Body: NewADFText("First comment")},
				{ID: "2", Body: NewADFText("Second comment")},
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	resp, err := c.ListComments("PROJ-1", 0, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Comments) != 2 {
		t.Fatalf("got %d comments, want 2", len(resp.Comments))
	}
	if resp.Total != 2 {
		t.Errorf("total = %d, want 2", resp.Total)
	}
}

func TestListAllComments_Paginated(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		startAt := r.URL.Query().Get("startAt")

		if startAt == "0" || startAt == "" {
			json.NewEncoder(w).Encode(CommentsResponse{
				StartAt:    0,
				MaxResults: 2,
				Total:      3,
				Comments: []Comment{
					{ID: "1"},
					{ID: "2"},
				},
			})
		} else {
			json.NewEncoder(w).Encode(CommentsResponse{
				StartAt:    2,
				MaxResults: 2,
				Total:      3,
				Comments: []Comment{
					{ID: "3"},
				},
			})
		}
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	comments, err := c.ListAllComments("PROJ-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(comments) != 3 {
		t.Fatalf("got %d comments, want 3", len(comments))
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

// --- Path builders ---

func TestAPIPath(t *testing.T) {
	p := apiPath("issue", "PROJ-1")
	if p != "/rest/api/3/issue/PROJ-1" {
		t.Errorf("apiPath = %q", p)
	}
}

func TestAgileAPIPath(t *testing.T) {
	p := agileAPIPath("board", "42", "sprint")
	if p != "/rest/agile/1.0/board/42/sprint" {
		t.Errorf("agileAPIPath = %q", p)
	}
}

// --- Test helpers ---

func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	c, err := NewClient(Config{
		BaseURL: baseURL,
		Email:   "user@test.com",
		Token:   "test-token",
	})
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}
	return c
}

// --- Additional coverage tests ---

func TestGetProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/project/PROJ" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(Project{
			ID:   "10000",
			Key:  "PROJ",
			Name: "My Project",
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	proj, err := c.GetProject("PROJ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if proj.Key != "PROJ" {
		t.Errorf("Key = %q, want PROJ", proj.Key)
	}
	if proj.Name != "My Project" {
		t.Errorf("Name = %q, want 'My Project'", proj.Name)
	}
}

func TestListProjectStatuses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/3/project/PROJ/statuses" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]ProjectIssueTypeStatuses{
			{
				ID:   "10000",
				Name: "Task",
				Statuses: []Status{
					{ID: "1", Name: "Open"},
					{ID: "3", Name: "In Progress"},
					{ID: "5", Name: "Done"},
				},
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	result, err := c.ListProjectStatuses("PROJ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("got %d issue types, want 1", len(result))
	}
	if len(result[0].Statuses) != 3 {
		t.Errorf("got %d statuses, want 3", len(result[0].Statuses))
	}
}

func TestSetHTTPClient(t *testing.T) {
	c := &Client{}
	customClient := &http.Client{}
	c.SetHTTPClient(customClient)
	if c.httpClient != customClient {
		t.Error("SetHTTPClient did not set the client")
	}
}

func TestAPIError_ErrorMessages(t *testing.T) {
	err := &APIError{
		StatusCode:    400,
		ErrorMessages: []string{"First error", "Second error"},
	}
	if err.Error() != "First error" {
		t.Errorf("Error() = %q, want 'First error'", err.Error())
	}
}

func TestAPIError_FieldErrors(t *testing.T) {
	err := &APIError{
		StatusCode: 400,
		Errors:     map[string]string{"summary": "Field is required"},
	}
	got := err.Error()
	if !strings.Contains(got, "summary") || !strings.Contains(got, "Field is required") {
		t.Errorf("Error() = %q, want to contain field error", got)
	}
}

func TestAPIError_Unknown(t *testing.T) {
	err := &APIError{StatusCode: 500}
	if err.Error() != "jira: unknown API error" {
		t.Errorf("Error() = %q, want 'jira: unknown API error'", err.Error())
	}
}

func TestClient_4xxNoRetry(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errorMessages": []string{"Forbidden"},
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.Get("/forbidden", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Errorf("4xx should not retry, got %d attempts", attempts)
	}
}

func TestCreateIssue_AllIssueTypes(t *testing.T) {
	for _, issueType := range []string{"Epic", "Story", "Task", "Bug", "Subtask"} {
		t.Run(issueType, func(t *testing.T) {
			var gotBody map[string]interface{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data, _ := io.ReadAll(r.Body)
				json.Unmarshal(data, &gotBody)
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(CreateIssueResponse{Key: "PROJ-99"})
			}))
			defer srv.Close()

			c := newTestClient(t, srv.URL)
			req := &CreateIssueRequest{
				Fields: CreateIssueFields{
					Project:   ProjectRef{Key: "PROJ"},
					IssueType: IssueTypeRef{Name: issueType},
					Summary:   "Test " + issueType,
				},
			}
			if issueType == "Subtask" {
				req.Fields.Parent = &IssueRef{Key: "PROJ-1"}
			}

			resp, err := c.CreateIssue(req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Key != "PROJ-99" {
				t.Errorf("Key = %q, want PROJ-99", resp.Key)
			}

			fields := gotBody["fields"].(map[string]interface{})
			it := fields["issuetype"].(map[string]interface{})
			if it["name"] != issueType {
				t.Errorf("issuetype.name = %v, want %s", it["name"], issueType)
			}
		})
	}
}

func TestListIssues_WithFilters(t *testing.T) {
	var gotJQL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req SearchRequest
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &req)
		gotJQL = req.JQL

		json.NewEncoder(w).Encode(SearchResponse{
			Issues: []Issue{{Key: "PROJ-1"}},
			IsLast: true,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	_, err := c.ListIssues(ListIssuesOptions{
		ProjectKey: "PROJ",
		IssueType:  "Story",
		Status:     "Open",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(gotJQL, "project = PROJ") {
		t.Errorf("JQL missing project clause: %q", gotJQL)
	}
	if !strings.Contains(gotJQL, `issuetype = "Story"`) {
		t.Errorf("JQL missing issuetype clause: %q", gotJQL)
	}
	if !strings.Contains(gotJQL, `status = "Open"`) {
		t.Errorf("JQL missing status clause: %q", gotJQL)
	}
}

func TestSearchJQL_DefaultMaxResults(t *testing.T) {
	var gotMax float64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &req)
		gotMax = req["maxResults"].(float64)

		json.NewEncoder(w).Encode(SearchResponse{IsLast: true})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	c.SearchJQL(&SearchRequest{JQL: "project = X"})

	if gotMax != 100 {
		t.Errorf("default maxResults = %v, want 100", gotMax)
	}
}

func TestNewADFWithHeading_InvalidLevel(t *testing.T) {
	doc := NewADFWithHeading(0, "Title", "")
	var attrs map[string]int
	json.Unmarshal(doc.Content[0].Attrs, &attrs)
	if attrs["level"] != 1 {
		t.Errorf("invalid level should default to 1, got %d", attrs["level"])
	}
}

func TestNewADFWithHeading_NoBody(t *testing.T) {
	doc := NewADFWithHeading(1, "Title Only", "")
	if len(doc.Content) != 1 {
		t.Errorf("expected 1 node (heading only), got %d", len(doc.Content))
	}
}

func TestNewADFCodeBlock_NoLanguage(t *testing.T) {
	doc := NewADFCodeBlock("plain code", "")
	if doc.Content[0].Attrs != nil {
		t.Error("expected nil attrs for empty language")
	}
}

func TestListBoards_NoProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("projectKeyOrId") != "" {
			t.Error("projectKeyOrId should not be set")
		}
		json.NewEncoder(w).Encode(BoardSearchResult{
			Values: []Board{{ID: 1}},
			IsLast: true,
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	boards, err := c.ListBoards("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(boards) != 1 {
		t.Errorf("got %d boards, want 1", len(boards))
	}
}
