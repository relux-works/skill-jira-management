package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/relux-works/skill-jira-management/internal/jira"
)

func screen() *jira.Transition {
	return &jira.Transition{
		ID:   "31",
		Name: "В работу",
		Fields: map[string]jira.TransitionField{
			"customfield_10500": {
				Required: true,
				Name:     "Вид деятельности",
				AllowedValues: []jira.TransitionOption{
					{ID: "10100", Value: "Разработка"},
					{ID: "10101", Value: "Аналитика"},
				},
			},
			"summary": {Name: "Summary"},
			"assignee": {
				Name:          "Assignee",
				AllowedValues: []jira.TransitionOption{{ID: "u1", Name: "aagrigore1"}},
			},
		},
	}
}

// Proves an option-typed field is sent in the shape Jira demands, addressed by
// its internal id, while the caller types the display name and label.
func TestBuildTransitionFields_OptionByDisplayName(t *testing.T) {
	got, err := buildTransitionFields(screen(), []string{"Вид деятельности=Разработка"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := got["customfield_10500"].(map[string]interface{})
	if !ok {
		t.Fatalf("field not set by id: %#v", got)
	}
	if v["id"] != "10100" || v["value"] != "Разработка" {
		t.Errorf("payload = %#v, want id 10100 and value Разработка", v)
	}
	if len(got) != 1 {
		t.Errorf("sent %d fields, want exactly the one requested: %#v", len(got), got)
	}
}

// The field id must work too: Jira's own refusals name the id, not the label.
func TestBuildTransitionFields_OptionByFieldID(t *testing.T) {
	got, err := buildTransitionFields(screen(), []string{"customfield_10500=Аналитика"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["customfield_10500"].(map[string]interface{})["id"] != "10101" {
		t.Errorf("payload = %#v", got)
	}
}

// An option carrying name rather than value keeps that shape; swapping them
// makes Jira reject the transition with an opaque message.
func TestBuildTransitionFields_NameShapedOption(t *testing.T) {
	got, err := buildTransitionFields(screen(), []string{"Assignee=aagrigore1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v := got["assignee"].(map[string]interface{})
	if v["name"] != "aagrigore1" {
		t.Errorf("payload = %#v, want a name-shaped option", v)
	}
	if _, present := v["value"]; present {
		t.Errorf("payload = %#v, must not invent a value key", v)
	}
}

// A field without an allowed set is a free-text field and must pass through
// unwrapped.
func TestBuildTransitionFields_FreeTextPassesThrough(t *testing.T) {
	got, err := buildTransitionFields(screen(), []string{"Summary=новое имя"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["summary"] != "новое имя" {
		t.Errorf("payload = %#v, want a plain string", got["summary"])
	}
}

// Negative cases. Each must be refused locally, with a message that names what
// is allowed: a value Jira rejects comes back as an opaque 400 otherwise.
func TestBuildTransitionFields_Refusals(t *testing.T) {
	cases := []struct {
		name    string
		spec    string
		wantErr string
	}{
		{"value outside the allowed set", "Вид деятельности=Бухгалтерия", "allowed: Разработка, Аналитика"},
		{"unknown field", "Нет такого=1", "has no field"},
		{"missing equals sign", "Вид деятельности", "is not name=value"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildTransitionFields(screen(), []string{tc.spec})
			if err == nil {
				t.Fatalf("expected a refusal, got %#v", got)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %v, want it to contain %q", err, tc.wantErr)
			}
		})
	}
}

// No --field must send no fields at all, not an empty map: an empty fields
// object changes what some workflows accept.
func TestBuildTransitionFields_NoneMeansNil(t *testing.T) {
	got, err := buildTransitionFields(screen(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("fields = %#v, want nil", got)
	}
}

// --field against a transition with no screen must say so rather than send a
// payload the workflow will ignore.
func TestBuildTransitionFields_NoScreenIsRefused(t *testing.T) {
	_, err := buildTransitionFields(&jira.Transition{ID: "1"}, []string{"Summary=x"})
	if err == nil {
		t.Fatal("expected a refusal for a transition with no screen fields")
	}
}

// --list-fields must put required fields first and print their allowed values:
// it exists to answer "what does this transition want from me".
func TestPrintTransitionFields(t *testing.T) {
	var buf bytes.Buffer
	printTransitionFields(&buf, screen())
	out := buf.String()

	req := strings.Index(out, "REQUIRED Вид деятельности")
	if req < 0 {
		t.Fatalf("required field not marked:\n%s", out)
	}
	if opt := strings.Index(out, "optional"); opt >= 0 && opt < req {
		t.Errorf("an optional field is printed before the required one:\n%s", out)
	}
	if !strings.Contains(out, "allowed: Разработка, Аналитика") {
		t.Errorf("allowed values not printed:\n%s", out)
	}
	if !strings.Contains(out, "customfield_10500") {
		t.Errorf("field id not printed, so a Jira refusal cannot be matched to it:\n%s", out)
	}
}
