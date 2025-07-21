package clientcredentials

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"
	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestHTTPAuthenticator(t *testing.T) {
	// Test server requires 'profile' scope
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
		wantError: autogold.Expect("Unauthorized: Inactive token\n"),
		wantLogs:  autogold.Expect([]string{"attempt to authenticate with inactive SAMS token"}),
	}, {
		name: "insufficient scopes",
		token: &sams.IntrospectTokenResponse{
			Active: true,
		},
		wantError: autogold.Expect("Forbidden: Missing required scope\n"),
		wantLogs:  autogold.Expect([]string{"attempt to authenticate using SAMS token without required scope"}),
	}, {
		name: "matches required scope",
		token: &sams.IntrospectTokenResponse{
			Active: true,
			Scopes: scopes.Scopes{"profile"},
		},
		wantError: autogold.Expect("OK"), // should not error!
		wantLogs:  autogold.Expect([]string{}),
	}, {
		name: "wrong scope",
		token: &sams.IntrospectTokenResponse{
			Active: true,
			Scopes: scopes.Scopes{"not-a-scope"},
		},
		wantError: autogold.Expect("Forbidden: Missing required scope\n"),
		wantLogs:  autogold.Expect([]string{"attempt to authenticate using SAMS token without required scope"}),
	}, {
		name: "SAT token with userID rejected",
		token: &sams.IntrospectTokenResponse{
			Active:   true,
			ClientID: "test-client",
			UserID:   "test-user",
			Scopes:   scopes.Scopes{"profile"},
		},
		wantError: autogold.Expect("Forbidden: User tokens not allowed\n"),
		wantLogs:  autogold.Expect([]string{"attempt to authenticate using SAMS token with user ID"}),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			logger, exportLogs := logtest.Captured(t)
			authenticator := NewHTTPAuthenticator(logger, &mockTokenIntrospector{
				response: tc.token,
			})
			mux := http.NewServeMux()
			mux.Handle("/", authenticator.RequireScopes(
				scopes.Scopes{scopes.Profile},
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(http.StatusText(http.StatusOK)))
				}),
			))
			srv := httptest.NewServer(mux)
			t.Cleanup(srv.Close)

			req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
			require.NoError(t, err)
			resp, err := oauth2.NewClient(
				context.Background(),
				oauth2.StaticTokenSource(&oauth2.Token{
					AccessToken: "foobar",
					TokenType:   "bearer",
				}),
			).Do(req)
			require.NoError(t, err)

			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			tc.wantError.Equal(t, string(body))
			tc.wantLogs.Equal(t, exportLogs().Messages())
		})
	}
}
