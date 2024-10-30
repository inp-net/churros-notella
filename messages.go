package notella

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	ll "github.com/ewen-lbh/label-logger-go"
)

func (msg Message) ShouldRun() bool {
	return time.Now().After(msg.SendAt)
}

func (msg Message) Run() error {
	users, err := Receivers(msg)
	if err != nil {
		return fmt.Errorf("could not determine who to send the notification to: %w", err)
	}

	subs, err := subscriptionsOfUsers(users)
	if err != nil {
		return fmt.Errorf("could not determine which subscriptions to send the notification to: %w", err)
	}

	if len(subs) == 0 {
		ll.Warn("no subscriptions to send notification [dim]%s[reset] ([bold]%s on %s[reset]) to", msg.Id, msg.Event, msg.ChurrosObjectId)
		return nil
	}

	group, err := msg.Group()
	if err != nil {
		return fmt.Errorf("could not get churros responsible group for %s: %w", msg.ChurrosObjectId, err)
	}

	ll.Log("Sending", "green", "notification for %s on %s to %d users (%d subscriptions)", msg.Event, msg.ChurrosObjectId, len(users), len(subs))

	var wg sync.WaitGroup

	wg.Add(len(subs))

	// Parallelize sending the notifications
	for _, sub := range subs {
		ll.Debug("sending notification to %#v", sub)
		go func(wg *sync.WaitGroup, sub Subscription) {
			if sub.IsWebpush() {
				err := msg.SendWebPush(group, sub)
				if err != nil {
					ll.ErrorDisplay("could not send webpush notification", err)
				}
			} else {
				err := msg.SendToFirebase(group, sub)
				if err != nil {
					ll.ErrorDisplay("could not send firebase notification", err)
				}
			}
			wg.Done()
		}(&wg, sub)
	}

	wg.Wait()
	return nil
}

func (msg Message) JSONString() string {
	out, err := json.Marshal(msg)
	if err != nil {
		return ""
	}
	return string(out)
}
