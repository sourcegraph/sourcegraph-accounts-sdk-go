package sams

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1/clientsv1connect"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/services"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"
)

// ServiceAccessTokensServiceV1 provides client methods to interact with the
// ServiceAccessTokensService API v1.
type ServiceAccessTokensServiceV1 struct {
	client *ClientV1
}

func (s *ServiceAccessTokensServiceV1) newClient(ctx context.Context) clientsv1connect.ServiceAccessTokensServiceClient {
	return clientsv1connect.NewServiceAccessTokensServiceClient(
		oauth2.NewClient(ctx, s.client.tokenSource),
		s.client.gRPCURL(),
		connect.WithInterceptors(s.client.defaultInterceptors...),
	)
}

// CreateServiceAccessTokenOptions represents the optional parameters for creating a service access token.
type CreateServiceAccessTokenOptions struct {
	// The human-friendly name of the token (optional).
	DisplayName string
	// The time the token will expire (optional, defaults to never expire).
	ExpiresAt *time.Time
}

// CreateServiceAccessTokenResponse represents the response from creating a service access token.
type CreateServiceAccessTokenResponse struct {
	Token  *clientsv1.ServiceAccessToken
	Secret string
}

// CreateServiceAccessToken creates a new service access token.
//
// Required scope: sams::service_access_token::write
func (s *ServiceAccessTokensServiceV1) CreateServiceAccessToken(ctx context.Context, service services.Service, tokenScopes []scopes.Scope, userID string, opts CreateServiceAccessTokenOptions) (*CreateServiceAccessTokenResponse, error) {
	if service == "" {
		return nil, errors.New("service cannot be empty")
	}
	if len(tokenScopes) == 0 {
		return nil, errors.New("scopes cannot be empty")
	}
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	token := &clientsv1.ServiceAccessToken{
		Service:     string(service),
		Scopes:      scopes.ToStrings(tokenScopes),
		UserId:      userID,
		DisplayName: opts.DisplayName,
	}

	if opts.ExpiresAt != nil {
		token.ExpireTime = timestamppb.New(*opts.ExpiresAt)
	}

	req := &clientsv1.CreateServiceAccessTokenRequest{Token: token}
	client := s.newClient(ctx)
	resp, err := parseResponseAndError(client.CreateServiceAccessToken(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}

	return &CreateServiceAccessTokenResponse{
		Token:  resp.Msg.Token,
		Secret: resp.Msg.Secret,
	}, nil
}

// ListServiceAccessTokensOptions represents the options for listing service access tokens.
type ListServiceAccessTokensOptions struct {
	// Maximum number of results to return (optional).
	PageSize int32
	// Page token for pagination (optional).
	PageToken string
	// Service filter (optional).
	Service string
	// User ID filter (optional).
	UserID string
	// Whether to include expired tokens (optional).
	ShowExpired bool
}

// ListServiceAccessTokens returns a list of service access tokens in reverse chronological
// order by creation time.
//
// Required scope: sams::service_access_token::read
func (s *ServiceAccessTokensServiceV1) ListServiceAccessTokens(ctx context.Context, opts ListServiceAccessTokensOptions) ([]*clientsv1.ServiceAccessToken, error) {
	req := &clientsv1.ListServiceAccessTokensRequest{
		PageSize:  opts.PageSize,
		PageToken: opts.PageToken,
	}

	// Build filters
	var filters []*clientsv1.ListServiceAccessTokensFilter
	if opts.Service != "" {
		filters = append(filters, &clientsv1.ListServiceAccessTokensFilter{
			Filter: &clientsv1.ListServiceAccessTokensFilter_Service{Service: opts.Service},
		})
	}
	if opts.UserID != "" {
		filters = append(filters, &clientsv1.ListServiceAccessTokensFilter{
			Filter: &clientsv1.ListServiceAccessTokensFilter_UserId{UserId: opts.UserID},
		})
	}
	if opts.ShowExpired {
		filters = append(filters, &clientsv1.ListServiceAccessTokensFilter{
			Filter: &clientsv1.ListServiceAccessTokensFilter_ShowExpired{ShowExpired: opts.ShowExpired},
		})
	}
	req.Filters = filters

	client := s.newClient(ctx)
	resp, err := parseResponseAndError(client.ListServiceAccessTokens(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}
	return resp.Msg.GetTokens(), nil
}

// RevokeServiceAccessToken revokes the specified service access token.
//
// Required scope: sams::service_access_token::delete
func (s *ServiceAccessTokensServiceV1) RevokeServiceAccessToken(ctx context.Context, tokenID string) error {
	if tokenID == "" {
		return errors.New("token ID cannot be empty")
	}

	req := &clientsv1.RevokeServiceAccessTokenRequest{Id: tokenID}
	client := s.newClient(ctx)
	_, err := parseResponseAndError(client.RevokeServiceAccessToken(ctx, connect.NewRequest(req)))
	return err
}
