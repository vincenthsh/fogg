package cmd

import (
	"testing"

	"github.com/chanzuckerberg/fogg/config/markers"
	"github.com/stretchr/testify/assert"
)

func TestMergeDependenciesWithOutputs(t *testing.T) {
	tests := []struct {
		name     string
		deps     []markers.Dependency
		inferred map[string][]string
		expected []markers.Dependency
	}{
		{
			name: "no_inference_needed",
			deps: []markers.Dependency{
				{GUID: "guid-1", Output: "vpc_id"},
				{GUID: "guid-1", Output: "vpc_name"},
			},
			inferred: map[string][]string{},
			expected: []markers.Dependency{
				{GUID: "guid-1", Output: "vpc_id"},
				{GUID: "guid-1", Output: "vpc_name"},
			},
		},
		{
			name: "infer_outputs_for_blank",
			deps: []markers.Dependency{
				{GUID: "guid-1"}, // blank output
			},
			inferred: map[string][]string{
				"guid-1": {"vpc_id", "vpc_name"},
			},
			expected: []markers.Dependency{
				{GUID: "guid-1", Output: "vpc_id"},
				{GUID: "guid-1", Output: "vpc_name"},
			},
		},
		{
			name: "merge_explicit_and_inferred",
			deps: []markers.Dependency{
				{GUID: "guid-1", Output: "vpc_id"},
				{GUID: "guid-1"}, // blank - will get inferred outputs
			},
			inferred: map[string][]string{
				"guid-1": {"vpc_name", "subnet_id"},
			},
			expected: []markers.Dependency{
				{GUID: "guid-1", Output: "subnet_id"},
				{GUID: "guid-1", Output: "vpc_id"},
				{GUID: "guid-1", Output: "vpc_name"},
			},
		},
		{
			name: "preserve_input_field",
			deps: []markers.Dependency{
				{GUID: "guid-1", Input: "my_input"},
			},
			inferred: map[string][]string{
				"guid-1": {"output_a", "output_b"},
			},
			expected: []markers.Dependency{
				{GUID: "guid-1", Output: "output_a", Input: "my_input"},
				{GUID: "guid-1", Output: "output_b", Input: "my_input"},
			},
		},
		{
			name: "multiple_guids_with_inference",
			deps: []markers.Dependency{
				{GUID: "guid-1"},
				{GUID: "guid-2", Output: "existing"},
			},
			inferred: map[string][]string{
				"guid-1": {"vpc_id"},
				"guid-2": {"subnet_id"},
			},
			expected: []markers.Dependency{
				{GUID: "guid-1", Output: "vpc_id"},
				{GUID: "guid-2", Output: "existing"},
				{GUID: "guid-2", Output: "subnet_id"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeDependenciesWithOutputs(tt.deps, tt.inferred)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestNeedsOutputInference(t *testing.T) {
	tests := []struct {
		name     string
		deps     []markers.Dependency
		expected bool
	}{
		{
			name: "all_have_outputs",
			deps: []markers.Dependency{
				{GUID: "guid-1", Output: "vpc_id"},
				{GUID: "guid-2", Output: "subnet"},
			},
			expected: false,
		},
		{
			name: "one_missing_output",
			deps: []markers.Dependency{
				{GUID: "guid-1", Output: "vpc_id"},
				{GUID: "guid-2"}, // blank output
			},
			expected: true,
		},
		{
			name: "all_missing_outputs",
			deps: []markers.Dependency{
				{GUID: "guid-1"},
				{GUID: "guid-2"},
			},
			expected: true,
		},
		{
			name:     "empty_deps",
			deps:     []markers.Dependency{},
			expected: false,
		},
		{
			name: "whitespace_only_output",
			deps: []markers.Dependency{
				{GUID: "guid-1", Output: "  "},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := needsOutputInference(tt.deps)
			assert.Equal(t, tt.expected, result)
		})
	}
}
