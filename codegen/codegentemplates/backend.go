package codegentemplates

const BackendCommandsDefinitions = `package {{.ModuleId}}

// WARNING: generated file

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"github.com/function61/eventkit/command"
{{if .CommandsImports.Date}}	"github.com/function61/eventkit/guts"{{end}}
)

type CommandHandlers interface { {{range .CommandSpecs}}
	{{.AsGoStructName}}(*{{.AsGoStructName}}, *command.Ctx) error{{end}}
}


// structs

{{range .CommandSpecs}}
{{.MakeStruct}}

func (x *{{.AsGoStructName}}) Validate() error {
	{{.MakeValidation}}

	return nil
}

func (x *{{.AsGoStructName}}) MiddlewareChain() string { return "{{.MiddlewareChain}}" }
func (x *{{.AsGoStructName}}) Key() string { return "{{.Command}}" }
func (x *{{.AsGoStructName}}) Invoke(ctx *command.Ctx, handlers interface{}) error {
	return handlers.(CommandHandlers).{{.AsGoStructName}}(x, ctx)
}
{{end}}

// builders

var Allocators = command.AllocatorMap{
{{range .CommandSpecs}}
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

func fieldLengthValidationError(fieldName string, maxLength int) error {
	return fmt.Errorf("field %s exceeded maximum length %d", fieldName, maxLength)
}
`

const BackendEventDefinitions = `package {{.ModuleId}}

import (
{{if .EventsImports.DateTime}}	"time"
{{end}}{{if .EventsImports.Date}}	"github.com/function61/eventkit/guts"
{{end}}	"github.com/function61/eventkit/event"
)

// WARNING: generated file

var Allocators = event.AllocatorMap{
{{range .EventDefs}}
	"{{.EventKey}}": func() event.Event { return &{{.GoStructName}}{meta: &event.EventMeta{}} },{{end}}
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
func (e *{{.GoStructName}}) Meta() *event.EventMeta { return e.meta }{{end}}

{{range .EventDefs}}
func (e *{{.GoStructName}}) MetaType() string { return "{{.EventKey}}" }{{end}}

{{range .EventDefs}}
func (e *{{.GoStructName}}) Serialize() string { return e.meta.Serialize(e) }{{end}}

// interface

type EventListener interface { {{range .EventDefs}}
	Apply{{.GoStructName}}(*{{.GoStructName}}) error{{end}}

	HandleUnknownEvent(event event.Event) error
}

func DispatchEvent(event event.Event, listener EventListener) error {
	switch e := event.(type) { {{range .EventDefs}}
	case *{{.GoStructName}}:
		return listener.Apply{{.GoStructName}}(e){{end}}
	default:
		return listener.HandleUnknownEvent(event)
	}
}
`

const BackendTypes = `package {{.ModuleId}}

import ( {{if .StringEnums}}
	"fmt"
	"encoding/json"{{end}}
{{if .TypesImports.Date}}	"github.com/function61/eventkit/guts"
{{end}}{{if .TypesImports.DateTime}}	"time"
{{end}}{{range .TypesImports.ModuleIds}}
	"{{$.Opts.BackendModulePrefix}}{{.}}"
{{end}}
)

{{range .ApplicationTypes.Types}}
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

{{range .ApplicationTypes.StringConsts}}
const {{.Key}} = "{{.Value}}";
{{end}}
`

const BackendRestEndpoints = `package {{.ModuleId}}

import (
	"encoding/json"
	"github.com/function61/gokit/httpauth"
	"net/http"
	"net/url"
)

type HttpHandlers interface { {{range .ApplicationTypes.Endpoints}}
	{{UppercaseFirst .Name}}(rctx *httpauth.RequestContext, {{if .Consumes}}input {{.Consumes.AsGoType}}, {{end}}w http.ResponseWriter, r *http.Request){{if .Produces}} *{{.Produces.AsGoType}}{{end}}{{end}}
}

// the following generated code brings type safety from all the way to the
// backend-frontend path (input/output structs and endpoint URLs) to the REST API
func RegisterRoutes(handlers HttpHandlers, mwares httpauth.MiddlewareChainMap, register func(method string, path string, fn http.HandlerFunc)) { {{range .ApplicationTypes.Endpoints}}
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

{{range .ApplicationTypes.Endpoints}}
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
