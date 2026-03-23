package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathsEnsureDirectories(t *testing.T) {
	tmp := t.TempDir()
	paths := &Paths{
		Home:           filepath.Join(tmp, "tumux"),
		WorkspacesRoot: filepath.Join(tmp, "tumux", "workspaces"),
		RegistryPath:   filepath.Join(tmp, "tumux", "projects.json"),
		MetadataRoot:   filepath.Join(tmp, "tumux", "workspaces-metadata"),
		ConfigPath:     filepath.Join(tmp, "tumux", "config.json"),
	}

	if err := paths.EnsureDirectories(); err != nil {
		t.Fatalf("EnsureDirectories() error = %v", err)
	}

	for _, dir := range []string{paths.Home, paths.WorkspacesRoot, paths.MetadataRoot} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected directory %s to exist: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", dir)
		}
	}
}
