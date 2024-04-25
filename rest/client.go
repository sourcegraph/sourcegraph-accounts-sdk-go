package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Client is a wrapper around SAMS primitive REST API. Most likely, you want to use
// the more robust service-to-service protobuf API.
//
// This API is needed when using SAMS to identify users, but not perform authorization
// checks. e.g. the caller will handle its own authorization checks based on the identity
// of the SAMS user. (The returned User.Subject, the SAMS account external ID.)
type Client interface {
	// GetUserDetails returns the basic user profile of the user whom the supplied SAMS access
	// token belongs to.
	//
	// If the supplied token is invalid, malformed, or expired, the error will contain
	// "unexpected status 401".
	GetUserDetails(ctx context.Context, token string) (*User, error)
}

// NewClient constructs a new SAMS REST client, pointed to the supplied SAMS host.
// e.g. "https://accounts.sourcegraph.com".
func NewClient(samsHost string) Client {
	// Canonicalize the host so we only need to check if it ends in a slash or not once.
	samsHost = strings.ToLower(samsHost)
	samsHost = strings.TrimSuffix(samsHost, "/")

	return &client{
		host: samsHost,
	}
}

type client struct {
	host string
}

func (c *client) GetUserDetails(ctx context.Context, token string) (*User, error) {
	url := fmt.Sprintf("%s/api/v1/user", c.host)

	req, err := http.NewRequest(http.MethodGet, url, nil /* body */)
	if err != nil {
		return nil, errors.Wrap(err, "creating SAMS user details request")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("User-Agent", "sourcegraph-accounts-sdk-go/1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "fetching user details")
	}
	if resp.Body == nil {
		return nil, errors.Wrap(err, "no response body")
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading response body")
	}
	if err = resp.Body.Close(); err != nil {
		return nil, errors.Wrap(err, "closing response body")
	}
	if resp.StatusCode != http.StatusOK {
		unexpectedRespErr := errors.Errorf(
			"unexpected status %d (response body: %s)",
			resp.StatusCode, string(bodyBytes))
		return nil, unexpectedRespErr
	}

	var user User
	if err = json.Unmarshal(bodyBytes, &user); err != nil {
		return nil, errors.Wrap(err, "unmarshalling response")
	}

	return &user, nil
}
