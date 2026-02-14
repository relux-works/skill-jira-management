package config

import (
	"errors"
	"testing"
)

// mockKeyring simulates an in-memory keychain for testing.
type mockKeyring struct {
	store map[string]map[string]string // service -> user -> password
}

func newMockKeyring() *mockKeyring {
	return &mockKeyring{store: make(map[string]map[string]string)}
}

func (m *mockKeyring) set(service, user, password string) error {
	if m.store[service] == nil {
		m.store[service] = make(map[string]string)
	}
	m.store[service][user] = password
	return nil
}

func (m *mockKeyring) get(service, user string) (string, error) {
	if users, ok := m.store[service]; ok {
		if pw, ok := users[user]; ok {
			return pw, nil
		}
	}
	return "", errors.New("not found")
}

func (m *mockKeyring) delete(service, user string) error {
	if users, ok := m.store[service]; ok {
		if _, ok := users[user]; ok {
			delete(users, user)
			return nil
		}
	}
	return errors.New("not found")
}

func newTestStore() (*KeychainStore, *mockKeyring) {
	mk := newMockKeyring()
	store := NewKeychainStore(mk.set, mk.get, mk.delete)
	return store, mk
}

func validCreds() Credentials {
	return Credentials{
		InstanceURL: "https://test.atlassian.net",
		Email:       "user@test.com",
		APIToken:    "test-token-123",
	}
}

