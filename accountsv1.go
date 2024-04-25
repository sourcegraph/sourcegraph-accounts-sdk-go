package sams

import (
	accountsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/accounts/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"
)

type AccountsV1Config struct {
	ConnConfig
	// TokenSource is the OAuth2 token source to use for authentication.
	//
	// If you have the SAMS user's Refresh token, using the oauth2.TokenSource
	// abstraction will take care of creating short-lived access tokens as
	// needed. But if you only have the access token, you will need to use a
	// StaticTokenSource instead.
	TokenSource oauth2.TokenSource
}

func (c AccountsV1Config) Validate() error {
	if err := c.ConnConfig.Validate(); err != nil {
		return errors.Wrap(err, "ConnConfig")
	}
	if c.TokenSource == nil {
		return errors.New("token source is required")
	}
	return nil
}

// NewAccountsV1 returns a new SAMS client for interacting with Accounts API V1.
func NewAccountsV1(config AccountsV1Config) (*accountsv1.Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return accountsv1.NewClient(config.getAPIURL(), config.TokenSource), nil
}
