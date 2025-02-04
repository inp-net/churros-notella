package notella

import (
	"context"
	"fmt"
	"strings"

	"git.inpt.fr/churros/notella/db"
	ll "github.com/gwennlbh/label-logger-go"
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

func (msg Message) CreateInDatabaseNotifications(groupId string, subs []Subscription) {
	if config.DryRunMode {
		ll.Warn("dry run mode enabled, not creating notifications in database")
		return
	}
	// Create sequentially: this is not something that has to be done fast, and parallelizing would swamp the database connections
	for _, sub := range subs {
		prisma.Notification.CreateOne(
			db.Notification.Subscription.Link(
				db.NotificationSubscription.Endpoint.Equals(sub.Webpush.Endpoint),
			),
			db.Notification.Title.Set(msg.Title),
			db.Notification.Body.Set(msg.Body),
			db.Notification.ID.Set(msg.Id),
			db.Notification.Channel.Set(msg.Channel()),
			db.Notification.Group.Link(db.Group.ID.Equals(groupId)),
			db.Notification.Goto.Set(msg.Action),
			db.Notification.Timestamp.Set(msg.SendAt),
		).Exec(context.Background())
	}
}

func ConnectToDababase() error {
		return prisma.Prisma.QueryRaw("SELECT 1").Exec(context.Background(), nil)
}

// Group returns the Churros group ID responsible for the notification
func (msg Message) Group() (string, error) {
	switch msg.Event {
	case EventNewPost:
		post, err := prisma.Article.FindUnique(
			db.Article.ID.Equals(msg.ChurrosObjectId),
		).Select(
			db.Article.GroupID.Field(),
		).Exec(context.Background())
		if err != nil {
			return "", fmt.Errorf("while getting the group responsible for the notification: %w", err)
		}

		return post.GroupID, nil
	case EventShotgunClosesSoon, EventShotgunOpensSoon:
		event, err := prisma.Event.FindUnique(
			db.Event.ID.Equals(msg.ChurrosObjectId),
		).Exec(context.Background())
		if err != nil {
			return "", fmt.Errorf("while getting the group responsible for the notification: %w", err)
		}
		return event.GroupID, nil
	case EventCustom, EventGodchildAccepted, EventGodchildRejected, EventGodchildRequest, EventTest:
		return "", nil
	}

	return "", fmt.Errorf("unknown event type %q", msg.Event)
}

func (msg Message) Channel() db.NotificationChannel {
	switch msg.Event {
	case EventNewPost:
		return db.NotificationChannelArticles
	case EventShotgunClosesSoon, EventShotgunOpensSoon:
		return db.NotificationChannelShotguns
	case EventGodchildRequest, EventGodchildAccepted, EventGodchildRejected:
		return db.NotificationChannelGodparentRequests
	}

	return db.NotificationChannelOther
}
