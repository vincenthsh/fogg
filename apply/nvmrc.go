package apply

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// semverPrefixRe pulls the leading numeric version out of either an .nvmrc body
// ("v22", "22.19.0", "lts/iron") or an engines.node range (">=22.19.0", "22").
var semverPrefixRe = regexp.MustCompile(`(\d+)(?:\.(\d+))?(?:\.(\d+))?`)

// parseVersionTriple returns the leading major/minor/patch of s and whether one
// was found. Absent components count as zero, so "v22" parses as 22.0.0 -- which
// is the behaviour we want: a bare major satisfies a ">=22.19.0" contract only if
// the runtime happens to be new enough, and we should warn rather than assume.
func parseVersionTriple(s string) ([3]int, bool) {
	m := semverPrefixRe.FindStringSubmatch(s)
	if m == nil {
		return [3]int{}, false
	}
	var out [3]int
	for i := 1; i <= 3; i++ {
		if m[i] != "" {
			out[i-1], _ = strconv.Atoi(m[i])
		}
	}
	return out, true
}

// warnNvmrcBelowNodeVersion warns when the repo's .nvmrc pins a Node older than
// the floor declared in the generated engines.node.
//
// .nvmrc ships as a ".create" file: written verbatim only when absent, so that a
// repo can pin its own runtime. The consequence is that raising the fogg default
// silently leaves existing repos behind -- engines.node moves but .nvmrc does
// not, and nothing enforces engines without engine-strict, so the mismatch only
// surfaces later as a confusing runtime or synth failure. Fogg cannot safely
// rewrite the file (that would clobber a deliberate pin), so it says so instead.
func warnNvmrcBelowNodeVersion(fs afero.Fs, nodeVersion string) {
	contents, err := afero.ReadFile(fs, ".nvmrc")
	if err != nil {
		return // no .nvmrc (turbo disabled, or not yet created)
	}
	pinned := strings.TrimSpace(string(contents))

	floor, ok := parseVersionTriple(nodeVersion)
	if !ok {
		return
	}
	have, ok := parseVersionTriple(pinned)
	if !ok {
		// Something like "lts/iron" -- we cannot compare, so leave it alone.
		return
	}

	if have[0] < floor[0] ||
		(have[0] == floor[0] && have[1] < floor[1]) ||
		(have[0] == floor[0] && have[1] == floor[1] && have[2] < floor[2]) {
		logrus.Warnf(
			".nvmrc pins node %s but engines.node requires %s -- update .nvmrc by hand "+
				"(it is only created when absent, so fogg will not overwrite your pin)",
			pinned, nodeVersion)
	}
}
