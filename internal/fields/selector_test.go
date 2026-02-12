package fields

import (
	"testing"

	"github.com/ivalx1s/skill-jira-management/internal/jira"
)

func TestNewSelector_Default(t *testing.T) {
	s, err := NewSelector(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Default should include key, summary, status, assignee
	if !s.Include("key") {
		t.Error("default should include 'key'")
	}
	if !s.Include("summary") {
		t.Error("default should include 'summary'")
	}
	if !s.Include("status") {
		t.Error("default should include 'status'")
	}
	if !s.Include("assignee") {
		t.Error("default should include 'assignee'")
	}
	if s.Include("description") {
		t.Error("default should not include 'description'")
	}
}

func TestNewSelector_Preset(t *testing.T) {
	s, err := NewSelector([]string{"minimal"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.Include("key") {
		t.Error("minimal should include 'key'")
	}
	if !s.Include("status") {
		t.Error("minimal should include 'status'")
	}
	if s.Include("summary") {
		t.Error("minimal should not include 'summary'")
	}
}

func TestNewSelector_Full(t *testing.T) {
	s, err := NewSelector([]string{"full"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// full should include everything
	if !s.Include("key") {
		t.Error("full should include 'key'")
	}
	if !s.Include("description") {
		t.Error("full should include 'description'")
	}
	if !s.Include("created") {
		t.Error("full should include 'created'")
	}
}

func TestNewSelector_IndividualFields(t *testing.T) {
	s, err := NewSelector([]string{"key", "priority"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.Include("key") {
		t.Error("should include 'key'")
	}
	if !s.Include("priority") {
		t.Error("should include 'priority'")
	}
	if s.Include("summary") {
		t.Error("should not include 'summary'")
	}
}

func TestNewSelector_UnknownField(t *testing.T) {
	_, err := NewSelector([]string{"foobar"})
	if err == nil {
		t.Error("expected error for unknown field")
	}
}

func TestSelector_Apply(t *testing.T) {
	s, err := NewSelector([]string{"key", "summary", "status", "assignee"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	issue := &jira.Issue{
		Key: "PROJ-123",
		Fields: jira.IssueFields{
			Summary: "Test issue",
			Status:  &jira.Status{Name: "In Progress"},
			Assignee: &jira.User{DisplayName: "John"},
			Priority: &jira.Priority{Name: "High"},
		},
	}

	result := s.Apply(issue)

	if result["key"] != "PROJ-123" {
		t.Errorf("expected key=PROJ-123, got %v", result["key"])
	}
	if result["summary"] != "Test issue" {
		t.Errorf("expected summary='Test issue', got %v", result["summary"])
	}
	if result["status"] != "In Progress" {
		t.Errorf("expected status='In Progress', got %v", result["status"])
	}
	if result["assignee"] != "John" {
		t.Errorf("expected assignee='John', got %v", result["assignee"])
	}
	if _, ok := result["priority"]; ok {
		t.Error("should not include priority")
	}
}

func TestSelector_ApplyNilFields(t *testing.T) {
	s, err := NewSelector([]string{"key", "status", "assignee", "priority", "parent", "reporter"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	issue := &jira.Issue{
		Key: "PROJ-1",
		Fields: jira.IssueFields{
			Summary: "Test",
		},
	}

	result := s.Apply(issue)

	if result["status"] != nil {
		t.Errorf("expected nil status, got %v", result["status"])
	}
	if result["assignee"] != nil {
		t.Errorf("expected nil assignee, got %v", result["assignee"])
	}
}

func TestSelector_ApplyMany(t *testing.T) {
	s, err := NewSelector([]string{"minimal"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	issues := []jira.Issue{
		{Key: "A-1", Fields: jira.IssueFields{Status: &jira.Status{Name: "Open"}}},
		{Key: "A-2", Fields: jira.IssueFields{Status: &jira.Status{Name: "Done"}}},
	}

	results := s.ApplyMany(issues)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0]["key"] != "A-1" {
		t.Errorf("expected key=A-1, got %v", results[0]["key"])
	}
	if results[1]["status"] != "Done" {
		t.Errorf("expected status=Done, got %v", results[1]["status"])
	}
}

func TestSelector_JiraAPIFields(t *testing.T) {
	s, err := NewSelector([]string{"key", "summary", "type"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	apiFields := s.JiraAPIFields()
	// key doesn't map to an API field (always returned)
	// summary -> summary, type -> issuetype
	found := make(map[string]bool)
	for _, f := range apiFields {
		found[f] = true
	}
	if !found["summary"] {
		t.Error("expected 'summary' in API fields")
	}
	if !found["issuetype"] {
		t.Error("expected 'issuetype' in API fields")
	}
}
