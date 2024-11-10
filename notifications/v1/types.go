package v1

import (
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/roles"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/services"
)

// ⚠️ WARNING: These types MUST match the SAMS implementation, at
// backend/internal/notification/types.go

const (
	nameUserDeleted        = "UserDeleted"
	nameUserRolesUpdated   = "UserRolesUpdated"
	nameSessionInvalidated = "SessionInvalidated"
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
type UserRolesUpdatedData struct {
	// AccountID is the SAMS external ID of the user whose roles have been updated.
	AccountID string `json:"account_id"`
	// Service is the service that the user's roles have been updated in.
	Service services.Service `json:"service"`
	// RoleID is the  role that has been updated
	RoleID roles.Role `json:"role"`
	// ResourceID is the ID of the resource the role has been updated on.
	ResourceID string `json:"resource_id,omitempty"`
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
