package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/client"
	"github.com/google/subcommands"
)

type kernelParamsCmd struct {
	os string
}

func (c *kernelParamsCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.os, "os", "coreos", "OS identifier")
}

func (c *kernelParamsCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	newc := newCommander(f, "kernel-params")
	newc.Register(kernelParamsGetCommand(c.os), "")
	newc.Register(kernelParamsSetCommand(c.os), "")
	return newc.Execute(ctx)
}

func kernelParamsCommand() subcommands.Command {
	return subcmd{
		&kernelParamsCmd{"coreos"},
		"kernel-params",
		"set/get kernel parameters",
		"kernel-params ACTION ...",
	}
}

type kernelParamsGetCmd struct {
	os string
}

func (c kernelParamsGetCmd) SetFlags(f *flag.FlagSet) {}

func (c kernelParamsGetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	kernelParams, err := client.KernelParamsGet(ctx, c.os)
	if err != nil {
		return handleError(err)
	}
	fmt.Println(kernelParams)

	return client.ExitSuccess
}

func kernelParamsGetCommand(os string) subcommands.Command {
	return subcmd{
		kernelParamsGetCmd{os: os},
		"get",
		"get kernel parameters",
		"get",
	}
}

type kernelParamsSetCmd struct {
	os string
}

func (c kernelParamsSetCmd) SetFlags(f *flag.FlagSet) {}

func (c kernelParamsSetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if len(f.Args()) != 1 {
		f.Usage()
		return client.ExitUsageError
	}

	err := client.KernelParamsSet(ctx, c.os, sabakan.KernelParams(f.Arg(0)))
	return handleError(err)
}

func kernelParamsSetCommand(os string) subcommands.Command {
	return subcmd{
		kernelParamsSetCmd{os: os},
		"set",
		"set kernel parameters",
		"set KERNEL_PARAMS",
	}
}
