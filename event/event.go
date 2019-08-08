// Event structure. Event has a name, occurrs at some point in time and is raised by some user. Serializes to JSON
package event

import (
	"github.com/function61/gokit/cryptorandombytes"
	"time"
)

type Event interface {
	Meta() *EventMeta
	MetaType() string
	Serialize() string
}

type EventMeta struct {
	Timestamp time.Time
	UserId    string
}

func Meta(timestamp time.Time, userId string) EventMeta {
	return EventMeta{
		Timestamp: timestamp,
		UserId:    userId,
	}
}

func RandomId() string {
	return cryptorandombytes.Hex(4)
}
