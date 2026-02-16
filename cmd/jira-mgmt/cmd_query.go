package main

import (
	"fmt"
	"strings"

	"github.com/relux-works/skill-agent-facing-api/agentquery"
	"github.com/relux-works/skill-jira-management/internal/jira"
	"github.com/relux-works/skill-jira-management/internal/query"
	"github.com/spf13/cobra"
)

// buildQueryCommand creates the "q" subcommand that parses and executes
// DSL queries via the agentquery schema.
func buildQueryCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "q '<query>'",
		Short: "Execute a DSL query against Jira",
		Long: `Execute one or more DSL queries against Jira Cloud.

Operations:
  get(ISSUE-KEY) { fields }             — single issue lookup
  list(project=X, type=epic) { fields } — filtered listing
  count(project=X, status=done)         — count matching issues
  summary()                             — project/board overview
  search(jql="...") { fields }          — JQL search
  schema()                              — introspect available operations, fields, presets

Field presets: minimal, default, overview, full
Batch: separate queries with semicolons.

Examples:
  jira-mgmt q 'get(PROJ-123) { overview }' --format json
  jira-mgmt q 'list(project=PROJ, type=epic) { minimal }' --format compact
  jira-mgmt q 'summary()' --format json
  jira-mgmt q 'search(jql="assignee = currentUser()") { default }' --format json
  jira-mgmt q 'get(PROJ-1) { minimal }; get(PROJ-2) { minimal }' --format json
  jira-mgmt q 'schema()' --format json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := buildJiraClientFromConfig()
			if err != nil {
				return err
			}

			schema := query.NewSchema(client, flagProject, flagBoard)
			return runQuery(cmd, schema, args[0], format)
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", `Output format: "json", "compact", or "llm"`)

	return cmd
}

// parseOutputMode converts a format flag value to an agentquery.OutputMode.
func parseOutputMode(s string) (agentquery.OutputMode, error) {
	switch strings.ToLower(s) {
	case "compact", "llm":
		return agentquery.LLMReadable, nil
	case "json":
		return agentquery.HumanReadable, nil
	default:
		return 0, fmt.Errorf("unknown format %q: use \"json\", \"compact\", or \"llm\"", s)
	}
}

// runQuery executes the query against the schema and writes output to the command's stdout.
func runQuery(cmd *cobra.Command, schema *agentquery.Schema[jira.Issue], queryStr string, format string) error {
	mode, err := parseOutputMode(format)
	if err != nil {
		return err
	}
	data, err := schema.QueryJSONWithMode(queryStr, mode)
	if err != nil {
		return err
	}
	_, err = cmd.OutOrStdout().Write(data)
	if err != nil {
		return err
	}
	_, err = cmd.OutOrStdout().Write([]byte("\n"))
	return err
}

func init() {
	rootCmd.AddCommand(buildQueryCommand())
}
