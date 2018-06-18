package main

import (
	"context"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan/client"
)

const (
	envSabakanURL = "SABAKAN_URL"
)

var (
	flagSerialFile = flag.String("serial-file", "/etc/neco/serial", "serial number file")
	flagServer     *string
)

func main() {
	serverDefault = os.Getenv(envSabakanURL)
	if len(serverDefault) == 0 {
		serverDefault = "http://localhost:10080"
	}
	flagServer = flag.String("server", serverDefault, "http://<Listen IP>:<Port number>")

	flag.Parse()
	cmd.LogConfig{}.Apply()

	client.Setup(*flagServer, &cmd.HTTPClient{
		Severity: log.LvDebug,
		Client:   &http.Client{},
	})

	var err error
	cmd.Go(func(ctx context.Context) error {
		err = execute(ctx)
		return nil
	})
	cmd.Stop()
	cmd.Wait()
	if err != nil {
		os.Exit(1)
	}
}

func getSerial(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	serialByte, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return strings.TrimSpace(string(serialByte)), nil
}

func execute(ctx context.Context) error {
	serial, err := getSerial(flagSerialFile)
	if err != nil {
		return err
	}
}
