package notella

var schedules = make(map[string]ScheduledJob)

func (job ScheduledJob) Unschedule() {
	delete(schedules, job.ID)
}

func (job ScheduledJob) Schedule() {
	schedules[job.ID] = job
}

func (job ScheduledJob) IsScheduled() bool {
	_, ok := schedules[job.ID]
	return ok
}

// StartScheduler starts the scheduler loop, which runs forever
func StartScheduler() {
	for {
		for _, job := range schedules {
			if job.ShouldRun() {
				job.Run()
			}
		}
	}
}
