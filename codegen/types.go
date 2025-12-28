package codegen

type ProcessedStringEnumMember struct {
	Key     string
	GoKey   string
	GoValue string
}

type ProcessedStringEnum struct {
	Definition    EnumDef
	MembersDigest string
	Members       []ProcessedStringEnumMember
}

type EnumDef struct {
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	StringMembers []string `json:"stringMembers"`
	Description   string   `json:"description"`
}

type StringConstDef struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

type DomainFile struct {
	Events []*EventSpec `json:"events"`
}

type EventDefForTpl struct {
	EventKey        string
	CtorArgs        string
	CtorAssignments string
	GoStructName    string
}

type Imports struct {
	// native types
	Date     bool
	DateTime bool
	Binary   bool

	ModuleIds []string // other modules whose types this module's types have dependencies to
}

func NewImports() Imports {
	return Imports{
		ModuleIds: []string{},
	}
}

// this is passed as data to each template that we'll render
type TplData struct {
	Module                 *Module
	Opts                   Opts
	AnyEndpointHasConsumes bool
	TypesImports           Imports
	CommandsImports        Imports
	CommandsImportsUi      Imports // UI only needs types that are mentioned in ctor
	EventsImports          Imports
	StringEnums            []ProcessedStringEnum
	EventStructsAsGoCode   string
	EventDefs              []EventDefForTpl
}

type EventSpec struct {
	Event     string            `json:"event"`
	CtorArgs  []string          `json:"ctor"`
	Changelog []string          `json:"changelog"`
	Fields    []*EventFieldSpec `json:"fields"`
}

type EventFieldSpec struct {
	Key   string      `json:"key"`
	Type  DatatypeDef `json:"type"`
	Notes string      `json:"notes"`
}
