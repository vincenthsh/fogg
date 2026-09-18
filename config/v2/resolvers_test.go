package v2_test

import (
	"testing"

	v2 "github.com/chanzuckerberg/fogg/config/v2"
	"github.com/stretchr/testify/require"
)

func TestResolveTfLint(t *testing.T) {
	r := require.New(t)
	tru := true
	fal := false

	data := []struct {
		def    *bool
		over   *bool
		output *bool
	}{
		{nil, nil, &fal},
		{nil, &tru, &tru},
		{nil, &fal, &fal},
		{&tru, nil, &tru},
		{&tru, &tru, &tru},
		{&tru, &fal, &fal},
		{&fal, nil, &fal},
		{&fal, &tru, &tru},
		{&fal, &fal, &fal},
	}
	for _, test := range data {
		tt := test
		t.Run("", func(t *testing.T) {
			def := v2.Common{Tools: &v2.Tools{TfLint: &v2.TfLint{Enabled: tt.def}}}
			over := v2.Common{Tools: &v2.Tools{TfLint: &v2.TfLint{Enabled: tt.over}}}
			result := v2.ResolveTfLint(def, over)
			r.Equal(tt.output, result.Enabled)
		})
	}
}

func TestResolveAtlantisCustomWorkflow(t *testing.T) {
	r := require.New(t)
	wf := func(name string) v2.Common {
		return v2.Common{Tools: &v2.Tools{Atlantis: &v2.Atlantis{CustomWorkflow: &name}}}
	}

	r.Nil(v2.ResolveOptionalString(v2.AtlantisCustomWorkflowGetter, v2.Common{}, v2.Common{}, v2.Common{}))
	r.Equal("defaults", *v2.ResolveOptionalString(v2.AtlantisCustomWorkflowGetter, wf("defaults"), v2.Common{}, v2.Common{}))
	r.Equal("env", *v2.ResolveOptionalString(v2.AtlantisCustomWorkflowGetter, wf("defaults"), wf("env"), v2.Common{}))
	r.Equal("component", *v2.ResolveOptionalString(v2.AtlantisCustomWorkflowGetter, wf("defaults"), wf("env"), wf("component")))
}
