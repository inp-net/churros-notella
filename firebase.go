package notella

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var firebaseClient *firebase.App

func (msg Message) SendToFirebase(groupId string, sub Subscription) error {
	fcm, err := firebaseClient.Messaging(context.Background())
	if err != nil {
		return fmt.Errorf("while initializing FCM client: %w", err)
	}

	message := msg.FirebaseMessage(groupId)
	message.Token = sub.FirebaseToken()
	_, err = fcm.Send(context.Background(), &message)

	return err
}

func (msg Message) FirebaseMessage(groupId string) messaging.Message {
	clickAction := ""
	if len(msg.Actions) > 0 {
		clickAction = msg.Actions[0].Label
	}
	return messaging.Message{
		Data: map[string]string{
			"original": msg.JSONString(),
		},
		Android: &messaging.AndroidConfig{
			RestrictedPackageName: config.AppPackageId,
			Notification: &messaging.AndroidNotification{
				VibrateTimingMillis: []int64{}, // TODO
				EventTimestamp:      nil,       //  TODO
				ClickAction:         clickAction,
			},
		},
		Notification: &messaging.Notification{
			Title:    msg.Title,
			Body:     msg.Body,
			ImageURL: msg.Image,
		},
	}
}

type firebaseServiceAccount struct {
	Type                    string `json:"type"`
	ProjectId               string `json:"project_id"`
	PrivateKeyId            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientId                string `json:"client_id"`
	AuthUri                 string `json:"auth_uri"`
	TokenUri                string `json:"token_uri"`
	AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
	ClientX509CertUrl       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

func setupFirebaseClient() (err error) {
	firebaseClient, err = firebase.NewApp(context.Background(),
		&firebase.Config{},
		option.WithCredentialsJSON([]byte(config.FirebaseServiceAccount)),
	)
	return
}

func (config Configuration) HasValidFirebaseServiceAccount() bool {
	var serviceAccount firebaseServiceAccount
	err := json.Unmarshal([]byte(config.FirebaseServiceAccount), &serviceAccount)
	if err != nil {
		return false
	}
	if serviceAccount.Type != "service_account" {
		return false
	}

	if err = setupFirebaseClient(); err != nil {
		return false
	}

	return true
}

func (sub Subscription) FirebaseToken() string {
	return strings.TrimPrefix(strings.TrimPrefix(sub.Webpush.Endpoint, "apns://"), "firebase://")
}

func (sub Subscription) IsNative() bool {
	return !sub.IsWebpush()
}
