package roles

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/services"
	"github.com/stretchr/testify/assert"
)

func TestAllowedGoldenList(t *testing.T) {
	autogold.Expect(AllowedRoles{Role("dotcom::site_admin"), Role("ssc::admin")}).Equal(t, Allowed())
}

func TestAllowedContains(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		expected bool
	}{
		{
			name:     "dotcom site admin literal",
			role:     "dotcom::site_admin",
			expected: true,
		},
		{
			name:     "dotcom site admin",
			role:     RoleDotcomSiteAdmin,
			expected: true,
		},
		{
			name:     "dotcom not a role",
			role:     "dotcom::not_a_role",
			expected: false,
		},
	}

	allowed := Allowed()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := allowed.Contains(test.role)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestAllowedRolesByService(t *testing.T) {
	tests := []struct {
		name     string
		service  services.Service
		expected autogold.Value
	}{
		{
			name:    "dotcom",
			service: services.Dotcom,
			expected: autogold.Expect([]Role{
				Role("dotcom::site_admin"),
			}),
		},
	}
	allowed := Allowed()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := allowed.ByService()
			test.expected.Equal(t, got[test.service])
		})
	}
}
