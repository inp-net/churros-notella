package notella

import (
	"context"
	"fmt"

	"git.inpt.fr/churros/notella/db"
	"github.com/SherClockHolmes/webpush-go"
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

	return subscriptions, nil
}
