package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/chanzuckerberg/fogg/config/markers"
	"github.com/chanzuckerberg/fogg/exp/tfscan"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

var uuidPattern = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

func resolveMarkerDependencies(marker LoadedMarker) ([]markers.Dependency, error) {
	deps := marker.Marker.Dependencies
	if !needsOutputInference(deps) {
		return deps, nil
	}

	dir := filepath.Dir(marker.Path)
	stateRefs, err := tfscan.Scan(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to scan terraform outputs in %s: %w", dir, err)
	}

	nameToGUID, err := mapRemoteStateGUIDs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to map remote state aliases to GUIDs in %s: %w", dir, err)
	}

	inferred := make(map[string][]string)
	for name, outputs := range stateRefs {
		guid := nameToGUID[name]
		if guid == "" {
			continue
		}
		outputList := make([]string, 0, len(outputs))
		for out := range outputs {
			outputList = append(outputList, out)
		}
		sort.Strings(outputList)
		inferred[guid] = appendUnique(inferred[guid], outputList...)
	}

	if len(inferred) == 0 {
		return deps, nil
	}

	return mergeDependenciesWithOutputs(deps, inferred), nil
}

func needsOutputInference(deps []markers.Dependency) bool {
	for _, dep := range deps {
		if strings.TrimSpace(dep.Output) == "" {
			return true
		}
	}
	return false
}

func mergeDependenciesWithOutputs(deps []markers.Dependency, inferred map[string][]string) []markers.Dependency {
	type depGroup struct {
		base     markers.Dependency
		outputs  []string
		hasBlank bool
	}

	groups := make(map[string]*depGroup)
	order := make([]string, 0)

	getGroup := func(dep markers.Dependency) *depGroup {
		key := strings.TrimSpace(dep.GUID) + "|" + dep.Input
		if g, ok := groups[key]; ok {
			return g
		}
		g := &depGroup{base: markers.Dependency{GUID: strings.TrimSpace(dep.GUID), Input: dep.Input}}
		groups[key] = g
		order = append(order, key)
		return g
	}

	for _, dep := range deps {
		g := getGroup(dep)
		if strings.TrimSpace(dep.Output) == "" {
			g.hasBlank = true
			continue
		}
		if !contains(g.outputs, dep.Output) {
			g.outputs = append(g.outputs, dep.Output)
		}
	}

	for _, g := range groups {
		inferredOutputs := inferred[g.base.GUID]
		for _, out := range inferredOutputs {
			if !contains(g.outputs, out) {
				g.outputs = append(g.outputs, out)
			}
		}
	}

	result := make([]markers.Dependency, 0, len(deps))
	for _, key := range order {
		group := groups[key]
		if len(group.outputs) == 0 {
			if group.hasBlank {
				result = append(result, group.base)
			}
			continue
		}
		sort.Strings(group.outputs)
		for _, out := range group.outputs {
			result = append(result, markers.Dependency{
				GUID:   group.base.GUID,
				Output: out,
				Input:  group.base.Input,
			})
		}
	}

	return result
}

func mapRemoteStateGUIDs(dir string) (map[string]string, error) {
	result := make(map[string]string)
	parser := hclparse.NewParser()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		switch {
		case strings.HasSuffix(entry.Name(), ".tf"):
			if err := collectRemoteStateGUIDsFromHCL(parser, path, result); err != nil {
				log.Printf("warning: failed to parse %s: %v", entry.Name(), err)
			}
		case strings.HasSuffix(entry.Name(), ".tf.json"):
			if err := collectRemoteStateGUIDsFromJSON(path, result); err != nil {
				log.Printf("warning: failed to parse %s: %v", entry.Name(), err)
			}
		}
	}

	return result, nil
}

func collectRemoteStateGUIDsFromHCL(parser *hclparse.Parser, path string, result map[string]string) error {
	file, diags := parser.ParseHCLFile(path)
	if diags.HasErrors() {
		return fmt.Errorf("%s", diags.Error())
	}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return nil
	}

	for _, block := range body.Blocks {
		if block.Type != "data" || len(block.Labels) < 2 || block.Labels[0] != "terraform_remote_state" {
			continue
		}

		configAttr, exists := block.Body.Attributes["config"]
		if !exists {
			continue
		}

		val, diags := configAttr.Expr.Value(nil)
		if diags.HasErrors() {
			continue
		}

		guid := extractGUIDFromCTY(val)
		if guid != "" {
			result[block.Labels[1]] = guid
		}
	}

	return nil
}

func collectRemoteStateGUIDsFromJSON(path string, result map[string]string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	rawData, ok := data["data"].(map[string]interface{})
	if !ok {
		return nil
	}

	rawRemoteState, ok := rawData["terraform_remote_state"].(map[string]interface{})
	if !ok {
		return nil
	}

	for name, block := range rawRemoteState {
		blockMap, ok := block.(map[string]interface{})
		if !ok {
			continue
		}

		config := blockMap["config"]
		guid := extractGUIDFromInterface(config)
		if guid != "" {
			result[name] = guid
		}
	}

	return nil
}

func extractGUIDFromCTY(val cty.Value) string {
	if !val.IsWhollyKnown() || val.IsNull() {
		return ""
	}

	if val.Type().Equals(cty.String) {
		return uuidPattern.FindString(val.AsString())
	}

	if val.Type().IsObjectType() || val.Type().IsMapType() {
		for _, v := range val.AsValueMap() {
			if guid := extractGUIDFromCTY(v); guid != "" {
				return guid
			}
		}
	}

	if val.Type().IsCollectionType() || val.Type().IsTupleType() {
		for it := val.ElementIterator(); it.Next(); {
			_, v := it.Element()
			if guid := extractGUIDFromCTY(v); guid != "" {
				return guid
			}
		}
	}

	return ""
}

func extractGUIDFromInterface(input interface{}) string {
	switch v := input.(type) {
	case string:
		return uuidPattern.FindString(v)
	case map[string]interface{}:
		for _, value := range v {
			if guid := extractGUIDFromInterface(value); guid != "" {
				return guid
			}
		}
	case []interface{}:
		for _, value := range v {
			if guid := extractGUIDFromInterface(value); guid != "" {
				return guid
			}
		}
	}
	return ""
}

func appendUnique(existing []string, items ...string) []string {
	for _, item := range items {
		if !contains(existing, item) {
			existing = append(existing, item)
		}
	}
	return existing
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
