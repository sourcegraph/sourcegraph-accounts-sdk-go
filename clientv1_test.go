package sams

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/hexops/autogold/v2"
	"github.com/hexops/valast"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
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
			name:            "no env",
			env:             staticEnvGetter{},
			want:            autogold.Expect(ClientV1ConnConfig{ExternalURL: "https://accounts.sourcegraph.com"}),
			wantValidateErr: autogold.Expect("empty client ID"),
		},
		{
			name: "only client credentials",
			env: staticEnvGetter{
				"SAMS_CLIENT_ID":     "fooclient",
				"SAMS_CLIENT_SECRET": "barsecret",
			},
			want: autogold.Expect(ClientV1ConnConfig{
				ExternalURL:  "https://accounts.sourcegraph.com",
				ClientID:     "fooclient",
				ClientSecret: "barsecret",
			}),
		},
		{
			name: "override API URL",
			env: staticEnvGetter{
				"SAMS_API_URL":       "https://my-internal-url.net",
				"SAMS_CLIENT_ID":     "fooclient",
				"SAMS_CLIENT_SECRET": "barsecret",
			},
			want: autogold.Expect(ClientV1ConnConfig{
				ExternalURL:  "https://accounts.sourcegraph.com",
				APIURL:       valast.Addr("https://my-internal-url.net").(*string),
				ClientID:     "fooclient",
				ClientSecret: "barsecret",
			}),
		},
		{
			name: "set all",
			env: staticEnvGetter{
				"SAMS_URL":           "https://my-external-url.net",
				"SAMS_API_URL":       "https://my-internal-url.net",
				"SAMS_CLIENT_ID":     "fooclient",
				"SAMS_CLIENT_SECRET": "barsecret",
			},
			want: autogold.Expect(ClientV1ConnConfig{
				ExternalURL:  "https://my-external-url.net",
				APIURL:       valast.Addr("https://my-internal-url.net").(*string),
				ClientID:     "fooclient",
				ClientSecret: "barsecret",
			}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := NewClientV1ConnectionConfigFromEnv(tc.env)
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

func TestNewClientV1(t *testing.T) {
	c, err := NewClientV1(
		ClientV1ConnConfig{
			ExternalURL:  "https://accounts.sourcegraph.com",
			ClientID:     "fooclient",
			ClientSecret: "barsecret",
		},
		[]scopes.Scope{scopes.Profile})
	require.NoError(t, err)
	assert.NotEmpty(t, c.defaultInterceptors)
}

func TestParseResponseAndError(t *testing.T) {
	tests := []struct {
		name    string
		err     func() error
		wantErr string
	}{
		{
			name: "no error",
			err: func() error {
				return nil
			},
			wantErr: "",
		},
		{
			name: "not found",
			err: func() error {
				return connect.NewError(connect.CodeNotFound, nil)
			},
			wantErr: ErrNotFound.Error(),
		},
		{
			name: "record mismatch",
			err: func() error {
				detail, err := connect.NewErrorDetail(&clientsv1.ErrorRecordMismatch{})
				require.NoError(t, err)
				grpcErr := connect.NewError(connect.CodeFailedPrecondition, nil)
				grpcErr.AddDetail(detail)
				return grpcErr
			},
			wantErr: ErrRecordMismatch.Error(),
		},
		{
			name: "other error",
			err: func() error {
				return errors.New("bar")
			},
			wantErr: "bar",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parseResponseAndError(connect.NewResponse(pointers.Ptr("foo")), test.err())
			if test.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, test.wantErr)
			}
		})
	}
}
