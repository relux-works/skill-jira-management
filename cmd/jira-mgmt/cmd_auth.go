package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/relux-works/skill-jira-management/internal/config"
	"github.com/relux-works/skill-jira-management/internal/jira"
	"github.com/spf13/cobra"
)

var (
	authFlagInstance string
	authFlagEmail    string
	authFlagToken    string
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Configure Jira authentication",
	Long: `Setup Jira authentication (Cloud or Server/DC).

Cloud (Basic auth — email + API token):
  jira-mgmt auth --instance https://mycompany.atlassian.net --email user@company.com --token API_TOKEN

Server/DC (Bearer auth — Personal Access Token):
  jira-mgmt auth --instance https://jira.company.com --token PAT_TOKEN

Interactive (prompts for input):
  jira-mgmt auth`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		instanceURL := authFlagInstance
		email := authFlagEmail
		apiToken := authFlagToken

		// If flags not provided, fall back to interactive prompts
		if instanceURL == "" || apiToken == "" {
			reader := bufio.NewReader(os.Stdin)

			fmt.Fprintln(out, "Jira Authentication Setup")
			fmt.Fprintln(out, "=========================")
			fmt.Fprintln(out)

			if instanceURL == "" {
				fmt.Fprint(out, "Instance URL (e.g. https://mycompany.atlassian.net): ")
				line, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("reading input: %w", err)
				}
				instanceURL = strings.TrimSpace(line)
			}

			if email == "" {
				fmt.Fprint(out, "Email (leave empty for Server/DC PAT auth): ")
				line, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("reading input: %w", err)
				}
				email = strings.TrimSpace(line)
			}

			if apiToken == "" {
				fmt.Fprint(out, "API Token / PAT: ")
				line, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("reading input: %w", err)
				}
				apiToken = strings.TrimSpace(line)
			}
		}

		// Determine auth type from presence of email.
		authType := "basic"
		if email == "" {
			authType = "bearer"
		}

		creds := config.Credentials{
			InstanceURL: instanceURL,
			Email:       email,
			APIToken:    apiToken,
			AuthType:    authType,
		}

		if err := creds.Validate(); err != nil {
			return fmt.Errorf("invalid credentials: %w", err)
		}

		// Validate against API
		fmt.Fprintln(out, "\nValidating credentials...")
		fmt.Fprintf(out, "Auth method: %s\n", authType)
		client, err := jira.NewClient(jira.Config{
			BaseURL:  instanceURL,
			Email:    email,
			Token:    apiToken,
			AuthType: jira.AuthType(authType),
		})
		if err != nil {
			return fmt.Errorf("creating client: %w", err)
		}

		// Use v2 myself endpoint — works on both Cloud and Server/DC.
		_, err = client.Get("/rest/api/2/myself", nil)
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Auto-detect instance type (Cloud vs Server/DC).
		fmt.Fprintln(out, "Detecting instance type...")
		instanceType, err := client.DetectInstanceType()
		if err != nil {
			return fmt.Errorf("detecting instance type: %w", err)
		}
		fmt.Fprintf(out, "Instance type: %s\n", instanceType)

		// Save credentials
		store := getCredentialStore()
		if err := store.Save(creds); err != nil {
			return fmt.Errorf("saving credentials: %w", err)
		}

		// Save instance URL and type to config
		cfgMgr, err := config.NewConfigManager()
		if err != nil {
			return fmt.Errorf("config manager: %w", err)
		}
		_ = cfgMgr.SetInstanceURL(instanceURL)
		_ = cfgMgr.SetInstanceType(string(instanceType))
		_ = cfgMgr.SetAuthType(authType)

		fmt.Fprintln(out, "Authentication configured successfully.")
		return nil
	},
}

func init() {
	authCmd.Flags().StringVar(&authFlagInstance, "instance", "", "Jira Cloud instance URL (e.g. https://mycompany.atlassian.net)")
	authCmd.Flags().StringVar(&authFlagEmail, "email", "", "Jira account email")
	authCmd.Flags().StringVar(&authFlagToken, "token", "", "Jira API token")
	rootCmd.AddCommand(authCmd)
}
