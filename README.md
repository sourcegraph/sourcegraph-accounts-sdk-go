# Sourcegraph Accounts SDK for Go

This repository contains the Go SDK for integrating with [Sourcegraph Accounts Management System (SAMS)](https://handbook.sourcegraph.com/departments/engineering/teams/core-services/sams/).

```zsh
go get github.com/sourcegraph/sourcegraph-accounts-sdk-go
```

> [!note]
> Please submit all issues to the [`sourcegraph/sourcegraph-accounts` repository](https://github.com/sourcegraph/sourcegraph-accounts/issues) 

## Authentication

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"

	samsauth "github.com/sourcegraph/sourcegraph-accounts-sdk-go/auth"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
)

func main() {
	stateSetter := func(w http.ResponseWriter, r *http.Request, state string) error {
		// TODO: Save to session data. For best security measures, store in a backend
		// component (e.g. Redis) or encrypted cookie. Unencrypted cookies are strongly
		// discouraged as they can be tampered with.
		return nil
	}
	stateGetter := func(r *http.Request) (string, error) {
		// TODO: Retrieve from session data.
		return "", nil
	}
	stateDeleter := func(w http.ResponseWriter, r *http.Request) {
		// TODO: Delete from session data.
	}

	samsauthHandler, err := samsauth.NewHandler(
		samsauth.Config{
			Issuer:         "https://accounts.sourcegraph.com",
			ClientID:       os.Getenv("SAMS_CLIENT_ID"),
			ClientSecret:   os.Getenv("SAMS_CLIENT_SECRET"),
			RequestScopes:  []scopes.Scope{scopes.OpenID, scopes.Email, scopes.Profile},
			RedirectURI:    os.Getenv("SAMS_REDIRECT_URI"),
			FailureHandler: samsauth.DefaultFailureHandler,
			StateSetter:    stateSetter,
			StateGetter:    stateGetter,
			StateDeleter:   stateDeleter,
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
