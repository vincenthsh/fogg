package cmd

// Integration tests for gridops that validate the fogg → gridops workflow.
//
// These tests use the golden test data from testdata/ to ensure that:
// 1. Markers are correctly scanned and validated
// 2. Dependencies are properly inferred from Terraform files
// 3. (Future) API calls to Grid match expected snapshots
//
// Usage:
//   make test-gridops-integration          # Run integration tests
//   make update-gridops-snapshots          # Update API call snapshots
//
// Note: These tests require golden files to be present. Run `make update-golden-files` first
// if the testdata directories are empty.

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/chanzuckerberg/fogg/util"
	"github.com/stretchr/testify/require"
)

var updateSnapshots = flag.Bool("update", false, "update API call snapshots")

// APICall represents a captured API call for snapshotting
type APICall struct {
	Method string                 `json:"method"`
	Path   string                 `json:"path"`
	Body   map[string]interface{} `json:"body,omitempty"`
}

// APICallSnapshot is the snapshot of all API calls
type APICallSnapshot struct {
	Calls []APICall `json:"calls"`
}

// mockGridServer creates an httptest server that mocks the Grid API
// and captures all API calls
func mockGridServer(t *testing.T, calls *[]APICall) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read and parse body if present
		var bodyMap map[string]interface{}
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			r.Body.Close()

			if len(bodyBytes) > 0 {
				err = json.Unmarshal(bodyBytes, &bodyMap)
				if err != nil {
					// If it's not JSON, store as string
					bodyMap = map[string]interface{}{
						"_raw": string(bodyBytes),
					}
				}
			}
		}

		// Capture the call
		*calls = append(*calls, APICall{
			Method: r.Method,
			Path:   r.URL.Path,
			Body:   bodyMap,
		})

		// Mock responses based on path patterns
		switch {
		case r.URL.Path == "/.well-known/grid/auth":
			// Auth discovery - return auth disabled for simplicity
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "auth not configured",
			})

		case strings.HasPrefix(r.URL.Path, "/api/v1/states/") && r.Method == "GET":
			// GetStateInfo - return 404 to simulate state doesn't exist yet
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "state not found",
			})

		case r.URL.Path == "/api/v1/states" && r.Method == "POST":
			// CreateState - return success
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"guid":      bodyMap["guid"],
				"logicalId": bodyMap["logicalId"],
			})

		case strings.HasPrefix(r.URL.Path, "/api/v1/states/") && strings.HasSuffix(r.URL.Path, "/labels") && r.Method == "PATCH":
			// UpdateStateLabels - return success
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "ok",
			})

		case strings.HasPrefix(r.URL.Path, "/api/v1/dependencies") && r.Method == "POST":
			// AddDependency - return success
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":   1,
				"from": bodyMap["from"],
				"to":   bodyMap["to"],
			})

		case strings.HasPrefix(r.URL.Path, "/api/v1/dependencies/") && r.Method == "DELETE":
			// RemoveDependency - return success
			w.WriteHeader(http.StatusNoContent)

		default:
			t.Logf("Unhandled request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))
}

