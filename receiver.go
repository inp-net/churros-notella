package notella

import (
	ll "github.com/ewen-lbh/label-logger-go"
	"github.com/nats-io/nats.go"
)

const StreamName = "notella:stream"
const SubjectName = "notella:notification"

func NatsReceiver(m *nats.Msg) {
	ll.Log("Received", "cyan", "message: %s", string(m.Data))
}
