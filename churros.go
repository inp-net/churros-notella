package notella

import (
	"context"
	"fmt"
	"strings"

	"git.inpt.fr/churros/notella/db"
	"github.com/SherClockHolmes/webpush-go"
)

var prisma = db.NewClient()

type ChurrosId struct {
	Type    string
	LocalID string
}

func (id ChurrosId) String() string {
	return fmt.Sprintf("%s:%s", id.Type, id.LocalID)
}

func ParseChurrosId(churrosId string) (ChurrosId, error) {
	parts := strings.Split(churrosId, ":")
	if len(parts) != 2 {
		return ChurrosId{}, fmt.Errorf("malformed churros global id: %q", churrosId)
	}

	return ChurrosId{
		Type:    parts[0],
		LocalID: parts[1],
	}, nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for ChurrosId.
func (id *ChurrosId) UnmarshalText(text []byte) error {
	s := string(text)

	parsed, err := ParseChurrosId(s)
	if err != nil {
		return err
	}

	id.Type = parsed.Type
	id.LocalID = parsed.LocalID

	return nil
}

func notificationSubscriptionsOf(userUid string) (subscriptions []webpush.Subscription, err error) {
	if err := prisma.Prisma.Connect(); err != nil {
		return nil, fmt.Errorf("could not connect to prisma: %w", err)
	}
	subs, err := prisma.NotificationSubscription.FindMany(
		db.NotificationSubscription.Owner.Where(db.User.UID.Equals(userUid)),
	).Exec(context.Background())

	if err != nil {
		return subscriptions, fmt.Errorf("while getting notification subscriptions from database: %w", err)
	}

	for _, sub := range subs {
		subscriptions = append(subscriptions, webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				Auth:   sub.AuthKey,
				P256dh: sub.P256DhKey,
			},
		})
	}

	return subscriptions, nil
}
