package main

import (
	"testing"

	"github.com/relux-works/skill-jira-management/internal/jira"
)

func TestBuildCancelFields_UsesCancelResolutionAndOtherReasonFallback(t *testing.T) {
	transition := &jira.Transition{
		ID:   "191",
		Name: "Cancel",
		Fields: map[string]jira.TransitionField{
			"resolution": {
				Required: true,
				AllowedValues: []jira.TransitionOption{
					{ID: "10200", Name: "Отменено"},
				},
			},
			"customfield_15501": {
				AllowedValues: []jira.TransitionOption{
					{ID: "20724", Value: "Другое (прокомментируйте что именно)"},
				},
			},
		},
	}

	fields, err := buildCancelFields(transition, "прекращение работы с ICONIA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resolution := fields["resolution"].(map[string]string)
	if resolution["id"] != "10200" {
		t.Fatalf("resolution id = %q, want 10200", resolution["id"])
	}

	reason := fields["customfield_15501"].(map[string]string)
	if reason["id"] != "20724" {
		t.Fatalf("customfield_15501 id = %q, want 20724", reason["id"])
	}
}

func TestBuildCancelFields_UsesExactReasonOptionMatch(t *testing.T) {
	transition := &jira.Transition{
		ID:   "191",
		Name: "Cancel",
		Fields: map[string]jira.TransitionField{
			"customfield_15501": {
				AllowedValues: []jira.TransitionOption{
					{ID: "20720", Value: "Внешние факторы у исполнителя"},
					{ID: "20724", Value: "Другое (прокомментируйте что именно)"},
				},
			},
		},
	}

	fields, err := buildCancelFields(transition, "Внешние факторы у исполнителя")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reason := fields["customfield_15501"].(map[string]string)
	if reason["id"] != "20720" {
		t.Fatalf("customfield_15501 id = %q, want 20720", reason["id"])
	}
}

func TestBuildCancelFields_RequiresCancelLikeResolutionWhenResolutionFieldExists(t *testing.T) {
	transition := &jira.Transition{
		ID:   "191",
		Name: "Cancel",
		Fields: map[string]jira.TransitionField{
			"resolution": {
				Required: true,
				AllowedValues: []jira.TransitionOption{
					{ID: "10002", Name: "Duplicate"},
				},
			},
		},
	}

	if _, err := buildCancelFields(transition, "irrelevant"); err == nil {
		t.Fatal("expected error, got nil")
	}
}
