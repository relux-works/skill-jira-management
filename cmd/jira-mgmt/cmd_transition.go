package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var transitionTo string

var transitionCmd = &cobra.Command{
	Use:   "transition <ISSUE-KEY>",
	Short: "Transition a Jira issue to a new status",
	Long: `Move a Jira issue to a new workflow status.

Examples:
  jira-mgmt transition PROJ-123 --to "In Progress"
  jira-mgmt transition PROJ-456 --to "Done"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		issueKey := args[0]

		if transitionTo == "" {
			return fmt.Errorf("--to is required: specify target status name")
		}

		client, err := buildJiraClientFromConfig()
		if err != nil {
			return err
		}

		transitions, err := client.GetTransitions(issueKey)
		if err != nil {
			return fmt.Errorf("getting transitions: %w", err)
		}

		var transitionID string
		var matchedName string
		for _, t := range transitions {
			if strings.EqualFold(t.Name, transitionTo) || strings.EqualFold(t.To.Name, transitionTo) {
				transitionID = t.ID
				matchedName = t.To.Name
				break
			}
		}

		if transitionID == "" {
			var available []string
			for _, t := range transitions {
				available = append(available, fmt.Sprintf("%s -> %s", t.Name, t.To.Name))
			}
			return fmt.Errorf("no transition to %q found for %s\navailable transitions:\n  %s",
				transitionTo, issueKey, strings.Join(available, "\n  "))
		}

		if err := client.DoTransition(issueKey, transitionID, nil); err != nil {
			return fmt.Errorf("executing transition: %w", err)
		}

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "%s -> %s\n", issueKey, matchedName)
		return nil
	},
}

func init() {
	transitionCmd.Flags().StringVar(&transitionTo, "to", "", "Target status name (required)")

	rootCmd.AddCommand(transitionCmd)
}
