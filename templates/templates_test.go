package templates

import (
	"io/fs"
	"testing"

	v2 "github.com/chanzuckerberg/fogg/config/v2"
	"github.com/chanzuckerberg/fogg/util"
	"github.com/stretchr/testify/require"
)

func TestOpenTemplate(t *testing.T) {
	temps := Templates

	type args struct {
		box  fs.FS
		path string
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"foo", args{temps.Components[v2.ComponentKindTerraform], "Makefile.tmpl"}, false},
	}

	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			r := require.New(t)
			f, err := tt.args.box.Open(tt.args.path)
			r.NoError(err)

			temp, err := util.OpenTemplate("foo", f, temps.Common)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			r.NotNil(temp.Templates())
		})
	}

}

// The index.d.ts property key must go through jsPropName. Terraform output names
// and fogg module prefixes may contain hyphens, which are invalid in a bare TS
// property key and produced a .d.ts that failed to parse. No golden fixture uses
// a hyphenated module prefix, which is why that shipped unnoticed -- so guard the
// template itself against a revert to raw interpolation.
func TestIndexDTsQuotesPropertyNames(t *testing.T) {
	r := require.New(t)

	b, err := fs.ReadFile(Templates.ModuleInvocation, "index.d.ts.tmpl")
	r.NoError(err)
	tmpl := string(b)

	r.Contains(tmpl, "jsPropName", "index.d.ts.tmpl must render property keys via jsPropName")
	r.NotContains(tmpl, "{{$outer.ModulePrefix}}{{.Name}}: any;",
		"raw interpolation emits unquoted hyphenated keys (TS1005/TS1131)")
}
