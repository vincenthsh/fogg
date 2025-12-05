package tfscan

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

// StateRefs maps a remote state name to a set of accessed output attributes
type StateRefs map[string]map[string]struct{}

// Tracker holds the results and analysis logic
type Tracker struct {
	refs StateRefs
}

func NewTracker() *Tracker {
	return &Tracker{
		refs: make(StateRefs),
	}
}

// Scan scans a directory for remote state references
func Scan(dir string) (StateRefs, error) {
	tracker := NewTracker()

	if err := scanDirectory(dir, tracker); err != nil {
		return nil, err
	}

	return tracker.refs, nil
}

func scanDirectory(dirPath string, tracker *Tracker) error {
	parser := hclparse.NewParser()

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fname := entry.Name()
		fullPath := filepath.Join(dirPath, fname)

		if strings.HasSuffix(fname, ".tf") {
			// Handle Native HCL
			if err := scanHCLFile(parser, fullPath, tracker); err != nil {
				log.Printf("Warning: processing %s: %v", fname, err)
			}
		} else if strings.HasSuffix(fname, ".tf.json") {
			// Handle JSON
			if err := scanJSONFile(fullPath, tracker); err != nil {
				log.Printf("Warning: processing %s: %v", fname, err)
			}
		}
	}

	return nil
}

// scanHCLFile parses .tf files using the HCL AST walker
func scanHCLFile(parser *hclparse.Parser, path string, tracker *Tracker) error {
	f, diags := parser.ParseHCLFile(path)
	if diags.HasErrors() {
		return fmt.Errorf(diags.Error())
	}

	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return nil
	}

	hclsyntax.Walk(body, &hclWalker{tracker: tracker})
	return nil
}

// scanJSONFile parses .tf.json files by traversing JSON structure and parsing expression strings
func scanJSONFile(path string, tracker *Tracker) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var data interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	// Recursively walk the JSON structure
	walkJSONStructure(data, path, tracker)
	return nil
}

// walkJSONStructure recursively looks for strings containing "${...}"
func walkJSONStructure(node interface{}, filename string, tracker *Tracker) {
	switch v := node.(type) {
	case map[string]interface{}:
		for _, val := range v {
			walkJSONStructure(val, filename, tracker)
		}
	case []interface{}:
		for _, val := range v {
			walkJSONStructure(val, filename, tracker)
		}
	case string:
		// Terraform interpolations in JSON always involve "${"
		if strings.Contains(v, "${") {
			// Parse the string as an HCL template
			expr, diags := hclsyntax.ParseTemplate([]byte(v), filename, hcl.Pos{Line: 1, Column: 1})
			if !diags.HasErrors() {
				// Extract variables from the parsed expression
				for _, traversal := range expr.Variables() {
					tracker.analyzeTraversal(traversal)
				}
			}
		}
	}
}

// hclWalker implements the hclsyntax.Walker interface for .tf files
type hclWalker struct {
	tracker *Tracker
}

func (w *hclWalker) Enter(node hclsyntax.Node) hcl.Diagnostics {
	expr, ok := node.(hclsyntax.Expression)
	if !ok {
		return nil
	}
	for _, traversal := range expr.Variables() {
		w.tracker.analyzeTraversal(traversal)
	}
	return nil
}

func (w *hclWalker) Exit(node hclsyntax.Node) hcl.Diagnostics {
	return nil
}

// analyzeTraversal checks if a variable reference matches data.terraform_remote_state...
func (t *Tracker) analyzeTraversal(tr hcl.Traversal) {
	// Looking for: data . terraform_remote_state . NAME . outputs . ATTRIBUTE
	if len(tr) < 5 {
		return
	}

	// 1. Root: "data"
	if !isRootName(tr[0], "data") {
		return
	}

	// 2. Resource Type: "terraform_remote_state"
	if !isTraverseName(tr[1], "terraform_remote_state") {
		return
	}

	// 3. State Name
	stateName := getTraverseName(tr[2])
	if stateName == "" {
		return
	}

	// 4. Attribute: "outputs"
	if !isTraverseName(tr[3], "outputs") {
		return
	}

	// 5. Output Attribute
	outputAttr := getTraverseName(tr[4])
	if outputAttr == "" {
		return
	}

	if t.refs[stateName] == nil {
		t.refs[stateName] = make(map[string]struct{})
	}
	t.refs[stateName][outputAttr] = struct{}{}
}

// --- Helper Functions ---

func isRootName(t hcl.Traverser, name string) bool {
	root, ok := t.(hcl.TraverseRoot)
	return ok && root.Name == name
}

func isTraverseName(t hcl.Traverser, name string) bool {
	attr, ok := t.(hcl.TraverseAttr)
	return ok && attr.Name == name
}

func getTraverseName(t hcl.Traverser) string {
	switch step := t.(type) {
	case hcl.TraverseAttr:
		return step.Name
	case hcl.TraverseIndex:
		if step.Key.Type() == cty.String {
			return step.Key.AsString()
		}
	}
	return ""
}
