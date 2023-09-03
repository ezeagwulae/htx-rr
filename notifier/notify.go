package notifier

import (
	"context"
	"encore.app/crossing"
	"encore.dev/pubsub"
	"encore.dev/rlog"
	"fmt"
	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/twilio"
)

var _ = pubsub.NewSubscription(crossing.CrossingTransitionTopic, "send-notification", pubsub.SubscriptionConfig[*crossing.CrossingTransitionEvent]{
	Handler: func(ctx context.Context, event *crossing.CrossingTransitionEvent) error {
		// skip if no subscribers
		if len(event.Subscribers) < 1 {
			rlog.Info("skipping, no subscribers found")
			return nil
		}
		msg := fmt.Sprintf("railroad crossing on %s is closed!", event.Crossing.Name)
		if event.Open {
			msg = fmt.Sprintf("railroad crossing on %s is back open.", event.Crossing.Name)
		}

		return SendNotification(msg, event.Subscribers)
	},
})

func SendNotification(message string, recipients []string) error {
	t := newTwilioService(recipients)
	// todo: verify an error in one receiver does not cancel other receivers
	if err := t.Send(context.Background(), "railroad crossing changed", message); err != nil {
		return err
	}

	rlog.Info("notification sent", "num_subscribers", len(recipients))
	return nil
}

func newTwilioService(recipients []string) *notify.Notify {
	svc, err := twilio.New(secrets.TwilioAccountSid, secrets.TwilioAuthToken, secrets.TwilioPhoneNumber)
	if err != nil {
		rlog.Error("[twilio] failed to instantiate", "err", err)
	}
	for _, r := range recipients {
		svc.AddReceivers(r)
	}
	n := notify.New()
	n.UseServices(svc)

	return n
}

var secrets struct {
	TwilioAccountSid  string // Twilio account sid
	TwilioAuthToken   string // Twilio auth token
	TwilioPhoneNumber string // Twilio phone number
}
