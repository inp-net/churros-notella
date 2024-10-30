//go:generate go run github.com/steebchen/prisma-client-go generate

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.inpt.fr/churros/notella"
	"github.com/common-nighthawk/go-figure"
	ll "github.com/ewen-lbh/label-logger-go"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

var Version = "DEV"

func main() {
	figure.NewColorFigure("Notella", "", "yellow", true).Print()
	fmt.Printf("%38s\n", fmt.Sprintf("美味しそう〜 v%s", Version))
	fmt.Println()

	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			ll.ErrorDisplay("could not load .env file", err)
		}
		ll.Info("loaded .env file")
	}

	config, _ := notella.LoadConfiguration()

	ll.Info("Running with config ")
	ll.Log("", "reset", "port:            [bold]%d[reset]", config.Port)
	ll.Log("", "reset", "contact email:   [bold]%s[reset]", config.ContactEmail)
	ll.Log("", "reset", "Churros DB URL:  [bold]%s[reset]", redactURL(config.ChurrosDatabaseURL))
	if config.VapidPublicKey != "" && config.VapidPrivateKey != "" {
		ll.Log("", "reset", "VAPID keys:      [bold][green]set[reset]")
	} else {
		ll.Log("", "reset", "VAPID keys:      [bold][red]not set[reset]")
	}
	fmt.Println()

	ll.Info("starting scheduler")
	go notella.StartScheduler()

	ll.Log("Connecting", "cyan", "to Churros database at [bold]%s[reset]", config.ChurrosDatabaseURL)
	err := notella.ConnectToDababase()
	if err != nil {
		ll.ErrorDisplay("could not connect to database", err)
	}

	ll.Log("Connecting", "cyan", "to NATS server at [bold]%s[reset]", nats.DefaultURL)
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		ll.ErrorDisplay("could not connect to NATS at %s", err, nats.DefaultURL)
		return
	}

	js, err := nc.JetStream()
	if err != nil {
		ll.ErrorDisplay("could not connect to Jetstream", err)
		return
	}

	ll.Log("Initializing", "cyan", "a Jetstream stream [bold]%s[reset], listening for subject [bold]%s[reset]", notella.StreamName, notella.SubjectName)

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     notella.StreamName,
		Subjects: []string{notella.SubjectName},
	})
	if err != nil {
		ll.ErrorDisplay("could not create stream", err)
		return
	}

	ll.Log("Initializing", "cyan", "Jetstream consumer [bold]NotellaConsumer[reset] with [bold]AckExplicitPolicy[reset]")

	_, err = js.AddConsumer(notella.StreamName, &nats.ConsumerConfig{
		Durable:   "NotellaConsumer",
		AckPolicy: nats.AckExplicitPolicy,
	})
	if err != nil {
		ll.ErrorDisplay("could not create consumer", err)
		return
	}

	ll.Log("Starting", "cyan", "consumer [bold]NotellaConsumer[reset]")

	sub, err := js.PullSubscribe(notella.SubjectName, "NotellaConsumer")
	if err != nil {
		ll.ErrorDisplay("could not start consumer", err)
		return
	}

	// Setup a context to handle graceful shutdowns
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture OS signals for graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		ll.Log("Shuting down", "magenta", "because of signal received")
		cancel()
	}()

	// Continuously fetch and process messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Fetch messages in batches
				msgs, err := sub.Fetch(10, nats.MaxWait(5*time.Second))
				if err != nil && err != nats.ErrTimeout {
					ll.ErrorDisplay("Could not fetch messages", err)
					time.Sleep(2 * time.Second) // Wait before retrying
					continue
				}

				// Process each message
				for _, msg := range msgs {
					err = notella.NatsReceiver(msg)
					if err != nil {
						ll.ErrorDisplay("Could not process message", err)
					}
					msg.Ack() // Acknowledge the message
				}
			}
		}
	}()

	// Block until the context is canceled (i.e., server shutdown signal received)
	<-ctx.Done()
	ll.Log("Stopped", "red", "server")
}
