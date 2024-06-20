// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: clients/v1/clients.proto

package clientsv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// UsersServiceName is the fully-qualified name of the UsersService service.
	UsersServiceName = "clients.v1.UsersService"
	// SessionsServiceName is the fully-qualified name of the SessionsService service.
	SessionsServiceName = "clients.v1.SessionsService"
	// TokensServiceName is the fully-qualified name of the TokensService service.
	TokensServiceName = "clients.v1.TokensService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// UsersServiceGetUserProcedure is the fully-qualified name of the UsersService's GetUser RPC.
	UsersServiceGetUserProcedure = "/clients.v1.UsersService/GetUser"
	// UsersServiceGetUsersProcedure is the fully-qualified name of the UsersService's GetUsers RPC.
	UsersServiceGetUsersProcedure = "/clients.v1.UsersService/GetUsers"
	// UsersServiceGetUserRolesProcedure is the fully-qualified name of the UsersService's GetUserRoles
	// RPC.
	UsersServiceGetUserRolesProcedure = "/clients.v1.UsersService/GetUserRoles"
	// SessionsServiceGetSessionProcedure is the fully-qualified name of the SessionsService's
	// GetSession RPC.
	SessionsServiceGetSessionProcedure = "/clients.v1.SessionsService/GetSession"
	// SessionsServiceSignOutSessionProcedure is the fully-qualified name of the SessionsService's
	// SignOutSession RPC.
	SessionsServiceSignOutSessionProcedure = "/clients.v1.SessionsService/SignOutSession"
	// TokensServiceIntrospectTokenProcedure is the fully-qualified name of the TokensService's
	// IntrospectToken RPC.
	TokensServiceIntrospectTokenProcedure = "/clients.v1.TokensService/IntrospectToken"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	usersServiceServiceDescriptor                 = v1.File_clients_v1_clients_proto.Services().ByName("UsersService")
	usersServiceGetUserMethodDescriptor           = usersServiceServiceDescriptor.Methods().ByName("GetUser")
	usersServiceGetUsersMethodDescriptor          = usersServiceServiceDescriptor.Methods().ByName("GetUsers")
	usersServiceGetUserRolesMethodDescriptor      = usersServiceServiceDescriptor.Methods().ByName("GetUserRoles")
	sessionsServiceServiceDescriptor              = v1.File_clients_v1_clients_proto.Services().ByName("SessionsService")
	sessionsServiceGetSessionMethodDescriptor     = sessionsServiceServiceDescriptor.Methods().ByName("GetSession")
	sessionsServiceSignOutSessionMethodDescriptor = sessionsServiceServiceDescriptor.Methods().ByName("SignOutSession")
	tokensServiceServiceDescriptor                = v1.File_clients_v1_clients_proto.Services().ByName("TokensService")
	tokensServiceIntrospectTokenMethodDescriptor  = tokensServiceServiceDescriptor.Methods().ByName("IntrospectToken")
)

// UsersServiceClient is a client for the clients.v1.UsersService service.
type UsersServiceClient interface {
	// GetUser returns the SAMS user with the given query. It returns connect.CodeNotFound
	// if no such user exists.
	//
	// Required scope: profile
	GetUser(context.Context, *connect.Request[v1.GetUserRequest]) (*connect.Response[v1.GetUserResponse], error)
	// GetUsers returns the list of SAMS users matching the provided IDs.
	//
	// NOTE: It silently ignores any invalid user IDs, i.e. the length of the return
	// slice may be less than the length of the input slice.
	//
	// Required scopes: profile
	GetUsers(context.Context, *connect.Request[v1.GetUsersRequest]) (*connect.Response[v1.GetUsersResponse], error)
	// GetUserRoles returns all roles that have been assigned to the SAMS user
	// with the given ID and scoped by the service.
	//
	// Required scopes: sams::user.roles::read
	GetUserRoles(context.Context, *connect.Request[v1.GetUserRolesRequest]) (*connect.Response[v1.GetUserRolesResponse], error)
}

