package client

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/coreos/container-linux-config-transpiler/config/types"
	"gopkg.in/yaml.v2"
)

const baseFileDir = "./files"
const baseSystemdDir = "./systemd"
const baseNetworkdDir = "./networkd"

type systemd struct {
	Enabled bool   `yaml:"enabled"`
	Source  string `yaml:"source"`
}

type ignitionSource struct {
	Passwd   string    `yaml:"passwd"`
	Files    []string  `yaml:"files"`
	Systemd  []systemd `yaml:"systemd"`
	Networkd []string  `yaml:"networkd"`
}

func constructCLConfigTemplate(fname string) (io.Reader, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, ErrorStatus(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, ErrorStatus(err)
	}

	var source ignitionSource
	err = yaml.Unmarshal(data, &source)
	if err != nil {
		return nil, ErrorStatus(err)
	}

	var clConf types.Config

	if source.Passwd != "" {
		err = constructPasswd(source.Passwd, &clConf)
		if err != nil {
			return nil, ErrorStatus(err)
		}
	}

	for _, file := range source.Files {
		err = constructFile(file, &clConf)
		if err != nil {
			return nil, ErrorStatus(err)
		}
	}

	for _, s := range source.Systemd {
		err = constructSystemd(s, &clConf)
		if err != nil {
			return nil, ErrorStatus(err)
		}
	}

	for _, n := range source.Networkd {
		err = constructNetworkd(n, &clConf)
		if err != nil {
			return nil, ErrorStatus(err)
		}
	}

	b, err := yaml.Marshal(clConf)
	if err != nil {
		return nil, ErrorStatus(err)
	}
	fmt.Println(string(b))

	return bytes.NewReader(b), nil
}

func constructPasswd(passwd string, clConf *types.Config) error {
	pf, err := os.Open(passwd)
	if err != nil {
		return err
	}
	defer pf.Close()
	passData, err := ioutil.ReadAll(pf)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(passData, &clConf.Passwd)
}

func constructFile(inputFile string, clConf *types.Config) error {
	p := filepath.Join(baseFileDir, inputFile)
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

	clConf.Storage.Files = append(clConf.Storage.Files, types.File{
		Path:       inputFile,
		Filesystem: "root",
		Mode:       &mode,
		Contents: types.FileContents{
			Inline: string(data),
		},
	})

	return nil
}

func constructSystemd(s systemd, clConf *types.Config) error {

	f, err := os.Open(filepath.Join(baseSystemdDir, s.Source))
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	clConf.Systemd.Units = append(clConf.Systemd.Units, types.SystemdUnit{
		Name:     s.Source,
		Enabled:  &s.Enabled,
		Contents: string(data),
	})

	return nil
}

func constructNetworkd(n string, clConf *types.Config) error {

	f, err := os.Open(filepath.Join(baseNetworkdDir, n))
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	clConf.Networkd.Units = append(clConf.Networkd.Units, types.NetworkdUnit{
		Name:     n,
		Contents: string(data),
	})

	return nil
}
