package codegen

import (
	"errors"
	"fmt"
	"strings"
)

type CommandSpecFile []*CommandSpec

func (c *CommandSpecFile) Validate() error {
	for _, spec := range *c {
		if err := spec.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type CommandSpec struct {
	Command                string              `json:"command"`
	Title                  string              `json:"title"`
	CrudNature             string              `json:"crudNature"`
	AdditionalConfirmation string              `json:"additional_confirmation"`
	MiddlewareChain        string              `json:"chain"`
	CtorArgs               []string            `json:"ctor"`
	Fields                 []*CommandFieldSpec `json:"fields"`
}

func (c *CommandSpec) AsGoStructName() string {
	// "user.Create" => "userCreate"
	dotRemoved := strings.Replace(c.Command, ".", "", -1)

	// "userCreate" => "UserCreate"
	titleCased := strings.Title(dotRemoved)

	return titleCased
}

func (c *CommandSpec) Validate() error {
	for _, field := range c.Fields {
		if err := field.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type CommandFieldSpec struct {
	Key                string `json:"key"`
	Type               string `json:"type"`
	ValidationRegex    string `json:"validation_regex"`
	MaxLength          *int   `json:"max_length"`
	Optional           bool   `json:"optional"`
	HideIfDefaultValue bool   `json:"hideIfDefaultValue"`
	Help               string `json:"help"`
}

func (c *CommandFieldSpec) AsGoField() string {
	return fmt.Sprintf("%s %s `json:\"%s\"`", c.Key, c.AsGoType(), c.Key)
}

func (c *CommandFieldSpec) AsValidationSnippet() string {
	goType := c.AsGoType()

	if goType == "string" || goType == "password" {
		maxLen := 128

		if c.MaxLength != nil {
			maxLen = *c.MaxLength
		} else if c.Type == "multiline" {
			maxLen = 4 * 1024
		}

		emptySnippet := ""

		if !c.Optional {
			emptySnippet = fmt.Sprintf(
				`if x.%s == "" {
		return fieldEmptyValidationError("%s")
	}
	`,
				c.Key,
				c.Key)
		}

		lengthSnippet := fmt.Sprintf(
			`if len(x.%s) > %d {
		return fieldLengthValidationError("%s", %d)
	}
	`,
			c.Key,
			maxLen,
			c.Key,
			maxLen)

		regexSnippet := ""
		if c.ValidationRegex != "" {
			regexSnippet = fmt.Sprintf(
				`if err := regexpValidation("%s", '%s', x.%s); err != nil {
		return err
	}
	`,
				c.Key,
				escapeStringInsideJsSingleQuotes(c.ValidationRegex),
				c.Key)
		}

		return emptySnippet + lengthSnippet + regexSnippet
	} else if goType == "bool" || goType == "int" || goType == "guts.Date" {
		// presence check not possible for these types
		return ""
	} else {
		panic(errors.New("validation not supported for type: " + goType))
	}
}

func (c *CommandFieldSpec) AsGoType() string {
	switch c.Type {
	case "text":
		return "string"
	case "multiline":
		return "string"
	case "password":
		return "string"
	case "checkbox":
		return "bool"
	case "integer":
		return "int"
	case "date":
		return "guts.Date"
	default:
		return ""
	}
}

func (c *CommandFieldSpec) AsTsType() string {
	switch c.Type {
	case "text":
		return "string"
	case "multiline":
		return "string"
	case "password":
		return "string"
	case "checkbox":
		return "boolean"
	case "integer":
		return "number"
	case "date":
		return "datetimeRFC3339"
	default:
		return ""
	}
}

func (c *CommandFieldSpec) Validate() error {
	if c.Type == "" {
		c.Type = "text"
	}

	if c.AsGoType() == "" || c.AsTsType() == "" {
		return errors.New("field " + c.Key + " has invalid type: " + c.Type)
	}

	return nil
}

func (c *CommandSpec) MakeStruct() string {
	template := `type %s struct {
	%s
}`

	fieldLines := []string{}

	for _, field := range c.Fields {
		fieldLines = append(fieldLines, field.AsGoField())
	}

	return fmt.Sprintf(
		template,
		c.AsGoStructName(),
		strings.Join(fieldLines, "\n\t"))
}

// returns Go code (as a string) for validating command inputs
func (c *CommandSpec) MakeValidation() string {
	validationSnippets := []string{}

	for _, field := range c.Fields {
		validationSnippet := field.AsValidationSnippet()
		if validationSnippet == "" {
			continue
		}

		validationSnippets = append(validationSnippets, validationSnippet)
	}

	return strings.Join(validationSnippets, "\n\t")
}

func (c *CommandSpec) FieldsForTypeScript() string {
	fields := []string{}

	for _, fieldSpec := range c.Fields {
		fieldSerialized := ""

		fieldToTypescript := func(fieldSpec *CommandFieldSpec, tsKind string, defValKey string) string {
			defVal := "undefined" // .. in literal TypeScript code
			if tsKind == "Checkbox" {
				defVal = "false"
			}

			for _, ctorArg := range c.CtorArgs {
				if ctorArg == fieldSpec.Key {
					defVal = fieldSpec.Key
					break
				}
			}

			return fmt.Sprintf(
				`{ Key: '%s', Required: %v, HideIfDefaultValue: %v, Kind: CommandFieldKind.%s, %s: %s, Help: '%s', ValidationRegex: '%s' },`,
				fieldSpec.Key,
				!fieldSpec.Optional,
				fieldSpec.HideIfDefaultValue,
				tsKind,
				defValKey,
				defVal,
				escapeStringInsideJsSingleQuotes(fieldSpec.Help),
				fieldSpec.ValidationRegex)
		}

		switch fieldSpec.Type {
		case "text":
			fieldSerialized = fieldToTypescript(fieldSpec, "Text", "DefaultValueString")
		case "multiline":
			fieldSerialized = fieldToTypescript(fieldSpec, "Multiline", "DefaultValueString")
		case "password":
			fieldSerialized = fieldToTypescript(fieldSpec, "Password", "DefaultValueString")
		case "checkbox":
			fieldSerialized = fieldToTypescript(fieldSpec, "Checkbox", "DefaultValueBoolean")
		case "integer":
			fieldSerialized = fieldToTypescript(fieldSpec, "Integer", "DefaultValueNumber")
		case "date":
			fieldSerialized = fieldToTypescript(fieldSpec, "Date", "DefaultValueString")
		default:
			panic(fmt.Errorf("Unsupported field type for UI: %s", fieldSpec.Type))
		}

		fields = append(fields, fieldSerialized)
	}

	return strings.Join(fields, "\n\t\t\t")
}

func (c *CommandSpec) CtorArgsForTypeScript() string {
	ctorArgs := []string{}

	for _, ctorArg := range c.CtorArgs {
		spec := c.fieldSpecByKey(ctorArg)
		if spec == nil {
			panic("field for CtorArg not found")
		}

		ctorArgs = append(ctorArgs, ctorArg+": "+spec.AsTsType())
	}

	return strings.Join(ctorArgs, ", ")
}

func (c *CommandSpec) fieldSpecByKey(key string) *CommandFieldSpec {
	for _, field := range c.Fields {
		if field.Key == key {
			return field
		}
	}

	return nil
}

func escapeStringInsideJsSingleQuotes(in string) string {
	return strings.ReplaceAll(strings.ReplaceAll(in, `\`, `\\`), `'`, `\'`)
}