// NewUsersServiceClient constructs a client for the clients.v1.UsersService service. By default, it
// uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewUsersServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) UsersServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &usersServiceClient{
		getUser: connect.NewClient[v1.GetUserRequest, v1.GetUserResponse](
			httpClient,
			baseURL+UsersServiceGetUserProcedure,
			connect.WithSchema(usersServiceGetUserMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		getUsers: connect.NewClient[v1.GetUsersRequest, v1.GetUsersResponse](
			httpClient,
			baseURL+UsersServiceGetUsersProcedure,
			connect.WithSchema(usersServiceGetUsersMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		getUserRoles: connect.NewClient[v1.GetUserRolesRequest, v1.GetUserRolesResponse](
			httpClient,
			baseURL+UsersServiceGetUserRolesProcedure,
			connect.WithSchema(usersServiceGetUserRolesMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// usersServiceClient implements UsersServiceClient.
type usersServiceClient struct {
	getUser      *connect.Client[v1.GetUserRequest, v1.GetUserResponse]
	getUsers     *connect.Client[v1.GetUsersRequest, v1.GetUsersResponse]
	getUserRoles *connect.Client[v1.GetUserRolesRequest, v1.GetUserRolesResponse]
}

// GetUser calls clients.v1.UsersService.GetUser.
func (c *usersServiceClient) GetUser(ctx context.Context, req *connect.Request[v1.GetUserRequest]) (*connect.Response[v1.GetUserResponse], error) {
	return c.getUser.CallUnary(ctx, req)
}

// GetUsers calls clients.v1.UsersService.GetUsers.
func (c *usersServiceClient) GetUsers(ctx context.Context, req *connect.Request[v1.GetUsersRequest]) (*connect.Response[v1.GetUsersResponse], error) {
	return c.getUsers.CallUnary(ctx, req)
}

// GetUserRoles calls clients.v1.UsersService.GetUserRoles.
func (c *usersServiceClient) GetUserRoles(ctx context.Context, req *connect.Request[v1.GetUserRolesRequest]) (*connect.Response[v1.GetUserRolesResponse], error) {
	return c.getUserRoles.CallUnary(ctx, req)
}

// UsersServiceHandler is an implementation of the clients.v1.UsersService service.
type UsersServiceHandler interface {
	// GetUser returns the SAMS user with the given query. It returns connect.CodeNotFound
	// if no such user exists.
	//
	// Required scope: profile
	GetUser(context.Context, *connect.Request[v1.GetUserRequest]) (*connect.Response[v1.GetUserResponse], error)
	// GetUsers returns the list of SAMS users matching the provided IDs.
	//
	// NOTE: It silently ignores any invalid user IDs, i.e. the length of the return
	// slice may be less than the length of the input slice.
	//
	// Required scopes: profile
	GetUsers(context.Context, *connect.Request[v1.GetUsersRequest]) (*connect.Response[v1.GetUsersResponse], error)
	// GetUserRoles returns all roles that have been assigned to the SAMS user
	// with the given ID and scoped by the service.
	//
	// Required scopes: sams::user.roles::read
	GetUserRoles(context.Context, *connect.Request[v1.GetUserRolesRequest]) (*connect.Response[v1.GetUserRolesResponse], error)
}

// NewUsersServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewUsersServiceHandler(svc UsersServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	usersServiceGetUserHandler := connect.NewUnaryHandler(
		UsersServiceGetUserProcedure,
		svc.GetUser,
		connect.WithSchema(usersServiceGetUserMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	usersServiceGetUsersHandler := connect.NewUnaryHandler(
		UsersServiceGetUsersProcedure,
		svc.GetUsers,
		connect.WithSchema(usersServiceGetUsersMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	usersServiceGetUserRolesHandler := connect.NewUnaryHandler(
		UsersServiceGetUserRolesProcedure,
		svc.GetUserRoles,
		connect.WithSchema(usersServiceGetUserRolesMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/clients.v1.UsersService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case UsersServiceGetUserProcedure:
			usersServiceGetUserHandler.ServeHTTP(w, r)
		case UsersServiceGetUsersProcedure:
			usersServiceGetUsersHandler.ServeHTTP(w, r)
		case UsersServiceGetUserRolesProcedure:
			usersServiceGetUserRolesHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedUsersServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedUsersServiceHandler struct{}

func (UnimplementedUsersServiceHandler) GetUser(context.Context, *connect.Request[v1.GetUserRequest]) (*connect.Response[v1.GetUserResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("clients.v1.UsersService.GetUser is not implemented"))
}

func (UnimplementedUsersServiceHandler) GetUsers(context.Context, *connect.Request[v1.GetUsersRequest]) (*connect.Response[v1.GetUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("clients.v1.UsersService.GetUsers is not implemented"))
}

func (UnimplementedUsersServiceHandler) GetUserRoles(context.Context, *connect.Request[v1.GetUserRolesRequest]) (*connect.Response[v1.GetUserRolesResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("clients.v1.UsersService.GetUserRoles is not implemented"))
}

// SessionsServiceClient is a client for the clients.v1.SessionsService service.
type SessionsServiceClient interface {
	// GetSession returns the SAMS session with the given ID. It returns
	// connect.CodeNotFound if no such session exists. The session's `User` field is
	// populated if the session is authenticated by a user.
	//
	// Required scope: sams::session::read
	GetSession(context.Context, *connect.Request[v1.GetSessionRequest]) (*connect.Response[v1.GetSessionResponse], error)
	// SignOutSession revokes the authenticated state of the session with the given
	// ID for the given user. It does not return error if the session does not exist
	// or is not authenticated. It returns clientsv1.ErrorRecordMismatch in the
	// error detail if the session is authenticated by a different user than the
	// given user.
	//
	// Required scope: sams::session::write
	SignOutSession(context.Context, *connect.Request[v1.SignOutSessionRequest]) (*connect.Response[v1.SignOutSessionResponse], error)
}

// NewSessionsServiceClient constructs a client for the clients.v1.SessionsService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewSessionsServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) SessionsServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &sessionsServiceClient{
		getSession: connect.NewClient[v1.GetSessionRequest, v1.GetSessionResponse](
			httpClient,
			baseURL+SessionsServiceGetSessionProcedure,
			connect.WithSchema(sessionsServiceGetSessionMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		signOutSession: connect.NewClient[v1.SignOutSessionRequest, v1.SignOutSessionResponse](
			httpClient,
			baseURL+SessionsServiceSignOutSessionProcedure,
			connect.WithSchema(sessionsServiceSignOutSessionMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// sessionsServiceClient implements SessionsServiceClient.
type sessionsServiceClient struct {
	getSession     *connect.Client[v1.GetSessionRequest, v1.GetSessionResponse]
	signOutSession *connect.Client[v1.SignOutSessionRequest, v1.SignOutSessionResponse]
}

// GetSession calls clients.v1.SessionsService.GetSession.
func (c *sessionsServiceClient) GetSession(ctx context.Context, req *connect.Request[v1.GetSessionRequest]) (*connect.Response[v1.GetSessionResponse], error) {
	return c.getSession.CallUnary(ctx, req)
}

// SignOutSession calls clients.v1.SessionsService.SignOutSession.
func (c *sessionsServiceClient) SignOutSession(ctx context.Context, req *connect.Request[v1.SignOutSessionRequest]) (*connect.Response[v1.SignOutSessionResponse], error) {
	return c.signOutSession.CallUnary(ctx, req)
}

// SessionsServiceHandler is an implementation of the clients.v1.SessionsService service.
type SessionsServiceHandler interface {
	// GetSession returns the SAMS session with the given ID. It returns
	// connect.CodeNotFound if no such session exists. The session's `User` field is
	// populated if the session is authenticated by a user.
	//
	// Required scope: sams::session::read
	GetSession(context.Context, *connect.Request[v1.GetSessionRequest]) (*connect.Response[v1.GetSessionResponse], error)
	// SignOutSession revokes the authenticated state of the session with the given
	// ID for the given user. It does not return error if the session does not exist
	// or is not authenticated. It returns clientsv1.ErrorRecordMismatch in the
	// error detail if the session is authenticated by a different user than the
	// given user.
	//
	// Required scope: sams::session::write
	SignOutSession(context.Context, *connect.Request[v1.SignOutSessionRequest]) (*connect.Response[v1.SignOutSessionResponse], error)
}

// NewSessionsServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewSessionsServiceHandler(svc SessionsServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	sessionsServiceGetSessionHandler := connect.NewUnaryHandler(
		SessionsServiceGetSessionProcedure,
		svc.GetSession,
		connect.WithSchema(sessionsServiceGetSessionMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	sessionsServiceSignOutSessionHandler := connect.NewUnaryHandler(
		SessionsServiceSignOutSessionProcedure,
		svc.SignOutSession,
		connect.WithSchema(sessionsServiceSignOutSessionMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/clients.v1.SessionsService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case SessionsServiceGetSessionProcedure:
			sessionsServiceGetSessionHandler.ServeHTTP(w, r)
		case SessionsServiceSignOutSessionProcedure:
			sessionsServiceSignOutSessionHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedSessionsServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedSessionsServiceHandler struct{}

func (UnimplementedSessionsServiceHandler) GetSession(context.Context, *connect.Request[v1.GetSessionRequest]) (*connect.Response[v1.GetSessionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("clients.v1.SessionsService.GetSession is not implemented"))
}

func (UnimplementedSessionsServiceHandler) SignOutSession(context.Context, *connect.Request[v1.SignOutSessionRequest]) (*connect.Response[v1.SignOutSessionResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("clients.v1.SessionsService.SignOutSession is not implemented"))
}

// TokensServiceClient is a client for the clients.v1.TokensService service.
type TokensServiceClient interface {
	// IntrospectToken takes a SAMS access token and returns relevant metadata.
	//
	// 🚨SECURITY: SAMS will return a successful result if the token is valid, but
	// is no longer active. It is critical that the caller not honor tokens where
	// `.Active == false`.
	IntrospectToken(context.Context, *connect.Request[v1.IntrospectTokenRequest]) (*connect.Response[v1.IntrospectTokenResponse], error)
}

// NewTokensServiceClient constructs a client for the clients.v1.TokensService service. By default,
// it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and
// sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC()
// or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewTokensServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) TokensServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &tokensServiceClient{
		introspectToken: connect.NewClient[v1.IntrospectTokenRequest, v1.IntrospectTokenResponse](
			httpClient,
			baseURL+TokensServiceIntrospectTokenProcedure,
			connect.WithSchema(tokensServiceIntrospectTokenMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// tokensServiceClient implements TokensServiceClient.
type tokensServiceClient struct {
	introspectToken *connect.Client[v1.IntrospectTokenRequest, v1.IntrospectTokenResponse]
}

// IntrospectToken calls clients.v1.TokensService.IntrospectToken.
func (c *tokensServiceClient) IntrospectToken(ctx context.Context, req *connect.Request[v1.IntrospectTokenRequest]) (*connect.Response[v1.IntrospectTokenResponse], error) {
	return c.introspectToken.CallUnary(ctx, req)
}

// TokensServiceHandler is an implementation of the clients.v1.TokensService service.
type TokensServiceHandler interface {
	// IntrospectToken takes a SAMS access token and returns relevant metadata.
	//
	// 🚨SECURITY: SAMS will return a successful result if the token is valid, but
	// is no longer active. It is critical that the caller not honor tokens where
	// `.Active == false`.
	IntrospectToken(context.Context, *connect.Request[v1.IntrospectTokenRequest]) (*connect.Response[v1.IntrospectTokenResponse], error)
}

// NewTokensServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewTokensServiceHandler(svc TokensServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	tokensServiceIntrospectTokenHandler := connect.NewUnaryHandler(
		TokensServiceIntrospectTokenProcedure,
		svc.IntrospectToken,
		connect.WithSchema(tokensServiceIntrospectTokenMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/clients.v1.TokensService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case TokensServiceIntrospectTokenProcedure:
			tokensServiceIntrospectTokenHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedTokensServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedTokensServiceHandler struct{}

func (UnimplementedTokensServiceHandler) IntrospectToken(context.Context, *connect.Request[v1.IntrospectTokenRequest]) (*connect.Response[v1.IntrospectTokenResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("clients.v1.TokensService.IntrospectToken is not implemented"))
}
