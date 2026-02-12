package main

import (
	"fmt"
	"os"

	"github.com/ivalx1s/skill-jira-management/internal/config"
	"github.com/spf13/cobra"
)

// Build-time variables set via ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Global flags.
var (
	flagProject string
	flagBoard   int
	flagFormat  string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "jira-mgmt",
	Short: "Jira management CLI for AI agents",
	Long:  "Agent-facing CLI for Jira Cloud: DSL queries, scoped grep, and write commands.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: persistentPreRun,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "jira-mgmt %s\n  commit: %s\n  built:  %s\n", version, commit, date)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagProject, "project", "", "Jira project key (overrides config)")
	rootCmd.PersistentFlags().IntVar(&flagBoard, "board", 0, "Jira board ID (overrides config)")
	rootCmd.PersistentFlags().StringVar(&flagFormat, "format", "json", "Output format: json or text")

	rootCmd.AddCommand(versionCmd)
}

func persistentPreRun(cmd *cobra.Command, args []string) error {
	loadConfigDefaults(cmd)

	// Skip first-run check for commands that don't need auth.
	name := cmd.Name()
	if name == "version" || name == "auth" || name == "config" || name == "set" || name == "show" || name == "help" || name == "jira-mgmt" {
		return nil
	}

	return checkFirstRun()
}

// loadConfigDefaults fills global flags from config when not set explicitly via CLI.
func loadConfigDefaults(cmd *cobra.Command) {
	mgr, err := config.NewConfigManager()
	if err != nil {
		return
	}

	cfg, err := mgr.GetConfig()
	if err != nil {
		return
	}

	if !cmd.Flags().Changed("project") && cfg.ActiveProject != "" {
		flagProject = cfg.ActiveProject
	}
	if !cmd.Flags().Changed("board") && cfg.ActiveBoard != 0 {
		flagBoard = cfg.ActiveBoard
	}
}

// checkFirstRun verifies that authentication is configured.
func checkFirstRun() error {
	cfgMgr, err := config.NewConfigManager()
	if err != nil {
		return nil
	}

	cfg, err := cfgMgr.GetConfig()
	if err != nil {
		return nil
	}

	if cfg.InstanceURL == "" {
		return fmt.Errorf("jira-mgmt is not configured\nRun 'jira-mgmt auth' to set up authentication")
	}

	return nil
}
