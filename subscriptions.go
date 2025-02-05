package notella

import (
	"context"
	"fmt"

	"git.inpt.fr/churros/notella/db"
	"github.com/SherClockHolmes/webpush-go"
	ll "github.com/gwennlbh/label-logger-go"
)

type SubscriptionOwner struct {
	Id        string `json:"id"`
	Uid       string `json:"uid"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Subscription struct {
	Webpush webpush.Subscription `json:"webpush"`
	Owner   SubscriptionOwner    `json:"owner"`
}

func subscriptionsOfUsers(ids []string) (subscriptions []Subscription, err error) {
	if err := prisma.Prisma.Connect(); err != nil {
		return nil, fmt.Errorf("could not connect to prisma: %w", err)
	}
	subs, err := prisma.NotificationSubscription.FindMany(
		db.NotificationSubscription.OwnerID.In(ids),
	).With(db.NotificationSubscription.Owner.Fetch()).Exec(context.Background())

	if err != nil {
		return subscriptions, fmt.Errorf("while getting notification subscriptions from database: %w", err)
	}

	for _, sub := range subs {
		subscriptions = append(subscriptions, Subscription{
			Webpush: webpush.Subscription{
				Endpoint: sub.Endpoint,
				Keys: webpush.Keys{
					Auth:   sub.AuthKey,
					P256dh: sub.P256DhKey,
				},
			},
			Owner: SubscriptionOwner{
				Id:        sub.OwnerID,
				Uid:       sub.Owner().UID,
				FirstName: sub.Owner().FirstName,
				LastName:  sub.Owner().LastName,
			},
		})
	}

	ll.Debug("Found %d subscriptions for %d users %v", len(subscriptions), len(ids), ids)

	return subscriptions, nil
}

func (sub Subscription) Destroy() error {
	_, err := prisma.NotificationSubscription.FindUnique(
		db.NotificationSubscription.Endpoint.Equals(sub.Webpush.Endpoint),
	).Delete().Exec(context.Background())
	return err
}

func FindSubscriptionByNativeToken(token string, subs []Subscription) (Subscription, bool) {
	for _, sub := range subs {
		if sub.FirebaseToken() == token {
			return sub, true
		}
	}
	return Subscription{}, false
}
