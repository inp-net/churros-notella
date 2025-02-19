package notella

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ll "github.com/gwennlbh/label-logger-go"
	cmap "github.com/orcaman/concurrent-map/v2"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

type Schedule struct {
	cmap.ConcurrentMap[string, Message]
}

// schedules stores the scheduled messages in memory, as a mapping of job.Id -> job
var schedules Schedule = Schedule{cmap.New[Message]()}

func (job Message) Unschedule() {
	ll.Debug("Unscheduling %s", job.Id)
	schedules.Remove(job.Id)
}

// RestoreSchedule restores the scheduled messages from Redis to memory
func RestoreSchedule(eager bool) error {
	if eager {
		ll.Log("Restoring", "blue", "schedule from Redis [red][bold]eagerly[reset]")
	} else {
		ll.Log("Restoring", "blue", "schedule from Redis")

	}
	keys, err := redisClient.Keys(context.Background(), "notella:message:*").Result()
	if err != nil {
		return fmt.Errorf("while getting notella:message:* keys from redis: %w", err)
	}

	keyCountBefore := schedules.Count()

	for _, key := range keys {
		value, err := redisClient.Get(context.Background(), key).Result()
		if err != nil {
			return fmt.Errorf("while restoring schedule: could not get value for Redis key %s: %w", key, err)
		}

		var job Message
		err = json.Unmarshal([]byte(value), &job)
		if err != nil {
			return fmt.Errorf("while restoring schedule: could not unmarshal value for Redis key %s: %w", key, err)
		}

		if !eager && job.SendAt.Before(time.Now()) {
			ll.Warn("skipping restoration of %s because it's in the past: %#v", job.Id, job)
			continue
		}

		schedules.Set(job.Id, job)
	}

	ll.Log("Restored", "green", "%d scheduled jobs from Redis", schedules.Count()-keyCountBefore)

	return nil
}

// SaveSchedule saves the in-memory scheduled messages to Redis
func SaveSchedule() {
	ll.Log("Saving", "blue", "%d scheduled jobs to Redis", schedules.Count())
	for key, job := range schedules.Items() {
		go func(key string, job Message) {
			status := redisClient.Set(context.Background(), fmt.Sprintf("notella:message:%s", key), job.JSONString(), 31*24*time.Hour)
			if status.Err() != nil {
				ll.ErrorDisplay("could not save %s to Redis", status.Err(), key)
			}
		}(key, job)
	}
}

func ClearSavedSchedule() {
	ll.Log("Clearing", "yellow", "all stored scheduled jobs in Redis")
	redisClient.Del(context.Background(), redisClient.Keys(context.Background(), "notella:message:*").Val()...)
}

func ClearInMemorySchedule() {
	ll.Log("Clearing", "yellow", "all scheduled jobs")
	for _, job := range schedules.Items() {
		job.Unschedule()
	}
}

// UnscheduleAllForObject unschedules all jobs for a given object ID. If any ofType is provided, only events of the types given will be unscheduled
func UnscheduleAllForObject(objectId string, ofType ...Event) {
	var filter func(Message) bool
	if len(ofType) > 0 {
		ll.Log("Unscheduling", "yellow", "all jobs for %s of type %v", objectId, ofType)

		filter = func(job Message) bool {
			for _, t := range ofType {
				if job.Event == t {
					return true
				}
			}
			return false
		}

	} else {
		ll.Log("Unscheduling", "yellow", "all jobs for %s", objectId)

		filter = func(Message) bool { return true }
	}

	for _, job := range schedules.Items() {
		if job.ChurrosObjectId == objectId && filter(job) {
			ll.Log("Unscheduling", "yellow", "%s | %s", job.Id, job.String())
			job.Unschedule()
		}
	}
}

func DisplaySchedule() {
	ll.Log("Showing", "magenta", "%d scheduled jobs", schedules.Count())
	ll.Log("", "reset", "[dim]%-15s | %-20s | %-20s | %s", "ID", "Event", "Object ID", "Fire at")
	for _, job := range schedules.Items() {
		ll.Log("", "reset", "%-15s | %-20s | %-20s | %s", job.Id, job.Event, job.ChurrosObjectId, job.SendAt)
	}
}

func (job Message) Schedule() {
	if !job.SendAt.IsZero() {
		ll.Log("Scheduling", "magenta", "%s for %s", job.Id, job.SendAt)
	}
	schedules.Set(job.Id, job)
}

func (job Message) IsScheduled() bool {
	return schedules.Has(job.Id)
}

// StartScheduler starts the scheduler loop, which runs forever
// TODO instead of having a in-memory scheduler, use jetstream:
// 1. Get the message
// 2. job.ShouldRun? if yes, run it
// 3. otherwise, put it back at then end of the stream
// this means that we'll have to do a lot of json marshalling/unmarshalling though, since we'll have to decode the message to check if we need to run it... is there a better way?
func StartScheduler() {
	for {
		for _, job := range schedules.Items() {
			if job.ShouldRun() {
				job.Unschedule()
				go func() {
					switch job.Event {
					case EventShowScheduledJobs:
						DisplaySchedule()

					case EventRestoreSchedule:
						err := RestoreSchedule(false)
						if err != nil {
							ll.ErrorDisplay("could not restore schedule", err)
						}

					case EventRestoreScheduleEager:
						err := RestoreSchedule(true)
						if err != nil {
							ll.ErrorDisplay("could not restore schedule", err)
						}

					case EventSaveSchedule:
						SaveSchedule()

					case EventClearStoredSchedule:
						ClearSavedSchedule()

					case EventClearSchedule:
						ClearInMemorySchedule()

					default:
						ll.Log("Running", "cyan", "[dim]%s[reset] job for %s on %s", job.Id, job.Event, job.ChurrosObjectId)
						err := job.Run()
						if err != nil {
							ll.ErrorDisplay("could not run job %s", err, job.Id)
						}
						ll.Log("Ran", "green", "[dim]%s[reset] job for %s on %s", job.Id, job.Event, job.ChurrosObjectId)
					}
				}()
			}
		}
	}
}
