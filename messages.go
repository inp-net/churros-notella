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

	// Separate native and webpush subscriptions
	nativeSubs := make([]Subscription, 0, len(subs))
	webpushSubs := make([]Subscription, 0, len(subs))
	for _, sub := range subs {
		if sub.IsNative() {
			nativeSubs = append(nativeSubs, sub)
		} else if sub.IsWebpush() {
			webpushSubs = append(webpushSubs, sub)
		} else {
			ll.Warn("invalid subscription %#v", sub)
		}
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		msg.CreateInDatabaseNotifications(group, subs)
	}(&wg)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := msg.SendToFirebase(group, nativeSubs)
		if err != nil {
			ll.ErrorDisplay("could not send notification via firebase", err)
		}
	}(&wg)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := msg.SendWebPush(group, webpushSubs)
		if err != nil {
			ll.ErrorDisplay("could not send notification via webpush", err)
		}
	}(&wg)
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

func (msg Message) JSONBytes() []byte {
	out, err := json.Marshal(msg)
	if err != nil {
		return nil
	}
	return out
}
