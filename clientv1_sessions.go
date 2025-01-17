package sams

import (
	"context"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"

	"github.com/hashicorp/golang-lru/v2/expirable"
	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1/clientsv1connect"
)

// SessionsServiceV1 provides client methods to interact with the
// SessionsService API v1.
type SessionsServiceV1 struct {
	client *ClientV1
	// sessionsCache may be nil if not enabled.
	sessionsCache *expirable.LRU[string, *clientsv1.Session]
}

func (s *SessionsServiceV1) newClient(ctx context.Context) clientsv1connect.SessionsServiceClient {
	return clientsv1connect.NewSessionsServiceClient(
		oauth2.NewClient(ctx, s.client.tokenSource),
		s.client.gRPCURL(),
		connect.WithInterceptors(s.client.defaultInterceptors...),
	)
}

// GetSessionByID returns the SAMS session with the given ID. It returns
// ErrNotFound if no such session exists. The session's `User` field is
// populated if the session is authenticated by a user.
//
// Required scope: sams::session::read
func (s *SessionsServiceV1) GetSessionByID(ctx context.Context, id string) (*clientsv1.Session, error) {
	if s.sessionsCache != nil {
		if cached, ok := s.sessionsCache.Get(id); ok {
			trace.SpanFromContext(ctx).
				SetAttributes(attribute.Bool("sams.session.fromCache", true))
			return cached, nil
		}
	}
	trace.SpanFromContext(ctx).
		SetAttributes(attribute.Bool("sams.session.fromCache", false))

	req := &clientsv1.GetSessionRequest{Id: id}
	client := s.newClient(ctx)
	resp, err := parseResponseAndError(client.GetSession(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}
	if s.sessionsCache != nil {
		_ = s.sessionsCache.Add(id, resp.Msg.Session)
	}
	return resp.Msg.Session, nil
}

// SignOutSession revokes the authenticated state of the session with the given
// ID for the given user. It does not return error if the session does not exist
// or is not authenticated. It returns ErrRecordMismatch if the session is
// authenticated by a different user than the given user.
//
// Required scope: sams::session::write
func (s *SessionsServiceV1) SignOutSession(ctx context.Context, sessionID, userID string) error {
	req := &clientsv1.SignOutSessionRequest{
		Id:     sessionID,
		UserId: userID,
	}
	client := s.newClient(ctx)
	_, err := parseResponseAndError(client.SignOutSession(ctx, connect.NewRequest(req)))
	if err == nil && s.sessionsCache != nil {
		_ = s.sessionsCache.Remove(sessionID)
	}
	return err
}
