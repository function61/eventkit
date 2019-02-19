package httpcommand

// Generic command dispatcher boilerplate. Takes in POST /api/command/... request and
// dispatches it to the command handler, and submits possible events to an event log.

import (
	"encoding/json"
	"github.com/function61/eventkit/command"
	"github.com/function61/eventkit/event"
	"github.com/function61/eventkit/eventlog"
	"github.com/function61/gokit/httpauth"
	"net/http"
	"time"
)

type HttpError struct {
	StatusCode  int // if 0, means errored but error response already sent by middleware
	ErrorCode   string
	Description string
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

	if errValidate := cmdStruct.Validate(); errValidate != nil {
		return badRequest("command_validation_failed", errValidate.Error())
	}

	ctx := command.NewCtx(
		event.Meta(time.Now(), userId),
		r.RemoteAddr,
		r.Header.Get("User-Agent"))

	if errInvoke := cmdStruct.Invoke(ctx, handlers); errInvoke != nil {
		return badRequest("command_failed", errInvoke.Error())
	}

	raisedEvents := ctx.GetRaisedEvents()

	if err := eventLog.Append(raisedEvents); err != nil {
		return customError("event_append_failed", err.Error(), http.StatusInternalServerError)
	}

	for _, cookie := range ctx.Cookies() {
		http.SetCookie(w, cookie)
	}

	return nil
}
