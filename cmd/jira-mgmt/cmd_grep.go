package main

import (
	"fmt"

	"github.com/ivalx1s/skill-jira-management/internal/jira"
	"github.com/ivalx1s/skill-jira-management/internal/search"
	"github.com/spf13/cobra"
)

var (
	grepScope           string
	grepCaseInsensitive bool
	grepContextLines    int
)

var grepCmd = &cobra.Command{
	Use:   "grep <pattern>",
	Short: "Search across Jira issues for matching text",
	Long: `Full-text regex search across Jira issues in the active project.

Scopes:
  issues   — search summary, description, labels
  comments — search issue comments
  all      — search everything (default)

Examples:
  jira-mgmt grep "authentication"
  jira-mgmt grep "TODO" --scope issues -i
  jira-mgmt grep "deploy" --scope comments -C 2`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]

		client, err := buildJiraClientFromConfig()
		if err != nil {
			return err
		}

		project := flagProject
		if project == "" {
			return fmt.Errorf("project required: use --project flag or configure via 'jira-mgmt config set project KEY'")
		}

		opts := search.GrepOptions{
			Scope:           grepScope,
			CaseInsensitive: grepCaseInsensitive,
			ContextLines:    grepContextLines,
		}

		// Fetch issues for the project
		issues, err := client.ListIssues(jira.ListIssuesOptions{
			ProjectKey: project,
			Fields:     []string{"summary", "description", "labels"},
		})
		if err != nil {
			return fmt.Errorf("fetching issues: %w", err)
		}

		var allMatches []search.Match

		// Search issues
		if grepScope == "issues" || grepScope == "all" || grepScope == "" {
			matches, err := search.GrepIssues(issues, pattern, opts)
			if err != nil {
				return err
			}
			allMatches = append(allMatches, matches...)
		}

		// Search comments
		if grepScope == "comments" || grepScope == "all" || grepScope == "" {
			for _, issue := range issues {
				comments, err := client.ListAllComments(issue.Key)
				if err != nil {
					continue
				}
				matches, err := search.GrepComments(comments, issue.Key, pattern, opts)
				if err != nil {
					continue
				}
				allMatches = append(allMatches, matches...)
			}
		}

		out := cmd.OutOrStdout()
		if flagFormat == "json" {
			data, err := search.PrintJSON(allMatches)
			if err != nil {
				return err
			}
			fmt.Fprintln(out, string(data))
		} else {
			fmt.Fprint(out, search.PrintText(allMatches))
		}

		return nil
	},
}

func init() {
	grepCmd.Flags().StringVar(&grepScope, "scope", "all", "Search scope: issues, comments, all")
	grepCmd.Flags().BoolVarP(&grepCaseInsensitive, "ignore-case", "i", false, "Case-insensitive search")
	grepCmd.Flags().IntVarP(&grepContextLines, "context", "C", 0, "Context lines around matches")

	rootCmd.AddCommand(grepCmd)
}
