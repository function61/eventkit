// Generic command dispatcher boilerplate. Takes in POST /api/command/... request and
// dispatches it to the command handler, and submits possible events to an event log.
package httpcommand

import (
	"encoding/json"
	"github.com/function61/eventkit/command"
	"github.com/function61/eventkit/event"
	"github.com/function61/eventkit/eventlog"
	"github.com/function61/gokit/httpauth"
	"net/http"
	"time"
)

const (
	CreatedRecordIdHeaderKey = "x-created-record-id"
)

type HttpError struct {
	StatusCode  int // if 0, means errored but error response already sent by middleware
	ErrorCode   string
	Description string
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
	return customError(errorCode, description, http.StatusBadRequest)
}

func noResponse() *HttpError {
	return &HttpError{}
}

func customError(errorCode string, description string, statusCode int) *HttpError {
	return &HttpError{
		ErrorCode:   errorCode,
		Description: description,
		StatusCode:  statusCode,
	}
}

func Serve(
	w http.ResponseWriter,
	r *http.Request,
	mwares httpauth.MiddlewareChainMap,
	commandName string,
	allocators command.AllocatorMap,
	handlers interface{},
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
		return noResponse() // middleware dealt with error response
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
		event.Meta(time.Now(), userId),
		r.RemoteAddr,
		r.Header.Get("User-Agent"))

	if herr := InvokeSkippingAuthorization(cmdStruct, ctx, handlers, eventLog); herr != nil {
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
	handlers interface{},
	eventLog eventlog.Log,
) *HttpError {
	if errValidate := cmdStruct.Validate(); errValidate != nil {
		return badRequest("command_validation_failed", errValidate.Error())
	}

	if errInvoke := cmdStruct.Invoke(ctx, handlers); errInvoke != nil {
		return badRequest("command_failed", errInvoke.Error())
	}

	raisedEvents := ctx.GetRaisedEvents()

	if err := eventLog.Append(raisedEvents); err != nil {
		return customError("event_append_failed", err.Error(), http.StatusInternalServerError)
	}

	return nil
}
