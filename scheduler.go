package notella

import (
	ll "github.com/ewen-lbh/label-logger-go"
	cmap "github.com/orcaman/concurrent-map/v2"
)


var schedules = cmap.New[ScheduledJob]()

func (job ScheduledJob) Unschedule() {
	schedules.Remove(job.ID)
}

func (job ScheduledJob) Schedule() {
	schedules.Set(job.ID, job)
}

func (job ScheduledJob) IsScheduled() bool {
	return schedules.Has(job.ID)
}

// StartScheduler starts the scheduler loop, which runs forever
func StartScheduler() {
	for {
		for _, job := range schedules.Items() {
			if job.ShouldRun() {
				ll.Log("Running", "cyan", "job for %s on %s", job.Event, job.Object)
				job.Unschedule()
				go job.Run()
			}
		}
	}
}
