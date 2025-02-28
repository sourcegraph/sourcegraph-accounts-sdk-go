package sams

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/structpb"

	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1/clientsv1connect"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// UsersServiceV1 provides client methods to interact with the UsersService API
// v1.
type UsersServiceV1 struct {
	client *ClientV1
}

func (s *UsersServiceV1) newClient() clientsv1connect.UsersServiceClient {
	return clientsv1connect.NewUsersServiceClient(
		s.client.httpClient(),
		s.client.gRPCURL(),
		connect.WithInterceptors(s.client.defaultInterceptors...),
	)
}

// GetUserByID returns the SAMS user with the given ID. It returns ErrNotFound
// if no such user exists.
//
// Required scope: profile
func (s *UsersServiceV1) GetUserByID(ctx context.Context, id string) (*clientsv1.User, error) {
	req := &clientsv1.GetUserRequest{Id: id}
	client := s.newClient()
	resp, err := parseResponseAndError(client.GetUser(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}
	return resp.Msg.User, nil
}

// GetUserByEmail returns the SAMS user with the given verified email. It returns
// ErrNotFound if no such user exists.
//
// Required scope: profile
func (s *UsersServiceV1) GetUserByEmail(ctx context.Context, email string) (*clientsv1.User, error) {
	req := &clientsv1.GetUserRequest{Email: email}
	client := s.newClient()
	resp, err := parseResponseAndError(client.GetUser(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}
	return resp.Msg.User, nil
}

// GetUsersByIDs returns the list of SAMS users matching the provided IDs.
//
// NOTE: It silently ignores any invalid user IDs, i.e. the length of the return
// slice may be less than the length of the input slice.
//
// Required scopes: profile
func (s *UsersServiceV1) GetUsersByIDs(ctx context.Context, ids []string) ([]*clientsv1.User, error) {
	req := &clientsv1.GetUsersRequest{Ids: ids}
	client := s.newClient()
	resp, err := parseResponseAndError(client.GetUsers(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}
	return resp.Msg.GetUsers(), nil
}

// GetUserRolesByID returns all roles that have been assigned to the SAMS user
// with the given ID and scoped by the service.
//
// Required scopes: sams::user.roles::read
func (s *UsersServiceV1) GetUserRolesByID(ctx context.Context, userID, service string) ([]*clientsv1.Role, error) {
	req := &clientsv1.GetUserRolesRequest{
		Id:      userID,
		Service: service,
	}
	client := s.newClient()
	resp, err := parseResponseAndError(client.GetUserRoles(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}
	return resp.Msg.GetUserRoles(), nil
}

// GetUserMetadata returns the metadata associated with the given user ID and
// metadata namespaces.
//
// Required scopes: sams::user.metadata.${NAMESPACE}::read for each of the
// requested namespaces.
func (s *UsersServiceV1) GetUserMetadata(ctx context.Context, userID string, namespaces []string) ([]*clientsv1.UserServiceMetadata, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}
	if len(namespaces) == 0 {
		return nil, errors.New("at least one namespace must be provided")
	}

	req := &clientsv1.GetUserMetadataRequest{
		Id:         userID,
		Namespaces: namespaces,
	}
	client := s.newClient()
	resp, err := parseResponseAndError(client.GetUserMetadata(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}
	return resp.Msg.GetMetadata(), nil
}

// UpdateUserMetadata updates the metadata associated with the given user ID
// and metadata namespace.
//
// Required scopes: sams::user.metadata.${NAMESPACE}::read for the namespace
// being updated.
func (s *UsersServiceV1) UpdateUserMetadata(ctx context.Context, userID, namespace string, metadata map[string]any) (*clientsv1.UserServiceMetadata, error) {
	if userID == "" || namespace == "" {
		return nil, errors.New("user ID and namespace cannot be empty")
	}

	md, err := structpb.NewStruct(metadata)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal user metadata")
	}
	req := &clientsv1.UpdateUserMetadataRequest{
		Metadata: &clientsv1.UserServiceMetadata{
			UserId:    userID,
			Namespace: namespace,
			Metadata:  md,
		},
	}
	client := s.newClient()
	resp, err := parseResponseAndError(client.UpdateUserMetadata(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}
	return resp.Msg.GetMetadata(), nil
}
