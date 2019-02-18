package eventlog

import (
	"github.com/function61/eventkit/event"
)

type Log interface {
	Append(events []event.Event) error
}