func TestCredentials_Validate(t *testing.T) {
	tests := []struct {
		name    string
		creds   Credentials
		wantErr bool
	}{
		{
			name:    "valid credentials",
			creds:   validCreds(),
			wantErr: false,
		},
		{
			name:    "missing instance URL",
			creds:   Credentials{Email: "a@b.com", APIToken: "tok"},
			wantErr: true,
		},
		{
			name:    "missing email",
			creds:   Credentials{InstanceURL: "https://x.atlassian.net", APIToken: "tok"},
			wantErr: true,
		},
		{
			name:    "missing token",
			creds:   Credentials{InstanceURL: "https://x.atlassian.net", Email: "a@b.com"},
			wantErr: true,
		},
		{
			name:    "all empty",
			creds:   Credentials{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.creds.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestKeychainStore_SaveAndLoad(t *testing.T) {
	store, _ := newTestStore()
	creds := validCreds()

	// Save
	if err := store.Save(creds); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load
	loaded, err := store.Load(creds.InstanceURL)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.InstanceURL != creds.InstanceURL {
		t.Errorf("InstanceURL = %q, want %q", loaded.InstanceURL, creds.InstanceURL)
	}
	if loaded.Email != creds.Email {
		t.Errorf("Email = %q, want %q", loaded.Email, creds.Email)
	}
	if loaded.APIToken != creds.APIToken {
		t.Errorf("APIToken = %q, want %q", loaded.APIToken, creds.APIToken)
	}
}

func TestKeychainStore_SaveInvalid(t *testing.T) {
	store, _ := newTestStore()

	err := store.Save(Credentials{})
	if err == nil {
		t.Fatal("Save() with empty credentials should fail")
	}
}

func TestKeychainStore_LoadNotFound(t *testing.T) {
	store, _ := newTestStore()

	_, err := store.Load("https://nonexistent.atlassian.net")
	if !errors.Is(err, ErrCredentialsNotFound) {
		t.Errorf("Load() error = %v, want ErrCredentialsNotFound", err)
	}
}

func TestKeychainStore_LoadEmptyURL(t *testing.T) {
	store, _ := newTestStore()

	_, err := store.Load("")
	if err == nil {
		t.Fatal("Load() with empty URL should fail")
	}
}

func TestKeychainStore_Delete(t *testing.T) {
	store, _ := newTestStore()
	creds := validCreds()

	// Save then delete
	if err := store.Save(creds); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := store.Delete(creds.InstanceURL); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Should not be found after deletion
	_, err := store.Load(creds.InstanceURL)
	if !errors.Is(err, ErrCredentialsNotFound) {
		t.Errorf("Load() after Delete() error = %v, want ErrCredentialsNotFound", err)
	}
}

func TestKeychainStore_DeleteNotFound(t *testing.T) {
	store, _ := newTestStore()

	err := store.Delete("https://nonexistent.atlassian.net")
	if err == nil {
		t.Fatal("Delete() of nonexistent credentials should fail")
	}
}

func TestKeychainStore_DeleteEmptyURL(t *testing.T) {
	store, _ := newTestStore()

	err := store.Delete("")
	if err == nil {
		t.Fatal("Delete() with empty URL should fail")
	}
}

func TestKeychainStore_Overwrite(t *testing.T) {
	store, _ := newTestStore()
	creds := validCreds()

	// Save original
	if err := store.Save(creds); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Overwrite with new token
	creds.APIToken = "new-token-456"
	if err := store.Save(creds); err != nil {
		t.Fatalf("Save() overwrite error = %v", err)
	}

	// Load should return updated token
	loaded, err := store.Load(creds.InstanceURL)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.APIToken != "new-token-456" {
		t.Errorf("APIToken = %q, want %q", loaded.APIToken, "new-token-456")
	}
}

func TestKeychainStore_MultipleInstances(t *testing.T) {
	store, _ := newTestStore()

	creds1 := Credentials{
		InstanceURL: "https://one.atlassian.net",
		Email:       "user1@one.com",
		APIToken:    "token-one",
	}
	creds2 := Credentials{
		InstanceURL: "https://two.atlassian.net",
		Email:       "user2@two.com",
		APIToken:    "token-two",
	}

	// Save both
	if err := store.Save(creds1); err != nil {
		t.Fatalf("Save(creds1) error = %v", err)
	}
	if err := store.Save(creds2); err != nil {
		t.Fatalf("Save(creds2) error = %v", err)
	}

	// Load each independently
	loaded1, err := store.Load(creds1.InstanceURL)
	if err != nil {
		t.Fatalf("Load(creds1) error = %v", err)
	}
	loaded2, err := store.Load(creds2.InstanceURL)
	if err != nil {
		t.Fatalf("Load(creds2) error = %v", err)
	}

	if loaded1.Email != creds1.Email {
		t.Errorf("creds1 Email = %q, want %q", loaded1.Email, creds1.Email)
	}
	if loaded2.Email != creds2.Email {
		t.Errorf("creds2 Email = %q, want %q", loaded2.Email, creds2.Email)
	}
}

func TestKeychainStore_KeyringSetError(t *testing.T) {
	failSet := func(service, user, password string) error {
		return errors.New("keychain unavailable")
	}
	mk := newMockKeyring()
	store := NewKeychainStore(failSet, mk.get, mk.delete)

	err := store.Save(validCreds())
	if err == nil {
		t.Fatal("Save() should fail when keyring set fails")
	}
}

func TestKeychainStore_KeyringGetCorruptData(t *testing.T) {
	mk := newMockKeyring()
	store := NewKeychainStore(mk.set, mk.get, mk.delete)

	// Manually insert corrupt JSON into current service name
	mk.store[serviceName] = map[string]string{
		"https://bad.atlassian.net": "not-json",
	}

	_, err := store.Load("https://bad.atlassian.net")
	if err == nil {
		t.Fatal("Load() should fail with corrupt JSON data")
	}
}

func TestKeychainStore_LegacyMigration(t *testing.T) {
	mk := newMockKeyring()
	store := NewKeychainStore(mk.set, mk.get, mk.delete)
	creds := validCreds()

	// Simulate legacy: save directly under old service name
	data := `{"instance_url":"https://test.atlassian.net","email":"user@test.com","api_token":"test-token-123"}`
	mk.store[legacyServiceName] = map[string]string{
		creds.InstanceURL: data,
	}

	// Load should find it via legacy fallback
	loaded, err := store.Load(creds.InstanceURL)
	if err != nil {
		t.Fatalf("Load() from legacy error = %v", err)
	}
	if loaded.APIToken != creds.APIToken {
		t.Errorf("APIToken = %q, want %q", loaded.APIToken, creds.APIToken)
	}

	// Should have migrated: present in new, gone from legacy
	if _, ok := mk.store[serviceName][creds.InstanceURL]; !ok {
		t.Error("credentials should be migrated to new service name")
	}
	if _, ok := mk.store[legacyServiceName][creds.InstanceURL]; ok {
		t.Error("credentials should be removed from legacy service name")
	}
}

func TestKeychainStore_LegacyNotUsedWhenNewExists(t *testing.T) {
	mk := newMockKeyring()
	store := NewKeychainStore(mk.set, mk.get, mk.delete)

	// Put different tokens in legacy and new
	legacyData := `{"instance_url":"https://test.atlassian.net","email":"user@test.com","api_token":"old-token"}`
	newData := `{"instance_url":"https://test.atlassian.net","email":"user@test.com","api_token":"new-token"}`

	mk.store[legacyServiceName] = map[string]string{
		"https://test.atlassian.net": legacyData,
	}
	mk.store[serviceName] = map[string]string{
		"https://test.atlassian.net": newData,
	}

	loaded, err := store.Load("https://test.atlassian.net")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should use new, not legacy
	if loaded.APIToken != "new-token" {
		t.Errorf("APIToken = %q, want %q (should prefer new over legacy)", loaded.APIToken, "new-token")
	}
}
