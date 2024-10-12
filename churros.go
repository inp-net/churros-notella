package notella

import (
	"fmt"
	"github.com/SherClockHolmes/webpush-go"
	"strings"
)

type ChurrosId struct {
	Type    string
	LocalID string
}

func ParseChurrosId(churrosId string) (ChurrosId, error) {
	parts := strings.Split(churrosId, ":")
	if len(parts) != 2 {
		return ChurrosId{}, fmt.Errorf("malformed churros global id: %q", churrosId)
	}

	return ChurrosId{
		Type:    parts[0],
		LocalID: parts[1],
	}, nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for ChurrosId.
func (id ChurrosId) UnmarshalText(text []byte) (err error) {
	s := string(text)

	id, err = ParseChurrosId(s)
	if err != nil {
		return err
	}

	return nil
}

func notificationSubscriptionsOf(userUid string) []webpush.Subscription {
	return []webpush.Subscription{}
}
