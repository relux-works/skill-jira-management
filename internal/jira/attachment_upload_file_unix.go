//go:build darwin || linux

package jira

import (
	"fmt"
	"os"
	"syscall"
)

// openAttachmentFile cannot block on a substituted special file and never follows a final symlink.
func openAttachmentFile(path string, expected os.FileInfo) (*os.File, error) {
	fd, err := syscall.Open(path, syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(expected, opened) {
		file.Close()
		return nil, fmt.Errorf("attachment file changed during validation")
	}
	// Only a verified regular descriptor can be switched to blocking reads.
	if err := syscall.SetNonblock(fd, false); err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}
