package scopes

import (
	"errors"
	"strings"
)

// PermissionToMetadataScope extracts the metadata scope from a permission.
// Only valid for 'user.metadata'-prefixed permissions.
//
// Examples:
//   - user.metadata => *
//   - user.metadata.cody => cody
//   - user.metadata.dotcom => dotcom
func PermissionToMetadataScope(permission Permission) (string, error) {
	const prefix = "user.metadata"
	if !strings.HasPrefix(string(permission), prefix) {
		return "", errors.New("permission is not a metadata permission")
	}
	if permission == prefix {
		return "*", nil
	}
	return strings.TrimPrefix(string(permission), prefix+"."), nil
}
