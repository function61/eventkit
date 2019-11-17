// Command running interface => command gets some context, raises some events when it runs
package command

import (
	"context"
	"github.com/function61/eventkit/event"
	"net/http"
)

type Ctx struct {
	Ctx context.Context // Go's cancellation context

	Meta event.EventMeta

	RemoteAddr string
	UserAgent  string

	raisedEvents []event.Event

	cookies []*http.Cookie

	// if you need to return to the client an ID of the record that was created
	createdRecordId string
}

func NewCtx(
	ctx context.Context,
	meta event.EventMeta,
	remoteAddr string,
	userAgent string,
) *Ctx {
	return &Ctx{
		Ctx:          ctx,
		Meta:         meta,
		RemoteAddr:   remoteAddr,
		UserAgent:    userAgent,
		raisedEvents: []event.Event{},
		cookies:      []*http.Cookie{},
	}
}

func (c *Ctx) GetRaisedEvents() []event.Event {
	return c.raisedEvents
}

func (c *Ctx) RaisesEvent(event event.Event) {
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

type Command interface {
	Key() string
	Validate() error
	MiddlewareChain() string
	Invoke(ctx *Ctx, handlers interface{}) error
}

// map keyed by command name (command.Key()) values are functions that allocates a new
// specific command struct
type AllocatorMap map[string]func() Command
