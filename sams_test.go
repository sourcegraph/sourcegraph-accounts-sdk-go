package sams

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/hexops/valast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type staticEnvGetter map[string]string

func (e staticEnvGetter) Get(name, defaultValue, _ string) string {
	v, ok := e[name]
	if !ok {
		return defaultValue
	}
	return v
}

func (e staticEnvGetter) GetOptional(name, description string) *string {
	v := e.Get(name, "", description)
	if v == "" {
		return nil
	}
	return &v
}

func TestNewClientV1ConnectionConfigFromEnv(t *testing.T) {
	for _, tc := range []struct {
		name string
		env  staticEnvGetter

		want            autogold.Value
		wantValidateErr autogold.Value
	}{
		{
			name: "no env",
			env:  staticEnvGetter{},
			want: autogold.Expect(ConnConfig{ExternalURL: "https://accounts.sourcegraph.com"}),
		},
		{
			name: "override API URL",
			env: staticEnvGetter{
				"SAMS_API_URL": "https://my-internal-url.net",
			},
			want: autogold.Expect(ConnConfig{
				ExternalURL: "https://accounts.sourcegraph.com",
				APIURL:      valast.Addr("https://my-internal-url.net").(*string),
			}),
		},
		{
			name: "set all",
			env: staticEnvGetter{
				"SAMS_URL":     "https://my-external-url.net",
				"SAMS_API_URL": "https://my-internal-url.net",
			},
			want: autogold.Expect(ConnConfig{
				ExternalURL: "https://my-external-url.net",
				APIURL:      valast.Addr("https://my-internal-url.net").(*string),
			}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := NewConnConfigFromEnv(tc.env)
			tc.want.Equal(t, got)

			t.Run("Validate", func(t *testing.T) {
				err := got.Validate()
				if tc.wantValidateErr != nil {
					require.Error(t, err)
					tc.wantValidateErr.Equal(t, err.Error())
				} else {
					assert.NoError(t, err)
				}
			})
		})
	}
}
