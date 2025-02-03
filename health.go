package notella

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"firebase.google.com/go/v4/messaging"
	ll "github.com/gwennlbh/label-logger-go"
	"github.com/nats-io/nats.go"
)

type HealthResponse struct {
	Redis           bool `json:"redis"`
	NATS            bool `json:"nats"`
	ChurrosDatabase bool `json:"churros_db"`
	Firebase        bool `json:"firebase"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	ll.Log("Checking", "cyan", "health due to request from %s", r.RemoteAddr)
	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Example response (you can modify this with your own business logic)
	response := HealthResponse{}

	if err := CheckRedisHealth(); err != nil {
		ll.ErrorDisplay("while checking Redis health", err)
	} else {
		response.Redis = true
	}

	if err := CheckNATSHealth(); err != nil {
		ll.ErrorDisplay("while checking NATS health", err)
	} else {
		response.NATS = true
	}

	if err := CheckChurrosDatabaseHealth(); err != nil {
		ll.ErrorDisplay("while checking Churros database health", err)
	} else {
		response.ChurrosDatabase = true
	}

	if err := CheckFirebaseHealth(); err != nil {
		ll.ErrorDisplay("while checking Firebase Cloud Messaging health", err)
	} else {
		response.Firebase = true
	}

	// Marshal the response to JSON and write it to the response writer
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Unable to encode JSON", http.StatusInternalServerError)
		return
	}
}

func StartHealthCheckEndpoint(port int) {
	// Set up route for the /health endpoint
	http.HandleFunc("/health", healthHandler)

	// Start the server and log any errors
	ll.Log("Starting", "cyan", "health check endpoint on :%d/health", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func CheckRedisHealth() error {
	return redisClient.Ping(context.Background()).Err()
}

func CheckNATSHealth() error {
	nc, err := nats.Connect(config.NatsURL)
	if err != nil {
		return fmt.Errorf("could not connect to NATS at %s: %w", config.NatsURL, err)
	}

	_, err = nc.JetStream()
	if err != nil {
		return fmt.Errorf("could not connect to Jetstream: %w", err)
	}

	return nil
}

func CheckChurrosDatabaseHealth() error {
	return ConnectToDababase()
}

func CheckFirebaseHealth() error {
	fcm, err := firebaseClient.Messaging(firebaseCtx)
	if err != nil {
		return fmt.Errorf("while initializing messaging client: %w", err)
	}

	_, err = fcm.SendDryRun(firebaseCtx, &messaging.Message{
		Notification: &messaging.Notification{
			Title: "Health check attempt",
			Body:  "This is a health check attempt to ensure that the FCM service is working properly. The notification is not supposed to be displayed to the user.",
		},
		Token: "invalid",
	})
	if err != nil && err.Error() == "The registration token is not a valid FCM registration token" {
		return nil
	}
	return err
}
