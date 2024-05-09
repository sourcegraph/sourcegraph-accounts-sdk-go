package v1

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/pubsub"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/lib/background"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"go.uber.org/atomic"
)

type subscriber struct {
	logger       log.Logger
	handlers     SubscriberHandlers
	subscription *pubsub.Subscription

	// state indicates the state of workers.
	state state
	// cancelContext is the function to cancel the context of the receiver that
	// effectively stops the receiver.
	cancelContext context.CancelFunc
}

type SubscriberOptions struct {
	// ProjectID is the GCP project ID that the Pub/Sub subscription belongs to. It
	// is almost always the same GCP project that the Cloud Run service is deployed
	// to.
	ProjectID string
	// SubscriptionID is the GCP Pub/Sub subscription ID to receive SAMS
	// notifications from.
	SubscriptionID string
	// ReceiveSettings is the settings for receiving messages of the subscription. A
	// zero value means to use the default settings.
	ReceiveSettings ReceiveSettings
	// Handlers is the collection of subscription handlers for each type of SAMS
	// notifications.
	Handlers SubscriberHandlers
}

func (opts SubscriberOptions) Validate() error {
	if opts.ProjectID == "" {
		return errors.New("ProjectID is required")
	}
	if opts.SubscriptionID == "" {
		return errors.New("SubscriptionID is required")
	}
	return nil
}

// NewSubscriber creates a new background routine for receiving SAMS
// notifications from given GCP project ID and Pub/Sub subscription ID.
//
// Users should prefer to use the top-level 'sams.NewNotificationsV1Subscriber'
// constructor instead.
func NewSubscriber(logger log.Logger, opts SubscriberOptions) (background.Routine, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	client, err := pubsub.NewClient(context.Background(), opts.ProjectID)
	if err != nil {
		return nil, errors.Wrap(err, "create GCP Pub/Sub client")
	}
	subscription := client.Subscription(opts.SubscriptionID)
	subscription.ReceiveSettings = opts.ReceiveSettings
	return &subscriber{
		logger:       logger.Scoped("notification.subscriber"),
		handlers:     opts.Handlers,
		subscription: subscription,
		state:        newState(),
	}, nil
}

func (s *subscriber) Start() {
	if err := s.state.toStarted(); err != nil {
		s.logger.Error("failed to start subscriber", log.Error(err))
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelContext = cancel
	err := s.subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var msgData struct {
			Name     string          `json:"name"`
			Metadata json.RawMessage `json:"metadata"`
		}
		err := json.Unmarshal(msg.Data, &msgData)
		if err != nil {
			s.logger.Error("failed to unmarshal notification message", log.Error(err))
			msg.Nack()
			return
		}

		err = s.handleReceive(msgData.Name, msgData.Metadata)
		if err == nil {
			msg.Ack()
		} else {
			s.logger.Error("failed to process notification message", log.Error(err))
			msg.Nack()
		}
	})
	if err != nil {
		s.logger.Error("failed to receive notifications", log.Error(err))
		return
	}
}

func (r *subscriber) Stop() {
	if err := r.state.toStopped(); err != nil {
		r.logger.Error("failed to stop subscriber", log.Error(err))
		return
	}

	r.cancelContext()
	r.logger.Info("subscriber stopped")
}

// SubscriberHandlers is a collection of subscription handlers for each type of
// SAMS notifications. If the handler of a notification is nil, the notification
// will be acknowledged automatically without any processing.
//
// If a handler returns an error, the notification will be unacknowledged and
// retried later.
type SubscriberHandlers struct {
	// OnUserDeleted is called when a "UserDeleted" notification is received.
	OnUserDeleted func(data *UserDeletedData) error
}

type ReceiveSettings = pubsub.ReceiveSettings

var DefaultReceiveSettings = pubsub.DefaultReceiveSettings

func (r *subscriber) handleReceive(name string, metadata json.RawMessage) error {
	switch name {
	case nameUserDeleted:
		if r.handlers.OnUserDeleted == nil {
			return nil
		}

		var data UserDeletedData
		if err := json.Unmarshal(metadata, &data); err != nil {
			return errors.Wrap(err, "unmarshal metadata")
		}
		return r.handlers.OnUserDeleted(&data)
	default:
		r.logger.Warn("acknowledging unknown notification name", log.String("name", name))
	}
	return nil
}

// state is a concurrent-safe state machine that transitions between "idle",
// "started", and "stopped" states.
type state struct {
	value *atomic.String
}

func newState() state {
	return state{value: atomic.NewString(stateIdle)}
}

func (s state) toStarted() error {
	if !s.value.CompareAndSwap(stateIdle, stateStarted) {
		return errors.Newf("not in %q state", stateIdle)
	}
	return nil
}

func (s state) toStopped() error {
	if !s.value.CompareAndSwap(stateStarted, stateStopped) {
		return errors.Newf("not in %q state", stateStarted)
	}
	return nil
}

const (
	stateIdle    = "idle"
	stateStarted = "started"
	stateStopped = "stopped"
)
