package main

import (
	"fmt"

	"github.com/relux-works/skill-jira-management/internal/jira"
	"github.com/spf13/cobra"
)

var (
	assignTo     string
	assignClear  bool
)

var assignCmd = &cobra.Command{
	Use:   "assign <ISSUE-KEY> [ISSUE-KEY...]",
	Short: "Assign Jira issues to a user",
	Long: `Set or clear the assignee on one or more Jira issues.

The addressing field differs between instances and they do not overlap: Cloud
uses accountId, Server/DC uses the username. The right one is chosen from the
detected instance type, so --to takes a username, an email or a display name
and is resolved before anything is written.

A query that matches more than one user is an error rather than a guess:
assigning somebody else's work to the wrong person is not a recoverable
mistake.

Examples:
  jira-mgmt assign PROJ-123 --to me
  jira-mgmt assign PROJ-123 PROJ-124 --to user@company.com
  jira-mgmt assign PROJ-123 --clear`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if assignClear && assignTo != "" {
			return fmt.Errorf("--to and --clear are mutually exclusive")
		}
		if !assignClear && assignTo == "" {
			return fmt.Errorf("--to is required (or --clear to unassign)")
		}

		client, err := buildJiraClientFromConfig()
		if err != nil {
			return err
		}

		var user *jira.User
		if !assignClear {
			if assignTo == "me" {
				user, err = client.Myself()
				if err != nil {
					return fmt.Errorf("resolving the authenticated user: %w", err)
				}
			} else {
				user, err = client.FindUser(assignTo)
				if err != nil {
					return err
				}
			}
		}

		out := cmd.OutOrStdout()
		var failures []string
		for _, issueKey := range args {
			if err := client.AssignIssue(issueKey, user); err != nil {
				fmt.Fprintf(out, "%s: %v\n", issueKey, err)
				failures = append(failures, issueKey)
				continue
			}
			if user == nil {
				fmt.Fprintf(out, "%s unassigned\n", issueKey)
			} else {
				fmt.Fprintf(out, "%s -> %s\n", issueKey, describeUser(user))
			}
		}
		if len(failures) > 0 {
			return fmt.Errorf("%d of %d issues not assigned: %v", len(failures), len(args), failures)
		}
		return nil
	},
}

func describeUser(u *jira.User) string {
	id := u.Name
	if id == "" {
		id = u.AccountID
	}
	if u.DisplayName != "" && u.DisplayName != id {
		return fmt.Sprintf("%s (%s)", id, u.DisplayName)
	}
	return id
}

func init() {
	assignCmd.Flags().StringVar(&assignTo, "to", "", "Username, email, display name, or \"me\"")
	assignCmd.Flags().BoolVar(&assignClear, "clear", false, "Clear the assignee instead of setting it")

	rootCmd.AddCommand(assignCmd)
}
