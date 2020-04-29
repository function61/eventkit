package codegentemplates

const DocsEvents = `Each and every *Event* additionally has the following common meta data:

- Timestamp, in UTC, the event was raised on
- ID of the user that caused that event


{{range .Module.Events.Events}}
{{.Event}}
-------

{{if .Changelog}}
Changelog:
{{range .Changelog}}
- {{.}}{{end}}
{{end}}

| key | type | notes |
|-----|------|-------|
{{range .Fields}}| {{.Key}} | {{.Type.Name}} | {{.Notes}} |
{{end}}
{{end}}
`

const DocsTypes = `{{if .Module.Types.StringConsts}}
Constants
---------

| const | value |
|-------|-------|
{{range .Module.Types.StringConsts}}| {{.Key}} | {{.Value}} |
{{end}}
{{end}}

{{range .Module.Types.Enums}}
enum {{.Name}}
---------

{{range .StringMembers}}
- {{.}}{{end}}
{{end}}

{{range .Module.Types.Types}}
{{.Name}}
---------

` + "```" + `
{{.AsTypeScriptCode}}
` + "```" + `
{{end}}
`

const DocsCommands = `Overview
--------

| Endpoint | Middleware | Title |
|----------|------------|-------| {{range .Module.Commands}}
| POST /command/{{.Command}} | {{.MiddlewareChain}} | {{.Title}} | {{end}}

{{range .Module.Commands}}
{{.Command}}
------------

| Field | Type | Required | Notes |
|-------|------|----------|-------|
{{range .Fields}}| {{.Key}} | {{.Type}} | {{not .Optional}} | {{.Help}} |
{{end}}
{{end}}
`

const DocsRestEndpoints = `Overview
--------

| Path | Middleware | Input | Output | Notes |
|------|------------|-------|--------|-------|
{{range .Module.Types.Endpoints}}| {{.HttpMethod}} {{.Path}} | {{.MiddlewareChain}} | {{if .Consumes}}{{.Consumes.AsTypeScriptType}}{{end}} | {{if .Produces}}{{.Produces.AsTypeScriptType}}{{end}} | {{.Description}} |
{{end}}

{{range .Module.Types.Endpoints}}
{{.HttpMethod}} {{.Path}}
-------------------------

| Detail           |                                                       |
|------------------|-------------------------------------------------------|
| Middleware chain | {{.MiddlewareChain}}                                  |
| Consumes         | {{if .Consumes}}{{.Consumes.AsTypeScriptType}}{{end}} |
| Produces         | {{if .Produces}}{{.Produces.AsTypeScriptType}}{{end}} |
| Description      | {{.Description}}                                      |

{{end}}
`
