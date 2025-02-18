package notella

import (
	"encoding/json"
	"fmt"

	ll "github.com/gwennlbh/label-logger-go"
	"github.com/nats-io/nats.go/jetstream"
)

const StreamName = "notella:stream"
const SubjectName = "notella:notification"
const ConsumerName = "NotellaConsumer"

func NatsReceiver(m jetstream.Msg) error {
	var message Message
	err := json.Unmarshal(m.Data(), &message)
	if err != nil {
		return fmt.Errorf("while unmarshaling received message: %w", err)
	}

	if message.Event != EventShowScheduledJobs {
		ll.Log("Received", "cyan", "%-10s | %-10s on %s", message.Id, message.Event, message.ChurrosObjectId)
	}

	if message.ClearSchedule {
		UnscheduleAllForObject(message.ChurrosObjectId)
	}

	message.Schedule()

	return nil
}
