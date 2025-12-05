package markers

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// Marker represents the content of a .grid-state.yaml file.
type Marker struct {
	GUID         string            `yaml:"guid"`
	LogicalID    string            `yaml:"logicalId"`
	Labels       map[string]string `yaml:"labels,omitempty"`
	Dependencies []Dependency      `yaml:"dependencies,omitempty"`
}

type Dependency struct {
	GUID   string `yaml:"guid"`
	Output string `yaml:"output,omitempty"`
	Input  string `yaml:"input,omitempty"`
}

// LoadMarker reads a Marker from the specified path.
func LoadMarker(path string) (*Marker, error) {
	return LoadMarkerFS(afero.NewOsFs(), path)
}

// LoadMarkerFS reads a Marker from the specified path using the provided filesystem.
func LoadMarkerFS(fs afero.Fs, path string) (*Marker, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read marker file: %w", err)
	}

	var marker Marker
	if err := yaml.Unmarshal(data, &marker); err != nil {
		return nil, fmt.Errorf("failed to unmarshal marker: %w", err)
	}

	return &marker, nil
}

// SaveMarker writes a Marker to the specified path.
// It creates the directory if it doesn't exist.
func SaveMarker(path string, marker *Marker) error {
	return SaveMarkerFS(afero.NewOsFs(), path, marker)
}

// SaveMarkerFS writes a Marker to the specified path using the provided filesystem.
// It creates the directory if it doesn't exist.
func SaveMarkerFS(fs afero.Fs, path string, marker *Marker) error {
	if err := fs.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := yaml.Marshal(marker)
	if err != nil {
		return fmt.Errorf("failed to marshal marker: %w", err)
	}

	if err := afero.WriteFile(fs, path, data, 0644); err != nil {
		return fmt.Errorf("failed to write marker file: %w", err)
	}

	return nil
}
