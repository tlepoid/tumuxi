package config

import (
	"os"
	"path/filepath"
)

// Paths holds all the file system paths used by the application
type Paths struct {
	Home           string // ~/.tumuxi
	WorkspacesRoot string // ~/.tumuxi/workspaces
	RegistryPath   string // ~/.tumuxi/projects.json
	MetadataRoot   string // ~/.tumuxi/workspaces-metadata
	ConfigPath     string // ~/.tumuxi/config.json
}

// DefaultPaths returns the default paths configuration
func DefaultPaths() (*Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	tumuxiHome := filepath.Join(home, ".tumuxi")

	return &Paths{
		Home:           tumuxiHome,
		WorkspacesRoot: filepath.Join(tumuxiHome, "workspaces"),
		RegistryPath:   filepath.Join(tumuxiHome, "projects.json"),
		MetadataRoot:   filepath.Join(tumuxiHome, "workspaces-metadata"),
		ConfigPath:     filepath.Join(tumuxiHome, "config.json"),
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
