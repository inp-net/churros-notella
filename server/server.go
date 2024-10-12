package main

import (
	"encoding/json"
	"io"
	"net/http"

	"git.inpt.fr/churros/notella"
	"git.inpt.fr/churros/notella/openapi"
	ll "github.com/ewen-lbh/label-logger-go"
	"github.com/google/uuid"
)

type Server struct{}

func NewServer() Server {
	return Server{}
}

func (Server) PostSchedule(w http.ResponseWriter, r *http.Request) {
	var req openapi.PostScheduleJSONRequestBody
	body, err := io.ReadAll(r.Body)
	if err != nil {
		ll.ErrorDisplay("could not read request body: %w", err)
		http.Error(w, "could not read request body", http.StatusBadRequest)
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		ll.ErrorDisplay("could not decode json", err)
		http.Error(w, "could not decode json", http.StatusBadRequest)
		return
	}

	ll.Debug("got request POST /schedule %+v", req)

	job := notella.ScheduledJob{
		ID:     uuid.New().String(),
		When:   req.When,
		Object: req.Resource,
		Event:  req.Event,
	}

	job.Schedule()
	w.WriteHeader(http.StatusCreated)

}
