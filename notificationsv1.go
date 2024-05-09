package sams

import (
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/lib/background"

	notificationsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/notifications/v1"
)

// NewNotificationsV1SubscriberOptions initializes configuration based on
// environment variables and supplied arguments.
func NewNotificationsV1SubscriberOptions(
	env envGetter,
	settings notificationsv1.ReceiveSettings,
	handlers notificationsv1.SubscriberHandlers,
) notificationsv1.SubscriberOptions {
	defaultProject := env.Get("GOOGLE_CLOUD_PROJECT", "", "The GCP project that the service is running in")
	return notificationsv1.SubscriberOptions{
		ProjectID:       env.Get("SAMS_NOTIFICATION_PROJECT", defaultProject, "GCP project ID that the Pub/Sub subscription belongs to"),
		SubscriptionID:  env.Get("SAMS_NOTIFICATION_SUBSCRIPTION", "sams_notifications", "GCP Pub/Sub subscription ID to receive SAMS notifications from"),
		ReceiveSettings: settings,
		Handlers:        handlers,
	}
}

// NewNotificationsV1Subscriber returns a new background routine for receiving
// SAMS notifications with v1 protocol.
func NewNotificationsV1Subscriber(logger log.Logger, opts notificationsv1.SubscriberOptions) (background.Routine, error) {
	return notificationsv1.NewSubscriber(logger, opts)
}
