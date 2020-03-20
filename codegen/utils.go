package codegen

import (
	"fmt"
	"github.com/function61/gokit/atomicfilewrite"
	"io"
	"net/url"
	"strings"
	"text/template"
)

func WriteTemplateFile(filename string, data interface{}, templateString string) error {
	templateFuncs := template.FuncMap{
		"StripQueryFromUrl":      stripQueryFromUrl,
		"UppercaseFirst":         func(input string) string { return strings.ToUpper(input[0:1]) + input[1:] },
		"EscapeForJsSingleQuote": func(input string) string { return strings.ReplaceAll(input, `'`, `\'`) },
	}

	tpl, err := template.New("").Funcs(templateFuncs).Parse(templateString)
	if err != nil {
		return fmt.Errorf("WriteTemplateFile Parse %s: %v", filename, err)
	}

	return atomicfilewrite.Write(filename, func(file io.Writer) error {
		if err := tpl.Execute(file, data); err != nil {
			return fmt.Errorf("WriteTemplateFile %s: %w", filename, err)
		}

		return nil
	})
}

// "/search?q={stuff}" => "/search"
func stripQueryFromUrl(input string) string {
	u, err := url.Parse(input)
	if err != nil {
		panic(err)
	}

	return u.Path
}

func allOk(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}