func TestGridOpsIntegration(t *testing.T) {
	testCases := []struct {
		name        string
		testdataDir string
	}{
		{
			name:        "v2_grid_inference",
			testdataDir: "v2_grid_inference",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			// Set up paths
			projectRoot := util.ProjectRoot()
			testdataPath := filepath.Join(projectRoot, "testdata", tc.testdataDir)
			snapshotPath := filepath.Join(testdataPath, ".gridops-snapshot.json")
			terraformPath := filepath.Join(testdataPath, "terraform")

			// Verify testdata exists
			_, err := os.Stat(terraformPath)
			r.NoError(err, "testdata directory should exist (run 'make update-golden-files' first)")

			// Set up mock Grid API server
			var capturedCalls []APICall
			mockServer := mockGridServer(t, &capturedCalls)
			defer mockServer.Close()

			// Save original working directory
			originalWd, err := os.Getwd()
			r.NoError(err)
			defer os.Chdir(originalWd)

			// Change to terraform directory
			err = os.Chdir(terraformPath)
			r.NoError(err)

			// Override global options for the test
			opts.serverURL = mockServer.URL
			opts.clientID = ""
			opts.clientSecret = ""

			// Run sync command
			ctx := context.Background()
			markers, err := ScanMarkers(".", excludeDirs)
			r.NoError(err)
			r.NotEmpty(markers, "should find at least one marker")

			// Validate markers
			issues := ValidateMarkers(markers)
			r.Empty(issues, "markers should be valid")

			// Create a mock client without auth (since our mock doesn't need it)
			session := sessionConfig{
				ServerURL:    mockServer.URL,
				ClientID:     "",
				ClientSecret: "",
			}

			// We can't use newGridClient since it tries to discover auth
			// Instead, we'll just verify markers were scanned correctly
			// The actual sync would happen through the SDK

			// For now, just verify the markers are correct
			t.Logf("Found %d markers", len(markers))
			for _, m := range markers {
				t.Logf("  - %s (GUID: %s)", m.Marker.LogicalID, m.Marker.GUID)
				if len(m.Marker.Dependencies) > 0 {
					t.Logf("    Dependencies:")
					for _, dep := range m.Marker.Dependencies {
						t.Logf("      - GUID: %s, Output: %s", dep.GUID, dep.Output)
					}
				}
			}

			// Verify expected markers
			r.Len(markers, 2, "should find 2 markers (vpc and existing)")

			// Find vpc and existing markers
			var vpcMarker, existingMarker *LoadedMarker
			for i := range markers {
				if markers[i].Marker.LogicalID == "proj-test-vpc" {
					vpcMarker = &markers[i]
				} else if markers[i].Marker.LogicalID == "proj-test-existing" {
					existingMarker = &markers[i]
				}
			}

			r.NotNil(vpcMarker, "should find vpc marker")
			r.NotNil(existingMarker, "should find existing marker")

			// Verify vpc has no dependencies
			r.Empty(vpcMarker.Marker.Dependencies, "vpc should have no dependencies")

			// Verify existing has dependency on vpc
			r.Len(existingMarker.Marker.Dependencies, 1, "existing should have 1 dependency")
			r.Equal("11111111-2222-3333-4444-555555555555", existingMarker.Marker.Dependencies[0].GUID)

			// Test dependency resolution (inference)
			resolvedDeps, err := resolveMarkerDependencies(*existingMarker)
			r.NoError(err, "should resolve dependencies")
			r.NotEmpty(resolvedDeps, "should infer at least one dependency with output")

			// Verify inferred output is vpc_name
			foundVPCName := false
			for _, dep := range resolvedDeps {
				if dep.GUID == "11111111-2222-3333-4444-555555555555" && dep.Output == "vpc_name" {
					foundVPCName = true
					break
				}
			}
			r.True(foundVPCName, "should infer vpc_name output from test.tf.json")

			t.Logf("Successfully validated dependency inference:")
			for _, dep := range resolvedDeps {
				t.Logf("  - GUID: %s -> Output: %s", dep.GUID, dep.Output)
			}

			// Note: We're not actually calling the sync command here because:
			// 1. The mock server needs proper Grid API protocol implementation
			// 2. The SDK client has its own auth requirements
			// 3. For now, we're testing the core logic: scanning, validation, and inference

			// If updateSnapshots is true, this is where we'd save the snapshot
			// For now, we'll skip the actual API call snapshotting since we need
			// a more complete mock implementation

			_ = session // unused for now
			_ = snapshotPath
			_ = capturedCalls
		})
	}
}

// normalizeSnapshot sorts calls for deterministic comparison
func normalizeSnapshot(snapshot *APICallSnapshot) {
	sort.Slice(snapshot.Calls, func(i, j int) bool {
		if snapshot.Calls[i].Method != snapshot.Calls[j].Method {
			return snapshot.Calls[i].Method < snapshot.Calls[j].Method
		}
		return snapshot.Calls[i].Path < snapshot.Calls[j].Path
	})
}

// compareSnapshots compares two snapshots and returns differences
func compareSnapshots(t *testing.T, expected, actual *APICallSnapshot) bool {
	normalizeSnapshot(expected)
	normalizeSnapshot(actual)

	expectedJSON, err := json.MarshalIndent(expected, "", "  ")
	require.NoError(t, err)

	actualJSON, err := json.MarshalIndent(actual, "", "  ")
	require.NoError(t, err)

	if !bytes.Equal(expectedJSON, actualJSON) {
		t.Errorf("API call snapshots differ:\nExpected:\n%s\n\nActual:\n%s\n",
			string(expectedJSON), string(actualJSON))
		return false
	}

	return true
}

func loadSnapshot(path string) (*APICallSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var snapshot APICallSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func saveSnapshot(path string, snapshot *APICallSnapshot) error {
	normalizeSnapshot(snapshot)

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
