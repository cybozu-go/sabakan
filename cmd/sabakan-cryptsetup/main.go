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
	serialFile    = "/sys/class/dmi/id/product_serial"
	modProbe      = "/sbin/cryptsetup"
)

var (
	flagServer *string
)

func main() {
	serverDefault := os.Getenv(envSabakanURL)
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
		log.ErrorExit(err)
	}
}

func getSerial() (string, error) {
	f, err := os.Open(serialFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	serialByte, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(serialByte)), nil
}

func execute(ctx context.Context) error {
	err := cmd.CommandContext(ctx, modProbe, "aesni-intel").Run()
	if err != nil {
		err := cmd.CommandContext(ctx, modProbe, "aes-x86_64").Run()
		if err != nil {
			return err
		}
	}

	serial, err := getSerial()
	if err != nil {
		return err
	}

	devices, err := detectStorageDevices(ctx, flag.Args())
	if err != nil {
		return err
	}

	for _, d := range devices {
		status := d.fetchKey(ctx, serial)

		// (1) if no problem, then do nothing
		if status == nil {
			continue
		}

		// (2) if error is not NotFound, then return error
		if status.Code() != client.ExitNotFound {
			return status
		}

		// (3) if error is NotFound, then initialize the device
		err = d.encrypt(ctx)
		if err != nil {
			return err
		}
		status = d.registerKey(ctx, serial)
		if status != nil {
			return status
		}
	}

	for _, d := range devices {
		err = d.decrypt(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
