package sabakan

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"

	"github.com/coreos/container-linux-config-transpiler/config"
	ignition "github.com/coreos/ignition/config/v2_2"
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

	cfg, ast, rpt := config.Parse(buf.Bytes())

	if len(rpt.Entries) > 0 {
		return "", fmt.Errorf("failed to parse container linux config file %s", rpt.String())
	}

	ignCfg, rpt := config.Convert(cfg, "", ast)

	if len(rpt.Entries) > 0 {
		return "", fmt.Errorf("failed to convert to ignition file %s", rpt.String())
	}

	dataOut, err := json.Marshal(&ignCfg)
	if err != nil {
		return "", err
	}

	return string(dataOut), nil
}
