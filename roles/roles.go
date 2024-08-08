package roles

import (
	"slices"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/services"
)

// Role is the string literal of a role.
type Role string

// RoleRegistration is a struct that contains the display name and roles for a service.
type RoleRegistration struct {
	DisplayName string
	Roles       []Role
}

var (
	// services.Dotcom
	dotcomDisplayName string = "Sourcegraph Dotcom"
	dotcomRoles              = []Role{
		RoleDotcomSiteadmin,
	}
)

const (
	// Roles for services.Dotcom

	// RoleDotcomSiteadmin is the role for site admins on dotcom
	RoleDotcomSiteadmin Role = "site_admin"
)

// Registered returns a map of all registered roles for all services.
type RegisteredRoles map[services.Service]RoleRegistration

func Registered() RegisteredRoles {
	registered := map[services.Service]RoleRegistration{}

	appendRoles := func(service services.Service, registration RoleRegistration) {
		registered[service] = registration
	}

	appendRoles(services.Dotcom, RoleRegistration{
		DisplayName: dotcomDisplayName,
		Roles:       dotcomRoles,
	})
	// 👉 ADD YOUR ROLES HERE

	return registered
}

func (r RegisteredRoles) Contains(service services.Service, role Role) bool {
	registration, ok := r[service]
	if !ok {
		return false
	}
	return slices.Contains(registration.Roles, role)
}
