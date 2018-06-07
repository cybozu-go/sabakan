package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"

	"github.com/cybozu-go/sabakan/client"
	"github.com/google/subcommands"
)

type ignitionsCmd struct{}

func (r ignitionsCmd) SetFlags(f *flag.FlagSet) {}

func (r ignitionsCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	cmdr := newCommander(f, "ignitions")
	cmdr.Register(ignitionsGetCommand(), "")
	cmdr.Register(ignitionsCatCommand(), "")
	cmdr.Register(ignitionsSetCommand(), "")
	cmdr.Register(ignitionsDeleteCommand(), "")
	return cmdr.Execute(ctx)
}

func ignitionsCommand() subcommands.Command {
	return subcmd{
		sub:      ignitionsCmd{},
		name:     "ignitions",
		synopsis: "manage ignitions",
		usage:    "ignitions ACTION ...",
	}
}

type ignitionsGetCmd struct{}

func (c ignitionsGetCmd) SetFlags(f *flag.FlagSet) {}

func (c ignitionsGetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if f.NArg() != 1 {
		f.Usage()
		return client.ExitUsageError
	}
	ids, status := client.IgnitionsGet(ctx, f.Arg(0))
	if status != nil {
		return handleError(status)
	}

	err := json.NewEncoder(os.Stdout).Encode(ids)
	if err != nil {
		return handleError(err)
	}
	return client.ExitSuccess
}

func ignitionsGetCommand() subcommands.Command {
	return subcmd{
		ignitionsGetCmd{},
		"get",
		"get IDs of ROLE",
		"get ROLE",
	}
}

type ignitionsCatCmd struct{}

func (c ignitionsCatCmd) SetFlags(f *flag.FlagSet) {}

func (c ignitionsCatCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if f.NArg() != 2 {
		f.Usage()
		return client.ExitUsageError
	}
	tmpl, status := client.IgnitionsCat(ctx, f.Arg(0), f.Arg(1))
	if status != nil {
		return handleError(status)
	}

	_, err := os.Stdout.WriteString(tmpl)
	if err != nil {
		return handleError(err)
	}

	return client.ExitSuccess
}

func ignitionsCatCommand() subcommands.Command {
	return subcmd{
		ignitionsCatCmd{},
		"cat",
		"show an ignition template for the ID and ROLE",
		"cat ROLE ID",
	}
}

type ignitionsSetCmd struct {
	file string
}

func (c *ignitionsSetCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.file, "f", "", "ignition template file")
}

func (c *ignitionsSetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if f.NArg() != 1 || c.file == "" {
		f.Usage()
		return client.ExitUsageError
	}

	data, status := client.IgnitionsSet(ctx, f.Arg(0), c.file)
	if status != nil {
		return handleError(status)
	}

	err := json.NewEncoder(os.Stdout).Encode(data)
	if err != nil {
		return handleError(err)
	}
	return client.ExitSuccess
}

func ignitionsSetCommand() subcommands.Command {
	return subcmd{
		&ignitionsSetCmd{},
		"set",
		"create ignition template",
		"set -f FILE ROLE ",
	}
}

type ignitionsDeleteCmd struct{}

func (c ignitionsDeleteCmd) SetFlags(f *flag.FlagSet) {}

func (c ignitionsDeleteCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if f.NArg() != 2 {
		f.Usage()
		return client.ExitUsageError
	}
	status := client.IgnitionsDelete(ctx, f.Arg(0), f.Arg(1))
	if status != nil {
		return handleError(status)
	}
	return client.ExitSuccess
}

func ignitionsDeleteCommand() subcommands.Command {
	return subcmd{
		ignitionsDeleteCmd{},
		"delete",
		"delete an ignition template for the ID and ROLE",
		"delete ROLE ID",
	}
}
