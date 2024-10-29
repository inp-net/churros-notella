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

	ll.Log("Received", "cyan", "%-10s | %s", message.Id, message.ChurrosObjectId)
	CreateInDatabaseNotification(message, "feur")
	return nil
}
