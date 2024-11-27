package clientcredentials

import (
	"context"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var tracer = otel.Tracer("sams/auth/clientcredentials")

type TokenIntrospector interface {
	// IntrospectToken takes a SAMS access token and returns relevant metadata.
	// This is generally implemented by *sams.TokensServiceV1.
	//
	// 🚨 SECURITY: SAMS will return a successful result if the token is valid, but
	// is no longer active. It is critical that the caller not honor tokens where
	// `.Active == false`.
	IntrospectToken(ctx context.Context, token string) (*sams.IntrospectTokenResponse, error)
}

func extractBearerContents(h http.Header) (string, error) {
	authHeader := h.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no token provided in Authorization header")
	}
	typ := strings.SplitN(authHeader, " ", 2)
	if len(typ) != 2 {
		return "", errors.New("token type missing in Authorization header")
	}
	if !strings.EqualFold(typ[0], "bearer") {
		return "", errors.Newf("invalid token type %s in Authorization header", typ[0])
	}
	return typ[1], nil
}
