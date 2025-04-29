package roles

import (
	"slices"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/services"
	"github.com/stretchr/testify/assert"
)

func TestGoldenList(t *testing.T) {
	got := List()
	slices.Sort(got)
	autogold.Expect([]Role{
		Role("dotcom::site_admin"), Role("enterprise_portal::customer_admin"),
		Role("enterprise_portal::service_admin"),
		Role("ssc::admin"),
		Role("workspaces::service_admin"),
	}).Equal(t, got)
}

func TestContains(t *testing.T) {
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := Contains(test.role)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestRolesByService(t *testing.T) {
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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ByService()[test.service]
			slices.Sort(got)
			test.expected.Equal(t, got)
		})
	}
}

func TestRolesByResourceType(t *testing.T) {
	tests := []struct {
		name     string
		resource ResourceType
		expected autogold.Value
	}{
		{
			name:     "service",
			resource: Service,
			expected: autogold.Expect([]Role{
				Role("dotcom::site_admin"), Role("enterprise_portal::service_admin"),
				Role("ssc::admin"),
				Role("workspaces::service_admin"),
			}),
		},
		{
			name:     "enterprise_subscription",
			resource: EnterpriseSubscription,
			expected: autogold.Expect([]Role{
				Role("enterprise_portal::customer_admin"),
			}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ByResourceType()[test.resource]
			slices.Sort(got)
			test.expected.Equal(t, got)
		})
	}
}

func TestServiceRolesByService(t *testing.T) {
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
		{
			name:    "ssc",
			service: services.SSC,
			expected: autogold.Expect([]Role{
				Role("ssc::admin"),
			}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ServiceRolesByService()[test.service]
			slices.Sort(got)
			test.expected.Equal(t, got)
		})
	}
}

func TestToService(t *testing.T) {
	tests := []struct {
		name string
		role Role
		want services.Service
	}{
		{
			name: "dotcom site admin",
			role: RoleDotcomSiteAdmin,
			want: services.Dotcom,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.role.Service()
			assert.Equal(t, test.want, got)
		})
	}
}

func TestDisplay(t *testing.T) {
	tests := []struct {
		name     string
		resource ResourceType
		want     string
	}{
		{
			name:     "service",
			resource: Service,
			want:     "Service",
		},
		{
			name:     "enterprise_subscription",
			resource: EnterpriseSubscription,
			want:     "Enterprise Subscription",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.resource.DisplayName()
			assert.Equal(t, test.want, got)
		})
	}
}
