package sams

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"

	"github.com/hashicorp/golang-lru/v2/expirable"
	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1/clientsv1connect"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
)

// TokensServiceV1 provides client methods to interact with the TokensService
// API v1.
type TokensServiceV1 struct {
	client *ClientV1
	// introspectTokenCache may be nil if not enabled.
	introspectTokenCache *expirable.LRU[string, *IntrospectTokenResponse]
}

func (s *TokensServiceV1) newClient(ctx context.Context) clientsv1connect.TokensServiceClient {
	return clientsv1connect.NewTokensServiceClient(
		oauth2.NewClient(ctx, s.client.tokenSource),
		s.client.gRPCURL(),
		connect.WithInterceptors(s.client.defaultInterceptors...),
	)
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
	if s.introspectTokenCache != nil {
		if cached, ok := s.introspectTokenCache.Get(token); ok && cached.ExpiresAt.Before(time.Now()) {
			trace.SpanFromContext(ctx).
				SetAttributes(attribute.Bool("sams.introspectToken.fromCache", true))
			return cached, nil
		}
	}
	trace.SpanFromContext(ctx).
		SetAttributes(attribute.Bool("sams.introspectToken.fromCache", false))

	req := &clientsv1.IntrospectTokenRequest{Token: token}
	client := s.newClient(ctx)
	resp, err := parseResponseAndError(client.IntrospectToken(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}

	tokenResponse := &IntrospectTokenResponse{
		Active:    resp.Msg.Active,
		Scopes:    scopes.ToScopes(resp.Msg.Scopes),
		ClientID:  resp.Msg.ClientId,
		ExpiresAt: resp.Msg.ExpiresAt.AsTime(),
	}
	if s.introspectTokenCache != nil && tokenResponse.ExpiresAt.After(time.Now()) {
		_ = s.introspectTokenCache.Add(token, tokenResponse)
	}
	return tokenResponse, nil
}
