package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var attachmentFetchCmd = &cobra.Command{
	Use:   "fetch ISSUE-KEY ATTACHMENT-ID",
	Short: "Download one attachment after verifying issue ownership",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		outPath, err := cmd.Flags().GetString("out")
		if err != nil {
			return err
		}
		if outPath == "" {
			return fmt.Errorf("--out is required")
		}
		maxBytes, err := cmd.Flags().GetInt64("max-bytes")
		if err != nil {
			return err
		}

		client, err := buildJiraClientFromConfig()
		if err != nil {
			return err
		}
		data, attachment, err := client.DownloadAttachment(args[0], args[1], maxBytes)
		if err != nil {
			return err
		}
		if err := writePrivateFile(outPath, data); err != nil {
			return err
		}

		result := map[string]any{
			"id":       attachment.ID,
			"filename": attachment.Filename,
			"size":     len(data),
			"out":      outPath,
		}
		if flagFormat == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Downloaded attachment %s to %s (%d bytes)\n", attachment.ID, outPath, len(data))
		return nil
	},
}

var attachmentCmd = &cobra.Command{
	Use:   "attachment",
	Short: "Read Jira issue attachments",
}

func writePrivateFile(path string, data []byte) error {
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("output already exists: %s", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect output path: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".jira-attachment-*")
	if err != nil {
		return fmt.Errorf("prepare output: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("protect output: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write output: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close output: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("publish output: %w", err)
	}
	return nil
}

func init() {
	attachmentFetchCmd.Flags().String("out", "", "Output file path (required; must not already exist)")
	attachmentFetchCmd.Flags().Int64("max-bytes", 32<<20, "Maximum accepted attachment size")
	attachmentCmd.AddCommand(attachmentFetchCmd)
	rootCmd.AddCommand(attachmentCmd)
}
