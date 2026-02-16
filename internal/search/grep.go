// Scoped grep: full-text regex search across Jira query results.
// Adapted from agent-facing-api skill reference implementation.

package search

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/relux-works/skill-jira-management/internal/jira"
)

// Match represents a single grep hit within an issue.
type Match struct {
	IssueKey string `json:"issue_key"`
	Field    string `json:"field"`    // which field matched (summary, description, comment, etc.)
	Content  string `json:"content"`  // the matched text
	Line     int    `json:"line"`     // line number within the field (1-indexed)
}

// GrepOptions controls grep behavior.
type GrepOptions struct {
	Scope           string // "issues" (default), "comments", "all"
	CaseInsensitive bool
	ContextLines    int
}

// GrepIssues searches across a slice of issues for lines matching pattern.
func GrepIssues(issues []jira.Issue, pattern string, opts GrepOptions) ([]Match, error) {
	if opts.CaseInsensitive {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}

	scope := opts.Scope
	if scope == "" {
		scope = "all"
	}

	var results []Match

	for i := range issues {
		issue := &issues[i]

		if scope == "issues" || scope == "all" {
			// Search summary
			if re.MatchString(issue.Fields.Summary) {
				results = append(results, Match{
					IssueKey: issue.Key,
					Field:    "summary",
					Content:  issue.Fields.Summary,
					Line:     1,
				})
			}

			// Search description (handles both ADF and plain string)
			descText := issue.Fields.DescriptionText()
			if descText != "" {
				lines := strings.Split(descText, "\n")
				for lineNum, line := range lines {
					if re.MatchString(line) {
						results = append(results, Match{
							IssueKey: issue.Key,
							Field:    "description",
							Content:  line,
							Line:     lineNum + 1,
						})
					}
				}
			}

			// Search labels
			for _, label := range issue.Fields.Labels {
				if re.MatchString(label) {
					results = append(results, Match{
						IssueKey: issue.Key,
						Field:    "labels",
						Content:  label,
						Line:     1,
					})
				}
			}
		}
	}

	if results == nil {
		results = []Match{}
	}
	return results, nil
}

// GrepComments searches across issue comments for lines matching pattern.
func GrepComments(comments []jira.Comment, issueKey, pattern string, opts GrepOptions) ([]Match, error) {
	if opts.CaseInsensitive {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex: %w", err)
	}

	var results []Match

	for _, comment := range comments {
		if comment.Body == nil {
			continue
		}
		text := extractADFText(comment.Body)
		lines := strings.Split(text, "\n")
		for lineNum, line := range lines {
			if re.MatchString(line) {
				results = append(results, Match{
					IssueKey: issueKey,
					Field:    fmt.Sprintf("comment/%s", comment.ID),
					Content:  line,
					Line:     lineNum + 1,
				})
			}
		}
	}

	if results == nil {
		results = []Match{}
	}
	return results, nil
}

// extractADFText recursively extracts plain text from an ADF document.
func extractADFText(doc *jira.ADFDoc) string {
	if doc == nil {
		return ""
	}
	var sb strings.Builder
	for _, node := range doc.Content {
		extractNodeText(&node, &sb)
	}
	return strings.TrimSpace(sb.String())
}

func extractNodeText(node *jira.ADFNode, sb *strings.Builder) {
	if node.Text != "" {
		sb.WriteString(node.Text)
	}
	for i := range node.Content {
		extractNodeText(&node.Content[i], sb)
	}
	// Add newlines after block-level nodes
	switch node.Type {
	case "paragraph", "heading", "codeBlock", "blockquote", "listItem":
		sb.WriteString("\n")
	}
}

// PrintJSON outputs matches as JSON array.
func PrintJSON(matches []Match) ([]byte, error) {
	return json.MarshalIndent(matches, "", "  ")
}

// PrintText outputs matches in grep-style format: issue_key:field:line:content
func PrintText(matches []Match) string {
	var sb strings.Builder
	for _, m := range matches {
		fmt.Fprintf(&sb, "%s:%s:%d:%s\n", m.IssueKey, m.Field, m.Line, m.Content)
	}
	return sb.String()
}
