package codegentemplates

const BackendCommandsDefinitions = `package {{.Module.Id}}

// WARNING: generated file

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"github.com/function61/eventkit/command"
{{if .CommandsImports.Date}}	"github.com/function61/eventkit/guts"{{end}}
)

// handlers

type CommandHandlers interface { {{range .Module.Commands}}
	{{.AsGoStructName}}(*{{.AsGoStructName}}, *command.Ctx) error{{end}}
}

// invoker

func CommandInvoker(handlers CommandHandlers) command.Invoker {
	return &invoker{handlers}
}

type invoker struct {
	handlers CommandHandlers
}

func (i *invoker) Invoke(cmdGeneric command.Command, ctx *command.Ctx) error {
	switch cmd := cmdGeneric.(type) { {{range .Module.Commands}}
	case *{{.AsGoStructName}}:
		return i.handlers.{{.AsGoStructName}}(cmd, ctx){{end}}
	default:
		// should not ever happen, because this is asserted in httpcommand
		return fmt.Errorf("unknown command: " + cmdGeneric.Key())
	}
}

// structs

{{range .Module.Commands}}
type {{.AsGoStructName}} struct { {{range .Fields}}
	{{.Key}} {{.AsGoType $.Module}} ` + "`json:\"{{.Key}}\"`" + `{{end}}
}

func (x *{{.AsGoStructName}}) Validate() error {
	{{.MakeValidation $.Module}}

	return nil
}

func (x *{{.AsGoStructName}}) MiddlewareChain() string { return "{{.MiddlewareChain}}" }
func (x *{{.AsGoStructName}}) Key() string { return "{{.Command}}" }
{{end}}

// allocators

var Allocators = command.Allocators{
{{range .Module.Commands}}
	"{{.Command}}": func() command.Command { return &{{.AsGoStructName}}{} },{{end}}
}

// util functions

func regexpValidation(fieldName string, pattern string, content string) error {
	if !regexp.MustCompile(pattern).MatchString(content) {
		return fmt.Errorf("field %s does not match pattern %s", fieldName, pattern)
	}

	return nil
}

func noNewlinesValidation(fieldName string, content string) error {
	if strings.ContainsAny(content, "\r\n") {
		return errors.New("single-line field " + fieldName + " contains newlines")
	}

	return nil
}

func fieldEmptyValidationError(fieldName string) error {
	return errors.New("field " + fieldName + " cannot be empty")
}

func fieldLengthValidationError(fieldName string, maxLength int, got int) error {
	return fmt.Errorf("field %s exceeded maximum length %d (got %d)", fieldName, maxLength, got)
}
`

const BackendEventDefinitions = `package {{.Module.Id}}

import (
{{if .EventsImports.DateTime}}	"time"
{{end}}{{if .EventsImports.Date}}	"github.com/function61/eventkit/guts"
{{end}}	"github.com/function61/eventhorizon/pkg/ehevent"
)

// WARNING: generated file

var EventTypes = ehevent.Allocators{
{{range .EventDefs}}
	"{{.EventKey}}": func() ehevent.Event { return &{{.GoStructName}}{meta: &ehevent.EventMeta{}} },{{end}}
}


{{.EventStructsAsGoCode}}


// constructors

{{range .EventDefs}}
func New{{.GoStructName}}({{.CtorArgs}}) *{{.GoStructName}} {
	return &{{.GoStructName}}{
		meta: &meta,
		{{.CtorAssignments}}
	}
}
{{end}}

{{range .EventDefs}}
func (e *{{.GoStructName}}) Meta() *ehevent.EventMeta { return e.meta }{{end}}

{{range .EventDefs}}
func (e *{{.GoStructName}}) MetaType() string { return "{{.EventKey}}" }{{end}}

// interface

type EventListener interface { {{range .EventDefs}}
	Apply{{.GoStructName}}(*{{.GoStructName}}) error{{end}}

	HandleUnknownEvent(event ehevent.Event) error
}

func DispatchEvent(event ehevent.Event, listener EventListener) error {
	switch e := event.(type) { {{range .EventDefs}}
	case *{{.GoStructName}}:
		return listener.Apply{{.GoStructName}}(e){{end}}
	default:
		return listener.HandleUnknownEvent(event)
	}
}
`

