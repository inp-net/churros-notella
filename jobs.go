package notella

import (
	"fmt"
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
	subscriptions, err := notificationSubscriptionsOf("versairea")
	if err != nil {
		return fmt.Errorf("while getting notification subscriptions for %s: %w", "versairea", err)
	}

	fmt.Printf("%+v\n", subscriptions)

	switch job.Event {
	// TODO
	}

	return nil
}
