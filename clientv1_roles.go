package sams

import (
	"context"
	"io"
	"slices"

	"connectrpc.com/connect"

	"github.com/google/uuid"
	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1/clientsv1connect"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/roles"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// RolesServiceV1 provides client methods to interact with the
// RolesService API v1.
type RolesServiceV1 struct {
	client *ClientV1
}

func (s *RolesServiceV1) newClient() clientsv1connect.RolesServiceClient {
	return clientsv1connect.NewRolesServiceClient(
		s.client.httpClient(),
		s.client.gRPCURL(),
		connect.WithInterceptors(s.client.defaultInterceptors...),
	)
}

// RegisterResourcesMetadata is the metadata for a set of resources to be registered.
type RegisterResourcesMetadata struct {
	ResourceType roles.ResourceType
}

func (r RegisterResourcesMetadata) validate() error {
	if !slices.Contains(roles.ResourceTypes(), r.ResourceType) {
		return errors.Newf("invalid resource type: %q", r.ResourceType)
	}
	return nil
}

// RegisterRoleResources registers the resources for a given resource type.
// `resourcesIterator` is a function that returns a page of resources to register.
// The function is invoked repeatedly until it produces an empty slice or an error.
// If another replica is already registering resources for the same resource type
// the function will return 0 with ErrAborted.
// ErrAborted means the request is safe to retry at a later time.
//
// Required scope: sams::roles.resources::write
func (s *RolesServiceV1) RegisterRoleResources(ctx context.Context, metadata RegisterResourcesMetadata, resourcesIterator func() ([]*clientsv1.RoleResource, error)) (uint64, error) {
	err := metadata.validate()
	if err != nil {
		return 0, errors.Wrap(err, "invalid metadata")
	}

	/// Generate a new revision for the request metadata.
	revision, err := uuid.NewV7()
	if err != nil {
		return 0, errors.Wrap(err, "failed to generate revision for request metadata")
	}

	client := s.newClient()
	stream := client.RegisterRoleResources(ctx)
	// Metadata must be submitted first in the stream.
	err = stream.Send(&clientsv1.RegisterRoleResourcesRequest{
		Payload: &clientsv1.RegisterRoleResourcesRequest_Metadata{
			Metadata: &clientsv1.RegisterRoleResourcesRequestMetadata{
				ResourceType: string(metadata.ResourceType),
				Revision:     revision.String(),
			},
		},
	})

	sendResources := true
	if err != nil {
		// The stream has been closed; skip sending resources.
		if errors.Is(err, io.EOF) {
			sendResources = false
		} else {
			return 0, errors.Wrap(err, "failed to send metadata")
		}
	}
	for sendResources {
		resources, err := resourcesIterator()
		if err != nil {
			return 0, errors.Wrap(err, "failed to get resources")
		}
		if len(resources) == 0 {
			sendResources = false
			continue
		}
		err = stream.Send(&clientsv1.RegisterRoleResourcesRequest{
			Payload: &clientsv1.RegisterRoleResourcesRequest_Resources_{
				Resources: &clientsv1.RegisterRoleResourcesRequest_Resources{
					Resources: resources,
				},
			},
		})
		if err != nil {
			// The stream has been closed, so we stop sending resources.
			if errors.Is(err, io.EOF) {
				sendResources = false
				continue
			}
			return 0, errors.Wrap(err, "failed to send resources")
		}
	}

	resp, err := parseResponseAndError(stream.CloseAndReceive())
	if err != nil {
		// Stream closed due to another replica registering the same resources.
		if errors.Is(err, ErrAborted) {
			return 0, nil
		}
		return 0, err
	}
	return resp.Msg.GetResourceCount(), nil
}
