package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
)

// Config contains the configuration for the SAMS authentication handler.
type Config struct {
	// Issuer is the SAMS instance URL, e.g. "https://accounts.sourcegraph.com".
	Issuer string
	// ClientID is the SAMS client ID, e.g. "sams_cid_xxx".
	ClientID string
	// ClientSecret is the SAMS client secret, e.g. "sams_cs_xxx".
	ClientSecret string
	// RequestScopes is the list of requested scopes for access tokens that are
	// issued to this client.
	RequestScopes []scopes.Scope
	// RedirectURI is the URL to redirect to after the user has authenticated.
	RedirectURI string

	// FailureHandler is the HTTP handler to call when an error occurs. Use
	// ErrorFromContext to extract the error.
	FailureHandler http.Handler
	// StateSetter sets the randomly-generated state to the per-user session.
	StateSetter func(w http.ResponseWriter, r *http.Request, state string) error
	// StateGetter gets the state from the per-user session.
	StateGetter func(r *http.Request) (string, error)
	// StateDeleter deletes the state from the per-user session.
	StateDeleter func(w http.ResponseWriter, r *http.Request)
}

// Error is an error that occurred during the authentication process.
type Error struct {
	// StatusCode is the HTTP status code to respond with.
	StatusCode int
	// Cause is the error that caused the failure.
	Cause error
}

// Handler is the SAMS authentication handler.
type Handler struct {
	config Config
}

// NewHandler returns a new SAMS authentication handler with the given
// configuration.
func NewHandler(config Config) (*Handler, error) {
	if config.FailureHandler == nil {
		return nil, errors.New("missing FailureHandler")
	} else if config.StateSetter == nil {
		return nil, errors.New("missing StateSetter")
	} else if config.StateGetter == nil {
		return nil, errors.New("missing StateGetter")
	} else if config.StateDeleter == nil {
		return nil, errors.New("missing StateDeleter")
	}
	return &Handler{config: config}, nil
}

// LoginHandler returns an HTTP handler that redirects the user to the SAMS
// authentication page.
//
// It passes through the "prompt" and "prompt_auth" query parameters to SAMS.
func (h *Handler) LoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Create a new OIDC provider from the issuer URL using OIDC discovery feature.
		p, err := oidc.NewProvider(ctx, h.config.Issuer)
		if err != nil {
			ctx = WithError(ctx, &Error{
				StatusCode: http.StatusInternalServerError,
				Cause:      errors.Wrap(err, "create new provider"),
			})
			h.config.FailureHandler.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Generate and store a random nonce to the session.
		nonce, err := randomString()
		if err != nil {
			ctx = WithError(ctx, &Error{
				StatusCode: http.StatusInternalServerError,
				Cause:      errors.Wrap(err, "generate nonce"),
			})
			h.config.FailureHandler.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		err = h.config.StateSetter(w, r, nonce)
		if err != nil {
			ctx = WithError(ctx, &Error{
				StatusCode: http.StatusInternalServerError,
				Cause:      errors.Wrap(err, "set state"),
			})
			h.config.FailureHandler.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// NOTE: Using same value for both state and nonce is fine because "state" is
		// for service self-validation (no CSRF) and "nonce" is for IdP validation (no
		// repeated attempts).
		params := url.Values{}
		params.Add("client_id", h.config.ClientID)
		params.Add("redirect_uri", h.config.RedirectURI)
		params.Add("state", nonce)
		params.Add("nonce", nonce)
		params.Add("response_type", "code")
		params.Add("scope", strings.Join(scopes.ToStrings(h.config.RequestScopes), " "))

		// Passthrough the IdP-aware query parameters.
		params.Add("prompt", r.URL.Query().Get("prompt"))
		params.Add("prompt_auth", r.URL.Query().Get("prompt_auth"))

		http.Redirect(w, r, p.Endpoint().AuthURL+"?"+params.Encode(), http.StatusFound)
	})
}

