package notella

import (
	"time"
)

type ScheduledJob struct {
	ID     string    `json:"id"`
	When   time.Time `json:"when"`
	Object ChurrosId `json:"object"`
	Event  Event     `json:"event"`
}

func (job ScheduledJob) ShouldRun() bool {
	return time.Now().After(job.When)
}

func (job ScheduledJob) Run() error {
	switch job.Event {
	// TODO
	}

	return nil
}
