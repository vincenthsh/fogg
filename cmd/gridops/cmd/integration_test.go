package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"connectrpc.com/connect"
	"github.com/chanzuckerberg/fogg/util"
	"github.com/stretchr/testify/require"
	"github.com/terraconstructs/grid/pkg/sdk"
)

var updateSnapshots = flag.Bool("update", false, "update API call snapshots")

type integrationTestCase struct {
	name           string
	testdataDir    string
	existingStates []seedState
	existingDeps   []sdk.DependencyEdge
}

type seedState struct {
	guid    string
	logicID string
	labels  sdk.LabelMap
}

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

// TestGridOpsIntegration tests the sync command with a fake Grid API client and seed data
func TestGridOpsIntegration(t *testing.T) {
	testCases := []integrationTestCase{
		{
			name:        "v2_grid_inference_happy_path",
			testdataDir: "v2_grid_inference",
		},
		{
			name:        "v2_grid_inference_existing_state",
			testdataDir: "v2_grid_inference",
			existingStates: []seedState{
				{
					guid:    "11111111-2222-3333-4444-555555555555",
					logicID: "proj-test-vpc",
					labels: sdk.LabelMap{
						"managed_by": "legacy",
						"legacy":     "true",
					},
				},
				{
					guid:    "019aa5fe-90bc-721b-8e3d-dde5a72aba70",
					logicID: "proj-test-existing",
					labels: sdk.LabelMap{
						"managed_by": "legacy",
					},
				},
			},
			existingDeps: []sdk.DependencyEdge{
				{
					ID: 1,
					From: sdk.StateReference{
						GUID: "11111111-2222-3333-4444-555555555555",
					},
					FromOutput: "vpc_id",
					To: sdk.StateReference{
						GUID: "019aa5fe-90bc-721b-8e3d-dde5a72aba70",
					},
					ToInputName: "",
				},
				{
					ID: 2,
					From: sdk.StateReference{
						GUID: "11111111-2222-3333-4444-555555555555",
					},
					FromOutput: "vpc_name",
					To: sdk.StateReference{
						GUID: "019aa5fe-90bc-721b-8e3d-dde5a72aba70",
					},
					ToInputName: "",
				},
				{
					ID: 3,
					From: sdk.StateReference{
						GUID: "deadbeef-dead-beef-dead-beefdeadbeef",
					},
					FromOutput: "unused_output",
					To: sdk.StateReference{
						GUID: "019aa5fe-90bc-721b-8e3d-dde5a72aba70",
					},
					ToInputName: "",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			projectRoot := util.ProjectRoot()
			testdataPath := filepath.Join(projectRoot, "testdata", tc.testdataDir)
			snapshotPath := filepath.Join(projectRoot, "cmd", "gridops", "testdata_snapshots", tc.name+".json")
			stdoutSnapshotPath := filepath.Join(projectRoot, "cmd", "gridops", "testdata_snapshots", tc.name+"-stdout.txt")
			// ensure snapshot Path directory exists
			err := os.MkdirAll(filepath.Dir(snapshotPath), 0o755)
			r.NoError(err, "failed to create snapshot directory")
			terraformPath := filepath.Join(testdataPath, "terraform")

			_, err = os.Stat(terraformPath)
			r.NoError(err, "testdata directory should exist (run 'make update-golden-files' first)")

			var capturedCalls []APICall
			fakeClient := newFakeGridClient(&capturedCalls)
			seedFakeGridClient(fakeClient, tc)

			originalFactory := gridClientFactory
			gridClientFactory = func(ctx context.Context, cfg sessionConfig) (gridAPIClient, error) {
				return fakeClient, nil
			}
			defer func() { gridClientFactory = originalFactory }()

			origOpts := opts
			defer func() { opts = origOpts }()

			stdout, stderr, err := runGridOpsCommand(t, testdataPath, "sync", "--server", "https://grid.mock")
			t.Logf("stdout:\n%s", stdout)
			t.Logf("stderr:\n%s", stderr)
			r.NoError(err)

			snapshot := &APICallSnapshot{Calls: capturedCalls}

			if *updateSnapshots {
				r.NoError(saveSnapshot(snapshotPath, snapshot))
				r.NoError(saveStdoutSnapshot(stdoutSnapshotPath, stdout))
			} else {
				expected, loadErr := loadSnapshot(snapshotPath)
				if errors.Is(loadErr, os.ErrNotExist) {
					t.Fatalf("snapshot %s missing; run 'make update-gridops-snapshots'", snapshotPath)
				}
				r.NoError(loadErr)
				r.True(compareSnapshots(t, expected, snapshot))

				expectedStdout, stdoutErr := loadStdoutSnapshot(stdoutSnapshotPath)
				if errors.Is(stdoutErr, os.ErrNotExist) {
					t.Fatalf("stdout snapshot %s missing; run 'make update-gridops-snapshots'", stdoutSnapshotPath)
				}
				r.NoError(stdoutErr)
				r.Equal(expectedStdout, stdout)
			}
		})
	}
}

func seedFakeGridClient(client *fakeGridClient, tc integrationTestCase) {
	for _, state := range tc.existingStates {
		client.states[state.guid] = &fakeState{
			guid:    state.guid,
			logicID: state.logicID,
			labels:  copyLabelMap(state.labels),
		}
	}

	for _, dep := range tc.existingDeps {
		edge := dep
		if edge.ID == 0 {
			edge.ID = client.nextEdgeID
			client.nextEdgeID++
		}
		if edge.ID >= client.nextEdgeID {
			client.nextEdgeID = edge.ID + 1
		}
		client.deps[edge.ID] = edge
	}
}

func runGridOpsCommand(t *testing.T, workdir string, args ...string) (string, string, error) {
	t.Helper()

	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(workdir))
	defer os.Chdir(originalWd)

	rootCmd.SetArgs(args)

	stdout, stderr, execErr := captureStdoutStderr(func() error {
		return rootCmd.Execute()
	})

	rootCmd.SetArgs(nil)
	return stdout, stderr, execErr
}

