package v1

// ⚠️ WARNING: These types MUST match the SAMS implementation, at
// backend/internal/notification/types.go

const (
	nameUserDeleted = "UserDeleted"
)

// UserDeletedData contains information of a "UserDeleted" notification.
type UserDeletedData struct {
	// UserID is the SAMS external user ID of the deleted user.
	UserID string `json:"user_id"`
	// Email is the email address of the deleted user.
	Email string `json:"email"`
}
