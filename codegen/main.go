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

	// computed state
	EventsSpec *DomainFile
	Types      *ApplicationTypesDefinition
	Commands   *CommandSpecFile
}

var moduleIdFromModulePathRe = regexp.MustCompile("[^/]+$")

func NewModule(
	modulePath string,
	typesFile string,
	eventsSpecFile string,
	commandSpecFile string,
) *Module {
	// "vstoserver/vstotypes" => "vstotypes"
	id := moduleIdFromModulePathRe.FindStringSubmatch(modulePath)[0]

	return &Module{
		Id:               id,
		Path:             modulePath,
		EventsSpecFile:   eventsSpecFile,
		CommandsSpecFile: commandSpecFile,
		TypesFile:        typesFile,
	}
}

type FileToGenerate struct {
	targetPath     string
	obtainTemplate func() (string, error)
}

func processModule(mod *Module, opts Opts) error {
	// should be ok with nil data
	mod.EventsSpec = &DomainFile{}
	mod.Types = &ApplicationTypesDefinition{}
	mod.Commands = &CommandSpecFile{}

	hasTypes := mod.TypesFile != ""
	hasEvents := mod.EventsSpecFile != ""
	hasCommands := mod.CommandsSpecFile != ""

	if hasEvents {
		if err := jsonfile.Read(mod.EventsSpecFile, mod.EventsSpec, true); err != nil {
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
		if err := mod.Commands.Validate(); err != nil {
			return err
		}
	}

	hasRestEndpoints := len(mod.Types.Endpoints) > 0

	// preprocessing
	eventDefs, eventStructsAsGoCode := ProcessEvents(mod.EventsSpec)

	uniqueTypes := mod.Types.UniqueDatatypesFlattened()

	typesImports := NewImports()
	typesImports.ModuleIds = uniqueModuleIdsFromDatatypes(uniqueTypes)

	commandsImports := NewImports()
	commandsImportsUi := NewImports()

	eventsImports := NewImports()

	for _, eventDef := range mod.EventsSpec.Events {
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
		ModuleId:               mod.Id,
		ModulePath:             mod.Path,
		Opts:                   opts,
		AnyEndpointHasConsumes: anyEndpointHasConsumes,
		TypesImports:           typesImports,
		CommandsImports:        commandsImports,
		CommandsImportsUi:      commandsImportsUi,
		EventsImports:          eventsImports,
		DomainSpecs:            mod.EventsSpec, // backwards compat
		CommandSpecs:           mod.Commands,   // backwards compat
		ApplicationTypes:       mod.Types,      // backwards compat
		StringEnums:            ProcessStringEnums(mod.Types.Enums),
		EventDefs:              eventDefs,
		EventStructsAsGoCode:   eventStructsAsGoCode,
	}

	maybeRenderOne := func(expr bool, path string, template string) error {
		if !expr {
			return nil
		}

		return ProcessFile(Inline(path, template), data)
	}

	docs := opts.AutogenerateModuleDocs

	return allOk(
		maybeRenderOne(hasCommands, backendPath("commanddefinitions.gen.go"), codegentemplates.BackendCommandsDefinitions),
		maybeRenderOne(hasCommands, frontendPath("commands.ts"), codegentemplates.FrontendCommandDefinitions),
		maybeRenderOne(hasCommands && docs, docPath("commands.md"), codegentemplates.DocsCommands),
		maybeRenderOne(hasEvents, backendPath("events.gen.go"), codegentemplates.BackendEventDefinitions),
		maybeRenderOne(hasEvents && docs, docPath("events.md"), codegentemplates.DocsEvents),
		maybeRenderOne(hasRestEndpoints, frontendPath("endpoints.ts"), codegentemplates.FrontendRestEndpoints),
		maybeRenderOne(hasRestEndpoints, backendPath("restendpoints.gen.go"), codegentemplates.BackendRestEndpoints),
		maybeRenderOne(hasRestEndpoints && docs, docPath("rest_endpoints.md"), codegentemplates.DocsRestEndpoints),
		maybeRenderOne(hasTypes, backendPath("types.gen.go"), codegentemplates.BackendTypes),
		maybeRenderOne(hasTypes, frontendPath("types.ts"), codegentemplates.FrontendDatatypes),
		maybeRenderOne(hasTypes && docs, docPath("types.md"), codegentemplates.DocsTypes),
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
