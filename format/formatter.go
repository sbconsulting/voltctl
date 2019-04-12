package format

import (
	"io"
	"reflect"
	"regexp"
	"strings"
	"text/tabwriter"
	"text/template"
	"text/template/parse"
)

var nameFinder = regexp.MustCompile(`\.([_A-Za-z0-9]*)}}`)

type Format string

func (f Format) IsTable() bool {
	return strings.HasPrefix(string(f), "table")
}

func (f Format) Execute(writer io.Writer, withHeaders bool, data interface{}) {
	var tabWriter *tabwriter.Writer = nil
	format := f

	if f.IsTable() {
		tabWriter = tabwriter.NewWriter(writer, 0, 4, 4, ' ', 0)
		format = Format(strings.TrimPrefix(string(f), "table"))
	}

	tmpl, err := template.New("output").Parse(string(format))
	if err != nil {
		panic(err)
	}

	if f.IsTable() && withHeaders {
		var header string
		for _, n := range tmpl.Tree.Root.Nodes {
			switch n.Type() {
			case parse.NodeText:
				header += n.String()
			case parse.NodeString:
				header += n.String()
			case parse.NodeAction:
				found := nameFinder.FindStringSubmatch(n.String())
				if len(found) == 2 {
					header += strings.ToUpper(found[1])
				}
			}
		}
		tabWriter.Write([]byte(header))
		tabWriter.Write([]byte("\n"))

		slice := reflect.ValueOf(data)
		if slice.Kind() == reflect.Slice {
			for i := 0; i < slice.Len(); i++ {
				tmpl.Execute(tabWriter, slice.Index(i).Interface())
				tabWriter.Write([]byte("\n"))
			}
		}
		tabWriter.Flush()
		return
	}

	slice := reflect.ValueOf(data)
	if slice.Kind() == reflect.Slice {
		for i := 0; i < slice.Len(); i++ {
			tmpl.Execute(writer, slice.Index(i).Interface())
			writer.Write([]byte("\n"))
		}
	}
}

func (f Format) ExtractHeader() {
}
