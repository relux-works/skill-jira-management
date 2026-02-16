package main

import (
	"fmt"

	"github.com/relux-works/skill-jira-management/internal/jira"
	"github.com/spf13/cobra"
)

var commentBody string

var commentCmd = &cobra.Command{
	Use:   "comment <ISSUE-KEY>",
	Short: "Add a comment to a Jira issue",
	Long: `Add a comment to an existing Jira issue.

Examples:
  jira-mgmt comment PROJ-123 --body "Fixed in commit abc123"
  jira-mgmt comment PROJ-456 --body "Blocked by dependency on auth service"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueKey := args[0]

		if commentBody == "" {
			return fmt.Errorf("--body is required")
		}

		client, err := buildJiraClientFromConfig()
		if err != nil {
			return err
		}

		adfBody := jira.NewADFText(commentBody)
		comment, err := client.AddComment(issueKey, adfBody)
		if err != nil {
			return fmt.Errorf("adding comment: %w", err)
		}

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Comment added to %s (id: %s)\n", issueKey, comment.ID)
		return nil
	},
}

func init() {
	commentCmd.Flags().StringVar(&commentBody, "body", "", "Comment text (required)")

	rootCmd.AddCommand(commentCmd)
}
