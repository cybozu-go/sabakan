package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/google/subcommands"
)

// const version is sabactl version
const version = "1.0.0"

type versionCmd struct{}

func (v versionCmd) SetFlags(f *flag.FlagSet) {}

func (v versionCmd) Execute(_ context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	fmt.Printf("%v\n", version)
	return subcommands.ExitSuccess
}

func versionCommand() subcommands.Command {
	return subcmd{
		versionCmd{},
		"version",
		"show sabactl version",
		"version ACTION ...",
	}
}
