package main

import (
	"context"
	"flag"

	"github.com/cybozu-go/sabakan/client"
	"github.com/google/subcommands"
)

type cryptsCmd struct{}

func (r cryptsCmd) SetFlags(f *flag.FlagSet) {}

func (r cryptsCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	newc := newCommander(f, "crypts")
	newc.Register(cryptsDeleteCommand(), "")
	return newc.Execute(ctx)
}

func cryptsCommand() subcommands.Command {
	return subcmd{
		cryptsCmd{},
		"crypts",
		"manage disk encryption key",
		"crypts ACTION ...",
	}
}

type cryptsDeleteCmd struct {
	force bool
}

func (c *cryptsDeleteCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&c.force, "force", false, "forces the removal of the disk encryption key")
}

func (c *cryptsDeleteCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if len(f.Args()) != 1 {
		f.Usage()
		return client.ExitUsageError
	}

	if !c.force {
		return client.ExitUsageError
	}

	err := client.CryptsDelete(ctx, f.Args()[0])
	return handleError(err)
}

func cryptsDeleteCommand() subcommands.Command {
	return subcmd{
		&cryptsDeleteCmd{},
		"delete",
		"delete all encryption keys of a machine",
		"delete --force SERIAL",
	}
}
