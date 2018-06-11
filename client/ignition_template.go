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

func constructIgnitionTemplate(fname string) (io.Reader, error) {
	absPath, err := filepath.Abs(fname)
	if err != nil {
		return nil, ErrorStatus(err)
	}
	baseDir := filepath.Dir(absPath)

	source, err := loadSource(absPath)
	if err != nil {
		return nil, ErrorStatus(err)
	}

	clConf := make(map[string]interface{})
	clConf["ignition"] = map[string]interface{}{
		"version": "2.2.0",
	}
	err = constructCLConf(baseDir, source, clConf)
	if err != nil {
		return nil, ErrorStatus(err)
	}

	b, err := yaml.Marshal(clConf)
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

func constructCLConf(baseDir string, source *ignitionSource, clConf map[string]interface{}) error {
	if source.Include != "" {
		include, err := loadSource(filepath.Join(baseDir, source.Include))
		if err != nil {
			return err
		}
		err = constructCLConf(baseDir, include, clConf)
		if err != nil {
			return err
		}
	}
	if source.Passwd != "" {
		err := constructPasswd(filepath.Join(baseDir, source.Passwd), clConf)
		if err != nil {
			return err
		}
	}

	for _, file := range source.Files {
		err := constructFile(baseDir, file, clConf)
		if err != nil {
			return err
		}
	}

	for _, s := range source.Systemd {
		err := constructSystemd(baseDir, s, clConf)
		if err != nil {
			return err
		}
	}

	for _, n := range source.Networkd {
		err := constructNetworkd(baseDir, n, clConf)
		if err != nil {
			return err
		}
	}
	return nil
}

func constructPasswd(passwd string, clConf map[string]interface{}) error {
	pf, err := os.Open(passwd)
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
	clConf["passwd"] = p

	return nil
}

func constructFile(baseDir, inputFile string, clConf map[string]interface{}) error {
	p := filepath.Join(baseDir, baseFileDir, inputFile)
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

	storage, ok := clConf["storage"].(map[string]interface{})
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
	clConf["storage"] = storage

	return nil
}

func constructSystemd(baseDir string, s systemd, clConf map[string]interface{}) error {

	f, err := os.Open(filepath.Join(baseDir, baseSystemdDir, s.Source))
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	systemd, ok := clConf["systemd"].(map[string]interface{})
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
	clConf["systemd"] = systemd

	return nil
}

func constructNetworkd(baseDir, n string, clConf map[string]interface{}) error {

	f, err := os.Open(filepath.Join(baseDir, baseNetworkdDir, n))
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	networkd, ok := clConf["networkd"].(map[string]interface{})
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
	clConf["networkd"] = networkd

	return nil
}
