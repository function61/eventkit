package eventlog

import (
	"github.com/function61/eventhorizon/pkg/ehevent"
)

// DEPRECATED: will soon be removed
type Log interface {
	Append(events []ehevent.Event) error
}
