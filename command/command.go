package command

import (
	"github.com/function61/eventkit/event"
	"net/http"
)

type Ctx struct {
	Meta event.EventMeta

	RemoteAddr string
	UserAgent  string

	raisedEvents []event.Event

	cookies []*http.Cookie
}

func NewCtx(meta event.EventMeta, remoteAddr string, userAgent string) *Ctx {
	return &Ctx{
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
