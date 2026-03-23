package config

import (
	"os"
	"path/filepath"
)

// Paths holds all the file system paths used by the application
type Paths struct {
	Home           string // ~/.tumux
	WorkspacesRoot string // ~/.tumux/workspaces
	RegistryPath   string // ~/.tumux/projects.json
	MetadataRoot   string // ~/.tumux/workspaces-metadata
	ConfigPath     string // ~/.tumux/config.json
}

// DefaultPaths returns the default paths configuration
func DefaultPaths() (*Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	tumuxHome := filepath.Join(home, ".tumux")

	return &Paths{
		Home:           tumuxHome,
		WorkspacesRoot: filepath.Join(tumuxHome, "workspaces"),
		RegistryPath:   filepath.Join(tumuxHome, "projects.json"),
		MetadataRoot:   filepath.Join(tumuxHome, "workspaces-metadata"),
		ConfigPath:     filepath.Join(tumuxHome, "config.json"),
	}, nil
}

// EnsureDirectories creates all required directories if they don't exist
func (p *Paths) EnsureDirectories() error {
	dirs := []string{
		p.Home,
		p.WorkspacesRoot,
		p.MetadataRoot,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return nil
}
