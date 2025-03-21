package front

import (
	_ "embed"
	"github.com/scylladb/scylla-operator/pkg/analyze/symptoms"
	"io"
	"text/template"
)

//go:embed front.tmpl
var tmpl_string string
var tmpl = template.Must(template.New("all").
	Funcs(template.FuncMap{
		"HasValidResources": hasValidResources,
	}).
	Parse(tmpl_string))

func hasValidResources(resources map[string]any) bool {
	for _, r := range resources {
		if r != nil {
			return true
		}
	}
	return false
}

func Print(writer io.Writer, issue symptoms.Issue) error {

	err := tmpl.ExecuteTemplate(writer, "issue", issue)

	return err
}
