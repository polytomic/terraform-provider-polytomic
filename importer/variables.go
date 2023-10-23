package importer

import (
	"bytes"
	"os"
	"text/template"
)

var varTmpl = `variable "{{ .Name }}" {
	type = {{ .Type }}
	{{- if .Default }}
	default = {{ .Default }}
	{{- end }}
	sensitive = {{ .Sensitive }}
}
`

var varTypeMap = map[string]string{
	"basetypes.StringType": "string",
	"basetypes.NumberType": "number",
	"basetypes.BoolType":   "bool",
}

type Variable struct {
	Name      string
	Default   any
	Type      string
	Sensitive bool
}

func generateVariables(f *os.File, vars []Variable) error {
	for _, v := range vars {
		tmpl, err := template.New("variable").Parse(varTmpl)
		if err != nil {
			return err
		}
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, v)
		if err != nil {
			return err
		}
		_, err = f.Write(buf.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}
