package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDict(t *testing.T) {
	m := make(map[string]string)
	m["foo"] = "bar"
	r := dict(m)
	require.NotNil(t, r)
	require.IsType(t, map[string]interface{}{}, r)
	require.Equal(t, "bar", r["foo"])
}

func TestJsPropName(t *testing.T) {
	r := require.New(t)

	// Valid bare identifiers pass through untouched.
	for _, ok := range []string{"foo", "_foo", "$foo", "foo_bar", "foo123", "F"} {
		r.Equal(ok, jsPropName(ok), ok)
	}

	// Terraform output names and fogg module prefixes routinely contain hyphens,
	// which are legal in HCL but not in a bare TS property key. Emitting them
	// unquoted produced a .d.ts that failed to parse (TS1005/TS1131).
	r.Equal(`"hs-api-app_ec2_instance_id_115C4EF3"`, jsPropName("hs-api-app_ec2_instance_id_115C4EF3"))
	r.Equal(`"foo-bar"`, jsPropName("foo-bar"))
	r.Equal(`"1leading"`, jsPropName("1leading"))
	r.Equal(`""`, jsPropName(""))
}
