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

// ParsedRole is a role parsed into its service and name.
type ParsedRole struct {
	Service services.Service
	Name    string
}

// ToRole creates a fully qualified role from a parsed role
// in the format of service::name.
func (p ParsedRole) ToRole() Role {
	return ToRole(p.Service, p.Name)
}

// Parse parses a role into its parts. It returns the service, and the role name.
func (r Role) Parse() (_ ParsedRole, valid bool) {
	i := strings.Index(string(r), "::")
	if i == -1 {
		return ParsedRole{}, false
	}

	service := r[:i]
	name := r[i+2:] // skip the "::"

	if service == "" || name == "" {
		return ParsedRole{}, false
	}

	return ParsedRole{
		Service: services.Service(service),
		Name:    string(name),
	}, true
}

var (
	// services.Dotcom

	// Dotcom site admin
	RoleDotcomSiteAdmin = ToRole(services.Dotcom, "site_admin")

	dotcomRoles = []Role{
		RoleDotcomSiteAdmin,
	}
)

var (
	// services.SSC

	// SSC admin
	RoleSSCAdmin = ToRole(services.SSC, "admin")

	sscRoles = []Role{
		RoleSSCAdmin,
	}
)

// AllowedRoles is a concrete list of allowed roles that can be granted to a user.
type AllowedRoles []Role

// Allowed returns all allowed roles that can be granted to a user. The caller
// should use AllowedRoles.Contains for matching requested roles.
func Allowed() AllowedRoles {
	var allowed AllowedRoles

	appendRoles := func(roles []Role) {
		allowed = append(allowed, roles...)
	}

	appendRoles(dotcomRoles)
	appendRoles(sscRoles)
	// 👉 ADD YOUR ROLES HERE

	return allowed
}

// Contains returns true if the role is in the list of allowed roles
func (r AllowedRoles) Contains(role Role) bool {
	return slices.Contains(r, role)
}

// ByService returns all allowed roles grouped by service.
func (r AllowedRoles) ByService() map[services.Service][]Role {
	byService := make(map[services.Service][]Role)
	for _, role := range Allowed() {
		parsed, valid := role.Parse()
		if !valid {
			continue
		}

		byService[parsed.Service] = append(byService[parsed.Service], role)
	}
	return byService
}
