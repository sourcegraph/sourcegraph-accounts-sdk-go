package sams

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
)

// TokensServiceV1 provides client methods to interact with the TokensService
// API v1.
type TokensServiceV1 struct {
	client *ClientV1
}

type IntrospectTokenResponse struct {
	// Active indicates whether the token is currently active. The value is "true"
	// if the token has been issued by the SAMS instance, has not been revoked, and
	// has not expired.
	Active bool
	// Scopes is the list of scopes granted by the token.
	Scopes scopes.Scopes
	// ClientID is the identifier of the SAMS client that the token was issued to.
	ClientID string
	// ExpiresAt indicates when the token expires.
	ExpiresAt time.Time
}

// IntrospectToken takes a SAMS access token and returns relevant metadata.
//
// 🚨SECURITY: SAMS will return a successful result if the token is valid, but
// is no longer active. It is critical that the caller not honor tokens where
// `.Active == false`.
func (s *TokensServiceV1) IntrospectToken(ctx context.Context, token string) (*IntrospectTokenResponse, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/oauth/introspect", s.client.rootURL),
		strings.NewReader("token="+token),
	)
	if err != nil {
		return nil, errors.Wrap(err, "create introspection request")
	}

	credentialStr := fmt.Sprintf(
		"%s:%s",
		s.client.clientCredentialsConfig.ClientID, s.client.clientCredentialsConfig.ClientSecret,
	)
	basicAuth := base64.StdEncoding.EncodeToString([]byte(credentialStr))
	req.Header.Set("Authorization", "Basic "+basicAuth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// NOTE: Always create a new client to avoid unintended global idle connection
	// sharing (http.DefaultClient).
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "request introspection endpoint")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read introspection response body")
	}

	// For full token introspection response format, see
	// https://www.oauth.com/oauth2-servers/token-introspection-endpoint/
	var result struct {
		Active     bool   `json:"active"`
		Scope      string `json:"scope"` // Space-separated list
		ClientID   string `json:"client_id"`
		Expiration int64  `json:"exp"` // Unix timestamp
	}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, errors.Wrap(err, "parse introspection response body")
	}
	return &IntrospectTokenResponse{
		Active:    result.Active,
		Scopes:    scopes.ToScopes(strings.Split(result.Scope, " ")),
		ClientID:  result.ClientID,
		ExpiresAt: time.Unix(result.Expiration, 0),
	}, nil
}
