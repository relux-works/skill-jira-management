package main

import (
	"fmt"

	"github.com/ivalx1s/skill-jira-management/internal/jira"
	"github.com/spf13/cobra"
)

var dodCriteria string

var dodCmd = &cobra.Command{
	Use:   "dod <ISSUE-KEY>",
	Short: "Set Definition of Done on a Jira issue",
	Long: `Add or update the Definition of Done criteria as a comment on a Jira issue.
The DoD is posted as a structured comment with a heading.

Examples:
  jira-mgmt dod PROJ-123 --set "All tests pass; Code reviewed; Deployed to staging"
  jira-mgmt dod PROJ-456 --set "Unit test coverage > 80%; API docs updated"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueKey := args[0]

		if dodCriteria == "" {
			return fmt.Errorf("--set is required: specify DoD criteria")
		}

		client, err := buildJiraClientFromConfig()
		if err != nil {
			return err
		}

		locale := getConfigLocale()
		heading := getLocaleString(locale, "dod_heading")
		adfBody := jira.NewADFWithHeading(3, heading, dodCriteria)

		comment, err := client.AddComment(issueKey, adfBody)
		if err != nil {
			return fmt.Errorf("setting DoD: %w", err)
		}

		out := cmd.OutOrStdout()
		msgFmt := getLocaleString(locale, "dod_set")
		fmt.Fprintf(out, msgFmt+"\n", issueKey, comment.ID)
		return nil
	},
}

func init() {
	dodCmd.Flags().StringVar(&dodCriteria, "set", "", "DoD criteria text (required)")

	rootCmd.AddCommand(dodCmd)
}
