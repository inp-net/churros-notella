package notella

import (
	"encoding/json"
	"fmt"

	ll "github.com/ewen-lbh/label-logger-go"
	"github.com/nats-io/nats.go"
)

const StreamName = "notella:stream"
const SubjectName = "notella:notification"

func NatsReceiver(m *nats.Msg) error {
	var message Message
	err := json.Unmarshal(m.Data, &message)
	if err != nil {
		return fmt.Errorf("while unmarshaling received message: %w", err)
	}

	if message.Event != EventShowScheduledJobs {
		ll.Log("Received", "cyan", "%-10s | %-10s on %s", message.Id, message.Event, message.ChurrosObjectId)
	}

	if message.Event == EventClearScheduledJobs {
		UnscheduleAllForObject(message.ChurrosObjectId)
		return nil
	}

	message.Schedule()

	return nil
}
