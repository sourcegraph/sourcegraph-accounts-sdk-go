package v1

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/pubsub"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Client is a client for receiving SAMS notifications from a GCP Pub/Sub
// subscription.
type Client struct {
	logger        log.Logger
	subscription  *pubsub.Subscription
	cancelContext context.CancelFunc
}

// NewClient creates a new Client for receiving SAMS notifications with given
// GCP project ID and Pub/Sub subscription ID.
func NewClient(logger log.Logger, projectID, subscriptionID string) (*Client, error) {
	client, err := pubsub.NewClient(context.Background(), projectID)
	if err != nil {
		return nil, errors.Wrap(err, "create GCP Pub/Sub client")
	}
	return &Client{
		logger:       logger.Scoped("notification.ClientV1"),
		subscription: client.Subscription(subscriptionID),
	}, nil
}

// ReceiveHandlers is a collection of receive handlers for each type of SAMS
// notifications. If the handler of a notification is nil, the notification will
// be acknowledged automatically without any processing.
//
// If a handler returns an error, the notification will be unacknowledged and
// retried later.
type ReceiveHandlers struct {
	// OnUserDeleted is called when a "UserDeleted" notification is received.
	OnUserDeleted func(data *UserDeletedData) error
}

// Receive starts receiving SAMS notifications and calls the corresponding
// handler for each notification. It blocks until the context is done (e.g.
// deadline exceed or canceled).
//
// ⚠️ WARNING: Each subscription can only have one active receiver at a time,
// i.e. there should only be one call to Receive for a given subscription.
func (c *Client) Receive(handler *ReceiveHandlers) error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancelContext = cancel
	return c.subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var msgData struct {
			Name     string          `json:"name"`
			Metadata json.RawMessage `json:"metadata"`
		}
		err := json.Unmarshal(msg.Data, &msgData)
		if err != nil {
			c.logger.Error("failed to unmarshal notification message", log.Error(err))
			msg.Nack()
			return
		}

		err = c.handleReceive(handler, msgData.Name, msgData.Metadata)
		if err != nil {
			c.logger.Error("failed to process notification message", log.Error(err))
			msg.Nack()
			return
		}
		msg.Ack()
	})
}

func (c *Client) handleReceive(handler *ReceiveHandlers, name string, metadata json.RawMessage) error {
	switch name {
	case nameUserDeleted:
		var data UserDeletedData
		err := json.Unmarshal(metadata, &data)
		if err != nil {
			return errors.Wrap(err, "unmarshal metadata")
		}
		if handler.OnUserDeleted != nil {
			return handler.OnUserDeleted(&data)
		}
	default:
		c.logger.Warn("acknowledging unknown notification name", log.String("name", name))
	}
	return nil
}

// Close stops the client from receiving notifications.
func (c *Client) Close() {
	c.cancelContext()
}
