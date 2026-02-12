# TASK-260212-2c11gw: implement-config-file-io

## Description
Implement config file read/write at ~/.config/jira-mgmt/config.yaml. Use YAML library. Handle missing directory (create on first write). Handle missing file (return defaults). Handle corrupted file (backup and recreate with defaults).

## Scope
(define task scope)

## Acceptance Criteria
- Config file read/write works at ~/.config/jira-mgmt/config.yaml
- Directory created automatically on first write
- Missing file returns defaults without error
- Corrupted file backed up and recreated
- YAML format is human-readable
