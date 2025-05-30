package clientcredentials

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

// See clientcredentials.NewInterceptor.
type Interceptor struct {
	logger       log.Logger
	introspector TokenIntrospector
	extension    *protoimpl.ExtensionInfo
}

// NewInterceptor creates a serverside ConnectRPC interceptor that ensures every
// incoming request has a valid client credential token with the required scopes
// indicated in the RPC method options. When used, required scopes CANNOT be
// empty - if no scopes are required, declare a separate service that does not
// use this interceptor.
//
// To declare required SAMS scopes in your RPC, add the following to your proto
// schema:
//
//	extend google.protobuf.MethodOptions {
//		// The SAMS scopes required to use this RPC.
//		//
//		// The range 50000-99999 is reserved for internal use within individual organizations
//		// so you can use numbers in this range freely for in-house applications.
//		repeated string sams_required_scopes = 50001;
//	}
//
// In your RPCs, add the `(sams_required_scopes)` option as a comma-delimited
// list:
//
//	rpc GetUserRoles(GetUserRolesRequest) returns (GetUserRolesResponse) {
//		option (sams_required_scopes) = "sams::user.roles::read";
//	};
//
// This will generate a variable called `E_SamsRequiredScopes` in your generated
// proto bindings. This variable should be provided to NewInterceptor to allow
// it to identify where to source the required scopes from.
//
// The provided logger is used to record internal-server errors.
func NewInterceptor(
	logger log.Logger,
	introspector TokenIntrospector,
	methodOptionsRequiredScopesExtension *protoimpl.ExtensionInfo,
) *Interceptor {
	return &Interceptor{
		logger:       logger.Scoped("clientcredentials"),
		introspector: introspector,
		extension:    methodOptionsRequiredScopesExtension,
	}
}

var _ connect.Interceptor = (*Interceptor)(nil)

func (i *Interceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if req.Spec().IsClient {
			return next(ctx, req) // no-op for clients
		}
		requiredScopes, err := extractSchemaRequiredScopes(req.Spec(), i.extension)
		if err != nil {
			return nil, connectInternalError(ctx, i.logger, err, "internal schema error") // invalid schema is internal error
		}
		info, err := i.requireScope(ctx, req.Header(), requiredScopes)
		if err != nil {
			return nil, err
		}
		return next(WithClientInfo(ctx, info), req)
	}
}

func (i *Interceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		return next(ctx, spec) // no-op for clients
	}
}

func (i *Interceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		if conn.Spec().IsClient {
			return next(ctx, conn) // no-op for clients
		}
		requiredScopes, err := extractSchemaRequiredScopes(conn.Spec(), i.extension)
		if err != nil {
			return connectInternalError(ctx, i.logger, err, "internal schema error") // invalid schema is internal error
		}
		info, err := i.requireScope(ctx, conn.RequestHeader(), requiredScopes)
		if err != nil {
			return err
		}
		return next(WithClientInfo(ctx, info), conn)
	}
}

// requireScope ensures the request context has a valid SAMS M2M token
// with requiredScope. It returns a ConnectRPC status error suitable to be
// returned directly from a ConnectRPC implementation.
func (i *Interceptor) requireScope(ctx context.Context, headers http.Header, requiredScopes scopes.Scopes) (_ *ClientInfo, err error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "clientcredentials.requireScope")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, "check failed")
		}
		span.End()
	}()

	token, err := extractBearerContents(headers)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated,
			errors.Wrap(err, "invalid authorization header"))
	}

	result, err := i.introspector.IntrospectToken(ctx, token)
	if err != nil {
		return nil, connectInternalError(ctx, i.logger, err, "unable to validate token")
	}
	span.SetAttributes(
		attribute.String("client_id", result.ClientID),
		attribute.String("token_expires_at", result.ExpiresAt.String()),
		attribute.StringSlice("token_scopes", scopes.ToStrings(result.Scopes)))
	info := &ClientInfo{
		ClientID:       result.ClientID,
		TokenExpiresAt: result.ExpiresAt,
		TokenScopes:    result.Scopes,
	}

	// Active encapsulates whether the token is active, including expiration.
	if !result.Active {
		// Record detailed error in span, and return an opaque one
		span.SetAttributes(attribute.String("full_error", "inactive token"))
		return info, connect.NewError(connect.CodePermissionDenied, errors.New("permission denied"))
	}

	// Check for our required scope.
	for _, required := range requiredScopes {
		if !result.Scopes.Match(required) {
			err = errors.Newf("got scopes %+v, required: %+v", result.Scopes, requiredScopes)
			span.SetAttributes(attribute.String("full_error", err.Error()))
			return info, connect.NewError(connect.CodePermissionDenied,
				errors.Wrap(err, "insufficient scopes"))
		}
	}

	return info, nil
}

func extractSchemaRequiredScopes(spec connect.Spec, extension *protoimpl.ExtensionInfo) (scopes.Scopes, error) {
	method, ok := spec.Schema.(protoreflect.MethodDescriptor)
	if !ok {
		return nil, errors.Newf("expected protoreflect.MethodDescriptor, got %T", spec.Schema)
	}

	value := method.Options().ProtoReflect().Get(extension.TypeDescriptor())
	if !value.IsValid() {
		return nil, errors.Newf("extension field %s not valid", extension.TypeDescriptor().FullName())
	}
	list := value.List()

	requiredScopes := make(scopes.Scopes, list.Len())
	for i := 0; i < list.Len(); i++ {
		requiredScopes[i] = scopes.Scope(list.Get(i).String())
	}
	return requiredScopes, nil
}

// connectInternalError logs an error, adds it to the trace, and returns a connect
// error with a safe message.
func connectInternalError(ctx context.Context, logger log.Logger, err error, safeMsg string) error {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("safe_msg", safeMsg),
		attribute.String("full_error", err.Error()))
	span.SetStatus(otelcodes.Error, err.Error())

	logger = logger.WithTrace(log.TraceContext{
		TraceID: trace.SpanContextFromContext(ctx).TraceID().String(),
		SpanID:  trace.SpanContextFromContext(ctx).SpanID().String(),
	}).AddCallerSkip(1)

	// Log at different levels and return different codes depending on the type
	// of this unexpected error.
	if errors.Is(err, context.Canceled) {
		code := connect.CodeCanceled
		logger.Warn(safeMsg,
			log.String("code", code.String()),
			log.Error(err))
		return connect.NewError(code, errors.New(safeMsg))
	}
	code := connect.CodeInternal
	logger.Error(safeMsg,
		log.String("code", code.String()),
		log.Error(err))
	return connect.NewError(code, errors.New(safeMsg))
}
