package main

import (
	"fmt"

	"github.com/relux-works/skill-jira-management/internal/config"
	"github.com/relux-works/skill-jira-management/internal/jira"
	"github.com/zalando/go-keyring"
)

// getCredentialResolver returns the platform-aware credential resolver.
// Tests can override it.
var getCredentialResolver = func() *config.Resolver {
	return config.NewResolver(
		config.Runtime{},
		config.NewKeychainStore(
			keyring.Set,
			keyring.Get,
			keyring.Delete,
		),
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

	resolver := getCredentialResolver()
	instanceURL := resolver.ResolveInstanceURL(cfg.InstanceURL)
	if instanceURL == "" {
		return nil, fmt.Errorf("not configured: run 'jira-mgmt auth set-access' first")
	}

	resolved, err := resolver.Resolve(config.SourceAuto, instanceURL)
	if err != nil {
		return nil, fmt.Errorf("loading credentials: %w (run 'jira-mgmt auth set-access' to configure)", err)
	}

	client, err := jira.NewClient(jira.Config{
		BaseURL:            resolved.Credentials.InstanceURL,
		Email:              resolved.Credentials.Email,
		Token:              resolved.Credentials.APIToken,
		InstanceType:       jira.InstanceType(cfg.InstanceType),
		AuthType:           jira.AuthType(resolved.Credentials.AuthType),
		InsecureSkipVerify: flagInsecure || cfg.TLSSkipVerify,
	})
	if err != nil {
		return nil, err
	}

	if cfg.InstanceType == "" {
		if instanceType, err := client.DetectInstanceType(); err == nil {
			_ = cfgMgr.SetInstanceType(string(instanceType))
		}
	}
	if cfg.AuthType == "" && resolved.ResolvedFrom != "env" {
		_ = cfgMgr.SetAuthType(resolved.Credentials.AuthType)
	}

	return client, nil
}
