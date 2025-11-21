package cmd

import (
	"path/filepath"
	"testing"

	"github.com/chanzuckerberg/fogg/config/markers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixtureDir(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("testdata", name)
}

func TestResolveMarkerDependencies(t *testing.T) {
	tests := []struct {
		name         string
		fixture      string
		inputDeps    []markers.Dependency
		expectedDeps []markers.Dependency
	}{
		{
			name:    "single_hcl",
			fixture: "hcl1",
			inputDeps: []markers.Dependency{
				{GUID: "11111111-1111-1111-1111-111111111111"},
			},
			expectedDeps: []markers.Dependency{
				{GUID: "11111111-1111-1111-1111-111111111111", Output: "found_output"},
			},
		},
		{
			name:    "split_hcl",
			fixture: "hcl2",
			inputDeps: []markers.Dependency{
				{GUID: "11111111-1111-1111-1111-111111111111"},
				{GUID: "22222222-2222-2222-2222-222222222222"},
			},
			expectedDeps: []markers.Dependency{
				{GUID: "11111111-1111-1111-1111-111111111111", Output: "name"},
				{GUID: "11111111-1111-1111-1111-111111111111", Output: "vpc_id"},
				{GUID: "22222222-2222-2222-2222-222222222222", Output: "public_subnets"},
			},
		},
		{
			name:    "cdktf",
			fixture: "cdktf",
			inputDeps: []markers.Dependency{
				{GUID: "33333333-3333-3333-3333-333333333333"},
				{GUID: "44444444-4444-4444-4444-444444444444"},
			},
			expectedDeps: []markers.Dependency{
				{GUID: "33333333-3333-3333-3333-333333333333", Output: "ami_id"},
				{GUID: "44444444-4444-4444-4444-444444444444", Output: "vpc_name"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := fixtureDir(t, tt.fixture)

			deps, err := resolveMarkerDependencies(LoadedMarker{
				Path: filepath.Join(dir, ".grid-state.yaml"),
				Marker: &markers.Marker{
					Dependencies: tt.inputDeps,
				},
			})
			require.NoError(t, err)

			require.Len(t, deps, len(tt.expectedDeps))
			assert.ElementsMatch(t, tt.expectedDeps, deps)
		})
	}
}
