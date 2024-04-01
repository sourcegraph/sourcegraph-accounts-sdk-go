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

type stateStore struct{
	// Authentication state is the unique identifier that is randomly-generated and
	// assigned to a particular authentication flow, they are used to prevent
	// authentication interception attacks and considered secrets, therefore it MUST
	// be stored in a backend component (e.g. Redis, database). The design of the
	// samsauth.StateStore interface explicitly disallowed storing state in the
	// cookie, as they can be tampered with when cookie values are stored
	// unencrypted.
}

func (s *stateStore) SetState(r *http.Request, state string) error {
	// TODO: Save to session data.
	return nil
}

func (s *stateStore) GetState(r *http.Request) (string, error) {
	// TODO: Retrieve from session data.
	return "", nil
}

func (s *stateStore) DeleteState(r *http.Request) {
	// TODO: Delete from session data.
}

func main() {
	samsauthHandler, err := samsauth.NewHandler(
		samsauth.Config{
			Issuer:         "https://accounts.sourcegraph.com",
			ClientID:       os.Getenv("SAMS_CLIENT_ID"),
			ClientSecret:   os.Getenv("SAMS_CLIENT_SECRET"),
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
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userInfo := samsauth.UserInfoFromContext(r.Context())
			// TODO: Save user info to somewhere.
		}),
	))

	// Continue setting up your server and use the mux.
}
```

## Clients API v1

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
