package notella

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	ll "github.com/gwennlbh/label-logger-go"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

var firebaseClient *firebase.App
var firebaseCtx = context.Background()

const MaxTokensPerRequest = 490

func (msg Message) SendToFirebase(groupId string, subs []Subscription) error {
	if firebaseClient == nil || !config.HasValidFirebaseServiceAccount() {
		return nil
	}

	fcm, err := firebaseClient.Messaging(firebaseCtx)
	if err != nil {
		return fmt.Errorf("while initializing FCM client: %w", err)
	}

	message := msg.FirebaseMessage(groupId)
	tokens := make([]string, len(subs))
	for i, sub := range subs {
		tokens[i] = sub.FirebaseToken()
	}

	for _, tokensChunk := range chunkBy(tokens, MaxTokensPerRequest) {
		go func(tokens []string) {
			if len(tokens) == 0 {
				return
			}
			message.Tokens = tokens
			resp, err := fcm.SendEachForMulticast(firebaseCtx, &message)
			if err != nil {
				ll.ErrorDisplay("while sending FCM message", err)
			} else if resp.FailureCount > 0 {
				fcmErrors := make([]string, 0, resp.FailureCount)
				for i, result := range resp.Responses {
					if !result.Success {
						if result.Error.Error() == "Requested entity was not found." {
							if sub, found := FindSubscriptionByNativeToken(tokens[i], subs); found {
								ll.Log("Deleting", "yellow", "invalid native subscription %s", tokens[i])
								sub.Destroy()
							}
						} else {
							fcmErrors = append(fcmErrors, fmt.Sprintf("%s: %s", tokens[i], result.Error))
						}
					}
				}
				if len(fcmErrors) > 0 {
					ll.ErrorDisplay(
						"some FCM messages failed for %d tokens",
						fmt.Errorf("- %s", strings.Join(fcmErrors, "\n- ")),
						resp.FailureCount,
					)
				}
			}
		}(tokensChunk)
	}

	return nil
}

func (msg Message) FirebaseMessage(groupId string) messaging.MulticastMessage {
	clickAction := ""
	if len(msg.Actions) > 0 {
		clickAction = msg.Actions[0].Label
	}
	return messaging.MulticastMessage{
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
	httpClient := http.DefaultClient
	if os.Getenv("DEBUG") == "1" {
		httpClient = &http.Client{
			Transport: debugTransport{t: http.DefaultTransport},
		}
	}

	ctxWithClient := context.WithValue(firebaseCtx, oauth2.HTTPClient, httpClient)
	creds, err := google.CredentialsFromJSON(ctxWithClient, []byte(config.FirebaseServiceAccount), "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		return fmt.Errorf("while setting credentials: %w", err)
	}

	client := &http.Client{
		Transport: &oauth2.Transport{
			Source: creds.TokenSource,
			Base:   httpClient.Transport,
		},
		Timeout: 10 * time.Second,
	}

	firebaseClient, err = firebase.NewApp(firebaseCtx, nil,
		option.WithCredentialsJSON([]byte(config.FirebaseServiceAccount)),
		option.WithHTTPClient(client),
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
