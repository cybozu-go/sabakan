package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/cybozu-go/sabakan"
	"github.com/google/subcommands"
)

type versionCmd struct{}

func (v versionCmd) SetFlags(f *flag.FlagSet) {}

func (v versionCmd) Execute(_ context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	fmt.Println(sabakan.Version)
	return subcommands.ExitSuccess
}

func versionCommand() subcommands.Command {
	return subcmd{
		versionCmd{},
		"version",
		"show sabactl version",
		"",
	}
}
