package notella

import (
	"context"
	"fmt"
	"strings"

	"git.inpt.fr/churros/notella/db"
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

func CreateInDatabaseNotification(notification Message, endpoint string) error {
	_, err := prisma.Notification.CreateOne(
		db.Notification.Subscription.Link(
			db.NotificationSubscription.Endpoint.Equals(endpoint),
		),
		db.Notification.Title.Set(notification.Title),
		db.Notification.Body.Set(notification.Body),
		db.Notification.ID.Set(notification.Id),
	).Exec(context.Background())

	return err
}

func ConnectToDababase() error {
	return prisma.Connect()
}
