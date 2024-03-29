package sams

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
)

func TestNewClientV1(t *testing.T) {
	c, err := NewClientV1(
		"https://accounts.sourcegraph.com",
		"fooclient",
		"barsecret",
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
