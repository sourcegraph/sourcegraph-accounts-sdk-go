package roles

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/services"
	"github.com/stretchr/testify/assert"
)

func TestRegisteredGoldenList(t *testing.T) {
	autogold.Expect(RegisteredRoles{
		services.Service("dotcom"): RoleRegistration{
			DisplayName: "Sourcegraph Dotcom",
			Roles:       []Role{Role("site_admin")},
		}}).Equal(t, Registered())
}

func TestRegisteredContains(t *testing.T) {
	tests := []struct {
		name     string
		service  services.Service
		role     Role
		expected bool
	}{
		{
			name:     "dotcom site admin",
			service:  services.Dotcom,
			role:     RoleDotcomSiteadmin,
			expected: true,
		},
		{
			name:     "dotcom site admin",
			service:  services.Dotcom,
			role:     Role("not_a_role"),
			expected: false,
		},
	}

	registered := Registered()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := registered.Contains(test.service, test.role)
			assert.Equal(t, test.expected, got)
		})
	}
}
