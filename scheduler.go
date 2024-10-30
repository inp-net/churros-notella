package notella

import (
	ll "github.com/ewen-lbh/label-logger-go"
	cmap "github.com/orcaman/concurrent-map/v2"
)

var schedules = cmap.New[Message]()

func (job Message) Unschedule() {
	schedules.Remove(job.Id)
}

func (job Message) Schedule() {
	ll.Log("Scheduling", "magenta", "%s for %s", job.Id, job.SendAt)
	schedules.Set(job.Id, job)
}

func (job Message) IsScheduled() bool {
	return schedules.Has(job.Id)
}

// StartScheduler starts the scheduler loop, which runs forever
func StartScheduler() {
	for {
		for _, job := range schedules.Items() {
			if job.ShouldRun() {
				ll.Log("Running", "cyan", "job for %s on %s", job.Event, job.ChurrosObjectId)
				job.Unschedule()
				go func() {
					err := job.Run()
					if err != nil {
						ll.ErrorDisplay("could not run job %s", err, job.Id)
					}
				}()
			}
		}
	}
}
