package codegentemplates

const BackendUiRoutes = `package {{.Module.Id}}

import (
	"github.com/gorilla/mux"
	"net/http"
)

func RegisterUiRoutes(routes *mux.Router, uiHandler http.HandlerFunc) { {{range .Module.UiRoutes}}
	routes.HandleFunc("{{.Path}}", uiHandler){{end}}
}
`

const FrontendUiRoutes = `// tslint:disable
// WARNING: generated file

import { parseQueryParams, queryParams, makeQueryParams } from 'f61ui/httputil';

export interface RouteHandlers { {{range .Module.UiRoutes}}
	{{.Id}}: ({{if .HasOpts}}opts: {{.TsOptsName}}{{end}}) => JSX.Element;{{end}}
	notFound: (url: string) => JSX.Element;
}

{{range .Module.UiRoutes}}
{{if .HasOpts}}export interface {{.TsOptsName}} { {{range .PathPlaceholders}}
	{{.}}: string;{{end}}{{range .QueryParams}}
	{{.Key}}{{if .Type.Nullable}}?{{end}}: {{if eq .Type.NameRaw "integer"}}number{{else}}string{{end}};{{end}}
}{{end}}

{{if .Title}}
export const {{.Id}}Title = '{{.Title}}';
{{end}}

// {{.Path}}
export function {{.Id}}URL({{if .HasOpts}}opts: {{.TsOptsName}}{{end}}): string {
	const query: queryParams = {};
{{range .QueryParams}}
{{if .Type.Nullable}}	if (opts.{{.Key}} !== undefined) {
	{{end}}	query.{{.Key}} = opts.{{.Key}}{{if eq .Type.NameRaw "integer"}}.toString(){{end}};{{if .Type.Nullable}}
	}{{end}}
{{end}}
 
	return makeQueryParams(` + "`{{.TsPath}}`" + `, query);
}

export function {{.Id}}Match(path: string, query: queryParams): {{if .HasOpts}}{{.TsOptsName}}{{else}}{}{{end}} | null {
	const matches = {{.PathReJavaScript}}.exec(path);
	if (matches == null) {
		return null;
	}

{{range .QueryParams}}{{if eq .Type.NameRaw "string"}}
	const {{.Key}}Par = query.{{.Key}};{{else}}
	let {{.Key}}Par: number | undefined;
	if (query.{{.Key}} !== undefined) {
		// parseInt() accepts garbage after the number, and "+" accepts empty string
		if (query.{{.Key}} === '') {
			throw new Error("Invalid URL param: '{{.Key}}'; expecting integer, got empty string")
		}
		const parsed = +query.{{.Key}};
		if (isNaN(parsed)) {
			throw new Error("Invalid URL param: '{{.Key}}'; expecting integer")
		}
		{{.Key}}Par = parsed;
	}{{end}}{{if not .Type.Nullable}}
	if ({{.Key}}Par === undefined) {
		throw new Error("Required URL param '{{.Key}}' missing");
	} {{end}}
{{end}}

	assertNoUnrecognizedKeys(Object.keys(query), [{{range .QueryParams}}'{{.Key}}', {{end}}]);

	return { {{range $idx, $key := .PathPlaceholders}}
		{{$key}}: matches[{{add $idx 1}}],{{end}}{{range .QueryParams}}
		{{.Key}}: {{.Key}}Par,{{end}}
	};
}

// ---------------
{{end}}

export function handle(url: string, handlers: RouteHandlers): JSX.Element {
	const qpos = url.indexOf('?');

	// "/search?query=foo" => "/search"
	const path = qpos === -1 ? url : url.substr(0, qpos);
	const queryPars = parseQueryParams(qpos === -1 ? '' : url.substr(qpos + 1));
{{range .Module.UiRoutes}}
	const {{.Id}}Opts = {{.Id}}Match(path, queryPars);
	if ({{.Id}}Opts) {
		return handlers.{{.Id}}({{if .HasOpts}}{{.Id}}Opts{{end}});
	}
{{end}}

	return handlers.notFound(path);
}

// for when you need to check if url can be routed to this route collection
export function hasRouteFor(url: string): boolean {
	const qpos = url.indexOf('?');

	// "/search?query=foo" => "/search"
	const path = qpos === -1 ? url : url.substr(0, qpos);
	const queryPars = parseQueryParams(qpos === -1 ? '' : url.substr(qpos + 1));
{{range .Module.UiRoutes}}
	if ({{.Id}}Match(path, queryPars)) {
		return true
	} {{end}}

	return false;
}

function assertNoUnrecognizedKeys(gotKeys: string[], allowedKeys: string[]) {
	const unrecognizedKeys = gotKeys.filter(key => allowedKeys.indexOf(key) === -1);
	if (unrecognizedKeys.length > 0) {
		throw new Error("Unrecognized keys in URL params: "+unrecognizedKeys.join(', '));
	}
}

`
