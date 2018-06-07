package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/client"
	"github.com/google/subcommands"
)

type dhcpCmd struct{}

func (r dhcpCmd) SetFlags(f *flag.FlagSet) {}

func (r dhcpCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	newc := newCommander(f, "dhcp")
	newc.Register(dhcpGetCommand(), "")
	newc.Register(dhcpSetCommand(), "")
	return newc.Execute(ctx)
}

func dhcpCommand() subcommands.Command {
	return subcmd{
		dhcpCmd{},
		"dhcp",
		"set/get DHCP configurations",
		"dhcp ACTION ...",
	}
}

type dhcpGetCmd struct{}

func (r dhcpGetCmd) SetFlags(f *flag.FlagSet) {}

func (r dhcpGetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	conf, err := client.DHCPConfigGet(ctx)
	if err != nil {
		return handleError(err)
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	e.Encode(conf)
	return client.ExitSuccess
}

func dhcpGetCommand() subcommands.Command {
	return subcmd{
		dhcpGetCmd{},
		"get",
		"get DHCP configurations",
		"get",
	}
}

type dhcpSetCmd struct {
	file string
}

func (r *dhcpSetCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.file, "f", "", "settings file in JSON")
}

func (r *dhcpSetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if r.file == "" {
		f.Usage()
		return client.ExitUsageError
	}
	file, err := os.Open(r.file)
	if err != nil {
		return handleError(err)
	}
	defer file.Close()

	var conf sabakan.DHCPConfig
	err = json.NewDecoder(file).Decode(&conf)
	if err != nil {
		return handleError(err)
	}

	errorStatus := client.DHCPConfigSet(ctx, &conf)
	return handleError(errorStatus)
}

func dhcpSetCommand() subcommands.Command {
	return subcmd{
		&dhcpSetCmd{""},
		"set",
		"set DHCP configurations",
		"set -f FILE",
	}
}
