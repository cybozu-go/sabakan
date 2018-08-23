package main

import (
	"context"
	"flag"
	"net/http"
	"os"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan/client"
	"github.com/google/subcommands"
)

var (
	flagServer = flag.String("server", "http://localhost:10080", "<Listen IP>:<Port number>")
)

func main() {
	subcommands.Register(subcommands.HelpCommand(), "misc")
	subcommands.Register(subcommands.FlagsCommand(), "misc")
	subcommands.Register(subcommands.CommandsCommand(), "misc")
	subcommands.Register(dhcpCommand(), "")
	subcommands.Register(ipamCommand(), "")
	subcommands.Register(machinesCommand(), "")
	subcommands.Register(imagesCommand(), "")
	subcommands.Register(assetsCommand(), "")
	subcommands.Register(ignitionsCommand(), "")
	subcommands.Register(cryptsCommand(), "")
	subcommands.Register(logsCommand(), "")
	subcommands.Register(kernelParamsCommand(), "")
	subcommands.Register(versionCommand(), "")

	flag.Parse()
	cmd.LogConfig{}.Apply()

	err := client.Setup(*flagServer, &cmd.HTTPClient{
		Severity: log.LvDebug,
		Client:   &http.Client{},
	})
	if err != nil {
		log.ErrorExit(err)
	}

	exitStatus := subcommands.ExitSuccess
	cmd.Go(func(ctx context.Context) error {
		exitStatus = subcommands.Execute(ctx)
		return nil
	})
	cmd.Stop()
	cmd.Wait()
	os.Exit(int(exitStatus))
}
