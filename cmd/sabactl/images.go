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

type imagesCmd struct {
	c  *client.Client
	os string
}

func (c *imagesCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.os, "os", "coreos", "OS identifier")
}

func (c *imagesCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	cmdr := newCommander(f, "images")
	cmdr.Register(imagesIndexCommand(c.c, c.os), "")
	//cmdr.Register(imagesUploadCommand(c.c, c.os), "")
	//cmdr.Register(imagesDeleteCommand(c.c, c.os), "")
	return cmdr.Execute(ctx)
}

func imagesCommand(c *client.Client) subcommands.Command {
	return subcmd{
		&imagesCmd{c, "coreos"},
		"images",
		"manage boot images",
		"images ACTION ...",
	}
}

type imagesIndexCmd struct {
	c  *client.Client
	os string
}

func (c imagesIndexCmd) SetFlags(f *flag.FlagSet) {}

func (c imagesIndexCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	index, err := c.c.ImagesIndex(ctx, c.os)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err.Code()
	}

	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	e.Encode(index)
	return client.ExitSuccess
}

func imagesIndexCommand(c *client.Client, os string) subcommands.Command {
	return subcmd{
		&imagesIndexCmd{c: c, os: os},
		"index",
		"get index of images",
		"index",
	}
}