func captureStdoutStderr(fn func() error) (string, string, error) {
	originalStdout := os.Stdout
	originalStderr := os.Stderr

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		return "", "", err
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		stdoutWriter.Close()
		stdoutReader.Close()
		return "", "", err
	}

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	execErr := fn()

	stdoutWriter.Close()
	stderrWriter.Close()

	os.Stdout = originalStdout
	os.Stderr = originalStderr

	stdoutBytes, readOutErr := io.ReadAll(stdoutReader)
	stdoutReader.Close()
	stderrBytes, readErrErr := io.ReadAll(stderrReader)
	stderrReader.Close()

	if readOutErr != nil {
		return "", "", readOutErr
	}
	if readErrErr != nil {
		return "", "", readErrErr
	}

	return string(stdoutBytes), string(stderrBytes), execErr
}

// normalizeSnapshot sorts calls for deterministic comparison
func normalizeSnapshot(snapshot *APICallSnapshot) {
	sort.Slice(snapshot.Calls, func(i, j int) bool {
		if snapshot.Calls[i].Method != snapshot.Calls[j].Method {
			return snapshot.Calls[i].Method < snapshot.Calls[j].Method
		}
		if snapshot.Calls[i].Path != snapshot.Calls[j].Path {
			return snapshot.Calls[i].Path < snapshot.Calls[j].Path
		}
		return fmt.Sprint(snapshot.Calls[i].Body) < fmt.Sprint(snapshot.Calls[j].Body)
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

	return os.WriteFile(path, data, 0o644)
}

func loadStdoutSnapshot(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func saveStdoutSnapshot(path string, stdout string) error {
	return os.WriteFile(path, []byte(stdout), 0o644)
}

type fakeGridClient struct {
	calls      *[]APICall
	states     map[string]*fakeState
	deps       map[int64]sdk.DependencyEdge
	nextEdgeID int64
}

type fakeState struct {
	guid    string
	logicID string
	labels  sdk.LabelMap
}

func newFakeGridClient(calls *[]APICall) *fakeGridClient {
	return &fakeGridClient{
		calls:      calls,
		states:     make(map[string]*fakeState),
		deps:       make(map[int64]sdk.DependencyEdge),
		nextEdgeID: 1,
	}
}

func (c *fakeGridClient) recordCall(method, path string, body map[string]interface{}) {
	if c.calls == nil {
		return
	}
	*c.calls = append(*c.calls, APICall{Method: method, Path: path, Body: body})
}

func (c *fakeGridClient) GetStateInfo(_ context.Context, ref sdk.StateReference) (*sdk.StateInfo, error) {
	c.recordCall("POST", "/state.v1.StateService/GetStateInfo", map[string]interface{}{
		"guid":    ref.GUID,
		"logicId": ref.LogicID,
	})

	state, ok := c.states[ref.GUID]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("state %s not found", ref.GUID))
	}

	return &sdk.StateInfo{
		State: sdk.StateReference{
			GUID:    state.guid,
			LogicID: state.logicID,
		},
		Labels:       copyLabelMap(state.labels),
		Dependencies: c.dependenciesFor(state.guid),
	}, nil
}

func (c *fakeGridClient) CreateState(_ context.Context, input sdk.CreateStateInput) (*sdk.State, error) {
	c.recordCall("POST", "/state.v1.StateService/CreateState", map[string]interface{}{
		"guid":    input.GUID,
		"logicId": input.LogicID,
		"labels":  input.Labels,
	})

	if input.GUID == "" {
		return nil, fmt.Errorf("guid is required")
	}

	if _, exists := c.states[input.GUID]; exists {
		return nil, fmt.Errorf("state %s already exists", input.GUID)
	}

	c.states[input.GUID] = &fakeState{
		guid:    input.GUID,
		logicID: input.LogicID,
		labels:  copyLabelMap(input.Labels),
	}

	return &sdk.State{GUID: input.GUID, LogicID: input.LogicID}, nil
}

func (c *fakeGridClient) UpdateStateLabels(_ context.Context, input sdk.UpdateStateLabelsInput) (*sdk.UpdateStateLabelsResult, error) {
	c.recordCall("POST", "/state.v1.StateService/UpdateStateLabels", map[string]interface{}{
		"stateId":  input.StateID,
		"adds":     input.Adds,
		"removals": input.Removals,
	})

	state, ok := c.states[input.StateID]
	if !ok {
		return nil, fmt.Errorf("state %s not found", input.StateID)
	}

	if state.labels == nil {
		state.labels = sdk.LabelMap{}
	}

	for k, v := range input.Adds {
		state.labels[k] = v
	}
	for _, key := range input.Removals {
		delete(state.labels, key)
	}

	return &sdk.UpdateStateLabelsResult{
		StateID: state.guid,
		Labels:  copyLabelMap(state.labels),
	}, nil
}

func (c *fakeGridClient) AddDependency(_ context.Context, input sdk.AddDependencyInput) (*sdk.AddDependencyResult, error) {
	c.recordCall("POST", "/state.v1.StateService/AddDependency", map[string]interface{}{
		"from": map[string]interface{}{
			"guid":    input.From.GUID,
			"logicId": input.From.LogicID,
		},
		"fromOutput": input.FromOutput,
		"to": map[string]interface{}{
			"guid":    input.To.GUID,
			"logicId": input.To.LogicID,
		},
		"toInputName": input.ToInputName,
	})

	if input.To.GUID == "" {
		return nil, fmt.Errorf("to guid is required")
	}

	key := depKey(input.From.GUID, input.FromOutput)
	for _, edge := range c.deps {
		if edge.To.GUID == input.To.GUID && depKey(edge.From.GUID, edge.FromOutput) == key && edge.ToInputName == input.ToInputName {
			return &sdk.AddDependencyResult{Edge: edge, AlreadyExists: true}, nil
		}
	}

	edge := sdk.DependencyEdge{
		ID:         c.nextEdgeID,
		From:       input.From,
		FromOutput: input.FromOutput,
		To: sdk.StateReference{
			GUID:    input.To.GUID,
			LogicID: input.To.LogicID,
		},
		ToInputName: input.ToInputName,
	}
	c.nextEdgeID++
	c.deps[edge.ID] = edge

	return &sdk.AddDependencyResult{Edge: edge}, nil
}

func (c *fakeGridClient) RemoveDependency(_ context.Context, edgeID int64) error {
	c.recordCall("POST", "/state.v1.StateService/RemoveDependency", map[string]interface{}{
		"edgeId": edgeID,
	})

	if _, ok := c.deps[edgeID]; !ok {
		return fmt.Errorf("edge %d not found", edgeID)
	}

	delete(c.deps, edgeID)
	return nil
}

func (c *fakeGridClient) dependenciesFor(toGUID string) []sdk.DependencyEdge {
	var edges []sdk.DependencyEdge
	for _, edge := range c.deps {
		if edge.To.GUID == toGUID {
			edges = append(edges, edge)
		}
	}
	return edges
}

func copyLabelMap(src sdk.LabelMap) sdk.LabelMap {
	if len(src) == 0 {
		return nil
	}
	dst := make(sdk.LabelMap, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
