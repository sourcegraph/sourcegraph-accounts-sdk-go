package clientcredentials

import (
	"context"
	"net/http"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"go.opentelemetry.io/otel/trace"
)

// See clientcredentials.NewHTTPMiddleware.
type HTTPAuthenticator struct {
	logger       log.Logger
	introspector TokenIntrospector
}

// NewHTTPAuthenticator provides a factor for auth middleware that uses SAMS
// service-to-service tokens to authenticate the requests.
//
// If you are using ConnectRPC, use clientcredentials.NewInterceptor() instead.
// HTTPAuthenticator should only be used for non-ConnectRPC APIs.
func NewHTTPAuthenticator(logger log.Logger, introspector TokenIntrospector) *HTTPAuthenticator {
	return &HTTPAuthenticator{
		logger:       logger,
		introspector: introspector,
	}
}

// RequireScopes performs an authorization check on the incoming HTTP request.
// It will return a 401 if the request does not have a valid SAMS access token,
// or a 403 if the token is valid but is missing ANY of the required scopes.
func (a *HTTPAuthenticator) RequireScopes(requiredScopes scopes.Scopes, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := a.logger.WithTrace(log.TraceContext{
			TraceID: trace.SpanContextFromContext(ctx).TraceID().String(),
			SpanID:  trace.SpanContextFromContext(ctx).SpanID().String(),
		})

		token, err := extractBearerContents(r.Header)
		if err != nil || token == "" {
			logger.Warn("error extracting bearer token", log.Error(err))
			const unauthorized = http.StatusUnauthorized
			http.Error(w, http.StatusText(unauthorized), unauthorized)
			return
		}

		result, err := a.introspector.IntrospectToken(ctx, token)
		if err != nil || result == nil {
			code := http.StatusInternalServerError
			if errors.Is(err, context.Canceled) {
				code = http.StatusBadRequest
				logger.Warn("error introspecting token", log.Error(err))
			} else {
				logger.Error("error introspecting token", log.Error(err))
			}
			http.Error(w, http.StatusText(code), code)
			return
		}

		if !result.Active {
			logger.Warn("attempt to authenticate with inactive SAMS token",
				log.String("client", result.ClientID))
			const unauthorized = http.StatusUnauthorized
			http.Error(w, "Unauthorized: Inactive token", unauthorized)
			return
		}

		if result.UserID != "" {
			logger.Warn("attempt to authenticate using SAMS token with user ID",
				log.String("client", result.ClientID),
				log.String("userID", result.UserID))
			http.Error(w, "Forbidden: User tokens not allowed", http.StatusForbidden)
			return
		}

		// Check for our required scope.
		for _, required := range requiredScopes {
			if !result.Scopes.Match(required) {
				logger.Warn("attempt to authenticate using SAMS token without required scope",
					log.Strings("gotScopes", scopes.ToStrings(result.Scopes)),
					log.String("requiredScope", string(required)))
				http.Error(w, "Forbidden: Missing required scope", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
