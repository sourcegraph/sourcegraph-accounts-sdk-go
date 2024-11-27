package clientcredentials

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/log/logtest"
	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1/clientsv1connect"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestInterceptor(t *testing.T) {
	// All tests based on UsersService.GetUser()
	for _, tc := range []struct {
		name      string
		token     *sams.IntrospectTokenResponse
		wantError autogold.Value
		wantLogs  autogold.Value
	}{{
		name: "inactive token",
		token: &sams.IntrospectTokenResponse{
			Active: false,
		},
		wantError: autogold.Expect("permission_denied: permission denied"),
		wantLogs:  autogold.Expect([]string{}),
	}, {
		name: "insufficient scopes",
		token: &sams.IntrospectTokenResponse{
			Active: true,
		},
		wantError: autogold.Expect("permission_denied: insufficient scopes: got scopes [], required: [profile]"),
		wantLogs:  autogold.Expect([]string{}),
	}, {
		name: "matches required scope",
		token: &sams.IntrospectTokenResponse{
			Active: true,
			Scopes: scopes.Scopes{"profile"},
		},
		wantError: autogold.Expect(nil), // should not error!
		wantLogs:  autogold.Expect([]string{}),
	}, {
		name: "wrong scope",
		token: &sams.IntrospectTokenResponse{
			Active: true,
			Scopes: scopes.Scopes{"not-a-scope"},
		},
		wantError: autogold.Expect("permission_denied: insufficient scopes: got scopes [not-a-scope], required: [profile]"),
		wantLogs:  autogold.Expect([]string{}),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			logger, exportLogs := logtest.Captured(t)
			interceptor := NewInterceptor(
				logger,
				&mockTokenIntrospector{
					response: tc.token,
				},
				clientsv1.E_SamsRequiredScopes,
			)
			mux := http.NewServeMux()
			mux.Handle(
				clientsv1connect.NewUsersServiceHandler(clientsv1connect.UnimplementedUsersServiceHandler{},
					connect.WithInterceptors(interceptor)),
			)
			srv := httptest.NewServer(mux)
			t.Cleanup(srv.Close)

			c := clientsv1connect.NewUsersServiceClient(
				oauth2.NewClient(
					context.Background(),
					oauth2.StaticTokenSource(&oauth2.Token{
						AccessToken: "foobar",
						TokenType:   "bearer",
					}),
				),
				srv.URL)
			_, err := c.GetUser(context.Background(), connect.NewRequest(&clientsv1.GetUserRequest{}))

			// Success cases are connect.CodeUnimplemented
			require.Error(t, err)

			var connectErr *connect.Error
			if errors.As(err, &connectErr) {
				if connectErr.Code() == connect.CodeUnimplemented {
					tc.wantError.Equal(t, nil) // should not expect an error
				} else {
					tc.wantError.Equal(t, err.Error())
				}
			} else {
				t.Errorf("error %q is not *connect.Error", err.Error())
			}

			tc.wantLogs.Equal(t, exportLogs().Messages())
		})
	}
}
