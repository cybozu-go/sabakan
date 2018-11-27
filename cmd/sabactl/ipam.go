package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"

	"github.com/cybozu-go/sabakan"
	"github.com/google/subcommands"
)

type ipamCmd struct{}

func (r ipamCmd) SetFlags(f *flag.FlagSet) {}

func (r ipamCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	newc := newCommander(f, "ipam")
	newc.Register(ipamGetCommand(), "")
	newc.Register(ipamSetCommand(), "")
	return newc.Execute(ctx)
}

func ipamCommand() subcommands.Command {
	return subcmd{
		ipamCmd{},
		"ipam",
		"set/get IPAM configurations",
		"ipam ACTION ...",
	}
}

type ipamGetCmd struct{}

func (r ipamGetCmd) SetFlags(f *flag.FlagSet) {}

func (r ipamGetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	conf, err := api.IPAMConfigGet(ctx)
	if err != nil {
		return handleError(err)
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	e.Encode(conf)
	return ExitSuccess
}

func ipamGetCommand() subcommands.Command {
	return subcmd{
		ipamGetCmd{},
		"get",
		"get IPAM configurations",
		"get",
	}
}

type ipamSetCmd struct {
	file string
}

func (r *ipamSetCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.file, "f", "", "settings file in JSON")
}

func (r *ipamSetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if r.file == "" {
		f.Usage()
		return ExitUsageError
	}
	file, err := os.Open(r.file)
	if err != nil {
		return handleError(err)
	}
	defer file.Close()

	var conf sabakan.IPAMConfig
	err = json.NewDecoder(file).Decode(&conf)
	if err != nil {
		return handleError(err)
	}

	errorStatus := api.IPAMConfigSet(ctx, &conf)
	return handleError(errorStatus)
}

func ipamSetCommand() subcommands.Command {
	return subcmd{
		&ipamSetCmd{""},
		"set",
		"set IPAM configurations",
		"set -f FILE",
	}
}
