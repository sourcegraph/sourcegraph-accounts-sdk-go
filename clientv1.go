package sams

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
)

// ClientV1 provides helpers to talk to a SAMS instance via Clients API v1.
type ClientV1 struct {
	rootURL                 string
	clientCredentialsConfig clientcredentials.Config
	tokenSource             oauth2.TokenSource

	// defaultInterceptors is a list of default interceptors to use with all
	// clients, generally providing enhanced diagnostics.
	defaultInterceptors []connect.Interceptor
}

type ClientV1ConnConfig struct {
	// ExternalURL is the configured default external ExternalURL of the relevant Sourcegraph
	// Accounts instance.
	ExternalURL string
	// APIURL is the URL to use for Sourcegraph Accounts API interactions. This
	// can be set to some internal URLs for private networking. If this is nil,
	// the client will fall back to URL instead.
	APIURL *string
	// Client credentials representing this Sourcegraph Accounts client
	ClientID     string
	ClientSecret string
}

// envGetter is a helper interface for getting environment variables based on
// the MSP runtime environment, matching the type
// github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/contract.Env.
type envGetter interface {
	// Get returns the value with the given name. If no value was supplied in the
	// environment, the given default is used in its place. If no value is available,
	// an error is added to the validation errors list.
	Get(name, defaultValue, description string) string
	// GetOptional returns the value with the given name, or nil if no value is
	// available.
	GetOptional(name, description string) *string
}

const defaultExternalURL = "https://accounts.sourcegraph.com"

// NewClientV1ConnectionConfigFromEnv initializes configuration for connecting
// to Sourcegraph Accounts using default standards for loading environment
// variables. This allows the Core Services team to more easily configure access.
func NewClientV1ConnectionConfigFromEnv(env envGetter) (c ClientV1ConnConfig) {
	c.ExternalURL = env.Get("SAMS_URL", defaultExternalURL, "External URL of the connected SAMS instance")
	c.APIURL = env.GetOptional("SAMS_API_URL", "URL to use for connecting to the API of a SAMS instance instead of SAMS_URL")
	c.ClientID = env.Get("SAMS_CLIENT_ID", "", "Client ID of the SAMS client")
	c.ClientSecret = env.Get("SAMS_CLIENT_SECRET", "", "Client secret of the SAMS client")
	return c
}

func (c ClientV1ConnConfig) Validate() error {
	if c.ExternalURL == "" {
		return errors.New("empty external URL")
	}
	if c.getAPIURL() == "" {
		return errors.New("evaluated API URL is empty")
	}
	if c.ClientID == "" {
		return errors.New("empty client ID")
	}
	if c.ClientSecret == "" {
		return errors.New("empty client secret")
	}
	
	return nil
}

func (c ClientV1ConnConfig) getAPIURL() string {
	if c.APIURL != nil {
		return *c.APIURL
	}
	return c.ExternalURL
}

// NewClientV1 returns a new SAMS client for interacting with Clients API v1
// using the given client credentials, and the scopes are used to as requested
// scopes for access tokens that are issued to this client.
func NewClientV1(config ClientV1ConnConfig, scopeList []scopes.Scope) (*ClientV1, error) {
	if err := config.Validate(); err != nil {
		return nil, errors.Wrap(err, "ClientV1ConnectionConfig is invalid")
	}

	otelinterceptor, err := otelconnect.NewInterceptor(
		// Start with simple, lower-cardinality metrics
		otelconnect.WithoutServerPeerAttributes(),
		// Start with lower-volume trace data
		otelconnect.WithoutTraceEvents())
	if err != nil {
		return nil, errors.Wrap(err, "initiate OTEL interceptor")
	}

	apiURL := config.getAPIURL()
	clientCredentialsConfig := clientcredentials.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		TokenURL:     fmt.Sprintf("%s/oauth/token", apiURL),
		Scopes:       scopes.ToStrings(scopeList),
	}
	return &ClientV1{
		rootURL:                 strings.TrimSuffix(apiURL, "/"),
		clientCredentialsConfig: clientCredentialsConfig,
		tokenSource:             clientCredentialsConfig.TokenSource(context.Background()),

		defaultInterceptors: []connect.Interceptor{otelinterceptor},
	}, nil
}

func parseResponseAndError[T any](resp *connect.Response[T], err error) (*connect.Response[T], error) {
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		// Not an error that we can extract information from.
		return resp, err
	}

	if connectErr.Code() == connect.CodeNotFound {
		return nil, ErrNotFound
	}

	// Cannot determine action solely based on status code, let's look at the error
	// details.
	for _, detail := range connectErr.Details() {
		value, err := detail.Value()
		if err != nil {
			return nil, errors.Wrap(err, "extract error detail value")
		}

		switch value.(type) {
		case *clientsv1.ErrorRecordMismatch:
			return nil, ErrRecordMismatch
		}
	}

	// Nothing juicy, return as-is.
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

var (
	ErrNotFound       = errors.New("not found")
	ErrRecordMismatch = errors.New("record mismatch")
)
