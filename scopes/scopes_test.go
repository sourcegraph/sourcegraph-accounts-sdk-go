package scopes

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
)

func TestAllowedGoldenList(t *testing.T) {
	// Generated golden list of all allowed scopes for ease of visual review.
	autogold.Expect(AllowedScopes{
		Scope("openid"), Scope("profile"), Scope("email"),
		Scope("offline_access"),
		Scope("client.ssc"),
		Scope("client.dotcom"),
		Scope("cody_gateway::flaggedprompts::read"),
		Scope("cody_gateway::flaggedprompts::write"),
		Scope("cody_gateway::flaggedprompts::delete"),
		Scope("sams::user::read"),
		Scope("sams::user::write"),
		Scope("sams::user::delete"),
		Scope("sams::user.profile::read"),
		Scope("sams::user.profile::write"),
		Scope("sams::user.profile::delete"),
		Scope("sams::user.roles::read"),
		Scope("sams::user.roles::write"),
		Scope("sams::user.roles::delete"),
		Scope("sams::user.metadata::read"),
		Scope("sams::user.metadata::write"),
		Scope("sams::user.metadata::delete"),
		Scope("sams::user.metadata.cody::read"),
		Scope("sams::user.metadata.cody::write"),
		Scope("sams::user.metadata.cody::delete"),
		Scope("sams::user.metadata.dotcom::read"),
		Scope("sams::user.metadata.dotcom::write"),
		Scope("sams::user.metadata.dotcom::delete"),
		Scope("sams::session::read"),
		Scope("sams::session::write"),
		Scope("sams::session::delete"),
		Scope("telemetry_gateway::events::read"),
		Scope("telemetry_gateway::events::write"),
		Scope("telemetry_gateway::events::delete"),
		Scope("enterprise_portal::subscription::read"),
		Scope("enterprise_portal::subscription::write"),
		Scope("enterprise_portal::subscription::delete"),
		Scope("enterprise_portal::permission.subscription::read"),
		Scope("enterprise_portal::permission.subscription::write"),
		Scope("enterprise_portal::permission.subscription::delete"),
		Scope("enterprise_portal::codyaccess::read"),
		Scope("enterprise_portal::codyaccess::write"),
		Scope("enterprise_portal::codyaccess::delete"),
		Scope("workspaces::workspace::read"),
		Scope("workspaces::workspace::write"),
		Scope("workspaces::workspace::delete"),
		Scope("workspaces::instance::read"),
		Scope("workspaces::instance::write"),
		Scope("workspaces::instance::delete"),
		Scope("workspaces::permission.workspace::read"),
		Scope("workspaces::permission.workspace::write"),
		Scope("workspaces::permission.workspace::delete"),
	}).Equal(t, Allowed())
}

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
