package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hexops/autogold/v2"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
)

// newMockServer returns a new mock server that mimics a SAMS instance for OIDC
// authentication flow.
func newMockServer(t *testing.T, redirectURI, clientID, subject, code, accessToken string, userinfo []byte) *httptest.Server {
	mux := http.NewServeMux()

	var openidConfig map[string]any
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(openidConfig)
		require.NoError(t, err)
	})

	var state, nonce string
	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		// Assert all desired parameters are present.
		assert.Equal(t, clientID, r.URL.Query().Get("client_id"))
		assert.Equal(t, redirectURI, r.URL.Query().Get("redirect_uri"))
		assert.Equal(t, "code", r.URL.Query().Get("response_type"))
		assert.NotEmpty(t, r.URL.Query().Get("scope"))
		assert.NotEmpty(t, r.URL.Query().Get("prompt"))
		assert.NotEmpty(t, r.URL.Query().Get("prompt_auth"))

		state = r.URL.Query().Get("state")
		assert.NotEmpty(t, state)
		nonce = r.URL.Query().Get("nonce")
		assert.NotEmpty(t, nonce)
		http.Redirect(w, r, fmt.Sprintf("%s?state=%s&nonce=%s&code=%s", redirectURI, state, nonce, code), http.StatusFound)
	})

	var rawIDToken string
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		vals, err := url.ParseQuery(string(body))
		require.NoError(t, err)

		assert.Equal(t, code, vals.Get("code"))
		assert.Equal(t, "authorization_code", vals.Get("grant_type"))
		assert.Equal(t, redirectURI, vals.Get("redirect_uri"))

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  accessToken,
			"token_type":    "Bearer",
			"refresh_token": "test-refresh-token",
			"expires_in":    3600,
			"id_token":      rawIDToken,
		})
		require.NoError(t, err)
	})
	mux.HandleFunc("/oauth/userinfo", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write(userinfo)
		require.NoError(t, err)
	})

	rs256, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	mux.HandleFunc("/oauth/discovery/keys", func(w http.ResponseWriter, _ *http.Request) {
		key, err := jwk.FromRaw(rs256.PublicKey)
		require.NoError(t, err)

		err = key.Set(jwk.KeyUsageKey, "sig")
		require.NoError(t, err)
		err = key.Set(jwk.AlgorithmKey, "RS256")
		require.NoError(t, err)
		marshalledKey, err := json.Marshal(key)
		require.NoError(t, err)

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]any{
			"keys": []json.RawMessage{marshalledKey},
		})
		require.NoError(t, err)
	})

	s := httptest.NewServer(mux)
	t.Cleanup(func() { s.Close() })

	openidConfig = map[string]any{
		"issuer":                                s.URL,
		"authorization_endpoint":                s.URL + "/oauth/authorize",
		"token_endpoint":                        s.URL + "/oauth/token",
		"userinfo_endpoint":                     s.URL + "/oauth/userinfo",
		"jwks_uri":                              s.URL + "/oauth/discovery/keys",
		"id_token_signing_alg_values_supported": []string{"RS256"},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodRS256,
		jwt.MapClaims{
			"iss":   s.URL,
			"sub":   subject,
			"aud":   clientID,
			"exp":   time.Now().Add(time.Hour).Unix(),
			"iat":   time.Now().Unix(),
			"nonce": nonce,
		},
	)
	rawIDToken, err = token.SignedString(rs256)
	require.NoError(t, err)
	return s
}

func TestHandler(t *testing.T) {
	const (
		testClientID    = "test-client-id"
		testCode        = "test-code"
		testAccessToken = "test-access-token"
		testSubject     = "018d21f2-0b8d-756a-84f0-13b942a2bae5"
		testName        = "John Doe"
		testEmail       = "john.doe@example.com"
	)
	someday, err := time.Parse(time.DateOnly, "2021-01-01")
	require.NoError(t, err)
	userinfo, err := json.Marshal(
		map[string]any{
			"sub":            testSubject,
			"name":           testName,
			"email":          testEmail,
			"email_verified": true,
			"picture":        "https://example.com/avatar.jpg",
			"created_at":     someday,
		},
	)
	require.NoError(t, err)

	// Set up the mock service server.
	mockState := ""
	mux := http.NewServeMux()
	mockServiceServer := httptest.NewServer(mux)
	t.Cleanup(func() { mockServiceServer.Close() })
	redirectURI := mockServiceServer.URL + "/callback"

	// Set up the mock SAMS server.
	mockSAMSServer := newMockServer(t, redirectURI, testClientID, testSubject, testCode, testAccessToken, userinfo)

	// Set up auth handlers for the mock service server.
	h, err := NewHandler(
		Config{
			Issuer:         mockSAMSServer.URL,
			ClientID:       testClientID,
			ClientSecret:   "test-client-secret",
			RequestScopes:  []scopes.Scope{scopes.OpenID, scopes.Profile, scopes.Email},
			RedirectURI:    redirectURI,
			FailureHandler: DefaultFailureHandler,
			StateSetter: func(_ http.ResponseWriter, _ *http.Request, state string) error {
				mockState = state
				return nil
			},
			StateGetter: func(*http.Request) (string, error) {
				return mockState, nil
			},
			StateDeleter: func(http.ResponseWriter, *http.Request) {
				mockState = ""
			},
		},
	)
	require.NoError(t, err)

	mux.Handle("/login", h.LoginHandler())
	mux.Handle("/callback", h.CallbackHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userInfo := UserInfoFromContext(r.Context())
			assert.NotNil(t, userInfo.Token)
			assert.NotNil(t, userInfo.IDToken)
			err := json.NewEncoder(w).Encode(userInfo)
			require.NoError(t, err)
		}),
	))

	// Simulate authentication flow.
	client := &http.Client{}
	resp, err := client.Get(mockServiceServer.URL + "/login?prompt=login&prompt_auth=github")
	require.NoError(t, err)
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	autogold.Expect(`{"sub":"018d21f2-0b8d-756a-84f0-13b942a2bae5","name":"John Doe","email":"john.doe@example.com","email_verified":true,"picture":"https://example.com/avatar.jpg","created_at":"2021-01-01T00:00:00Z"}
`).Equal(t, string(respBody))
}
