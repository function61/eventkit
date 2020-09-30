// Generic command dispatcher boilerplate. Takes in POST /api/command/... request and
// dispatches it to the command handler, and submits possible events to an event log.
package httpcommand

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/function61/eventhorizon/pkg/ehevent"
	"github.com/function61/eventkit/command"
	"github.com/function61/eventkit/eventlog"
	"github.com/function61/gokit/httpauth"
)

const (
	CreatedRecordIdHeaderKey = "x-created-record-id"
)

var noResponse = NewHttpError(0, "", "")

// error that is sent as a JSON-formatted error
type HttpError struct {
	StatusCode  int // if 0, means errored but error response already sent by middleware
	ErrorCode   string
	Description string
}

func NewHttpError(statusCode int, errorCode string, description string) *HttpError {
	return &HttpError{statusCode, errorCode, description}
}

func (r *HttpError) Error() string {
	if r.Description != "" {
		return r.ErrorCode + ": " + r.Description
	} else {
		return r.ErrorCode
	}
}

func (r *HttpError) ErrorResponseAlreadySentByMiddleware() bool {
	return r.StatusCode == 0
}

func badRequest(errorCode string, description string) *HttpError {
	return NewHttpError(http.StatusBadRequest, errorCode, description)
}

func Serve(
	w http.ResponseWriter,
	r *http.Request,
	mwares httpauth.MiddlewareChainMap,
	commandName string,
	allocators command.Allocators,
	invoker command.Invoker,
	eventLog eventlog.Log,
) *HttpError {
	allocator, commandExists := allocators[commandName]
	if !commandExists {
		return badRequest("unsupported_command", "")
	}

	cmdStruct := allocator()

	middlewareChain := mwares[cmdStruct.MiddlewareChain()]
	reqCtx := middlewareChain(w, r)
	if reqCtx == nil {
		return noResponse // middleware dealt with error response
	}

	userId := ""
	if reqCtx.User != nil {
		userId = reqCtx.User.Id
	}

	if r.Header.Get("Content-Type") != "application/json" {
		return badRequest("expecting_content_type_json", "expecting Content-Type header with application/json")
	}

	jsonDecoder := json.NewDecoder(r.Body)
	jsonDecoder.DisallowUnknownFields()
	if errJson := jsonDecoder.Decode(cmdStruct); errJson != nil {
		return badRequest("json_parsing_failed", errJson.Error())
	}

	ctx := command.NewCtx(
		r.Context(),
		ehevent.Meta(time.Now(), userId),
		r.RemoteAddr,
		r.Header.Get("User-Agent"))

	if herr := InvokeSkippingAuthorization(cmdStruct, ctx, invoker, eventLog); herr != nil {
		return herr
	}

	for _, cookie := range ctx.Cookies() {
		http.SetCookie(w, cookie)
	}

	if id := ctx.GetCreatedRecordId(); id != "" {
		w.Header().Set(CreatedRecordIdHeaderKey, id)
	}

	return nil
}

// validates command, invokes it and pushes raised events to event log
//
// "SkippingAuthorization" suffix to warn that no authorization checks are performed
// (i.e. this should only be used for events not triggered by user)
func InvokeSkippingAuthorization(
	cmdStruct command.Command,
	ctx *command.Ctx,
	invoker command.Invoker,
	eventLog eventlog.Log,
) *HttpError {
	if errValidate := cmdStruct.Validate(); errValidate != nil {
		return badRequest("command_validation_failed", errValidate.Error())
	}

	if errInvoke := invoker.Invoke(cmdStruct, ctx); errInvoke != nil {
		// see if returned error is already an *HttpError
		if httpErr, is := errInvoke.(*HttpError); is {
			return httpErr // use as-is
		} else {
			return badRequest("command_failed", errInvoke.Error())
		}
	}

	if err := eventLog.Append(ctx.GetRaisedEvents()); err != nil {
		return NewHttpError(http.StatusInternalServerError, "event_append_failed", err.Error())
	}

	return nil
}
