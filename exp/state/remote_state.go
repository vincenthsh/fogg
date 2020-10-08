package state

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chanzuckerberg/fogg/config"
	v2 "github.com/chanzuckerberg/fogg/config/v2"
	"github.com/chanzuckerberg/go-misc/sets"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/lang"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// liberal borrowing from https://github.com/hashicorp/terraform-config-inspect/blob/c481b8bfa41ea9dca417c2a8a98fd21bd0399e14/tfconfig/load_hcl.go#L16
func Run(fs afero.Fs, configFile, path string) error {
	// figure out which component or account we are talking about
	conf, err := config.FindAndReadConfig(fs, configFile)
	if err != nil {
		return err
	}

	component, err := conf.PathToComponentType(path)
	if err != nil {
		return err
	}
	logrus.Debugf("component kind %s", component.Kind)

	// collect remote state references
	references, err := collectRemoteStateReferences(path)
	if err != nil {
		return err
	}
	logrus.Debugf("in %s found references %#v", path, references)

	// for each reference, figure out if it is an account or component, since those are separate in our configs
	accounts := []string{}
	components := []string{}

	// we do accounts for both accounts and env components
	for _, r := range references {
		if _, found := conf.Accounts[r]; found {
			accounts = append(accounts, r)
		}
	}

	if component.Kind == "envs" {
		env := conf.Envs[component.Env]

		for _, r := range references {
			if _, found := env.Components[r]; found {
				components = append(components, r)
			}
		}
	}

	// update fogg.yml with new references
	logrus.Debugf("found accounts %#v", accounts)
	logrus.Debugf("found components %#v", components)

	switch component.Kind {
	case "accounts":
		if len(accounts) > 0 {
			c := conf.Accounts[component.Name]
			if c.Common.DependsOn == nil {
				c.Common.DependsOn = &v2.DependsOn{}
			}

			c.DependsOn.Accounts = accounts
			conf.Accounts[component.Name] = c
		}

	case "env":
		if len(accounts) > 0 || len(components) > 0 {
			c := conf.Envs[component.Env].Components[component.Name]

			if c.Common.DependsOn == nil {
				c.Common.DependsOn = &v2.DependsOn{}
			}

			c.DependsOn.Accounts = accounts
			c.DependsOn.Components = components

			conf.Envs[component.Env].Components[component.Name] = c
		}

	default:
		return fmt.Errorf("unknown component.Kind: %s", component.Kind)
	}

	return conf.Write(fs, configFile)
}

func collectRemoteStateReferences(path string) ([]string, error) {
	fs := tfconfig.NewOsFs()

	references := sets.StringSet{}

	primaryPaths, err := dirFiles(fs, path)
	if err != nil {
		return nil, err
	}

	parser := hclparse.NewParser()

	for _, filename := range primaryPaths {
		logrus.Debugf("reading file %s", filename)
		b, err := fs.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		var file *hcl.File
		var fileDiags hcl.Diagnostics

		if strings.HasSuffix(filename, ".json") {
			file, fileDiags = parser.ParseJSON(b, filename)
		} else {
			file, fileDiags = parser.ParseHCL(b, filename)
		}
		if fileDiags.HasErrors() {
			return nil, fileDiags
		}

		if file == nil {
			continue
		}

		content, _, contentDiags := file.Body.PartialContent(rootSchema)
		if contentDiags.HasErrors() {
			return nil, contentDiags
		}

		logrus.Debugf("len(content.Blocks) %v", len(content.Blocks))
		for _, block := range content.Blocks {
			logrus.Debugf("block type: %v", block.Type)

			attrs, _ := block.Body.JustAttributes()

			for _, v := range attrs {
				refs, _ := lang.ReferencesInExpr(v.Expr)

				for _, r := range refs {
					if r == nil {
						continue
					}
					logrus.Debugf("ref: %v", r)
					if resource, ok := r.Subject.(addrs.ResourceInstance); ok {
						if resource.Resource.Type == "terraform_remote_state" {
							references.Add(resource.Resource.Name)
						}
					}
				}
			}
		}
	}

	refNames := references.List()

	sort.Strings(refNames)
	return refNames, nil
}

// taken from https://github.com/hashicorp/terraform-config-inspect/blob/c481b8bfa41ea9dca417c2a8a98fd21bd0399e14/tfconfig/load.go#L81
func dirFiles(fs tfconfig.FS, dir string) ([]string, error) {
	var primary []string

	infos, err := fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var override []string
	for _, info := range infos {
		if info.IsDir() {
			// We only care about files
			continue
		}

		name := info.Name()
		ext := fileExt(name)
		if ext == "" || isIgnoredFile(name) {
			continue
		}

		baseName := name[:len(name)-len(ext)] // strip extension
		isOverride := baseName == "override" || strings.HasSuffix(baseName, "_override")

		fullPath := filepath.Join(dir, name)
		if isOverride {
			override = append(override, fullPath)
		} else {
			primary = append(primary, fullPath)
		}
	}

	// We are assuming that any _override files will be logically named,
	// and processing the files in alphabetical order. Primaries first, then overrides.
	primary = append(primary, override...)

	return primary, nil
}

// fileExt returns the Terraform configuration extension of the given
// path, or a blank string if it is not a recognized extension.
func fileExt(path string) string {
	if strings.HasSuffix(path, ".tf") {
		return ".tf"
	} else if strings.HasSuffix(path, ".tf.json") {
		return ".tf.json"
	} else {
		return ""
	}
}

// isIgnoredFile returns true if the given filename (which must not have a
// directory path ahead of it) should be ignored as e.g. an editor swap file.
func isIgnoredFile(name string) bool {
	return strings.HasPrefix(name, ".") || // Unix-like hidden files
		strings.HasSuffix(name, "~") || // vim
		strings.HasPrefix(name, "#") && strings.HasSuffix(name, "#") // emacs
}