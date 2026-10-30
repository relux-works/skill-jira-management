package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/relux-works/skill-jira-management/internal/jira"
	"github.com/spf13/cobra"
)

var attachmentUploadCmd = &cobra.Command{
	Use:   "upload ISSUE-KEY FILE",
	Short: "Upload a local file after verifying the issue exists",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		if cmd.Flags().Changed("name") && name == "" {
			return fmt.Errorf("--name must be a plain base name")
		}
		maxBytes, err := cmd.Flags().GetInt64("max-bytes")
		if err != nil {
			return err
		}
		upload, err := jira.PrepareAttachmentUpload(args[0], args[1], name, maxBytes)
		if err != nil {
			return err
		}
		client, err := buildJiraClientFromConfig()
		if err != nil {
			return err
		}
		attachment, err := client.UploadAttachment(upload)
		if err != nil {
			return err
		}
		result := struct {
			Issue    string `json:"issue"`
			ID       string `json:"id"`
			Filename string `json:"filename"`
			Size     int64  `json:"size"`
		}{args[0], attachment.ID, attachment.Filename, attachment.Size}
		if flagFormat == "json" {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(result)
		}
		_, err = fmt.Fprintf(cmd.OutOrStdout(), "Uploaded attachment %s to %s: %s (%d bytes)\n", attachment.ID, args[0], attachment.Filename, attachment.Size)
		return err
	},
}

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
	Short: "Download and upload Jira issue attachments",
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
	attachmentUploadCmd.Flags().String("name", "", "Uploaded filename (defaults to the local base name)")
	attachmentUploadCmd.Flags().Int64("max-bytes", jira.DefaultAttachmentUploadBytes, "Maximum accepted upload file size")
	attachmentCmd.AddCommand(attachmentUploadCmd)
	attachmentFetchCmd.Flags().String("out", "", "Output file path (required; must not already exist)")
	attachmentFetchCmd.Flags().Int64("max-bytes", 32<<20, "Maximum accepted attachment size")
	attachmentCmd.AddCommand(attachmentFetchCmd)
	rootCmd.AddCommand(attachmentCmd)
}
