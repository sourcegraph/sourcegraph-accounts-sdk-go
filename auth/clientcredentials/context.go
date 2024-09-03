package clientcredentials

import (
	"context"
	"strings"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
)

type contextKey int

const (
	clientInfoKey contextKey = iota
)

type ClientInfo struct {
	ClientID       string
	TokenExpiresAt time.Time
	TokenScopes    scopes.Scopes
}

// LogFields represents a standard log representation of a client, for use in
// propagting in loggers for auditing purposes. It is safe to use on a nil
// *ClientInfo.
func (c *ClientInfo) LogFields() []log.Field {
	if c == nil {
		return []log.Field{log.Stringp("client", nil)}
	}
	return []log.Field{
		log.String("client.clientID", c.ClientID),
		log.Time("client.tokenExpiresAt", c.TokenExpiresAt),
		log.String("client.tokenScopes", strings.Join(scopes.ToStrings(c.TokenScopes), " ")),
	}
}

// ClientInfoFromContext returns client info from the given context. This is
// generally set by clientcredentials.Interceptor.
func ClientInfoFromContext(ctx context.Context) *ClientInfo {
	return ctx.Value(clientInfoKey).(*ClientInfo)
}

// WithClientInfo returns a new context with the given client info.
func WithClientInfo(ctx context.Context, info *ClientInfo) context.Context {
	return context.WithValue(ctx, clientInfoKey, info)
}
