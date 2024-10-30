package notella

import (
	"encoding/json"
	"fmt"

	"git.inpt.fr/churros/notella/db"
	"github.com/SherClockHolmes/webpush-go"
	ll "github.com/ewen-lbh/label-logger-go"
)

type WebPushNotification struct {
	Title              string                  `json:"title"`
	Actions            []webpushAction         `json:"actions"`
	Badge              string                  `json:"badge"`
	Icon               string                  `json:"icon"`
	Image              string                  `json:"image"`
	Body               string                  `json:"body"`
	Renotify           bool                    `json:"renotify"`
	RequireInteraction bool                    `json:"requireInteraction"`
	Silent             bool                    `json:"silent"`
	Tag                string                  `json:"tag"`
	Timestamp          int64                   `json:"timestamp"`
	Vibrate            []int                   `json:"vibrate"`
	Data               webpushNotificationData `json:"data"`
}

type webpushAction struct {
	Action string `json:"action"`
	Label  string `json:"label"`
	Icon   string `json:"icon"`
}

type webpushNotificationData struct {
	Group            string                 `json:"group"`
	Channel          db.NotificationChannel `json:"channel"`
	SubscriptionName string                 `json:"subscriptionName"`
	Goto             string                 `json:"goto"`
}

func (msg Message) WebPush(groupId string) WebPushNotification {
	actions := make([]webpushAction, len(msg.Actions))
	for i, action := range msg.Actions {
		actions[i] = webpushAction{
			Action: action.Action,
			Label:  action.Label,
			Icon:   "",
		}
	}

	return WebPushNotification{
		Title:   msg.Title,
		Actions: actions,
		Badge:   "",
		Icon:    "",
		Image:   msg.Image,
		Body:    msg.Body,
		Data: webpushNotificationData{
			Group:            groupId,
			Channel:          msg.Channel(),
			SubscriptionName: "",
			Goto:             msg.Action,
		},
	}
}

func (msg Message) SendWebPush(groupId string, sub Subscription) error {
	jsoned, err := json.Marshal(msg.WebPush(groupId))
	if err != nil {
		ll.ErrorDisplay("could not marshal notification to json", err)
	}

	resp, err := webpush.SendNotification(jsoned, &sub.Webpush, &webpush.Options{
		TTL:             30,
		Subscriber:      config.ContactEmail,
		VAPIDPublicKey:  config.VapidPublicKey,
		VAPIDPrivateKey: config.VapidPrivateKey,
	})
	if err != nil {
		return fmt.Errorf("could not send notification to %s: %w", sub.Owner.Uid, err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("could not send notification to %s: %s", sub.Owner.Uid, resp.Status)
	}

	return nil
}

func (sub Subscription) IsWebpush() bool {
	return sub.Webpush.Endpoint != ""
}
