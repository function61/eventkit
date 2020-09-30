package codegen

import (
	"sort"
	"testing"

	"github.com/function61/gokit/assert"
)

func TestBeginsWithUppercaseLetter(t *testing.T) {
	mkDatatype := func(name string) *DatatypeDef {
		return &DatatypeDef{NameRaw: name}
	}

	assert.Assert(t, mkDatatype("Foo").isCustomType())
	assert.Assert(t, !mkDatatype("foo").isCustomType())

	assert.Assert(t, !mkDatatype("!perkele").isCustomType())
}

func TestFlattenDatatype(t *testing.T) {
	person := &DatatypeDef{
		NameRaw: "object",
		Fields: map[string]*DatatypeDef{
			"Name": &DatatypeDef{NameRaw: "string"},
			"Age":  &DatatypeDef{NameRaw: "boolean"},
		},
	}

	flattened := flattenDatatype(person)
	sort.Slice(flattened, func(i, j int) bool { return flattened[i].NameRaw < flattened[j].NameRaw })
	assert.Assert(t, len(flattened) == 3)
	assert.EqualString(t, flattened[0].NameRaw, "boolean")
	assert.EqualString(t, flattened[1].NameRaw, "object")
	assert.EqualString(t, flattened[2].NameRaw, "string")
}
