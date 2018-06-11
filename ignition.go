package sabakan

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"

	ignition "github.com/coreos/ignition/config/v2_2"
	"github.com/vincent-petithory/dataurl"
	yaml "gopkg.in/yaml.v2"
)

// MaxIgnitions is a number of the ignitions to keep on etcd
const MaxIgnitions = 10

// ValidateIgnitionTemplate validates if the tmpl is a template for a valid ignition.
// The method returns nil if valid template is given, otherwise returns an error.
// The method returns template by tmpl nil value of Machine.
func ValidateIgnitionTemplate(tmpl string, ipam *IPAMConfig) error {
	mc := &Machine{
		Serial: "1234abcd",
		Rack:   1,
	}

	ipam.GenerateIP(mc)
	ign, err := RenderIgnition(tmpl, mc)

	if err != nil {
		return err
	}
	_, rpt, err := ignition.Parse([]byte(ign))
	if err != nil {
		return err
	}
	if len(rpt.Entries) > 0 {
		return errors.New(rpt.String())
	}
	return nil
}

// RenderIgnition returns the rendered ignition from the template and a machine
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

	var ign interface{}
	err = yaml.Unmarshal(buf.Bytes(), &ign)
	if err != nil {
		return "", err
	}

	ign = convert(ign)
	ignMap, ok := ign.(map[string]interface{})
	if !ok {
		return "", errors.New("invalid ignition, failed to convert")
	}

	if storageMap, ok := ignMap["storage"].(map[string]interface{}); ok {
		if files, ok := storageMap["files"].([]interface{}); ok {
			for _, elem := range files {
				if f, ok := elem.(map[string]interface{}); ok {
					if contents, ok := f["contents"].(map[string]interface{}); ok {
						if source, ok := contents["source"].(string); ok {
							contents["source"] = fmt.Sprintf("data:,%s", dataurl.EscapeString(source))
						}
					}
				}
			}
		}
	}

	dataOut, err := json.Marshal(&ign)
	if err != nil {
		return "", err
	}

	return string(dataOut), nil
}

func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}
