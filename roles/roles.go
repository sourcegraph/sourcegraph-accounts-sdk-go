package roles

import (
	"slices"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/services"
)

// Role is always the full qualified role name.
type Role string

// ToRole returns a role string in the format of
// "service::role", which comprises a fully qualified role name.
func ToRole(service services.Service, name string) Role {
	return Role(string(service) + "::" + name)
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
func ToRoles(strings []string) []Role {
	roles := make([]Role, len(strings))
	for i, s := range strings {
		roles[i] = Role(s)
	}
	return roles
}

var (
	// services.Dotcom
	dotcomRoles = []string{
		RoleNameDotcomSiteAdmin,
	}
)

const (
	// Role names for services.Dotcom

	// Dotcom site admin
	RoleNameDotcomSiteAdmin = "site_admin"
)

// AllowedRoles is a concrete list of allowed roles that can be granted to a user.
type AllowedRoles []Role

type registeredRoles map[services.Service]AllowedRoles

func registered() registeredRoles {
	registered := make(registeredRoles)

	appendRoles := func(service services.Service, roleNames []string) {
		var roles AllowedRoles
		for _, role := range roleNames {
			roles = append(roles, ToRole(service, role))
		}
		registered[service] = roles

	}

	appendRoles(services.Dotcom, dotcomRoles)
	// 👉 ADD YOUR ROLES HERE

	return registered
}

// Allowed returns all allowed roles that can be granted to a user. The caller
// should use AllowedRoles.Contains for matching requested roles.
func Allowed() AllowedRoles {
	var allowed AllowedRoles
	for _, roles := range registered() {
		allowed = append(allowed, roles...)
	}
	return allowed
}

// Contains returns true if the role is in the list of allowed roles
func (r AllowedRoles) Contains(role Role) bool {
	return slices.Contains(r, role)
}

// EntitleRoles is a map of services to a list of roles that can be granted to a user
type EntitleRoles registeredRoles

// AllowedEntitle returns a map of services to a list of roles that can be granted to a user.
func AllowedEntitle() EntitleRoles {
	return EntitleRoles(registered())
}
