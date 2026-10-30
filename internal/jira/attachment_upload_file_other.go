//go:build !darwin && !linux

package jira

import (
	"fmt"
	"os"
)

// openAttachmentFile preserves descriptor identity checks on other supported platforms.
func openAttachmentFile(path string, expected os.FileInfo) (*os.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(expected, opened) {
		file.Close()
		return nil, fmt.Errorf("attachment file changed during validation")
	}
	return file, nil
}