// CallbackHandler returns an HTTP handler that handles the SAMS callback and
// calls the success handler upon successful authentication. Use
// UserInfoFromContext to extract the user information.
func (h *Handler) CallbackHandler(success http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		nonce, err := h.config.StateGetter(r)
		h.config.StateDeleter(w, r) // Delete the state after getting it to make sure it's one-time use.
		if err != nil {
			ctx = WithError(ctx, &Error{
				StatusCode: http.StatusInternalServerError,
				Cause:      errors.Wrap(err, "get state"),
			})
			h.config.FailureHandler.ServeHTTP(w, r.WithContext(ctx))
			return
		}
		if got := r.URL.Query().Get("state"); nonce != got {
			ctx = WithError(ctx, &Error{
				StatusCode: http.StatusBadRequest,
				Cause:      errors.Errorf("mismatched state, want %q but got %q", nonce, got),
			})
			h.config.FailureHandler.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		userInfo, err := h.getUserInfo(r)
		if err != nil {
			ctx = WithError(ctx, &Error{
				StatusCode: http.StatusInternalServerError,
				Cause:      err,
			})
			h.config.FailureHandler.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ctx = WithUserInfo(ctx, userInfo)
		success.ServeHTTP(w, r.WithContext(ctx))
	})
}

// DefaultFailureHandler responds with the status code and message based on the
// error extracted from the context.
var DefaultFailureHandler = http.HandlerFunc(failureHandler)

func failureHandler(w http.ResponseWriter, r *http.Request) {
	err := ErrorFromContext(r.Context())
	http.Error(w, err.Cause.Error(), err.StatusCode)
}

// randomString returns a base64-encoded random 32 byte string.
func randomString() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// UserInfo contains the information about the authenticated user.
type UserInfo struct {
	// ID is the unique identifier of the user.
	ID string `json:"sub"`
	// Name is the display name of the user.
	Name string `json:"name"`
	// Email is the email address of the user.
	Email string `json:"email"`
	// EmailVerified is true if the email address has been verified.
	EmailVerified bool `json:"email_verified"`
	// AvatarURL is the URL to the user's avatar.
	AvatarURL string `json:"picture"`
	// CreatedAt is the time when the user account was created.
	CreatedAt time.Time `json:"created_at"`

	// Token is the OAuth2 access token.
	Token *oauth2.Token `json:"-"`
	// IDToken is the OpenID Connect ID token.
	IDToken *oidc.IDToken `json:"-"`
}

func (h *Handler) getUserInfo(r *http.Request) (*UserInfo, error) {
	ctx := r.Context()
	p, err := oidc.NewProvider(ctx, h.config.Issuer)
	if err != nil {
		return nil, errors.Wrap(err, "create new provider")
	}

	nonce, err := h.config.StateGetter(r)
	if err != nil {
		return nil, errors.Wrap(err, "get state")
	}

	oauth2Config := oauth2.Config{
		ClientID:     h.config.ClientID,
		ClientSecret: h.config.ClientSecret,
		RedirectURL:  h.config.RedirectURI,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: p.Endpoint(),
		Scopes:   scopes.ToStrings(h.config.RequestScopes),
	}

	code := r.URL.Query().Get("code")
	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, errors.Wrap(err, "exchange token")
	}

	// Extract the ID Token from the access token, see http://openid.net/specs/openid-connect-core-1_0.html#TokenResponse.
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New(`missing "id_token" from the issuer's authorization response`)
	}

	verifier := p.Verifier(&oidc.Config{ClientID: oauth2Config.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, errors.Wrap(err, "verify raw ID Token")
	}
	if nonce != idToken.Nonce {
		return nil, errors.Errorf("mismatched nonce, want %q but got %q", nonce, idToken.Nonce)
	}

	rawUserInfo, err := p.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, errors.Wrap(err, "fetch user info")
	}

	userInfo := &UserInfo{
		Token:   token,
		IDToken: idToken,
	}
	err = rawUserInfo.Claims(&userInfo)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal claims")
	}

	if userInfo.ID == "" {
		return nil, errors.New(`the field "sub" is not found in claims or has empty value`)
	} else if userInfo.Name == "" {
		return nil, errors.New(`the field "name" is not found in claims or has empty value`)
	} else if userInfo.Email == "" {
		return nil, errors.New(`the field "email" is not found in claims or has empty value`)
	} else if !userInfo.EmailVerified {
		return nil, errors.New(`the field "email_verified" is not "true"`)
	} else if userInfo.AvatarURL == "" {
		return nil, errors.New(`the field "picture" is not found in claims or has empty value`)
	} else if userInfo.CreatedAt.IsZero() {
		return nil, errors.New(`the field "created_at" is not found in claims or has zero value`)
	}
	return userInfo, nil
}
