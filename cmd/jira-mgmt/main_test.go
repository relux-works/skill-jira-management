package main

import (
	"bytes"
	"strings"
	"testing"
)

func executeCommand(args ...string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestRootCommand_NoArgs(t *testing.T) {
	out, err := executeCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Agent-facing CLI for Jira Cloud") {
		t.Errorf("expected help text in output, got: %s", out)
	}
}

func TestVersionCommand(t *testing.T) {
	out, err := executeCommand("version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "jira-mgmt") {
		t.Errorf("expected 'jira-mgmt' in version output, got: %s", out)
	}
	if !strings.Contains(out, "commit:") {
		t.Errorf("expected 'commit:' in version output, got: %s", out)
	}
}

func TestGlobalFlags_Format(t *testing.T) {
	// Default format is json
	rootCmd.SetArgs([]string{"version", "--format", "text"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flagFormat != "text" {
		t.Errorf("expected flagFormat='text', got %q", flagFormat)
	}

	// Reset
	rootCmd.SetArgs([]string{"version", "--format", "json"})
	_ = rootCmd.Execute()
	if flagFormat != "json" {
		t.Errorf("expected flagFormat='json', got %q", flagFormat)
	}
}

func TestGlobalFlags_Project(t *testing.T) {
	rootCmd.SetArgs([]string{"version", "--project", "MYPROJ"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flagProject != "MYPROJ" {
		t.Errorf("expected flagProject='MYPROJ', got %q", flagProject)
	}
}

func TestGlobalFlags_Board(t *testing.T) {
	rootCmd.SetArgs([]string{"version", "--board", "42"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flagBoard != 42 {
		t.Errorf("expected flagBoard=42, got %d", flagBoard)
	}
}
