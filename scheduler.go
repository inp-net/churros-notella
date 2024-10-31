package notella

import (
	ll "github.com/ewen-lbh/label-logger-go"
	cmap "github.com/orcaman/concurrent-map/v2"
)

var schedules = cmap.New[Message]()

func (job Message) Unschedule() {
	schedules.Remove(job.Id)
}

func UnscheduleAllForObject(objectId string) {
	ll.Log("Unscheduling", "yellow", "all jobs for %s", objectId)
	for _, job := range schedules.Items() {
		if job.ChurrosObjectId == objectId {
			job.Unschedule()
		}
	}
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
				ll.Log("Running", "cyan", "[dim]%s[reset] job for %s on %s", job.Id, job.Event, job.ChurrosObjectId)
				job.Unschedule()
				go func() {
					err := job.Run()
					if err != nil {
						ll.ErrorDisplay("could not run job %s", err, job.Id)
					}
					ll.Log("Ran", "green", "[dim]%s[reset] job for %s on %s", job.Id, job.Event, job.ChurrosObjectId)
				}()
			}
		}
	}
}
