package plan

import (
	"io/fs"
	"strconv"
	"strings"
	"testing"

	v2 "github.com/chanzuckerberg/fogg/config/v2"
	"github.com/chanzuckerberg/fogg/templates"
	"github.com/stretchr/testify/require"
)

// The generated repo declares its Node floor in two places that cannot be
// derived from one another: engines.node is templated from TurboConfig, while
// .nvmrc is a ".create" file copied verbatim and only when absent, so it has no
// access to template data. These tests keep the pair honest.

func TestNvmrcMatchesNodeVersionFloor(t *testing.T) {
	r := require.New(t)

	contents, err := fs.ReadFile(templates.Templates.TurboRoot, ".nvmrc.create")
	r.NoError(err)

	r.Equal("v"+nodeVersionFloor+"\n", string(contents),
		".nvmrc.create must pin the same Node version as nodeVersionFloor in plan/ci.go")
}

func TestNodeVersionDefaultUsesFloor(t *testing.T) {
	r := require.New(t)

	p := &Plan{}
	turbo := p.buildTurboRootConfig(&v2.Config{})

	// A bare major (">=22" or "22") would admit 22.0.0, which cdktn-cli 0.24
	// rejects. The default must carry the full floor.
	r.Equal(">="+nodeVersionFloor, turbo.NodeVersion)
	r.Equal(pnpmVersion, turbo.PnpmVersion)
}

// pnpm 10 stopped reading the "pnpm" key from package.json, which is where
// fogg used to emit overrides. Anything below 10 would silently re-enable the
// old location and mask the move to pnpm-workspace.yaml.
func TestPnpmVersionIsAtLeast10(t *testing.T) {
	r := require.New(t)

	major, err := strconv.Atoi(strings.SplitN(pnpmVersion, ".", 2)[0])
	r.NoError(err)
	r.GreaterOrEqual(major, 10, "pnpm must be >=10 so overrides are read from pnpm-workspace.yaml")
}

func TestPnpmAllowBuildsDefaultsAndMerge(t *testing.T) {
	r := require.New(t)
	p := &Plan{}

	// @swc/core must be allowed out of the box: it is a generated-component
	// devDependency with a native postinstall, and pnpm >=10 halts without it.
	base := p.buildTurboRootConfig(&v2.Config{})
	r.Equal([]string{"@swc/core"}, base.PnpmAllowBuilds)
	r.Nil(base.PnpmMinReleaseAgeExempt, "cooldown exemptions must be opt-in")

	tr := true
	merged := p.buildTurboRootConfig(&v2.Config{
		Turbo: &v2.TurboConfig{
			Enabled:                      &tr,
			PnpmAllowBuilds:              []string{"esbuild", "@swc/core"},
			PnpmMinimumReleaseAgeExclude: []string{"terraconstructs", "cdktn"},
		},
	})

	// union, de-duplicated, sorted -- so the generated file is stable.
	r.Equal([]string{"@swc/core", "esbuild"}, merged.PnpmAllowBuilds)
	r.Equal([]string{"cdktn", "terraconstructs"}, merged.PnpmMinReleaseAgeExempt)
}

func TestMergeSortedEmptyIsNil(t *testing.T) {
	r := require.New(t)
	r.Nil(mergeSorted(nil, nil))
	r.Nil(mergeSorted([]string{""}, nil), "blank entries must not emit a key")
}
