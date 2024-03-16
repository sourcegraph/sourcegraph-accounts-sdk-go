package sams

import (
	"context"

	"connectrpc.com/connect"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1/clientsv1connect"
)

// SessionsServiceV1 provides client methods to interact with the
// SessionsService API v1.
type SessionsServiceV1 struct {
	client *ClientV1
}

func (s *SessionsServiceV1) newClient(ctx context.Context) clientsv1connect.SessionsServiceClient {
	return clientsv1connect.NewSessionsServiceClient(
		oauth2.NewClient(ctx, s.client.tokenSource),
		s.client.gRPCURL(),
	)
}

// GetSessionByID returns the SAMS session with the given ID. It returns
// ErrNotFound if no such session exists. The session's `User` field is
// populated if the session is authenticated by a user.
func (s *SessionsServiceV1) GetSessionByID(ctx context.Context, id string) (*clientsv1.Session, error) {
	req := &clientsv1.GetSessionRequest{Id: id}
	client := s.newClient(ctx)
	resp, err := parseResponseAndError(client.GetSession(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}
	return resp.Msg.Session, nil
}