const BackendTypes = `package {{.Module.Id}}

import ( {{if .StringEnums}}
	"fmt"
	"encoding/json"{{end}}
{{if .TypesImports.Date}}	"github.com/function61/eventkit/guts"
{{end}}{{if .TypesImports.DateTime}}	"time"
{{end}}{{range .TypesImports.ModuleIds}}
	"{{$.Opts.BackendModulePrefix}}{{.}}"
{{end}}
)

{{range .Module.Types.Types}}
{{.AsToGoCode}}
{{end}}

{{range $_, $enum := .StringEnums}}
type {{$enum.Name}} string
const (
{{range $_, $member := $enum.Members}}
	{{$member.GoKey}} {{$enum.Name}} = "{{$member.GoValue}}"{{end}}
)

var {{$enum.Name}}Members = []{{$enum.Name}}{ {{range $_, $member := $enum.Members}}
	{{$member.GoKey}},{{end}}
}

func (e *{{$enum.Name}}) MarshalJSON() ([]byte, error) {
	str := string(*e)
	return json.Marshal(&str)
}

func (e *{{$enum.Name}}) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	validated, err := {{$enum.Name}}Validate(str)
	if err != nil {
		return err
	}
	*e = validated
	return nil
}

func {{$enum.Name}}Validate(input string) ({{$enum.Name}}, error) {
	for _, member := range {{$enum.Name}}Members {
		if member == {{$enum.Name}}(input) {
			return member, nil
		}
	}

	return "", fmt.Errorf("invalid {{$enum.Name}} member: %s", input)
}

// digest in name because there's no easy way to make exhaustive Enum pattern matching
// in Go, so we hack around it by calling this generated function everywhere we want
// to do the pattern match, and when enum members change the digest changes and thus
// it forces you to manually review and fix each place
func {{$enum.Name}}Exhaustive{{$enum.MembersDigest}}(in {{$enum.Name}}) {{$enum.Name}} {
	return in
}
{{end}}

{{range .Module.Types.StringConsts}}
const {{.Key}} = "{{.Value}}";
{{end}}
`

const BackendRestEndpoints = `package {{.Module.Id}}

import (
	"encoding/json"
	"github.com/function61/gokit/net/http/httpauth"
	"net/http"
	"net/url"
)

type HttpHandlers interface { {{range .Module.Types.Endpoints}}
	{{UppercaseFirst .Name}}(rctx *httpauth.RequestContext, {{if .Consumes}}input {{.Consumes.AsGoType}}, {{end}}w http.ResponseWriter, r *http.Request){{if .Produces}} *{{.Produces.AsGoType}}{{end}}{{end}}
}

// the following generated code brings type safety from all the way to the
// backend-frontend path (input/output structs and endpoint URLs) to the REST API
func RegisterRoutes(handlers HttpHandlers, mwares httpauth.MiddlewareChainMap, register func(method string, path string, fn http.HandlerFunc)) { {{range .Module.Types.Endpoints}}
	register("{{.HttpMethod}}", "{{StripQueryFromUrl .Path}}", func(w http.ResponseWriter, r *http.Request) {
		rctx := mwares["{{.MiddlewareChain}}"](w, r)
		if rctx == nil {
			return // middleware aborted request handing and handled error response itself
		}
{{if .Consumes}}		input := &{{.Consumes.AsGoType}}{}
		if ok := parseJsonInput(w, r, input); !ok {
			return // parseJsonInput handled error message
		} {{end}}
{{if .Produces}}
		if out := handlers.{{UppercaseFirst .Name}}(rctx, {{if .Consumes}}*input, {{end}}w, r); out != nil { handleJsonOutput(w, out) } {{else}}
		handlers.{{UppercaseFirst .Name}}(rctx, {{if .Consumes}}*input, {{end}}w, r) {{end}}
	})
{{end}}
}

func handleJsonOutput(w http.ResponseWriter, output interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(output); err != nil {
		panic(err)
	}
}

func parseJsonInput(w http.ResponseWriter, r *http.Request, input interface{}) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "expecting Content-Type with application/json header", http.StatusBadRequest)
		return false
	}

	if err := decoder.Decode(input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}

	return true
}

type RestClientUrlBuilder struct {
	baseUrl string
}

func NewRestClientUrlBuilder(baseUrl string) *RestClientUrlBuilder {
	return &RestClientUrlBuilder{baseUrl}
}

{{range .Module.Types.Endpoints}}
// {{.Path}}
func (r *RestClientUrlBuilder) {{UppercaseFirst .Name}}({{.GoArgs}}) string {
	return r.baseUrl + "{{.GoPath}}"
}
{{end}}

// a hack so we don't have to conditionally import net/url module
// FIXME: path components should be escaped differently than query comps
func queryEscape(s string) string {
	return url.QueryEscape(s)
}

`
