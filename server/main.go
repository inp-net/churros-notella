//go:generate go run github.com/steebchen/prisma-client-go generate

package main

import (
	"fmt"
	"net/http"
	"os"

	"git.inpt.fr/churros/notella"
	"git.inpt.fr/churros/notella/openapi"
	"github.com/caarlos0/env/v11"
	"github.com/common-nighthawk/go-figure"
	ll "github.com/ewen-lbh/label-logger-go"
	"github.com/joho/godotenv"
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

	ll.Info("starting server on port %d", config.Port)
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: openapi.HandlerFromMux(NewServer(), mux),
	}
	server.ListenAndServe()
}
