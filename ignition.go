package sabakan

import (
	"bytes"
	"text/template"
)

func isValidIgnitionTemplate(templateString string) (bool) {
	return false
	//t, err := template.New("ignition").Parse(templateString)
	//if err != nil {
	//	return false
	//}
	//buf := new(bytes.Buffer)
	//err = t.Execute(buf, m)
	//if err != nil {
	//	return false
	//}
	//ignitionString := buf.String()
	//return
}

func RenderIgnition(tmpl string, m *Machine) (string, error) {
	t, err := template.New("ignition").Parse(tmpl)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, m)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
