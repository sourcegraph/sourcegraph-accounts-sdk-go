package sams

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
)

// ClientV1 provides helpers to talk to a SAMS instance via Clients API v1.
type ClientV1 struct {
	rootURL                 string
	clientCredentialsConfig clientcredentials.Config
	tokenSource             oauth2.TokenSource
}

// NewV1 returns a new SAMS client for interacting with Clients API v1 using the
// given client credentials, and the scopes are used to as requested scopes for
// access tokens that are issued to this client.
func NewV1(url, clientID, clientSecret string, scopeList []scopes.Scope) (*ClientV1, error) {
	if url == "" {
		return nil, errors.New("empty URL")
	} else if clientID == "" || clientSecret == "" {
		return nil, errors.New("empty client ID or secret")
	}

	clientCredentialsConfig := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth/token", url),
		Scopes:       scopes.ToStrings(scopeList),
	}
	return &ClientV1{
		rootURL:                 strings.TrimSuffix(url, "/"),
		clientCredentialsConfig: clientCredentialsConfig,
		tokenSource:             clientCredentialsConfig.TokenSource(context.Background()),
	}, nil
}

func parseResponseAndError[T any](resp *connect.Response[T], err error) (*connect.Response[T], error) {
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		switch connectErr.Code() {
		case connect.CodeNotFound:
			return resp, ErrNotFound
		}
		return resp, err
	}
	return resp, err
}

func (c *ClientV1) gRPCURL() string {
	return c.rootURL + "/api/grpc"
}

// Users returns a client handler to interact with the UsersServiceV1 API.
func (c *ClientV1) Users() *UsersServiceV1 {
	return &UsersServiceV1{client: c}
}

// Sessions returns a client handler to interact with the SessionsServiceV1 API.
func (c *ClientV1) Sessions() *SessionsServiceV1 {
	return &SessionsServiceV1{client: c}
}

// Tokens returns a client handler to interact with the TokensServiceV1 API.
func (c *ClientV1) Tokens() *TokensServiceV1 {
	return &TokensServiceV1{client: c}
}

// TokenSource returns a valid access token to send requests authenticated with
// a SAMS access token. Internally, the token returned is reused. So that new
// tokens are only created when needed. (Provided this `Client` is long-lived.)
//
// To send outbound requests to other SAMS clients, you would use:
//
//	```go
//	httpClient := oauth2.NewClient(ctx, samsClient.TokenSource())
//	```
func (c *ClientV1) TokenSource() oauth2.TokenSource {
	return c.clientCredentialsConfig.TokenSource(context.Background())
}

var ErrNotFound = errors.New("not found")
