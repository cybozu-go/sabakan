package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"

	"github.com/cybozu-go/sabakan/client"
	"github.com/google/subcommands"
)

type imagesCmd struct {
	os string
}

func (c *imagesCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.os, "os", "coreos", "OS identifier")
}

func (c *imagesCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	cmdr := newCommander(f, "images")
	cmdr.Register(imagesIndexCommand(c.os), "")
	cmdr.Register(imagesUploadCommand(c.os), "")
	cmdr.Register(imagesDeleteCommand(c.os), "")
	return cmdr.Execute(ctx)
}

func imagesCommand() subcommands.Command {
	return subcmd{
		&imagesCmd{"coreos"},
		"images",
		"manage boot images",
		"images ACTION ...",
	}
}

type imagesIndexCmd struct {
	os string
}

func (c imagesIndexCmd) SetFlags(f *flag.FlagSet) {}

func (c imagesIndexCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	index, err := client.ImagesIndex(ctx, c.os)
	if err != nil {
		return handleError(err)
	}

	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	e.Encode(index)
	return ExitSuccess
}

func imagesIndexCommand(os string) subcommands.Command {
	return subcmd{
		imagesIndexCmd{os: os},
		"index",
		"get index of images",
		"index",
	}
}

type imagesUploadCmd struct {
	os string
}

func (c imagesUploadCmd) SetFlags(f *flag.FlagSet) {}

func (c imagesUploadCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if len(f.Args()) != 3 {
		f.Usage()
		return ExitUsageError
	}

	err := client.ImagesUpload(ctx, c.os, f.Arg(0), f.Arg(1), f.Arg(2))
	return handleError(err)
}

func imagesUploadCommand(os string) subcommands.Command {
	return subcmd{
		imagesUploadCmd{os: os},
		"upload",
		"upload image",
		"upload ID KERNEL INITRD",
	}
}

type imagesDeleteCmd struct {
	os string
}

func (c imagesDeleteCmd) SetFlags(f *flag.FlagSet) {}

func (c imagesDeleteCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if len(f.Args()) != 1 {
		f.Usage()
		return ExitUsageError
	}

	err := client.ImagesDelete(ctx, c.os, f.Arg(0))
	return handleError(err)
}

func imagesDeleteCommand(os string) subcommands.Command {
	return subcmd{
		imagesDeleteCmd{os: os},
		"delete",
		"delete image",
		"delete ID",
	}
}
