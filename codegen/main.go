package codegen

import (
	"github.com/function61/eventkit/codegen/codegentemplates"
	"github.com/function61/gokit/jsonfile"
	"github.com/function61/gokit/sliceutil"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type Module struct {
	// config
	Id   string // "vstotypes"
	Path string // "vstoserver/vstotypes"

	// input files
	EventsSpecFile   string
	CommandsSpecFile string
	TypesFile        string
	UiRoutesFile     string

	// computed state
	Events   *DomainFile
	Types    *ApplicationTypesDefinition
	Commands *CommandSpecFile
	UiRoutes []uiRouteSpec
}

func (m *Module) HasEnum(name string) bool {
	for _, enumDef := range m.Types.Enums {
		if enumDef.Name == name {
			return true
		}
	}

	return false
}

var moduleIdFromModulePathRe = regexp.MustCompile("[^/]+$")

func NewModule(
	modulePath string,
	typesFile string,
	eventsSpecFile string,
	commandSpecFile string,
	uiRoutesFile string,
) *Module {
	// "vstoserver/vstotypes" => "vstotypes"
	id := moduleIdFromModulePathRe.FindStringSubmatch(modulePath)[0]

	return &Module{
		Id:               id,
		Path:             modulePath,
		EventsSpecFile:   eventsSpecFile,
		CommandsSpecFile: commandSpecFile,
		TypesFile:        typesFile,
		UiRoutesFile:     uiRoutesFile,
	}
}

type FileToGenerate struct {
	targetPath     string
	obtainTemplate func() (string, error)
}

func processModule(mod *Module, opts Opts) error {
	// should be ok with nil data
	mod.Events = &DomainFile{}
	mod.Types = &ApplicationTypesDefinition{}
	mod.Commands = &CommandSpecFile{}

	hasTypes := mod.TypesFile != ""
	hasEvents := mod.EventsSpecFile != ""
	hasCommands := mod.CommandsSpecFile != ""
	hasUiRoutes := mod.UiRoutesFile != ""

	if hasEvents {
		if err := jsonfile.Read(mod.EventsSpecFile, mod.Events, true); err != nil {
			return err
		}
	}

	if hasTypes {
		if err := jsonfile.Read(mod.TypesFile, mod.Types, true); err != nil {
			return err
		}
		if err := mod.Types.Validate(); err != nil {
			return err
		}
	}

	if hasCommands {
		if err := jsonfile.Read(mod.CommandsSpecFile, mod.Commands, true); err != nil {
			return err
		}
		if err := mod.Commands.Validate(mod); err != nil {
			return err
		}
	}

	if hasUiRoutes {
		if err := jsonfile.Read(mod.UiRoutesFile, &mod.UiRoutes, true); err != nil {
			return err
		}
	}

	hasRestEndpoints := len(mod.Types.Endpoints) > 0

	// preprocessing
	eventDefs, eventStructsAsGoCode := ProcessEvents(mod.Events)

	uniqueTypes := mod.Types.UniqueDatatypesFlattened()

	typesImports := NewImports()
	typesImports.ModuleIds = uniqueModuleIdsFromDatatypes(uniqueTypes)

	commandsImports := NewImports()
	commandsImportsUi := NewImports()

	eventsImports := NewImports()

	for _, eventDef := range mod.Events.Events {
		for _, field := range eventDef.Fields {
			for _, datatype := range flattenDatatype(&field.Type) {
				switch datatype.NameRaw {
				case "date":
					eventsImports.Date = true
				case "datetime":
					eventsImports.DateTime = true
				}
			}
		}
	}

	for _, datatype := range uniqueTypes {
		switch datatype.NameRaw {
		case "date":
			typesImports.Date = true
		case "datetime":
			typesImports.DateTime = true
		case "binary":
			typesImports.Binary = true
		}
	}

	anyEndpointHasConsumes := false
	for _, endpoint := range mod.Types.Endpoints {
		if endpoint.Consumes != nil {
			anyEndpointHasConsumes = true
			break
		}
	}

	for _, command := range *mod.Commands {
		for _, field := range command.Fields {
			if field.Type == "date" {
				commandsImports.Date = true

				if sliceutil.ContainsString(command.CtorArgs, field.Key) {
					commandsImportsUi.Date = true
				}
			}
		}
	}

	backendPath := func(file string) string {
		return "pkg/" + mod.Path + "/" + file
	}

	frontendPath := func(file string) string {
		return "frontend/generated/" + mod.Path + "_" + file
	}

	docPath := func(file string) string {
		return "docs/" + mod.Path + "/" + file
	}

	data := &TplData{
		Module:                 mod,
		Opts:                   opts,
		AnyEndpointHasConsumes: anyEndpointHasConsumes,
		TypesImports:           typesImports,
		CommandsImports:        commandsImports,
		CommandsImportsUi:      commandsImportsUi,
		EventsImports:          eventsImports,
		StringEnums:            ProcessStringEnums(mod.Types.Enums),
		EventDefs:              eventDefs,
		EventStructsAsGoCode:   eventStructsAsGoCode,
	}

	renderOneIf := func(expr bool, path string, template string) error {
		if !expr {
			return nil
		}

		return ProcessFile(Inline(path, template), data)
	}

	docs := opts.AutogenerateModuleDocs

	return allOk(
		renderOneIf(hasCommands, backendPath("commanddefinitions.gen.go"), codegentemplates.BackendCommandsDefinitions),
		renderOneIf(hasCommands, frontendPath("commands.ts"), codegentemplates.FrontendCommandDefinitions),
		renderOneIf(hasCommands && docs, docPath("commands.md"), codegentemplates.DocsCommands),
		renderOneIf(hasEvents, backendPath("events.gen.go"), codegentemplates.BackendEventDefinitions),
		renderOneIf(hasEvents && docs, docPath("events.md"), codegentemplates.DocsEvents),
		renderOneIf(hasRestEndpoints, frontendPath("endpoints.ts"), codegentemplates.FrontendRestEndpoints),
		renderOneIf(hasRestEndpoints, backendPath("restendpoints.gen.go"), codegentemplates.BackendRestEndpoints),
		renderOneIf(hasRestEndpoints && docs, docPath("rest_endpoints.md"), codegentemplates.DocsRestEndpoints),
		renderOneIf(hasTypes, backendPath("types.gen.go"), codegentemplates.BackendTypes),
		renderOneIf(hasTypes, frontendPath("types.ts"), codegentemplates.FrontendDatatypes),
		renderOneIf(hasTypes && docs, docPath("types.md"), codegentemplates.DocsTypes),
		renderOneIf(hasUiRoutes, backendPath("ui-routes.gen.go"), codegentemplates.BackendUiRoutes),
		renderOneIf(hasUiRoutes, frontendPath("uiroutes.ts"), codegentemplates.FrontendUiRoutes),
	)
}

type Opts struct {
	BackendModulePrefix    string // "github.com/myorg/myproject/pkg/"
	FrontendModulePrefix   string // "generated/"
	AutogenerateModuleDocs bool
}

func ProcessModules(modules []*Module, opts Opts) error {
	for _, mod := range modules {
		if err := processModule(mod, opts); err != nil {
			return err
		}
	}

	return nil
}

// companion file means that for each of these files their corresponding .template file
// exists and will be rendered which will end up as the filename given
func CompanionFile(targetPath string) FileToGenerate {
	return FileToGenerate{
		targetPath: targetPath,
		obtainTemplate: func() (string, error) {
			templateContent, readErr := ioutil.ReadFile(targetPath + ".template")
			if readErr != nil {
				return "", readErr
			}

			return string(templateContent), nil
		},
	}
}

func Inline(targetPath string, inline string) FileToGenerate {
	return FileToGenerate{
		targetPath: targetPath,
		obtainTemplate: func() (string, error) {
			return inline, nil
		},
	}
}

func ProcessFile(target FileToGenerate, data interface{}) error {
	templateContent, err := target.obtainTemplate()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(target.targetPath), 0755); err != nil {
		return err
	}

	return WriteTemplateFile(target.targetPath, data, templateContent)
}
