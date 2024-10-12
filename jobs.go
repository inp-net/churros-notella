package notella

import "time"

type ScheduledJob struct {
	ID              string
	When            time.Time
	ChurrosObjectId ChurrosId
}

func (job ScheduledJob) ShouldRun() bool {
	return time.Now().After(job.When)
}

func (job ScheduledJob) Run() {

}
