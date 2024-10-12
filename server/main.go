package main

import (
	"fmt"
	"net/http"
	"time"

	"git.inpt.fr/churros/notella"
	"github.com/caarlos0/env/v11"
	"github.com/common-nighthawk/go-figure"
	ll "github.com/ewen-lbh/label-logger-go"
)

type Configuration struct {
	Port          int    `env:"PORT" envDefault:"8080"`
	ChurrosApiUrl string `env:"CHURROS_API_URL" envDefault:"http://localhost:4000/graphql"`
	PollInterval  int    `env:"POLL_INTERVAL_MS" envDefault:"500"`
}

type PostScheduleRequest struct {
	When      time.Time         `json:"when"`
	Ressource notella.ChurrosId `json:"ressource"`
}

func main() {
	figure.NewColorFigure("Notella", "", "yellow", true).Print()
	fmt.Println()

	config := Configuration{}
	err := env.Parse(&config)
	if err != nil {
		ll.ErrorDisplay("could not load env variables", err)
	}

	ll.Info("Running with config ")
	ll.Log("", "reset", "port:            [bold]%d[reset]", config.Port)
	ll.Log("", "reset", "Churros API URL: [bold]%s[reset]", config.ChurrosApiUrl)
	ll.Log("", "reset", "Poll interval:   [bold]%d[reset] ms", config.PollInterval)
	fmt.Println()

	http.HandleFunc("POST /schedule/{when}", func(w http.ResponseWriter, r *http.Request) {

	})

	ll.Info("starting scheduler")
	go notella.StartScheduler()
	ll.Info("starting server on port %d", config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}
