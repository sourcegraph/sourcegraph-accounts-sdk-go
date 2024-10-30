package roles

import (
	"slices"
	"strings"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/services"
)

// Role is always the full qualified role name, e.g. "dotcom::site_admin".
type Role string

// ToRole returns a role string in the format of
// "service::role", which comprises a fully qualified role name.
func ToRole(service services.Service, name string) Role {
	return Role(string(service) + "::" + name)
}

// Service returns the service that the role belongs to.
func (r Role) Service() services.Service {
	return services.Service(r[:strings.Index(string(r), "::")])
}

// ToStrings converts a list of roles to a list of strings.
func ToStrings(roles []Role) []string {
	ss := make([]string, len(roles))
	for i, role := range roles {
		ss[i] = string(role)
	}
	return ss
}

// ToRoles converts a list of strings to a list of roles.
// It does not validate each input value.
func ToRoles(strings []string) []Role {
	roles := make([]Role, len(strings))
	for i, s := range strings {
		roles[i] = Role(s)
	}
	return roles
}

// ResourceType is the type of resource that a role is associated with.
type ResourceType string

const (
	// Service type is a special type used for service level roles.
	Service ResourceType = "service"
	// Subscription resources for Enterprise Portal.
	Subscription ResourceType = "subscription"
)

// IsService returns true if the resource type is a service.
// This is a special helper function as service level roles have special handling.
func (r ResourceType) IsService() bool {
	return r == Service
}

// roleInfo is the sdk internal representation of a role.
type roleInfo struct {
	// id is the fully qualified role name. e.g. "dotcom::site_admin"
	id Role
	//  service is the service that the role belongs to.
	service services.Service
	// resourceType is the type of resource that the role is associated with.
	resourceType ResourceType
}

// services.Dotcom
var (
	// Dotcom site admin
	RoleDotcomSiteAdmin = ToRole(services.Dotcom, "site_admin")

	dotcomRoles = []roleInfo{
		{
			id:           RoleDotcomSiteAdmin,
			service:      services.Dotcom,
			resourceType: Service,
		},
	}
)

// services.SSC
var (
	// SSC admin
	RoleSSCAdmin = ToRole(services.SSC, "admin")

	sscRoles = []roleInfo{
		{
			id:           RoleSSCAdmin,
			service:      services.SSC,
			resourceType: Service,
		},
	}
)

// services.EnterprisePortal
var (
	// Enterprise Portal customer admin
	RoleEnterprisePortalCustomerAdmin = ToRole(services.EnterprisePortal, "customer_admin")

	enterprisePortalRoles = []roleInfo{
		{
			id:           RoleEnterprisePortalCustomerAdmin,
			service:      services.EnterprisePortal,
			resourceType: Subscription,
		},
	}
)

var registeredRoles = func() []roleInfo {
	var registered []roleInfo

	appendRoles := func(roles []roleInfo) {
		registered = append(registered, roles...)
	}

	appendRoles(dotcomRoles)
	appendRoles(sscRoles)
	appendRoles(enterprisePortalRoles)
	// 👉 ADD YOUR ROLES HERE

	return registered
}()

// List returns a list of all List
func List() []Role {
	var roles []Role
	for _, role := range registeredRoles {
		roles = append(roles, role.id)
	}
	return roles
}

// Contains returns true if the role is in the list of allowed roles
func Contains(role Role) bool {
	return slices.Contains(List(), role)
}

// ByService returns all allowed roles grouped by service.
func ByService() map[services.Service][]Role {
	byService := make(map[services.Service][]Role)
	for _, role := range registeredRoles {
		byService[role.service] = append(byService[role.service], role.id)
	}
	return byService
}

// ByResourceType returns all allowed roles grouped by resource type.
func ByResourceType() map[ResourceType][]Role {
	byResourceType := make(map[ResourceType][]Role)
	for _, role := range registeredRoles {
		byResourceType[role.resourceType] = append(byResourceType[role.resourceType], role.id)
	}
	return byResourceType
}
