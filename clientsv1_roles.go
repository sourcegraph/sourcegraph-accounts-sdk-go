package sams

import (
	"context"
	"slices"

	"connectrpc.com/connect"
	"golang.org/x/oauth2"

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

func (s *RolesServiceV1) newClient(ctx context.Context) clientsv1connect.RolesServiceClient {
	return clientsv1connect.NewRolesServiceClient(
		oauth2.NewClient(ctx, s.client.tokenSource),
		s.client.gRPCURL(),
		connect.WithInterceptors(s.client.defaultInterceptors...),
	)
}

// RegisterResourcesMetadata is the metadata for registering resources.
type RegisterResourcesMetadata struct {
	ResourceType roles.ResourceType
	Revision     uuid.UUID
}

func (r RegisterResourcesMetadata) validate() error {
	if !slices.Contains(roles.ResourceTypes(), r.ResourceType) {
		return errors.New("invalid resource type")
	}
	return nil
}

// RegisterRoleResources registers the resources for a given resource type.
// `resourcesIterator` is a function that returns a list of resources to register.
// The function is invoked repeatedly until it produces an empty slice or an error.
//
// Required scope: sams::roles.resources::write
func (s *RolesServiceV1) RegisterRoleResources(ctx context.Context, metadata RegisterResourcesMetadata, resourcesIterator func() ([]*clientsv1.RoleResource, error)) (uint64, error) {
	err := metadata.validate()
	if err != nil {
		return 0, errors.Wrap(err, "invalid metadata")
	}

	client := s.newClient(ctx)
	stream := client.RegisterRoleResources(ctx)
	err = stream.Send(&clientsv1.RegisterRoleResourcesRequest{
		Payload: &clientsv1.RegisterRoleResourcesRequest_Metadata{
			Metadata: &clientsv1.RegisterRoleResourcesRequestMetadata{
				ResourceType: string(metadata.ResourceType),
				Revision:     metadata.Revision.String(),
			},
		},
	})
	if err != nil {
		return 0, errors.Wrap(err, "failed to send metadata")
	}
	for {
		resources, err := resourcesIterator()
		if err != nil {
			return 0, errors.Wrap(err, "failed to get resources")
		}
		if len(resources) == 0 {
			break
		}
		err = stream.Send(&clientsv1.RegisterRoleResourcesRequest{
			Payload: &clientsv1.RegisterRoleResourcesRequest_Resources_{
				Resources: &clientsv1.RegisterRoleResourcesRequest_Resources{
					Resources: resources,
				},
			},
		})
		if err != nil {
			return 0, errors.Wrap(err, "failed to send resources")
		}
	}
	resp, err := stream.CloseAndReceive()
	if err != nil {
		return 0, errors.Wrap(err, "failed to close stream")
	}
	return resp.Msg.GetResourceCount(), nil
}
