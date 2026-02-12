package query

import (
	"testing"
)

func TestParseQuery_Get(t *testing.T) {
	q, err := ParseQuery(`get(PROJ-123) { minimal }`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(q.Statements))
	}
	s := q.Statements[0]
	if s.Operation != "get" {
		t.Errorf("expected operation 'get', got %q", s.Operation)
	}
	if len(s.Args) != 1 || s.Args[0].Value != "PROJ-123" {
		t.Errorf("expected arg PROJ-123, got %v", s.Args)
	}
	// minimal preset should expand to key, status
	if len(s.Fields) != 2 {
		t.Errorf("expected 2 fields from minimal preset, got %d: %v", len(s.Fields), s.Fields)
	}
}

func TestParseQuery_GetNoFields(t *testing.T) {
	q, err := ParseQuery(`get(PROJ-456)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := q.Statements[0]
	if s.Fields != nil {
		t.Errorf("expected nil fields, got %v", s.Fields)
	}
}

func TestParseQuery_List(t *testing.T) {
	q, err := ParseQuery(`list(project=MYPROJ, type=epic) { overview }`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := q.Statements[0]
	if s.Operation != "list" {
		t.Errorf("expected operation 'list', got %q", s.Operation)
	}
	if len(s.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(s.Args))
	}
	if s.Args[0].Key != "project" || s.Args[0].Value != "MYPROJ" {
		t.Errorf("expected project=MYPROJ, got %s=%s", s.Args[0].Key, s.Args[0].Value)
	}
	if s.Args[1].Key != "type" || s.Args[1].Value != "epic" {
		t.Errorf("expected type=epic, got %s=%s", s.Args[1].Key, s.Args[1].Value)
	}
	// overview preset: key, summary, status, assignee, type, priority, parent
	if len(s.Fields) != 7 {
		t.Errorf("expected 7 fields from overview preset, got %d: %v", len(s.Fields), s.Fields)
	}
}

func TestParseQuery_Summary(t *testing.T) {
	q, err := ParseQuery(`summary()`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := q.Statements[0]
	if s.Operation != "summary" {
		t.Errorf("expected operation 'summary', got %q", s.Operation)
	}
	if len(s.Args) != 0 {
		t.Errorf("expected 0 args, got %d", len(s.Args))
	}
}

func TestParseQuery_Search(t *testing.T) {
	q, err := ParseQuery(`search(jql="assignee = currentUser()") { default }`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := q.Statements[0]
	if s.Operation != "search" {
		t.Errorf("expected operation 'search', got %q", s.Operation)
	}
	if len(s.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(s.Args))
	}
	if s.Args[0].Key != "jql" {
		t.Errorf("expected key 'jql', got %q", s.Args[0].Key)
	}
	if s.Args[0].Value != "assignee = currentUser()" {
		t.Errorf("expected JQL value, got %q", s.Args[0].Value)
	}
}

func TestParseQuery_Batch(t *testing.T) {
	q, err := ParseQuery(`get(PROJ-1) { minimal }; get(PROJ-2) { minimal }`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(q.Statements))
	}
	if q.Statements[0].Args[0].Value != "PROJ-1" {
		t.Errorf("expected PROJ-1, got %q", q.Statements[0].Args[0].Value)
	}
	if q.Statements[1].Args[0].Value != "PROJ-2" {
		t.Errorf("expected PROJ-2, got %q", q.Statements[1].Args[0].Value)
	}
}

func TestParseQuery_ThreeBatched(t *testing.T) {
	q, err := ParseQuery(`get(A-1); list(project=B); summary()`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(q.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(q.Statements))
	}
}

func TestParseQuery_FullPreset(t *testing.T) {
	q, err := ParseQuery(`get(X-1) { full }`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := q.Statements[0]
	// full preset: all fields
	if len(s.Fields) != 13 {
		t.Errorf("expected 13 fields from full preset, got %d: %v", len(s.Fields), s.Fields)
	}
}

func TestParseQuery_IndividualFields(t *testing.T) {
	q, err := ParseQuery(`get(X-1) { key summary status }`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := q.Statements[0]
	if len(s.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d: %v", len(s.Fields), s.Fields)
	}
}

func TestParseQuery_FieldDeduplication(t *testing.T) {
	q, err := ParseQuery(`get(X-1) { key minimal }`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := q.Statements[0]
	// key + minimal(key, status) should deduplicate key
	if len(s.Fields) != 2 {
		t.Errorf("expected 2 unique fields, got %d: %v", len(s.Fields), s.Fields)
	}
}

func TestParseQuery_ErrorEmpty(t *testing.T) {
	_, err := ParseQuery("")
	if err == nil {
		t.Error("expected error for empty query")
	}
}

func TestParseQuery_ErrorUnknownOp(t *testing.T) {
	_, err := ParseQuery(`delete(PROJ-1)`)
	if err == nil {
		t.Error("expected error for unknown operation")
	}
}

func TestParseQuery_ErrorUnknownField(t *testing.T) {
	_, err := ParseQuery(`get(X-1) { foobar }`)
	if err == nil {
		t.Error("expected error for unknown field")
	}
}

func TestParseQuery_ErrorUnterminatedString(t *testing.T) {
	_, err := ParseQuery(`search(jql="unterminated)`)
	if err == nil {
		t.Error("expected error for unterminated string")
	}
}

func TestParseQuery_ErrorMissingParen(t *testing.T) {
	_, err := ParseQuery(`get(PROJ-1`)
	if err == nil {
		t.Error("expected error for missing closing paren")
	}
}

func TestParseQuery_StatusFilter(t *testing.T) {
	q, err := ParseQuery(`list(project=PROJ, type=story, status=open) { default }`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := q.Statements[0]
	if len(s.Args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(s.Args))
	}
	if s.Args[2].Key != "status" || s.Args[2].Value != "open" {
		t.Errorf("expected status=open, got %s=%s", s.Args[2].Key, s.Args[2].Value)
	}
}

func TestParseQuery_PositionalArg(t *testing.T) {
	q, err := ParseQuery(`summary(MYPROJ)`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := q.Statements[0]
	if len(s.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(s.Args))
	}
	if s.Args[0].Key != "" {
		t.Errorf("expected positional arg (empty key), got key=%q", s.Args[0].Key)
	}
	if s.Args[0].Value != "MYPROJ" {
		t.Errorf("expected value MYPROJ, got %q", s.Args[0].Value)
	}
}
