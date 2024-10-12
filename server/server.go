package main

import (
	"net/http"
	"sync"

	"git.inpt.fr/churros/notella"
	"git.inpt.fr/churros/notella/openapi"
	"github.com/google/uuid"
)

type Server struct{}

func NewServer() Server {
	return Server{}
}

func (Server) PostSchedule(w http.ResponseWriter, r *http.Request) {
	var req openapi.PostScheduleJSONRequestBody
	decodeRequest(w, r, &req)

	job := notella.ScheduledJob{
		ID:     uuid.New().String(),
		When:   req.When,
		Object: req.Resource,
		Event:  req.Event,
	}

	job.Schedule()
	w.WriteHeader(http.StatusCreated)
}

func (Server) PostScheduleBatch(w http.ResponseWriter, r *http.Request) {
	var req openapi.PostScheduleBatchJSONRequestBody
	decodeRequest(w, r, &req)

	var wg sync.WaitGroup

	for _, schedule := range req {
		wg.Add(1)
		job := notella.ScheduledJob{
			ID:     uuid.New().String(),
			When:   schedule.When,
			Object: schedule.Resource,
			Event:  schedule.Event,
		}

		go func(job notella.ScheduledJob, wg *sync.WaitGroup) {
			job.Schedule()
			wg.Done()
		}(job, &wg)
	}

	wg.Wait()
	w.WriteHeader(http.StatusCreated)
}
