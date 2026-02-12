package main

import (
	"encoding/json"
	"fmt"

	"github.com/ivalx1s/skill-jira-management/internal/query"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "q '<query>'",
	Short: "Execute a DSL query against Jira",
	Long: `Execute one or more DSL queries against Jira Cloud.

Operations:
  get(ISSUE-KEY) { fields }             — single issue lookup
  list(project=X, type=epic) { fields } — filtered listing
  summary()                             — project/board overview
  search(jql="...") { fields }          — JQL search

Field presets: minimal, default, overview, full
Batch: separate queries with semicolons.

Examples:
  jira-mgmt q 'get(PROJ-123) { overview }'
  jira-mgmt q 'list(project=PROJ, type=epic) { minimal }'
  jira-mgmt q 'summary()'
  jira-mgmt q 'search(jql="assignee = currentUser()") { default }'
  jira-mgmt q 'get(PROJ-1) { minimal }; get(PROJ-2) { minimal }'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		queryStr := args[0]

		parsed, err := query.ParseQuery(queryStr)
		if err != nil {
			return fmt.Errorf("parse error: %w", err)
		}

		client, err := buildJiraClientFromConfig()
		if err != nil {
			return err
		}

		executor := query.NewExecutor(client, flagProject, flagBoard)
		results, err := executor.Execute(parsed)
		if err != nil {
			return err
		}

		return outputQueryResults(cmd, results)
	},
}

func outputQueryResults(cmd *cobra.Command, results []json.RawMessage) error {
	out := cmd.OutOrStdout()

	if len(results) == 1 {
		var pretty json.RawMessage
		if err := json.Unmarshal(results[0], &pretty); err != nil {
			return err
		}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(pretty)
	}

	combined := make([]json.RawMessage, len(results))
	copy(combined, results)
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(combined)
}

func init() {
	rootCmd.AddCommand(queryCmd)
}
