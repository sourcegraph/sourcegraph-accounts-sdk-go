# Sourcegraph Accounts SDK for Go

This repository contains the Go SDK for integrating with [Sourcegraph Accounts Management System (SAMS)](https://handbook.sourcegraph.com/departments/engineering/teams/core-services/sams/).

```zsh
go get github.com/sourcegraph/sourcegraph-accounts-sdk-go
```

> [!note]
> Please submit all issues to the [`sourcegraph/sourcegraph-accounts` repository](https://github.com/sourcegraph/sourcegraph-accounts/issues) 

## Authentication

The following example demonstrates how to use the SDK to set up user authentication flow with SAMS for your service.

In particular,

- The route `/auth/login` is where the user should be redirected to start a new authentication flow.
- The route `/auth/callback` is where the user will be redirected back to the service after completing the authentication on the SAMS side.

```go
package main

import (
	"log"
	"net/http"
	"os"

	samsauth "github.com/sourcegraph/sourcegraph-accounts-sdk-go/auth"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
)

type secretStore struct{
	// Authentication state is the unique identifier that is randomly-generated and
	// assigned to a particular authentication flow, they are used to prevent
	// authentication interception attacks and considered secrets, therefore it MUST
	// be stored in a backend component (e.g. Redis, database). The design of the
	// samsauth.StateStore interface explicitly disallowed storing state in the
	// cookie, as they can be tampered with when cookie values are stored
	// unencrypted.
	//
	// Authentication nonce is a unique identifier that is randomly-generated to
	// make sure the ID Token we get back from SAMS is intended for the same
	// authentication flow that we started. It is also a secret and MUST be stored
	// in a backend component.
}

func (s *secretStore) SetState(r *http.Request, state string) error {
	// TODO: Save state to session data.
	return nil
}

func (s *secretStore) GetState(r *http.Request) (string, error) {
	// TODO: Retrieve state from session data.
	return "", nil
}

func (s *secretStore) DeleteState(r *http.Request) {
	// TODO: Delete state from session data.
}

func (s *secretStore) SetNonce(r *http.Request, nonce string) error {
	// TODO: Save nonce to session data.
	return nil
}

func (s *secretStore) GetNonce(r *http.Request) (string, error) {
	// TODO: Retrieve nonce from session data.
	return "", nil
}

func (s *secretStore) DeleteNonce(r *http.Request) {
	// TODO: Delete nonce from session data.
}

func main() {
	samsauthHandler, err := samsauth.NewHandler(
		samsauth.Config{
			Issuer:         "https://accounts.sourcegraph.com",
			ClientID:       os.Getenv("SAMS_CLIENT_ID"),
			ClientSecret:   os.Getenv("SAMS_CLIENT_SECRET"),
			// RequestScopes needs to include all the scopes that the service needs to
			// access on behalf of the user. Scopes that are only used for Clients API are
			// not needed here.
			RequestScopes:  []scopes.Scope{scopes.OpenID, scopes.Email, scopes.Profile},
			RedirectURI:    os.Getenv("SAMS_REDIRECT_URI"),
			FailureHandler: samsauth.DefaultFailureHandler,
			StateStore:     &stateStore{},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/auth/login", samsauthHandler.LoginHandler())
	mux.Handle("/auth/callback", samsauthHandler.CallbackHandler(
		// The SAMS auth handler will handle the callback and complete the
		// authentication flow. And if successful, the `samsauth.UserInfo` will be
		// accessible from the request context.
		//
		// You can safely assume the `samsauth.UserInfo` will be present when this
		// user-supplied handler is being invoked. If any error is encountered, the
		// provided FailureHandler will be invoked instead.
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userInfo := samsauth.UserInfoFromContext(r.Context())
			// TODO: Save user info to somewhere.
		}),
	))

	// Continue setting up your server and use the mux.
}
```

## Clients API v1

The SAMS Clients API is for SAMS clients to obtain information directly from SAMS. For example,
authorizing a request based on the scopes attached to a token. Or looking up a user's profile
information based on the SAMS external account ID.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
)

func main() {
	samsClient, err := sams.NewClientV1(
		"https://accounts.sourcegraph.com",
		os.Getenv("SAMS_CLIENT_ID"),
		os.Getenv("SAMS_CLIENT_SECRET"),
		[]scopes.Scope{
			scopes.OpenID,
			scopes.Profile,
			scopes.Email,
			"sams::user.roles::read",
			"sams::session::read",
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	
	user, err := samsClient.Users().GetUserByID(context.Background(), "user-id")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(user)
}
```

## SAMS Accounts API

The SAMS Accounts API is for user-oriented operations like inspecting your own account details. These APIs are
much simpler in nature, as most integrations will make use of the Clients API. However, the Accounts API is
required if the service is not governing access based on the SAMS token scope, but instead using its own
authorization mechanism. e.g. governing access based on the SAMS external account ID.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2"
	accountsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/accounts/v1"
)

func main() {
	// e.g. the SAMS token prefixed with "sams_at_".
	rawToken := os.Getenv("SAMS_USER_ACCESS_TOKEN")

	// If you have the SAMS user's Refresh token, using the oauth2.TokenSource abstraction
	// will take care of creating short-lived access tokens as needed. But if you only have
	// the access token, you will need to use a StaticTokenSource instead.
	token := oauth2.Token{
		AccessToken: rawToken,
	}
	tokenSource := oauth2.StaticTokenSource(t)

	client := accountsv1.NewClient("https://accounts.sourcegraph.com", tokenSource)
	user, err := client.GetUser(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User Details: %+v", user)
}
```

## Development

[Buf](https://buf.build) and [Connect](https://connectrpc.com/) are used for gRPC and Protocol Buffers code generation.

```zsh
go install github.com/bufbuild/buf/cmd/buf@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
```

After making any changes to the `.proto` files, in the direction that contains the `buf.gen.yaml` file,  run:

```zsh
buf generate
```
