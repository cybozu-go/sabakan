package main

import (
	"context"
	"flag"
	"net/http"
	"os"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
	"github.com/google/subcommands"
)

var (
	flagServer = flag.String("server", "http://localhost:8888", "<Listen IP>:<Port number>")
)

func main() {
	c := sabakan.NewClient(*flagServer, &cmd.HTTPClient{
		Client:   &http.Client{},
		Severity: log.LvDebug,
	})

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&remoteConfigCmd{c: c}, "")

	flag.Parse()

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
