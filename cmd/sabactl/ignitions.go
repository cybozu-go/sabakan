package main

import (
	"bytes"
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
		return ExitUsageError
	}
	index, status := api.IgnitionsGet(ctx, f.Arg(0))
	if status != nil {
		return handleError(status)
	}

	err := json.NewEncoder(os.Stdout).Encode(index)
	return handleError(err)
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
		return ExitUsageError
	}
	status := api.IgnitionsCat(ctx, f.Arg(0), f.Arg(1), os.Stdout)
	return handleError(status)
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
	meta mapFlags
}

func (c *ignitionsSetCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.file, "f", "", "ignition template file")
	f.Var(&c.meta, "meta", "optional metadata <KEY>=<VALUE>")
}

func (c *ignitionsSetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if f.NArg() != 2 || c.file == "" {
		f.Usage()
		return ExitUsageError
	}

	buf := new(bytes.Buffer)
	err := client.AssembleIgnitionTemplate(c.file, buf)
	if err != nil {
		return handleError(err)
	}
	err = api.IgnitionsSet(ctx, f.Arg(0), f.Arg(1), buf, c.meta)
	return handleError(err)
}

func ignitionsSetCommand() subcommands.Command {
	return subcmd{
		&ignitionsSetCmd{meta: mapFlags{}},
		"set",
		"create ignition template",
		"set -f FILE ROLE ID",
	}
}

type ignitionsDeleteCmd struct{}

func (c ignitionsDeleteCmd) SetFlags(f *flag.FlagSet) {}

func (c ignitionsDeleteCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if f.NArg() != 2 {
		f.Usage()
		return ExitUsageError
	}
	status := api.IgnitionsDelete(ctx, f.Arg(0), f.Arg(1))
	return handleError(status)
}

func ignitionsDeleteCommand() subcommands.Command {
	return subcmd{
		ignitionsDeleteCmd{},
		"delete",
		"delete an ignition template for the ID and ROLE",
		"delete ROLE ID",
	}
}
