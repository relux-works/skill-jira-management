package main

import (
	"fmt"

	"github.com/relux-works/skill-jira-management/internal/config"
	"github.com/relux-works/skill-jira-management/internal/jira"
	"github.com/zalando/go-keyring"
)

// getCredentialStore returns the keychain-backed credential store.
// Uses a mock keyring for testing when overridden.
var getCredentialStore = func() config.CredentialStore {
	return config.NewKeychainStore(
		keyring.Set,
		keyring.Get,
		keyring.Delete,
	)
}

// buildJiraClientFromConfig creates a Jira client from stored config and credentials.
func buildJiraClientFromConfig() (*jira.Client, error) {
	cfgMgr, err := config.NewConfigManager()
	if err != nil {
		return nil, fmt.Errorf("config manager: %w", err)
	}

	cfg, err := cfgMgr.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	if cfg.InstanceURL == "" {
		return nil, fmt.Errorf("not configured: run 'jira-mgmt auth' first")
	}

	store := getCredentialStore()
	creds, err := store.Load(cfg.InstanceURL)
	if err != nil {
		return nil, fmt.Errorf("loading credentials: %w (run 'jira-mgmt auth' to configure)", err)
	}

	return jira.NewClient(jira.Config{
		BaseURL:      creds.InstanceURL,
		Email:        creds.Email,
		Token:        creds.APIToken,
		InstanceType: jira.InstanceType(cfg.InstanceType),
		AuthType:     jira.AuthType(cfg.AuthType),
	})
}
