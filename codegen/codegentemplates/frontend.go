package codegentemplates

const FrontendDatatypes = `// tslint:disable
// WARNING: generated file

{{range .TypesImports.ModuleIds}}
import * as {{.}} from '{{$.Opts.FrontendModulePrefix}}{{.}}_types';{{end}}
{{if .TypesImports.Date}}import {dateRFC3339} from 'f61ui/types';
{{end}}
{{if .TypesImports.DateTime}}import {datetimeRFC3339} from 'f61ui/types';
{{end}}
{{if .TypesImports.Binary}}import {binaryBase64} from 'f61ui/types';
{{end}}

{{range .StringEnums}}
export enum {{.Name}} {
{{range .Members}}
	{{.Key}} = '{{.GoValue}}',{{end}}
}
{{end}}
{{range .Module.Types.StringConsts}}
export const {{.Key}} = '{{EscapeForJsSingleQuote .Value}}';{{end}}
{{range .Module.Types.Types}}
{{.AsTypeScriptCode}}
{{end}}
`

const FrontendRestEndpoints = `// tslint:disable
// WARNING: generated file

// WHY: wouldn't make sense complicating code generation to check
// if we need template string or not in path string

import { {{range .Module.Types.EndpointsProducesAndConsumesTypescriptTypes}}
	{{.}},{{end}}
} from '{{$.Opts.FrontendModulePrefix}}{{.Module.Path}}_types';
import {
	getJson,
{{if .AnyEndpointHasConsumes}}	postJson,{{end}}
} from 'f61ui/httputil';

{{range .Module.Types.Endpoints}}
// {{.Path}}
export function {{.Name}}({{.TypescriptArgs}}) {
	return {{if .Consumes}}postJson<{{if .Consumes}}{{.Consumes.AsTypeScriptType}}{{else}}void{{end}}, {{if .Produces}}{{.Produces.AsTypeScriptType}}{{else}}void{{end}}>{{else}}getJson<{{if .Produces}}{{.Produces.AsTypeScriptType}}{{else}}void{{end}}>{{end}}(` + "`{{.TypescriptPath}}`" + `{{if .Consumes}}, body{{end}});
}
{{if not .Consumes}}
export function {{.Name}}Url({{.TypescriptArgs}}): string {
	return ` + "`{{.TypescriptPath}}`" + `;
}{{end}}
{{end}}
`

const FrontendCommandDefinitions = `// tslint:disable
// WARNING: generated file

{{if .Module.Commands.ImportedCustomFieldTypes}}import { {{range .Module.Commands.ImportedCustomFieldTypes}}
	{{.}},{{end}}
} from '{{$.Opts.FrontendModulePrefix}}{{.Module.Path}}_types';{{end}}
import {CommandDefinition, CommandFieldKind, CommandSettings, CrudNature} from 'f61ui/commandtypes';
{{if .CommandsImportsUi.Date}}import {dateRFC3339} from 'f61ui/types';
{{end}}
{{if .CommandsImportsUi.DateTime}}import {datetimeRFC3339} from 'f61ui/types';
{{end}}

{{range .Module.Commands}}
export function {{.AsGoStructName}}({{if .CtorArgsForTypeScript}}{{.CtorArgsForTypeScript}}, {{end}}settings: CommandSettings = {}): CommandDefinition {
	return {
		key: '{{.Command}}',{{if .AdditionalConfirmation}}
		additional_confirmation: '{{EscapeForJsSingleQuote .AdditionalConfirmation}}',
{{end}}		title: '{{EscapeForJsSingleQuote .Title}}',
		crudNature: CrudNature.{{.CrudNature}},
		info: {{if .Info}}[{{range .Info}}'{{EscapeForJsSingleQuote .}}',{{end}}]{{else}}[]{{end}},
		fields: [
{{.FieldsForTypeScript}}
		],
		settings: settings,
	};
}
{{end}}
`

const FrontendVersion = `// tslint:disable
// WARNING: generated file

export const version = '{{.Version}}';

export const isDevVersion = {{if eq .Version "dev"}}true{{else}}false{{end}};
`
