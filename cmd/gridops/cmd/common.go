package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chanzuckerberg/fogg/config/markers"
)

// LoadedMarker contains the marker data and its source path.
type LoadedMarker struct {
	Path   string
	Marker *markers.Marker
}

// ScanMarkers finds all .grid-state.yaml files in the current directory and subdirectories,
// skipping directories that match any of the exclusion patterns.
func ScanMarkers(root string, excludePatterns []string) ([]LoadedMarker, error) {
	var results []LoadedMarker

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if d.IsDir() && shouldExclude(d.Name(), excludePatterns) {
			return filepath.SkipDir
		}

		// Check if this is a marker file
		if !d.IsDir() && d.Name() == ".grid-state.yaml" {
			marker, err := markers.LoadMarker(path)
			if err != nil {
				return fmt.Errorf("failed to load marker at %s: %w", path, err)
			}
			results = append(results, LoadedMarker{
				Path:   path,
				Marker: marker,
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan markers: %w", err)
	}

	return results, nil
}

// shouldExclude returns true if the directory name matches any exclusion pattern.
func shouldExclude(dirName string, patterns []string) bool {
	for _, pattern := range patterns {
		// Use simple glob matching
		matched, err := filepath.Match(pattern, dirName)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// ValidateMarkers checks for required fields and conflicts.
// Returns a slice of human-readable issues. Empty slice means no issues found.
func ValidateMarkers(markers []LoadedMarker) []string {
	var issues []string
	guidMap := make(map[string][]LoadedMarker)
	logicIDMap := make(map[string][]LoadedMarker)

	for _, m := range markers {
		if strings.TrimSpace(m.Marker.GUID) == "" {
			issues = append(issues, fmt.Sprintf("%s is missing guid", m.Path))
		}
		if strings.TrimSpace(m.Marker.LogicalID) == "" {
			issues = append(issues, fmt.Sprintf("%s is missing logicalId", m.Path))
		}

		guidMap[m.Marker.GUID] = append(guidMap[m.Marker.GUID], m)
		logicIDMap[m.Marker.LogicalID] = append(logicIDMap[m.Marker.LogicalID], m)

		for i, dep := range m.Marker.Dependencies {
			if strings.TrimSpace(dep.GUID) == "" {
				issues = append(issues, fmt.Sprintf("%s dependency[%d] is missing guid", m.Path, i))
			}
			// Note: Missing outputs are OK - they will be inferred from Terraform files
		}
	}

	for guid, entries := range guidMap {
		if guid == "" {
			continue
		}
		if len(entries) > 1 {
			issues = append(issues, fmt.Sprintf("duplicate guid '%s' in %s", guid, joinPaths(entries)))
		}
	}

	for logicID, entries := range logicIDMap {
		if logicID == "" {
			continue
		}
		if len(entries) > 1 {
			issues = append(issues, fmt.Sprintf("duplicate logicalId '%s' in %s", logicID, joinPaths(entries)))
		}
	}

	return issues
}

func joinPaths(markers []LoadedMarker) string {
	paths := make([]string, 0, len(markers))
	for _, m := range markers {
		paths = append(paths, m.Path)
	}
	return strings.Join(paths, ", ")
}
