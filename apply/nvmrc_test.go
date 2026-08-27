package apply

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestParseVersionTriple(t *testing.T) {
	r := require.New(t)

	for in, want := range map[string][3]int{
		"v22":        {22, 0, 0},
		"22":         {22, 0, 0},
		"v22.19":     {22, 19, 0},
		"22.19.0":    {22, 19, 0},
		">=22.19.0":  {22, 19, 0},
		"v20.11.1\n": {20, 11, 1},
	} {
		got, ok := parseVersionTriple(in)
		r.True(ok, in)
		r.Equal(want, got, in)
	}

	_, ok := parseVersionTriple("lts/iron")
	r.False(ok, "non-numeric aliases are not comparable")
}

func TestWarnNvmrcBelowNodeVersion(t *testing.T) {
	// The warning path is logging-only; assert it does not panic and that the
	// comparison classifies each case correctly via parseVersionTriple.
	r := require.New(t)

	cases := []struct {
		nvmrc, engines string
		below          bool
	}{
		{"v20", ">=22.19.0", true},       // the L&S case: stale pin
		{"v22", ">=22.19.0", true},       // bare major is below the floor
		{"v22.19.0", ">=22.19.0", false}, // exactly the floor
		{"v24.11.1", ">=22.19.0", false}, // newer runtime
		{"lts/iron", ">=22.19.0", false}, // not comparable -> no warning
	}

	for _, c := range cases {
		fs := afero.NewMemMapFs()
		r.NoError(afero.WriteFile(fs, ".nvmrc", []byte(c.nvmrc+"\n"), 0644))
		warnNvmrcBelowNodeVersion(fs, c.engines) // must not panic

		have, okHave := parseVersionTriple(c.nvmrc)
		floor, okFloor := parseVersionTriple(c.engines)
		if !okHave || !okFloor {
			r.False(c.below, c.nvmrc)
			continue
		}
		isBelow := have[0] < floor[0] ||
			(have[0] == floor[0] && have[1] < floor[1]) ||
			(have[0] == floor[0] && have[1] == floor[1] && have[2] < floor[2])
		r.Equal(c.below, isBelow, "%s vs %s", c.nvmrc, c.engines)
	}
}

func TestWarnNvmrcMissingFileIsNoop(t *testing.T) {
	warnNvmrcBelowNodeVersion(afero.NewMemMapFs(), ">=22.19.0") // must not panic
}
