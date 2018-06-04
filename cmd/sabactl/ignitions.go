package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/cybozu-go/sabakan/client"
	"github.com/google/subcommands"
)

type ignitionsCmd struct {
	c *client.Client
}

func (r ignitionsCmd) SetFlags(f *flag.FlagSet) {

}

func (r ignitionsCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	cmdr := newCommander(f, "ignitions")
	cmdr.Register(ignitionsGetCommand(r.c), "")
	cmdr.Register(ignitionsCatCommand(r.c), "")
	cmdr.Register(ignitionsSetCommand(r.c), "")
	cmdr.Register(ignitionsDeleteCommand(r.c), "")
	return cmdr.Execute(ctx)
}

func ignitionsCommand(c *client.Client) subcommands.Command {
	return subcmd{
		sub:      ignitionsCmd{c},
		name:     "ignitions",
		synopsis: "manage ignitions",
		usage:    "ignitions ACTION ...",
	}
}

func ignitionsGetCommand(c *client.Client) subcommands.Command {
	return subcmd{
		&ignitionsGetCmd{c},
		"get",
		"get IDs of ROLE",
		"get ROLE",
	}
}

type ignitionsGetCmd struct {
	c *client.Client
}

func (c ignitionsGetCmd) SetFlags(f *flag.FlagSet) {}

func (c ignitionsGetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if len(f.Args()) != 1 {
		return client.ExitUsageError
	}
	ids, status := c.c.IgnitionsGet(ctx, f.Arg(0))
	if status != nil {
		fmt.Fprintln(os.Stderr, status)
		return status.Code()
	}

	err := json.NewEncoder(os.Stdout).Encode(ids)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return client.ExitFailure
	}
	return client.ExitSuccess
}

func ignitionsCatCommand(c *client.Client) subcommands.Command {
	return subcmd{
		&ignitionsCatCmd{c},
		"cat",
		"show an ignition template for the ID and ROLE",
		"cat ROLE ID",
	}
}

type ignitionsCatCmd struct {
	c *client.Client
}

func (c ignitionsCatCmd) SetFlags(f *flag.FlagSet) {}

func (c ignitionsCatCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if len(f.Args()) != 2 {
		return client.ExitUsageError
	}
	tmpl, status := c.c.IgnitionsCat(ctx, f.Arg(0), f.Arg(1))
	if status != nil {
		fmt.Fprintln(os.Stderr, status)
		return status.Code()
	}

	fmt.Printf(tmpl)
	return client.ExitSuccess
}

func ignitionsSetCommand(c *client.Client) subcommands.Command {
	return subcmd{
		&ignitionsSetCmd{c: c},
		"set",
		"create ignition template",
		"set ROLE -f FILE",
	}
}

type ignitionsSetCmd struct {
	c    *client.Client
	file string
}

func (c *ignitionsSetCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.file, "f", "", "ignition template file")
}

func (c *ignitionsSetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if f.NArg() != 1 {
		return client.ExitUsageError
	}

	data, status := c.c.IgnitionsSet(ctx, f.Arg(0), c.file)
	if status != nil {
		fmt.Fprintln(os.Stderr, status)
		return status.Code()
	}

	err := json.NewEncoder(os.Stdout).Encode(data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return client.ExitFailure
	}
	return client.ExitSuccess
}

func ignitionsDeleteCommand(c *client.Client) subcommands.Command {
	return subcmd{
		&ignitionsDeleteCmd{c},
		"delete",
		"delete an ignition template for the ID and ROLE",
		"delete ROLE ID",
	}
}

type ignitionsDeleteCmd struct {
	c *client.Client
}

func (c ignitionsDeleteCmd) SetFlags(f *flag.FlagSet) {}

func (c ignitionsDeleteCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if len(f.Args()) != 2 {
		return client.ExitUsageError
	}
	status := c.c.IgnitionsDelete(ctx, f.Arg(0), f.Arg(1))
	if status != nil {
		fmt.Fprintln(os.Stderr, status)
		return status.Code()
	}
	return client.ExitSuccess
}
