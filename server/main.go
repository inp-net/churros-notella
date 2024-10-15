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
	"github.com/caarlos0/env/v11"
	"github.com/common-nighthawk/go-figure"
	ll "github.com/ewen-lbh/label-logger-go"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Configuration struct {
	Port               int    `env:"PORT" envDefault:"8080"`
	ChurrosApiUrl      string `env:"CHURROS_API_URL" envDefault:"http://localhost:4000/graphql"`
	PollInterval       int    `env:"POLL_INTERVAL_MS" envDefault:"500"`
	RedisURL           string `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
	ChurrosDatabaseURL string `env:"DATABASE_URL"`
}

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

	config := Configuration{}
	err := env.Parse(&config)
	if err != nil {
		ll.ErrorDisplay("could not load env variables", err)
	}

	ll.Info("Running with config ")
	ll.Log("", "reset", "port:            [bold]%d[reset]", config.Port)
	ll.Log("", "reset", "Churros API URL: [bold]%s[reset]", redactURL(config.ChurrosApiUrl))
	ll.Log("", "reset", "Churros DB URL:  [bold]%s[reset]", redactURL(config.ChurrosDatabaseURL))
	ll.Log("", "reset", "Redis URL:       [bold]%s[reset]", redactURL(config.RedisURL))
	ll.Log("", "reset", "Poll interval:   [bold]%d[reset] ms", config.PollInterval)
	fmt.Println()

	ll.Info("starting scheduler")
	go notella.StartScheduler()

	ll.Log("Connecting", "cyan", "to NATS server at [bold]%s[reset]", nats.DefaultURL)
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		ll.ErrorDisplay("could not connect to NATS at %s", err, nats.DefaultURL)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
	defer cancel()

	js, err := jetstream.New(nc)
	if err != nil {
		ll.ErrorDisplay("could not connect to Jetstream", err)
		return
	}

	ll.Log("Initializing", "cyan", "a Jetstream stream [bold]notella:stream[reset], listening for [bold]notella:*[reset] subjects")

	s, err := js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     "notella:stream",
		Subjects: []string{"notella:*"},
	})
	if err != nil {
		ll.ErrorDisplay("could not create stream", err)
		return
	}

	ll.Log("Initializing", "cyan", "Jetstream consumer [bold]NotellaConsumer[reset] with [bold]AckExplicitPolicy[reset]")

	cons, err := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:   "NotellaConsumer",
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		ll.ErrorDisplay("could not create consumer", err)
		return
	}

	ll.Log("Starting", "cyan", "consumer [bold]NotellaConsumer[reset]")

	cc, err := cons.Consume(func(msg jetstream.Msg) {
		fmt.Println(string(msg.Data()))
		msg.Ack()
	}, jetstream.ConsumeErrHandler(func(consumeCtx jetstream.ConsumeContext, err error) {
		fmt.Println(err)
	}))
	if err != nil {
		ll.ErrorDisplay("could not start consumer", err)
		return
	}
	defer cc.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
