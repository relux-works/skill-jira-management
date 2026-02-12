package main

import (
	"fmt"
	"strconv"

	"github.com/ivalx1s/skill-jira-management/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage jira-mgmt configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Keys:
  project  — active Jira project key (e.g. PROJ)
  board    — active board ID (e.g. 42)
  locale   — content locale: en or ru

Examples:
  jira-mgmt config set project MYPROJ
  jira-mgmt config set board 42
  jira-mgmt config set locale en`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		cfgMgr, err := config.NewConfigManager()
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()

		switch key {
		case "project":
			if err := cfgMgr.SetActiveProject(value); err != nil {
				return err
			}
			fmt.Fprintf(out, "Active project set to %s\n", value)

		case "board":
			boardID, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("board must be a number: %w", err)
			}
			if err := cfgMgr.SetActiveBoard(boardID); err != nil {
				return err
			}
			fmt.Fprintf(out, "Active board set to %d\n", boardID)

		case "locale":
			if err := cfgMgr.SetLocale(config.Locale(value)); err != nil {
				return err
			}
			fmt.Fprintf(out, "Locale set to %s\n", value)

		default:
			return fmt.Errorf("unknown config key %q (supported: project, board, locale)", key)
		}

		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgMgr, err := config.NewConfigManager()
		if err != nil {
			return err
		}

		cfg, err := cfgMgr.GetConfig()
		if err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprintln(out, "Configuration")
		fmt.Fprintln(out, "=============")
		fmt.Fprintf(out, "  config file:    %s\n", cfgMgr.ConfigPath())
		fmt.Fprintf(out, "  instance:       %s\n", valueOrNone(cfg.InstanceURL))
		fmt.Fprintf(out, "  active project: %s\n", valueOrNone(cfg.ActiveProject))
		if cfg.ActiveBoard != 0 {
			fmt.Fprintf(out, "  active board:   %d\n", cfg.ActiveBoard)
		} else {
			fmt.Fprintf(out, "  active board:   (none)\n")
		}
		fmt.Fprintf(out, "  locale:         %s\n", cfg.Locale)

		return nil
	},
}

func valueOrNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}
