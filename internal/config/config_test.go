package config

import (
	"os"
	"path/filepath"
	"testing"
)

func tempConfigPath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return filepath.Join(dir, "config.yaml")
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Locale != LocaleEN {
		t.Errorf("default Locale = %q, want %q", cfg.Locale, LocaleEN)
	}
	if cfg.ActiveProject != "" {
		t.Errorf("default ActiveProject = %q, want empty", cfg.ActiveProject)
	}
	if cfg.ActiveBoard != 0 {
		t.Errorf("default ActiveBoard = %d, want 0", cfg.ActiveBoard)
	}
}

func TestConfigManager_GetConfig_NoFile(t *testing.T) {
	mgr := NewConfigManagerWithPath(tempConfigPath(t))

	cfg, err := mgr.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	// Should return defaults when file doesn't exist
	if cfg.Locale != LocaleEN {
		t.Errorf("Locale = %q, want %q", cfg.Locale, LocaleEN)
	}
}

func TestConfigManager_SetActiveProject(t *testing.T) {
	mgr := NewConfigManagerWithPath(tempConfigPath(t))

	if err := mgr.SetActiveProject("MYPROJ"); err != nil {
		t.Fatalf("SetActiveProject() error = %v", err)
	}

	cfg, err := mgr.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if cfg.ActiveProject != "MYPROJ" {
		t.Errorf("ActiveProject = %q, want %q", cfg.ActiveProject, "MYPROJ")
	}
}

func TestConfigManager_SetActiveBoard(t *testing.T) {
	mgr := NewConfigManagerWithPath(tempConfigPath(t))

	if err := mgr.SetActiveBoard(42); err != nil {
		t.Fatalf("SetActiveBoard() error = %v", err)
	}

	cfg, err := mgr.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if cfg.ActiveBoard != 42 {
		t.Errorf("ActiveBoard = %d, want 42", cfg.ActiveBoard)
	}
}

func TestConfigManager_SetLocale(t *testing.T) {
	mgr := NewConfigManagerWithPath(tempConfigPath(t))

	if err := mgr.SetLocale(LocaleRU); err != nil {
		t.Fatalf("SetLocale(ru) error = %v", err)
	}

	cfg, err := mgr.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if cfg.Locale != LocaleRU {
		t.Errorf("Locale = %q, want %q", cfg.Locale, LocaleRU)
	}
}

func TestConfigManager_SetLocaleInvalid(t *testing.T) {
	mgr := NewConfigManagerWithPath(tempConfigPath(t))

	err := mgr.SetLocale("fr")
	if err == nil {
		t.Fatal("SetLocale('fr') should fail for unsupported locale")
	}
}

func TestConfigManager_PreservesOtherFields(t *testing.T) {
	mgr := NewConfigManagerWithPath(tempConfigPath(t))

	// Set project first
	if err := mgr.SetActiveProject("PROJ1"); err != nil {
		t.Fatalf("SetActiveProject() error = %v", err)
	}

	// Set board — should not overwrite project
	if err := mgr.SetActiveBoard(99); err != nil {
		t.Fatalf("SetActiveBoard() error = %v", err)
	}

	// Set locale — should not overwrite project or board
	if err := mgr.SetLocale(LocaleRU); err != nil {
		t.Fatalf("SetLocale() error = %v", err)
	}

	cfg, err := mgr.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if cfg.ActiveProject != "PROJ1" {
		t.Errorf("ActiveProject = %q, want %q", cfg.ActiveProject, "PROJ1")
	}
	if cfg.ActiveBoard != 99 {
		t.Errorf("ActiveBoard = %d, want 99", cfg.ActiveBoard)
	}
	if cfg.Locale != LocaleRU {
		t.Errorf("Locale = %q, want %q", cfg.Locale, LocaleRU)
	}
}

func TestConfigManager_CorruptYAML(t *testing.T) {
	path := tempConfigPath(t)
	mgr := NewConfigManagerWithPath(path)

	// Write garbage to config file
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir error = %v", err)
	}
	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0o644); err != nil {
		t.Fatalf("write error = %v", err)
	}

	_, err := mgr.GetConfig()
	if err == nil {
		t.Fatal("GetConfig() should fail with corrupt YAML")
	}
}

func TestConfigManager_YAMLRoundTrip(t *testing.T) {
	path := tempConfigPath(t)
	mgr := NewConfigManagerWithPath(path)

	// Write a full config
	if err := mgr.SetActiveProject("TEST"); err != nil {
		t.Fatalf("SetActiveProject() error = %v", err)
	}
	if err := mgr.SetActiveBoard(7); err != nil {
		t.Fatalf("SetActiveBoard() error = %v", err)
	}
	if err := mgr.SetLocale(LocaleEN); err != nil {
		t.Fatalf("SetLocale() error = %v", err)
	}

	// Read with a fresh manager pointing to the same file
	mgr2 := NewConfigManagerWithPath(path)
	cfg, err := mgr2.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if cfg.ActiveProject != "TEST" {
		t.Errorf("ActiveProject = %q, want %q", cfg.ActiveProject, "TEST")
	}
	if cfg.ActiveBoard != 7 {
		t.Errorf("ActiveBoard = %d, want 7", cfg.ActiveBoard)
	}
	if cfg.Locale != LocaleEN {
		t.Errorf("Locale = %q, want %q", cfg.Locale, LocaleEN)
	}
}

func TestConfigManager_CreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "nested", "config.yaml")
	mgr := NewConfigManagerWithPath(path)

	// Should create intermediate directories
	if err := mgr.SetActiveProject("PROJ"); err != nil {
		t.Fatalf("SetActiveProject() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}
}

func TestConfigManager_OverwriteProject(t *testing.T) {
	mgr := NewConfigManagerWithPath(tempConfigPath(t))

	if err := mgr.SetActiveProject("FIRST"); err != nil {
		t.Fatalf("SetActiveProject() error = %v", err)
	}
	if err := mgr.SetActiveProject("SECOND"); err != nil {
		t.Fatalf("SetActiveProject() error = %v", err)
	}

	cfg, err := mgr.GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if cfg.ActiveProject != "SECOND" {
		t.Errorf("ActiveProject = %q, want %q", cfg.ActiveProject, "SECOND")
	}
}

func TestNewConfigManager(t *testing.T) {
	mgr, err := NewConfigManager()
	if err != nil {
		t.Fatalf("NewConfigManager() error = %v", err)
	}
	if mgr == nil {
		t.Fatal("NewConfigManager() returned nil")
	}
	if mgr.configPath == "" {
		t.Fatal("configPath is empty")
	}
}
