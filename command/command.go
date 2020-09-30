// Command running interface => command gets some context, raises some events when it runs
package command

import (
	"context"
	"net/http"

	"github.com/function61/eventhorizon/pkg/ehevent"
)

type Command interface {
	Key() string
	Validate() error
	MiddlewareChain() string
}

// can invoke any command in typesafe manner
type Invoker interface {
	Invoke(cmdGeneric Command, ctx *Ctx) error
}

// map keyed by command name (command.Key()) values are functions that allocates a new
// specific command struct
type Allocators map[string]func() Command

// context for invoking command
type Ctx struct {
	Ctx context.Context // Go's cancellation context

	Meta ehevent.EventMeta

	RemoteAddr string
	UserAgent  string

	raisedEvents []ehevent.Event

	cookies []*http.Cookie

	// if you need to return to the client an ID of the record that was created
	createdRecordId string
}

func NewCtx(
	ctx context.Context,
	meta ehevent.EventMeta,
	remoteAddr string,
	userAgent string,
) *Ctx {
	return &Ctx{
		Ctx:          ctx,
		Meta:         meta,
		RemoteAddr:   remoteAddr,
		UserAgent:    userAgent,
		raisedEvents: []ehevent.Event{},
		cookies:      []*http.Cookie{},
	}
}

func (c *Ctx) GetRaisedEvents() []ehevent.Event {
	return c.raisedEvents
}

func (c *Ctx) RaisesEvent(event ehevent.Event) {
	c.raisedEvents = append(c.raisedEvents, event)
}

func (c *Ctx) AddCookie(cookie *http.Cookie) {
	c.cookies = append(c.cookies, cookie)
}

func (c *Ctx) CreatedRecordId(id string) {
	c.createdRecordId = id
}

func (c *Ctx) GetCreatedRecordId() string {
	return c.createdRecordId
}

func (c *Ctx) Cookies() []*http.Cookie {
	return c.cookies
}
