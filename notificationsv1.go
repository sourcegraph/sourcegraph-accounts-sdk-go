package sams

import (
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	notificationsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/notifications/v1"
)

type NotificationsV1Config struct {
	// ProjectID is the GCP project ID that the Pub/Sub subscription belongs to. It
	// is almost always the same GCP project that the Cloud Run service is deployed
	// to.
	ProjectID string
	// SubscriptionID is the GCP Pub/Sub subscription ID to receive SAMS
	// notifications from.
	SubscriptionID string
}

// NewNotificationsV1ConfigFromEnv initializes configuration based on
// environment variables for setting up a new client for receiving SAMS
// notifications with v1 protocol.
func NewNotificationsV1ConfigFromEnv(env envGetter) NotificationsV1Config {
	defaultProject := env.Get("GOOGLE_CLOUD_PROJECT", "", "The GCP project that the service is running in")
	return NotificationsV1Config{
		ProjectID:      env.Get("SAMS_NOTIFICATION_PROJECT", defaultProject, "GCP project ID that the Pub/Sub subscription belongs to"),
		SubscriptionID: env.Get("SAMS_NOTIFICATION_SUBSCRIPTION", "sams_notifications", "GCP Pub/Sub subscription ID to receive SAMS notifications from"),
	}
}

func (c NotificationsV1Config) Validate() error {
	if c.ProjectID == "" {
		return errors.New("ProjectID is required")
	}
	if c.SubscriptionID == "" {
		return errors.New("SubscriptionID is required")
	}
	return nil
}

// NewNotificationsV1 returns a new client for receiving SAMS notifications with
// v1 protocol.
func NewNotificationsV1(logger log.Logger, config NotificationsV1Config) (*notificationsv1.Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return notificationsv1.NewClient(logger, config.ProjectID, config.SubscriptionID)
}
