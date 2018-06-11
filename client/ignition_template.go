package client

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

const baseFileDir = "files"
const baseSystemdDir = "systemd"
const baseNetworkdDir = "networkd"

type systemd struct {
	Enabled bool   `yaml:"enabled"`
	Source  string `yaml:"source"`
}

type ignitionSource struct {
	Passwd   string    `yaml:"passwd"`
	Files    []string  `yaml:"files"`
	Systemd  []systemd `yaml:"systemd"`
	Networkd []string  `yaml:"networkd"`
	Include  string    `yaml:"include"`
}

type ignitionBuilder struct {
	baseDir  string
	ignition map[string]interface{}
}

func generateIgnitionYAML(fname string) (io.Reader, error) {
	absPath, err := filepath.Abs(fname)
	if err != nil {
		return nil, ErrorStatus(err)
	}
	baseDir := filepath.Dir(absPath)
	builder := ignitionBuilder{baseDir: baseDir, ignition: make(map[string]interface{})}

	source, err := loadSource(absPath)
	if err != nil {
		return nil, ErrorStatus(err)
	}

	builder.ignition["ignition"] = map[string]interface{}{
		"version": "2.2.0",
	}
	err = builder.constructIgnitionYAML(source)
	if err != nil {
		return nil, ErrorStatus(err)
	}

	b, err := yaml.Marshal(builder.ignition)
	if err != nil {
		return nil, ErrorStatus(err)
	}

	return bytes.NewReader(b), nil
}

func loadSource(fname string) (*ignitionSource, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var source ignitionSource
	err = yaml.Unmarshal(data, &source)
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (b *ignitionBuilder) constructIgnitionYAML(source *ignitionSource) error {
	if source.Include != "" {
		include, err := loadSource(filepath.Join(b.baseDir, source.Include))
		if err != nil {
			return err
		}
		err = b.constructIgnitionYAML(include)
		if err != nil {
			return err
		}
	}
	if source.Passwd != "" {
		err := b.constructPasswd(source.Passwd)
		if err != nil {
			return err
		}
	}

	for _, file := range source.Files {
		err := b.constructFile(file)
		if err != nil {
			return err
		}
	}

	for _, s := range source.Systemd {
		err := b.constructSystemd(s)
		if err != nil {
			return err
		}
	}

	for _, n := range source.Networkd {
		err := b.constructNetworkd(n)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *ignitionBuilder) constructPasswd(passwd string) error {
	pf, err := os.Open(filepath.Join(b.baseDir, passwd))
	if err != nil {
		return err
	}
	defer pf.Close()
	passData, err := ioutil.ReadAll(pf)
	if err != nil {
		return err
	}

	var p interface{}
	err = yaml.Unmarshal(passData, &p)
	if err != nil {
		return err
	}
	b.ignition["passwd"] = p

	return nil
}

func (b *ignitionBuilder) constructFile(inputFile string) error {
	p := filepath.Join(b.baseDir, baseFileDir, inputFile)
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	fi, err := os.Stat(p)
	if err != nil {
		return err
	}
	mode := int(fi.Mode())

	storage, ok := b.ignition["storage"].(map[string]interface{})
	if !ok {
		storage = make(map[string]interface{})
	}
	files, ok := storage["files"].([]interface{})
	if !ok {
		files = make([]interface{}, 0)
	}
	files = append(files, map[string]interface{}{
		"path":       inputFile,
		"filesystem": "root",
		"mode":       mode,
		"contents": map[string]interface{}{
			"source": string(data),
		},
	})

	storage["files"] = files
	b.ignition["storage"] = storage

	return nil
}

func (b *ignitionBuilder) constructSystemd(s systemd) error {

	f, err := os.Open(filepath.Join(b.baseDir, baseSystemdDir, s.Source))
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	systemd, ok := b.ignition["systemd"].(map[string]interface{})
	if !ok {
		systemd = make(map[string]interface{})
	}
	units, ok := systemd["units"].([]interface{})
	if !ok {
		units = make([]interface{}, 0)
	}
	units = append(units, map[string]interface{}{
		"name":     s.Source,
		"enabled":  s.Enabled,
		"contents": string(data),
	})
	systemd["units"] = units
	b.ignition["systemd"] = systemd

	return nil
}

func (b *ignitionBuilder) constructNetworkd(n string) error {
	f, err := os.Open(filepath.Join(b.baseDir, baseNetworkdDir, n))
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	networkd, ok := b.ignition["networkd"].(map[string]interface{})
	if !ok {
		networkd = make(map[string]interface{})
	}
	units, ok := networkd["units"].([]interface{})
	if !ok {
		units = make([]interface{}, 0)
	}
	units = append(units, map[string]interface{}{
		"name":     n,
		"contents": string(data),
	})
	networkd["units"] = units
	b.ignition["networkd"] = networkd

	return nil
}
