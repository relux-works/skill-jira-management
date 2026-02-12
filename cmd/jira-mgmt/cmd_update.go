package main

import (
	"fmt"

	"github.com/ivalx1s/skill-jira-management/internal/jira"
	"github.com/spf13/cobra"
)

var (
	updateSummary     string
	updateDescription string
)

var updateCmd = &cobra.Command{
	Use:   "update ISSUE-KEY",
	Short: "Update issue fields (summary, description)",
	Long: `Update fields on an existing issue.

Examples:
  jira-mgmt update PROJ-123 --summary "New title"
  jira-mgmt update PROJ-123 --description "New description"
  jira-mgmt update PROJ-123 --summary "New title" --description "New description"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueKey := args[0]
		out := cmd.OutOrStdout()

		if updateSummary == "" && updateDescription == "" {
			return fmt.Errorf("at least one of --summary or --description is required")
		}

		client, err := buildJiraClientFromConfig()
		if err != nil {
			return err
		}

		fields := make(map[string]interface{})
		if updateSummary != "" {
			fields["summary"] = updateSummary
		}
		if updateDescription != "" {
			// Server/DC v2 accepts plain string, Cloud v3 needs ADF.
			if client.IsCloud() {
				fields["description"] = jira.NewADFText(updateDescription)
			} else {
				fields["description"] = updateDescription
			}
		}

		req := &jira.UpdateIssueRequest{Fields: fields}
		if err := client.UpdateIssue(issueKey, req); err != nil {
			return fmt.Errorf("updating %s: %w", issueKey, err)
		}

		fmt.Fprintf(out, "Updated %s\n", issueKey)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateSummary, "summary", "", "New issue summary/title")
	updateCmd.Flags().StringVar(&updateDescription, "description", "", "New issue description")
	rootCmd.AddCommand(updateCmd)
}
