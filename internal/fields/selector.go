// Field selection and projection for Jira issues.
// Adapted from agent-facing-api skill reference implementation.

package fields

import (
	"fmt"

	"github.com/ivalx1s/skill-jira-management/internal/jira"
)

// ValidFields — all recognized field names for Jira issues.
var ValidFields = map[string]bool{
	"key":         true,
	"summary":     true,
	"status":      true,
	"assignee":    true,
	"type":        true,
	"priority":    true,
	"parent":      true,
	"description": true,
	"labels":      true,
	"reporter":    true,
	"created":     true,
	"updated":     true,
	"project":     true,
	"subtasks":    true,
}

// Presets — named bundles of fields for common access patterns.
var Presets = map[string][]string{
	"minimal":  {"key", "status"},
	"default":  {"key", "summary", "status", "assignee"},
	"overview": {"key", "summary", "status", "assignee", "type", "priority", "parent"},
	"full":     {"key", "summary", "status", "assignee", "type", "priority", "parent", "description", "labels", "reporter", "created", "updated", "project", "subtasks"},
}

// Selector controls which fields appear in the response.
type Selector struct {
	fields map[string]bool
	all    bool
}

// NewSelector creates a Selector from requested field names.
// Empty input defaults to the "default" preset.
// Presets are expanded inline. Unknown fields return an error.
func NewSelector(requested []string) (*Selector, error) {
	if len(requested) == 0 {
		return &Selector{
			fields: map[string]bool{"key": true, "summary": true, "status": true, "assignee": true},
		}, nil
	}

	s := &Selector{fields: make(map[string]bool)}

	for _, f := range requested {
		if expanded, ok := Presets[f]; ok {
			if f == "full" {
				s.all = true
			}
			for _, ef := range expanded {
				s.fields[ef] = true
			}
			continue
		}
		if !ValidFields[f] {
			return nil, fmt.Errorf("unknown field: %s", f)
		}
		s.fields[f] = true
	}

	return s, nil
}

// Include returns true if the field should be in the response.
func (s *Selector) Include(field string) bool {
	if s.all {
		return true
	}
	return s.fields[field]
}

// Apply builds a response map for a Jira issue, including only selected fields.
func (s *Selector) Apply(issue *jira.Issue) map[string]interface{} {
	result := make(map[string]interface{})

	if s.Include("key") {
		result["key"] = issue.Key
	}
	if s.Include("summary") {
		result["summary"] = issue.Fields.Summary
	}
	if s.Include("status") {
		if issue.Fields.Status != nil {
			result["status"] = issue.Fields.Status.Name
		} else {
			result["status"] = nil
		}
	}
	if s.Include("assignee") {
		if issue.Fields.Assignee != nil {
			result["assignee"] = issue.Fields.Assignee.DisplayName
		} else {
			result["assignee"] = nil
		}
	}
	if s.Include("type") {
		result["type"] = issue.Fields.IssueType.Name
	}
	if s.Include("priority") {
		if issue.Fields.Priority != nil {
			result["priority"] = issue.Fields.Priority.Name
		} else {
			result["priority"] = nil
		}
	}
	if s.Include("parent") {
		if issue.Fields.Parent != nil {
			result["parent"] = issue.Fields.Parent.Key
		} else {
			result["parent"] = nil
		}
	}
	if s.Include("description") {
		result["description"] = issue.Fields.DescriptionText()
	}
	if s.Include("labels") {
		result["labels"] = issue.Fields.Labels
	}
	if s.Include("reporter") {
		if issue.Fields.Reporter != nil {
			result["reporter"] = issue.Fields.Reporter.DisplayName
		} else {
			result["reporter"] = nil
		}
	}
	if s.Include("created") {
		result["created"] = issue.Fields.Created
	}
	if s.Include("updated") {
		result["updated"] = issue.Fields.Updated
	}
	if s.Include("project") {
		result["project"] = issue.Fields.Project.Key
	}
	if s.Include("subtasks") && len(issue.Fields.Subtasks) > 0 {
		subs := make([]map[string]interface{}, len(issue.Fields.Subtasks))
		for i, st := range issue.Fields.Subtasks {
			sub := map[string]interface{}{
				"key":     st.Key,
				"summary": st.Fields.Summary,
			}
			if st.Fields.Status != nil {
				sub["status"] = st.Fields.Status.Name
			}
			subs[i] = sub
		}
		result["subtasks"] = subs
	}

	return result
}

// ApplyMany applies field selection to a slice of issues.
func (s *Selector) ApplyMany(issues []jira.Issue) []map[string]interface{} {
	results := make([]map[string]interface{}, len(issues))
	for i := range issues {
		results[i] = s.Apply(&issues[i])
	}
	return results
}

// JiraAPIFields returns the Jira REST API field names needed for the selected fields.
// This is used to optimize API calls by requesting only needed fields.
func (s *Selector) JiraAPIFields() []string {
	fieldMap := map[string]string{
		"key":         "", // always returned
		"summary":     "summary",
		"status":      "status",
		"assignee":    "assignee",
		"type":        "issuetype",
		"priority":    "priority",
		"parent":      "parent",
		"description": "description",
		"labels":      "labels",
		"reporter":    "reporter",
		"created":     "created",
		"updated":     "updated",
		"project":     "project",
		"subtasks":    "subtasks",
	}

	seen := make(map[string]bool)
	var apiFields []string
	for field := range s.fields {
		apiField := fieldMap[field]
		if apiField != "" && !seen[apiField] {
			seen[apiField] = true
			apiFields = append(apiFields, apiField)
		}
	}
	if s.all {
		// Return all mapped fields
		for _, v := range fieldMap {
			if v != "" && !seen[v] {
				seen[v] = true
				apiFields = append(apiFields, v)
			}
		}
	}
	return apiFields
}
