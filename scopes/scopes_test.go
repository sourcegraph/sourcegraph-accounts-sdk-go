package scopes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllowed(t *testing.T) {
	allowedScopes := Allowed()

	t.Run("user.metadata", func(t *testing.T) {
		tests := []struct {
			scope   Scope
			allowed bool
		}{
			{"sams::user.metadata::read", true},
			{"sams::user.metadata::write", true},
			{"sams::user.metadata.cody::read", true},
			{"sams::user.metadata.dotcom::delete", true},

			{"sams::user.not_a_permission::read", false},
			{"sams::user.roles::badaction", false},
			{"not_a_service::user.roles::read", false},
		}
		for _, test := range tests {
			t.Run(string(test.scope), func(t *testing.T) {
				got := allowedScopes.Contains(test.scope)
				assert.Equal(t, test.allowed, got)
			})
		}
	})

	t.Run("all allowed scopes are spec-compliant", func(t *testing.T) {
		for _, scope := range allowedScopes {
			t.Run(string(scope), func(t *testing.T) {
				if scope == OpenID || scope == Profile || scope == Email || scope == OfflineAccess {
					t.Skip("Builtin scopes for OAuth and OIDC")
				}

				if scope == ClientSSC || scope == ClientDotcom {
					t.Skip("Legacy scopes to be replaced")
				}

				assert.True(t, len(scope) <= 255, "scope length")

				parsedScope, valid := ParseScope(scope)
				assert.True(t, valid)
				assert.True(t, ServiceRegex.MatchString(string(parsedScope.Service)))
				assert.True(t, PermissionRegex.MatchString(string(parsedScope.Permission)))
				assert.True(t, ActionRegex.MatchString(string(parsedScope.Action)))
			})
		}
	})
}

func TestStrategy(t *testing.T) {
	tests := []struct {
		name     string
		matchers []string
		needle   string
		expected bool
	}{
		{
			name:     "exact match",
			matchers: []string{"sams::user::read"},
			needle:   "sams::user::read",
			expected: true,
		},
		{
			name:     "permission prefix matching",
			matchers: []string{"sams::user::read"},
			needle:   "sams::user.metadata::read",
			expected: true,
		},
		{
			name:     "alias matching",
			matchers: []string{string(Profile)},
			needle:   "sams::user.profile::read",
			expected: true,
		},
		{
			name:     "complex alias matching",
			matchers: []string{"sams::user::read", string(Profile), "sams::user.roles::read"},
			needle:   "sams::user.profile::read",
			expected: true,
		},
		{
			name:     "default scopes matching - profile",
			matchers: []string{string(OpenID), string(Profile), string(Email), string(OfflineAccess)},
			needle:   string(Profile),
			expected: true,
		},
		{
			name:     "default scopes matching - openid",
			matchers: []string{string(OpenID), string(Profile), string(Email), string(OfflineAccess)},
			needle:   string(OpenID),
			expected: true,
		},
		{
			name:     "default scopes matching - email",
			matchers: []string{string(OpenID), string(Profile), string(Email), string(OfflineAccess)},
			needle:   string(Email),
			expected: true,
		},
		{
			name:     "default scopes matching - offline_access",
			matchers: []string{string(OpenID), string(Profile), string(Email), string(OfflineAccess)},
			needle:   string(OfflineAccess),
			expected: true,
		},

		{
			name:     "legacy scopes - client.dotcom",
			matchers: []string{string(ClientDotcom), string(ClientSSC)},
			needle:   string(ClientDotcom),
			expected: true,
		},

		{
			name:     "legacy scopes - client.ssc",
			matchers: []string{string(ClientDotcom), string(ClientSSC)},
			needle:   string(ClientSSC),
			expected: true,
		},

		{
			name:     "service mismatch but rest matching",
			matchers: []string{"ssc::user::read"},
			needle:   "sams::user::read",
			expected: false,
		},
		{
			name:     "service mismatch but rest matching with prefix",
			matchers: []string{"ssc::user::read"},
			needle:   "sams::user.roles::read",
			expected: false,
		},
		{
			name:     "permission mismatch but rest matching",
			matchers: []string{"ssc::user::read"},
			needle:   "ssc::cody::read",
			expected: false,
		},
		{
			name:     "permission mismatch with prefix",
			matchers: []string{"ssc::user::read"},
			needle:   "ssc::cody.subscriptions::read",
			expected: false,
		},
		{
			name:     "action mismatch but rest matching",
			matchers: []string{"sams::user::read"},
			needle:   "sams::user::write",
			expected: false,
		},
		{
			name:     "action mismatch but rest matching with prefix",
			matchers: []string{"sams::user::read"},
			needle:   "sams::user.roles::write",
			expected: false,
		},
		{
			name:     "narrower permission cannot match broader permission",
			matchers: []string{"sams::user.metadata::read"},
			needle:   "sams::user::read",
			expected: false,
		},
		{
			name:     "alias and action mismatch",
			matchers: []string{string(Profile)},
			needle:   "sams::user.profile::write",
			expected: false,
		},

		{
			name:     "malformed matcher",
			matchers: []string{"", "sams::user", "sams::::read", "user::", "::user::read"},
			needle:   "sams::user::read",
			expected: false,
		},
		{
			name:     "malformed needle",
			matchers: []string{string(Profile)},
			needle:   "sams::user.profile",
			expected: false,
		},
		{
			name:     "malformed matcher and needle but matching literals",
			matchers: []string{"", "sams::user", "sams::::read", "user::", "::user::read"},
			needle:   "sams::user",
			expected: false,
		},
		{
			name:     "awkward but legal matcher",
			matchers: nil,
			needle:   "sams::user::read",
			expected: false,
		},
		{
			name:     "awkward but legal matcher",
			matchers: []string{},
			needle:   "sams::user::read",
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Strategy(test.matchers, test.needle)
			assert.Equal(t, test.expected, got)
		})
	}
}

// go test -bench=. -benchmem -cpu=4
// Worth matching the CPU number we allocate to the Cloud Run container.
func BenchmarkStrategy_Match(b *testing.B) {
	matchers := []string{
		"profile",
		"ssc::subscriptions::read",
		"sams::user.roles::read",
		"sams::user::write",
		"sams::user.metadata::read",
	}
	needle := "sams::user.metadata.cody::read"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Strategy(matchers, needle)
	}
}

func BenchmarkStrategy_NoMatch(b *testing.B) {
	matchers := []string{
		"profile",
		"ssc::subscriptions::read",
		"sams::user.roles::read",
		"sams::user::write",
		"sams::user.metadata.cody::read",
	}
	needle := "sams::user.metadata.dotcom::read"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Strategy(matchers, needle)
	}
}
