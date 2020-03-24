package codegentemplates

const BackendUiRoutes = `package {{.Module.Id}}

import (
	"github.com/gorilla/mux"
	"net/http"
)

func RegisterUiRoutes(routes *mux.Router, uiHandler http.HandlerFunc) { {{range .Module.UiRoutes}}
	routes.HandleFunc("{{.PathWithoutQuery}}", uiHandler){{end}}
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
{{if .HasOpts}}export interface {{.TsOptsName}} { {{range .PathAndQueryOpts}}
	{{.}}: string;{{end}}
}{{end}}

// {{.Path}}
export function {{.Id}}Url({{if .HasOpts}}opts: {{.TsOptsName}}{{end}}): string {
	const query: queryParams = {}
{{range .QueryPlaceholders}}
	if (opts.{{.}}) {
		query.{{.}} = opts.{{.}};
	} {{end}}
 
	return makeQueryParams(` + "`{{.TsPath}}`" + `, query);
}

// @ts-ignore
export function {{.Id}}Match(path: string, query: queryParams): {{if .HasOpts}}{{.TsOptsName}}{{else}}{}{{end}} | null {
	const matches = {{.PathReJavaScript}}.exec(path);
	if (matches == null) {
		return null;
	}
{{range .QueryPlaceholders}}
	if (!query.{{.}}) {
		return null;
	}{{end}}

	return { {{range $idx, $key := .PathPlaceholders}}
		{{$key}}: matches[{{add $idx 1}}],{{end}}{{range .QueryPlaceholders}}
		{{.}}: query.{{.}},{{end}}
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

`
