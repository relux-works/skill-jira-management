# Secure Token Storage Research

**Date:** 2025-02-12
**Task:** TASK-260212-p52iok
**Author:** agent-config

## Problem

We need to securely store Jira API credentials (instance URL, email, API token) on macOS. The token is the most sensitive piece — it grants full API access.

## Options Evaluated

### Option 1: `github.com/zalando/go-keyring`

**What it is:** Pure Go library that wraps OS-native credential stores.

**Platforms supported:**
- macOS: Keychain (via Security framework / `security` CLI)
- Linux: Secret Service (GNOME Keyring, KDE Wallet)
- Windows: Credential Manager (wincred)

**API:**
```go
// Store
go_keyring.Set(service, user, password)

// Retrieve
password, err := go_keyring.Get(service, user)

// Delete
go_keyring.Delete(service, user)
```

**Pros:**
- Cross-platform out of the box
- Simple API — three functions
- Well-maintained (Zalando)
- Uses OS-native secure storage (Keychain on macOS)
- No CGO required on macOS (uses `security` CLI under the hood)
- ~1.3k stars, actively maintained

**Cons:**
- Stores only string values (key-value pairs)
- Each field must be stored separately OR serialized into one string
- No built-in encryption — relies entirely on OS keychain
- Limited metadata support

### Option 2: Direct `security` CLI usage

**What it is:** macOS ships with `/usr/bin/security` CLI for Keychain operations.

**Commands:**
```bash
# Add
security add-generic-password -a "account" -s "service" -w "password" -U

# Find
security find-generic-password -a "account" -s "service" -w

# Delete
security delete-generic-password -a "account" -s "service"
```

**Pros:**
- Zero dependencies
- Native macOS Keychain
- Full control over keychain item attributes

**Cons:**
- macOS only — no cross-platform
- Requires shelling out to `security` CLI via `os/exec`
- Parsing CLI output is fragile
- Error handling is messy (exit codes + stderr parsing)
- Essentially reimplementing what go-keyring already does

### Option 3: Encrypted file (`~/.config/jira-mgmt/credentials`)

**What it is:** Store credentials in an encrypted file using AES-256-GCM or similar.

**Pros:**
- Fully cross-platform
- No OS dependencies
- Full control over format

**Cons:**
- Must manage encryption key — where do you store THAT key?
- Key derivation from password = need to prompt user every time
- Reinventing the wheel — OS keychains exist for this exact purpose
- Security is only as good as our implementation

### Option 4: Plain text file with permissions

**What it is:** Store credentials in a YAML/JSON file with `0600` permissions.

**Pros:**
- Simplest implementation
- Easy to debug

**Cons:**
- Not secure — any process running as the user can read it
- Not acceptable for API tokens

## Decision

**Go with `github.com/zalando/go-keyring`.**

### Rationale:

1. **Cross-platform for free.** Even though we target macOS primarily, Linux support comes free. No reason to lock ourselves in.
2. **Battle-tested.** Zalando uses it in production. Well-maintained.
3. **Simple API.** Three functions — Set, Get, Delete. Exactly what we need.
4. **No CGO.** On macOS it uses the `security` CLI internally, so no CGO dependency. Clean builds.
5. **OS-native security.** Keychain is the right place for API tokens on macOS.

### Storage scheme:

We'll use a single keychain entry per credential set:

- **Service name:** `jira-mgmt` (constant)
- **Account:** instance URL (e.g., `https://mycompany.atlassian.net`)
- **Password:** JSON-serialized credentials blob

```go
type Credentials struct {
    InstanceURL string `json:"instance_url"`
    Email       string `json:"email"`
    APIToken    string `json:"api_token"`
}
```

Serialize the full `Credentials` struct to JSON and store as the "password" value. This way we use one keychain entry instead of three separate ones.

### Interface design:

```go
type CredentialStore interface {
    Save(creds Credentials) error
    Load(instanceURL string) (Credentials, error)
    Delete(instanceURL string) error
}
```

This allows:
- `KeychainStore` — real implementation using go-keyring
- `MockStore` — for testing without touching the real keychain

### Testing approach:

- Define `CredentialStore` interface
- Tests use a mock/in-memory implementation
- Integration tests (if needed) can use real keychain on CI with a test service name
- No dependency on real keychain in unit tests

## Dependencies to Add

```
github.com/zalando/go-keyring v0.2.6
```

## File Structure

```
internal/config/
├── auth.go          # Credentials type + CredentialStore interface + KeychainStore
├── config.go        # Config type + file I/O
├── auth_test.go     # Tests for credential storage (mocked)
└── config_test.go   # Tests for config file operations
```
