package clientcredentials

import (
	"context"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
)

type mockTokenIntrospector struct {
	response *sams.IntrospectTokenResponse
}

func (m *mockTokenIntrospector) IntrospectToken(ctx context.Context, token string) (*sams.IntrospectTokenResponse, error) {
	return m.response, nil
}
