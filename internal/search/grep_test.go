package search

import (
	"testing"

	"github.com/relux-works/skill-jira-management/internal/jira"
)

func TestGrepIssues_Summary(t *testing.T) {
	issues := []jira.Issue{
		{Key: "A-1", Fields: jira.IssueFields{Summary: "Fix authentication bug"}},
		{Key: "A-2", Fields: jira.IssueFields{Summary: "Add new feature"}},
		{Key: "A-3", Fields: jira.IssueFields{Summary: "Update auth module"}},
	}

	matches, err := GrepIssues(issues, "auth", GrepOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].IssueKey != "A-1" {
		t.Errorf("expected first match A-1, got %s", matches[0].IssueKey)
	}
	if matches[0].Field != "summary" {
		t.Errorf("expected field 'summary', got %s", matches[0].Field)
	}
}

func TestGrepIssues_CaseInsensitive(t *testing.T) {
	issues := []jira.Issue{
		{Key: "A-1", Fields: jira.IssueFields{Summary: "Fix Authentication"}},
		{Key: "A-2", Fields: jira.IssueFields{Summary: "no match here"}},
	}

	matches, err := GrepIssues(issues, "authentication", GrepOptions{CaseInsensitive: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}

func TestGrepIssues_Description(t *testing.T) {
	issues := []jira.Issue{
		{
			Key: "A-1",
			Fields: jira.IssueFields{
				Summary: "No match",
				Description: &jira.ADFDoc{
					Type:    "doc",
					Version: 1,
					Content: []jira.ADFNode{
						{
							Type: "paragraph",
							Content: []jira.ADFNode{
								{Type: "text", Text: "This has a secret pattern"},
							},
						},
						{
							Type: "paragraph",
							Content: []jira.ADFNode{
								{Type: "text", Text: "Another line"},
							},
						},
					},
				},
			},
		},
	}

	matches, err := GrepIssues(issues, "secret", GrepOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Field != "description" {
		t.Errorf("expected field 'description', got %s", matches[0].Field)
	}
}

func TestGrepIssues_Labels(t *testing.T) {
	issues := []jira.Issue{
		{Key: "A-1", Fields: jira.IssueFields{
			Summary: "No match",
			Labels:  []string{"frontend", "backend", "urgent"},
		}},
	}

	matches, err := GrepIssues(issues, "front", GrepOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Field != "labels" {
		t.Errorf("expected field 'labels', got %s", matches[0].Field)
	}
}

func TestGrepIssues_NoMatches(t *testing.T) {
	issues := []jira.Issue{
		{Key: "A-1", Fields: jira.IssueFields{Summary: "Nothing here"}},
	}

	matches, err := GrepIssues(issues, "zzzzz", GrepOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(matches))
	}
}

func TestGrepIssues_InvalidRegex(t *testing.T) {
	_, err := GrepIssues(nil, "[invalid", GrepOptions{})
	if err == nil {
		t.Error("expected error for invalid regex")
	}
}

func TestGrepComments(t *testing.T) {
	comments := []jira.Comment{
		{
			ID: "1001",
			Body: &jira.ADFDoc{
				Type:    "doc",
				Version: 1,
				Content: []jira.ADFNode{
					{
						Type: "paragraph",
						Content: []jira.ADFNode{
							{Type: "text", Text: "Deployed to staging"},
						},
					},
				},
			},
		},
		{
			ID: "1002",
			Body: &jira.ADFDoc{
				Type:    "doc",
				Version: 1,
				Content: []jira.ADFNode{
					{
						Type: "paragraph",
						Content: []jira.ADFNode{
							{Type: "text", Text: "No match"},
						},
					},
				},
			},
		},
	}

	matches, err := GrepComments(comments, "PROJ-1", "staging", GrepOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Field != "comment/1001" {
		t.Errorf("expected field 'comment/1001', got %s", matches[0].Field)
	}
}

func TestPrintText(t *testing.T) {
	matches := []Match{
		{IssueKey: "A-1", Field: "summary", Line: 1, Content: "Fix auth"},
		{IssueKey: "A-2", Field: "description", Line: 3, Content: "auth module"},
	}

	text := PrintText(matches)
	expected := "A-1:summary:1:Fix auth\nA-2:description:3:auth module\n"
	if text != expected {
		t.Errorf("unexpected text output:\n  got:    %q\n  expect: %q", text, expected)
	}
}

func TestPrintJSON(t *testing.T) {
	matches := []Match{
		{IssueKey: "A-1", Field: "summary", Line: 1, Content: "test"},
	}

	data, err := PrintJSON(matches)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty JSON output")
	}
}

func TestExtractADFText(t *testing.T) {
	doc := &jira.ADFDoc{
		Type:    "doc",
		Version: 1,
		Content: []jira.ADFNode{
			{
				Type: "paragraph",
				Content: []jira.ADFNode{
					{Type: "text", Text: "Hello "},
					{Type: "text", Text: "World"},
				},
			},
			{
				Type: "heading",
				Content: []jira.ADFNode{
					{Type: "text", Text: "Title"},
				},
			},
		},
	}

	text := extractADFText(doc)
	if text != "Hello World\nTitle" {
		t.Errorf("unexpected ADF text: %q", text)
	}
}

func TestExtractADFText_Nil(t *testing.T) {
	text := extractADFText(nil)
	if text != "" {
		t.Errorf("expected empty string for nil doc, got %q", text)
	}
}
