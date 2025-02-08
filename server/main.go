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
	ll "github.com/gwennlbh/label-logger-go"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var Version = "DEV"

func main() {
	figure.NewColorFigure("Notella", "", "yellow", true).Print()
	fmt.Printf("%38s\n", fmt.Sprintf("美味しそう〜 v%s", Version))
	fmt.Println()

	// Setup a context to handle graceful shutdowns
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, _ := notella.LoadConfiguration()

	ll.Info("Server time is %s", time.Now().Format("2006-01-02 15:04:05 -07:00:00"))
	if config.DryRunMode && len(config.DryRunExceptions) > 0 {
		ll.Info("Running [bold]in dry run mode, [red]except for %+v[reset] with", config.DryRunExceptions)
	} else if config.DryRunMode {
		ll.Info("Running [bold]in dry run mode[reset] with")
	} else {
		ll.Info("Running with config")
	}
	ll.Log("", "reset", "Schedule recovery: [bold][dim]at startup [reset][bold]%s[reset]", config.StartupScheduleRestoration)
	ll.Log("", "reset", "contact email:     [bold]%s[reset]", config.ContactEmail)
	ll.Log("", "reset", "NATS URL:          [bold]%s[reset]", redactURL(config.NatsURL))
	ll.Log("", "reset", "Churros DB URL:    [bold]%s[reset]", redactURL(config.ChurrosDatabaseURL))
	ll.Log("", "reset", "Redis URL:         [bold]%s[reset]", redactURL(config.RedisURL))
	ll.Log("", "reset", "Health check on:   [bold]:%d/health[reset]", config.HealthCheckPort)
	ll.Log("", "reset", "App Package ID:    [bold]%s[reset]", config.AppPackageId)
	if config.VapidPublicKey != "" && config.VapidPrivateKey != "" {
		ll.Log("", "reset", "VAPID keys:        [bold][green]set[reset]")
	} else {
		ll.Log("", "reset", "VAPID keys:        [bold][red]not set[reset]")
	}
	if config.HasValidFirebaseServiceAccount() {
		ll.Log("", "reset", "Firebase:          [bold][green]configured[reset]")
	} else {
		ll.Log("", "reset", "Firebase:          [bold][red]unconfigured[reset]")
	}
	fmt.Println()

	if config.StartupScheduleRestoration != "disabled" {
		notella.RestoreSchedule(config.StartupScheduleRestoration == "eager")
	}
	notella.DisplaySchedule()

	ll.Info("starting scheduler")
	go notella.StartScheduler()

	ll.Log("Connecting", "cyan", "to Churros database at [bold]%s[reset]", config.ChurrosDatabaseURL)
	err := notella.ConnectToDababase()
	if err != nil {
		ll.ErrorDisplay("could not connect to database", err)
	}

	ll.Log("Connecting", "cyan", "to NATS server at [bold]%s[reset]", config.NatsURL)
	nc, err := nats.Connect(config.NatsURL)
	if err != nil {
		ll.ErrorDisplay("could not connect to NATS at %s", err, config.NatsURL)
		return
	}

	js, err := jetstream.New(nc)
	if err != nil {
		ll.ErrorDisplay("could not connect to Jetstream", err)
		return
	}

	ll.Log("Initializing", "cyan", "a Jetstream stream [bold]%s[reset], listening for subject [bold]%s[reset]", notella.StreamName, notella.SubjectName)

	stream, err := js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     notella.StreamName,
		Subjects: []string{notella.SubjectName},
	})
	if err != nil {
		ll.ErrorDisplay("could not create stream", err)
		return
	}

	ll.Log("Initializing", "cyan", "Jetstream consumer [bold]NotellaConsumer[reset] with [bold]AckExplicitPolicy[reset]")

	consumer, err := stream.CreateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:   "NotellaConsumer",
		Name:      "NotellaConsumer",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		ll.ErrorDisplay("could not create consumer", err)
		return
	}

	ll.Log("Starting", "cyan", "consumer [bold]NotellaConsumer[reset]")

	cc, err := consumer.Consume(
		func(msg jetstream.Msg) {
			err := notella.NatsReceiver(msg)
			if err != nil {
				ll.ErrorDisplay("Could not process message", err)
			}
			msg.Ack() // Acknowledge the message
		},
		jetstream.ConsumeErrHandler(func(_ jetstream.ConsumeContext, err error) {
			ll.ErrorDisplay("Error while consuming message", err)
		}),
	)
	if err != nil {
		ll.ErrorDisplay("could not start consumer", err)
		return
	}

	defer cc.Stop()

	// Capture OS signals for graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		ll.Log("Shuting down", "magenta", "because of signal received")
		cancel()
	}()

	// Start healthcheck endpoint
	go notella.StartHealthCheckEndpoint(config.HealthCheckPort)

	// Send EventShowScheduledJobs to the stream every 5 minutes and save schedule to redis
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(5 * time.Minute)
				notella.DisplaySchedule()
				notella.SaveSchedule()
			}
		}
	}()

	// Block until the context is canceled (i.e., server shutdown signal received)
	<-ctx.Done()
	ll.Log("Stopped", "red", "server")
}
