package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/client"
	"github.com/google/subcommands"
)

type dhcpCmd struct {
	c *client.Client
}

func (r dhcpCmd) SetFlags(f *flag.FlagSet) {}

func (r dhcpCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	newc := newCommander(f, "dhcp")
	newc.Register(dhcpGetCommand(r.c), "")
	newc.Register(dhcpSetCommand(r.c), "")
	return newc.Execute(ctx)
}

func dhcpCommand(c *client.Client) subcommands.Command {
	return subcmd{
		dhcpCmd{c},
		"dhcp",
		"set/get DHCP configurations",
		"dhcp ACTION ...",
	}
}

type dhcpGetCmd struct {
	c *client.Client
}

func (r dhcpGetCmd) SetFlags(f *flag.FlagSet) {}

func (r dhcpGetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	conf, err := r.c.DHCPConfigGet(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err.Code()
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	e.Encode(conf)
	return client.ExitSuccess
}

func dhcpGetCommand(c *client.Client) subcommands.Command {
	return subcmd{
		dhcpGetCmd{c},
		"get",
		"get DHCP configurations",
		"get",
	}
}

type dhcpSetCmd struct {
	c    *client.Client
	file string
}

func (r *dhcpSetCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.file, "f", "", "settings file in JSON")
}

func (r *dhcpSetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	file, err := os.Open(r.file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return client.ExitFailure
	}
	defer file.Close()

	var conf sabakan.DHCPConfig
	err = json.NewDecoder(file).Decode(&conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return client.ExitInvalidParams
	}

	errorStatus := r.c.DHCPConfigSet(ctx, &conf)
	if errorStatus != nil {
		fmt.Fprintln(os.Stderr, errorStatus)
		return errorStatus.Code()
	}
	return client.ExitSuccess
}

func dhcpSetCommand(c *client.Client) subcommands.Command {
	return subcmd{
		&dhcpSetCmd{c, ""},
		"set",
		"set DHCP configurations",
		"set",
	}
}
