package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/relux-works/skill-jira-management/internal/jira"
	"github.com/spf13/cobra"
)

var (
	transitionTo         string
	transitionFields     []string
	transitionListFields bool
)

var transitionCmd = &cobra.Command{
	Use:   "transition <ISSUE-KEY>",
	Short: "Transition a Jira issue to a new status",
	Long: `Move a Jira issue to a new workflow status.

Examples:
  jira-mgmt transition PROJ-123 --to "In Progress"
  jira-mgmt transition PROJ-456 --to "Done"

A workflow screen may demand fields that the issue does not yet carry; without
them Jira refuses the transition and the message names an internal field id
rather than anything the caller typed. List what a transition wants, then
supply it:

  jira-mgmt transition PROJ-123 --to "In Progress" --list-fields
  jira-mgmt transition PROJ-123 --to "In Progress" --field "Вид деятельности=Разработка"

--field takes the field's display name or its id, and is repeatable. A value
outside a field's allowed set is refused here, with the allowed values named,
rather than being sent for Jira to reject.`,
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

		var matched *jira.Transition
		for i := range transitions {
			if transitions[i].ID == transitionID {
				matched = &transitions[i]
				break
			}
		}

		out := cmd.OutOrStdout()
		if transitionListFields {
			printTransitionFields(out, matched)
			return nil
		}

		fields, err := buildTransitionFields(matched, transitionFields)
		if err != nil {
			return err
		}

		if err := client.DoTransition(issueKey, transitionID, fields); err != nil {
			return fmt.Errorf("executing transition: %w\n\nIf Jira names a required field, run the same command with --list-fields to see what it wants", err)
		}

		fmt.Fprintf(out, "%s -> %s\n", issueKey, matchedName)
		return nil
	},
}

func init() {
	transitionCmd.Flags().StringVar(&transitionTo, "to", "", "Target status name (required)")
	transitionCmd.Flags().StringArrayVar(&transitionFields, "field", nil, "Screen field as name=value or id=value; repeatable")
	transitionCmd.Flags().BoolVar(&transitionListFields, "list-fields", false, "Print the transition's screen fields and allowed values, and change nothing")

	rootCmd.AddCommand(transitionCmd)
}

// printTransitionFields describes the transition screen. Required fields come
// first: they are the ones that block the call.
//
// "optional" here is Jira's own metadata and it is not always the truth. A
// workflow validator can enforce a field the screen reports as optional —
// observed on MTS Server/DC, where "Вид деятельности" is metadata-optional and
// the transition still fails with «Поле … должно быть заполнено». So the list
// is what the transition can take, not a promise about what it will accept.
func printTransitionFields(out interface{ Write([]byte) (int, error) }, t *jira.Transition) {
	if t == nil || len(t.Fields) == 0 {
		fmt.Fprintln(out, "This transition has no screen fields.")
		return
	}
	defer fmt.Fprintln(out, "\nNote: a workflow validator may enforce a field listed here as\noptional. If the transition is refused, supply that field with --field.")
	var ids []string
	for id := range t.Fields {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := t.Fields[ids[i]], t.Fields[ids[j]]
		if a.Required != b.Required {
			return a.Required
		}
		return ids[i] < ids[j]
	})
	for _, id := range ids {
		f := t.Fields[id]
		label := "optional"
		if f.Required {
			label = "REQUIRED"
		}
		fmt.Fprintf(out, "%-8s %s  (id: %s)\n", label, f.Name, id)
		if len(f.AllowedValues) > 0 {
			var vals []string
			for _, v := range f.AllowedValues {
				if v.Value != "" {
					vals = append(vals, v.Value)
				} else if v.Name != "" {
					vals = append(vals, v.Name)
				}
			}
			fmt.Fprintf(out, "         allowed: %s\n", strings.Join(vals, ", "))
		}
	}
}

// buildTransitionFields turns --field arguments into the transition payload.
//
// Resolution is by display name or field id, because the caller sees the
// display name on the screen while Jira's own error messages name the id.
// An option-typed field is wrapped in the shape its schema demands; a plain
// string is sent as-is.
func buildTransitionFields(t *jira.Transition, specs []string) (map[string]interface{}, error) {
	if len(specs) == 0 {
		return nil, nil
	}
	if t == nil || len(t.Fields) == 0 {
		return nil, fmt.Errorf("this transition has no screen fields, so --field has nothing to set")
	}

	fields := map[string]interface{}{}
	for _, spec := range specs {
		name, value, ok := strings.Cut(spec, "=")
		if !ok {
			return nil, fmt.Errorf("--field %q is not name=value", spec)
		}
		name = strings.TrimSpace(name)
		value = strings.TrimSpace(value)

		id, f, err := resolveTransitionField(t, name)
		if err != nil {
			return nil, err
		}

		payload, err := coerceTransitionValue(f, value)
		if err != nil {
			return nil, err
		}
		fields[id] = payload
	}
	return fields, nil
}

func resolveTransitionField(t *jira.Transition, name string) (string, jira.TransitionField, error) {
	if f, ok := t.Fields[name]; ok {
		return name, f, nil
	}
	var matches []string
	for id, f := range t.Fields {
		if strings.EqualFold(f.Name, name) {
			matches = append(matches, id)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], t.Fields[matches[0]], nil
	case 0:
		var known []string
		for id, f := range t.Fields {
			known = append(known, fmt.Sprintf("%s (id: %s)", f.Name, id))
		}
		sort.Strings(known)
		return "", jira.TransitionField{}, fmt.Errorf(
			"this transition has no field %q; it has: %s", name, strings.Join(known, ", "))
	default:
		sort.Strings(matches)
		return "", jira.TransitionField{}, fmt.Errorf(
			"%q matches several fields (%s); use the id", name, strings.Join(matches, ", "))
	}
}

func coerceTransitionValue(f jira.TransitionField, value string) (interface{}, error) {
	if len(f.AllowedValues) == 0 {
		return value, nil
	}
	var allowed []string
	for _, v := range f.AllowedValues {
		label := v.Value
		if label == "" {
			label = v.Name
		}
		allowed = append(allowed, label)
		if strings.EqualFold(label, value) {
			out := map[string]interface{}{}
			if v.ID != "" {
				out["id"] = v.ID
			}
			if v.Value != "" {
				out["value"] = v.Value
			} else if v.Name != "" {
				out["name"] = v.Name
			}
			return out, nil
		}
	}
	return nil, fmt.Errorf("%q is not an allowed value for %q; allowed: %s",
		value, f.Name, strings.Join(allowed, ", "))
}
