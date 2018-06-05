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

type ipamCmd struct {
	c *client.Client
}

func (r ipamCmd) SetFlags(f *flag.FlagSet) {}

func (r ipamCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	newc := newCommander(f, "ipam")
	newc.Register(ipamGetCommand(r.c), "")
	newc.Register(ipamSetCommand(r.c), "")
	return newc.Execute(ctx)
}

func ipamCommand(c *client.Client) subcommands.Command {
	return subcmd{
		ipamCmd{c},
		"ipam",
		"set/get IPAM configurations",
		"ipam ACTION ...",
	}
}

type ipamGetCmd struct {
	c *client.Client
}

func (r ipamGetCmd) SetFlags(f *flag.FlagSet) {}

func (r ipamGetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	conf, err := r.c.IPAMConfigGet(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err.Code()
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	e.Encode(conf)
	return client.ExitSuccess
}

func ipamGetCommand(c *client.Client) subcommands.Command {
	return subcmd{
		ipamGetCmd{c},
		"get",
		"get IPAM configurations",
		"get",
	}
}

type ipamSetCmd struct {
	c    *client.Client
	file string
}

func (r *ipamSetCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.file, "f", "", "settings file in JSON")
}

func (r *ipamSetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if r.file == "" {
		f.Usage()
		return client.ExitUsageError
	}
	file, err := os.Open(r.file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return client.ExitFailure
	}
	defer file.Close()

	var conf sabakan.IPAMConfig
	err = json.NewDecoder(file).Decode(&conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return client.ExitInvalidParams
	}

	errorStatus := r.c.IPAMConfigSet(ctx, &conf)
	if errorStatus != nil {
		fmt.Fprintln(os.Stderr, errorStatus)
		return errorStatus.Code()
	}
	return client.ExitSuccess
}

func ipamSetCommand(c *client.Client) subcommands.Command {
	return subcmd{
		&ipamSetCmd{c, ""},
		"set",
		"set IPAM configurations",
		"set",
	}
}
