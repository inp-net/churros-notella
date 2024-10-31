package notella

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ll "github.com/ewen-lbh/label-logger-go"
	cmap "github.com/orcaman/concurrent-map/v2"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

type Schedule struct {
	cmap.ConcurrentMap[string, Message]
}

var schedules Schedule = Schedule{cmap.New[Message]()}

func (job Message) Unschedule() {
	ll.Debug("Unscheduling %s", job.Id)
	schedules.Remove(job.Id)
}

// RestoreSchedule restores the scheduled messages from Redis to memory
func RestoreSchedule() error {
	ll.Log("Restoring", "blue", "schedule from Redis")
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

		if job.SendAt.Before(time.Now()) {
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

func UnscheduleAllForObject(objectId string) {
	ll.Log("Unscheduling", "yellow", "all jobs for %s", objectId)
	for _, job := range schedules.Items() {
		if job.ChurrosObjectId == objectId {
			job.Unschedule()
		}
	}
}

func DisplaySchedule() {
	ll.Log("Showing", "magenta", "%d scheduled jobs", schedules.Count())
	ll.Log("", "reset", "[dim]%-15s | %-20s | %-20s", "ID", "Event", "Object ID")
	for _, job := range schedules.Items() {
		ll.Log("", "reset", "%-15s | %-20s | %-20s", job.Id, job.Event, job.ChurrosObjectId)
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
						err := RestoreSchedule()
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
