package main

import (
	"fmt"
	"strings"

	"github.com/relux-works/skill-jira-management/internal/jira"
	"github.com/spf13/cobra"
)

var (
	cancelReason          string
	cancelCascadeSubtasks bool
)

var cancelCmd = &cobra.Command{
	Use:   "cancel <ISSUE-KEY>",
	Short: "Cancel a Jira issue with workflow-aware fields",
	Long: `Cancel a Jira issue via its workflow "Cancel" transition.

This command is aware of common Jira workflow requirements such as:
  - selecting a cancel-like resolution (for example "Отменено")
  - filling select fields like "Причина переноса / отмены" with a best-effort fallback to "Другое"
  - optionally cascading the cancel operation to direct subtasks first
  - optionally adding a reason comment after the transition succeeds

Examples:
  jira-mgmt cancel PROJ-123
  jira-mgmt cancel PROJ-123 --reason "прекращение работы с ICONIA"
  jira-mgmt cancel PROJ-123 --reason "прекращение работы с ICONIA" --cascade-subtasks`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := buildJiraClientFromConfig()
		if err != nil {
			return err
		}

		issueKey := args[0]
		if err := runCancelIssue(client, issueKey, cancelOptions{
			Reason:          cancelReason,
			CascadeSubtasks: cancelCascadeSubtasks,
		}, map[string]struct{}{}); err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "%s cancelled\n", issueKey)
		return nil
	},
}

type cancelOptions struct {
	Reason          string
	CascadeSubtasks bool
}

func init() {
	cancelCmd.Flags().StringVar(&cancelReason, "reason", "", "Optional cancel reason comment")
	cancelCmd.Flags().BoolVar(&cancelCascadeSubtasks, "cascade-subtasks", false, "Cancel direct subtasks before the parent issue")

	rootCmd.AddCommand(cancelCmd)
}

func runCancelIssue(client *jira.Client, issueKey string, opts cancelOptions, visited map[string]struct{}) error {
	if _, ok := visited[issueKey]; ok {
		return nil
	}
	visited[issueKey] = struct{}{}

	fields := []string{"status"}
	if opts.CascadeSubtasks {
		fields = append(fields, "subtasks")
	}

	issue, err := client.GetIssue(issueKey, fields)
	if err != nil {
		return fmt.Errorf("getting issue %s: %w", issueKey, err)
	}

	if isDoneStatus(issue.Fields.Status) {
		return nil
	}

	if opts.CascadeSubtasks {
		for _, subtask := range issue.Fields.Subtasks {
			if subtask.Key == "" || isDoneStatus(subtask.Fields.Status) {
				continue
			}
			if err := runCancelIssue(client, subtask.Key, cancelOptions{
				Reason:          opts.Reason,
				CascadeSubtasks: false,
			}, visited); err != nil {
				return fmt.Errorf("canceling subtask %s: %w", subtask.Key, err)
			}
		}
	}

	transitions, err := client.GetTransitions(issueKey)
	if err != nil {
		return fmt.Errorf("getting transitions for %s: %w", issueKey, err)
	}

	cancelTransition, err := findCancelTransition(transitions)
	if err != nil {
		return fmt.Errorf("finding cancel transition for %s: %w", issueKey, err)
	}

	transitionFields, err := buildCancelFields(cancelTransition, opts.Reason)
	if err != nil {
		return fmt.Errorf("building cancel fields for %s: %w", issueKey, err)
	}

	if err := client.DoTransition(issueKey, cancelTransition.ID, transitionFields); err != nil {
		return fmt.Errorf("canceling %s: %w", issueKey, err)
	}

	if strings.TrimSpace(opts.Reason) != "" {
		if _, err := client.AddComment(issueKey, jira.NewADFText(opts.Reason)); err != nil {
			return fmt.Errorf("adding cancel reason comment to %s: %w", issueKey, err)
		}
	}

	return nil
}

func findCancelTransition(transitions []jira.Transition) (*jira.Transition, error) {
	for _, transition := range transitions {
		if strings.EqualFold(transition.Name, "Cancel") || strings.EqualFold(transition.To.Name, "Cancel") {
			t := transition
			return &t, nil
		}
	}

	var available []string
	for _, transition := range transitions {
		available = append(available, fmt.Sprintf("%s -> %s", transition.Name, transition.To.Name))
	}

	if len(available) == 0 {
		return nil, fmt.Errorf("no transitions available")
	}

	return nil, fmt.Errorf("no cancel transition found; available transitions:\n  %s", strings.Join(available, "\n  "))
}

func buildCancelFields(transition *jira.Transition, reason string) (map[string]interface{}, error) {
	if transition == nil {
		return nil, fmt.Errorf("transition is required")
	}

	fields := map[string]interface{}{}

	if resolutionField, ok := transition.Fields["resolution"]; ok {
		resolution, found := findTransitionOption(resolutionField.AllowedValues,
			[]string{"Отменено", "Cancelled", "Canceled", "Cancel"})
		if !found {
			return nil, fmt.Errorf("cancel transition requires resolution, but no cancel-like option was found")
		}
		fields["resolution"] = transitionOptionRef(resolution)
	}

	for fieldID, field := range transition.Fields {
		if fieldID == "resolution" {
			continue
		}

		option, found := chooseCancelReasonOption(field.AllowedValues, reason)
		if found {
			fields[fieldID] = transitionOptionRef(option)
			continue
		}

		if field.Required {
			return nil, fmt.Errorf("cancel transition requires unsupported field %q (%s)", fieldID, field.Name)
		}
	}

	if len(fields) == 0 {
		return nil, nil
	}

	return fields, nil
}

func chooseCancelReasonOption(options []jira.TransitionOption, reason string) (jira.TransitionOption, bool) {
	if len(options) == 0 {
		return jira.TransitionOption{}, false
	}

	if strings.TrimSpace(reason) != "" {
		if exact, found := findTransitionOption(options, []string{reason}); found {
			return exact, true
		}
	}

	return findTransitionOption(options, []string{
		"Другое (прокомментируйте что именно)",
		"Другое",
		"Other (comment with details)",
		"Other",
		"Прочее",
	})
}

func findTransitionOption(options []jira.TransitionOption, candidates []string) (jira.TransitionOption, bool) {
	for _, candidate := range candidates {
		for _, option := range options {
			if option.Disabled {
				continue
			}
			if strings.EqualFold(option.Name, candidate) || strings.EqualFold(option.Value, candidate) {
				return option, true
			}
		}
	}

	return jira.TransitionOption{}, false
}

func transitionOptionRef(option jira.TransitionOption) map[string]string {
	switch {
	case option.ID != "":
		return map[string]string{"id": option.ID}
	case option.Value != "":
		return map[string]string{"value": option.Value}
	default:
		return map[string]string{"name": option.Name}
	}
}

func isDoneStatus(status *jira.Status) bool {
	return status != nil && status.StatusCategory != nil && strings.EqualFold(status.StatusCategory.Key, "done")
}
