package report

import (
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"
)

const (
	markdownTmplText = `
# {{.Title}}

## Updated {{date .Range.Start .Config.DateLayout}}
{{- range $ek, $e := .Envs}}
### {{$e.Name}}
	{{- range $ek, $j := $e.Jobs}}
- [{{$j.Name}}](https://testgrid.k8s.io/redhat-osd-{{$e.Name}}#{{$j.Name}})
		{{- range $rn, $r := $j.Runs}}
   * [#{{$r.BuildNum}}](https://prow.k8s.io/view/gcs/{{$j.Prefix}}/{{$r.BuildNum}})
			{{- range $k, $v := $r.Finished.Metadata}}
				{{- if eq $k "cluster-id"}}
      + Cluster ID: {{$v}}
				{{- end}}
			{{- end}}
      + Hive logs: {{hiveLogs $r}}
      + **Failures**:
			{{- range $fn, $f := $r.Failures}}
         - Test Name: {{$f.Name}}
{{failureTxt $f | indent 11}}
			{{- end}}
		{{- end}}
	{{- end}}
{{- end}}
`
)

var (
	reportTmpl = template.Must(template.New("report").
		Funcs(template.FuncMap{
			"indent":     indent,
			"date":       printDate,
			"hiveLogs":   hiveLogs,
			"failureTxt": failureTxt,
		}).Parse(markdownTmplText))
)

func (r *Report) Markdown(w io.Writer) error {
	err := reportTmpl.Execute(w, r)
	if err != nil {
		return fmt.Errorf("couldn't render report: %v", err)
	}
	return nil
}

func hiveLogs(run Run) string {
	if run.HiveLogURL != "" {
		return run.HiveLogURL
	}
	return "Could not be found!!!"
}

func printDate(t time.Time, layout string) string {
	return t.Format(layout)
}

func failureTxt(f Failure) string {
	failTxt := f.Message(0)
	return fmt.Sprintf("```%s\n```", failTxt)
}

func indent(spaces int, s string) string {
	padding := strings.Repeat(" ", spaces)
	return padding + strings.Replace(s, "\n", "\n"+padding, -1)
}
