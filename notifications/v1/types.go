package v1

import (
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/roles"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/services"
)

// ⚠️ WARNING: These types MUST match the SAMS implementation, at
// backend/internal/notification/types.go

const (
	nameUserDeleted         = "UserDeleted"
	nameUserRolesUpdated    = "UserRolesUpdated"
	nameUserMetadataUpdated = "UserMetadataUpdated"
	nameSessionInvalidated  = "SessionInvalidated"
)

// UserDeletedData contains information of a "UserDeleted" notification.
type UserDeletedData struct {
	// AccountID is the SAMS external ID of the deleted user.
	AccountID string `json:"account_id"`
	// Email is the email address of the deleted user.
	Email string `json:"email"`
}

// UserRolesUpdatedData contains information of a "UserRolesUpdated" notification.
// When a user's roles have been updated it is neccessary to query SAMS to get the
// updated roles to determine if it was granted/revoked.
//
// For more details see:
// https://sourcegraph.notion.site/SAMS-Roles-Resources-13ca8e11265880f9a573cac77070ca0c
type UserRolesUpdatedData struct {
	// AccountID is the SAMS external ID of the user whose roles have been updated.
	AccountID string `json:"account_id"`
	// Service is the service that the user's roles have been updated in.
	Service services.Service `json:"service"`
	// RoleID is the  role that has been updated.
	RoleID roles.Role `json:"role"`
	// ResourceID is the ID of the resource the role has been updated on,
	// if applicable. When ResourceID is empty, the role is a service-level
	// role that does not apply to a specific resource.
	ResourceID string `json:"resource_id,omitempty"`
	// ResourceType is the type of the resource the role has been updated on,
	// if applicable. When ResourceType is empty, the role is a service-level
	// role that does not apply to a specific resource.
	ResourceType roles.ResourceType `json:"resource_type,omitempty"`
}

type UserMetadataUpdatedData struct {
	// AccountID is the SAMS external ID of the user whose metadata has been
	// updated.
	AccountID string `json:"account_id"`
	// Namespace is the metadata scope that the user's roles have been updated in.
	Namespace string `json:"namespace"`
}

// SessionInvalidatedData contains information of a "SessionInvalidated"
// notification.
type SessionInvalidatedData struct {
	// AccountID is the SAMS external ID of the user whose session has been
	// invalidated.
	AccountID string `json:"account_id"`
	// SessionID is the ID of the invalidated session.
	SessionID string `json:"session_id"`
}

var tracer = otel.Tracer("sams.notifications.v1")
