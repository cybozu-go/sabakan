package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/cybozu-go/sabakan"
	"github.com/google/subcommands"
)

type remoteConfigCmd struct {
	c *client
}

func (r *remoteConfigCmd) Name() string     { return "remote-config" }
func (r *remoteConfigCmd) Synopsis() string { return "Configure a sabakan server." }
func (r *remoteConfigCmd) Usage() string {
	return `Usage:
	remote-config get
	remote-config set -f <config-file.json>
`
}
func (r *remoteConfigCmd) SetFlags(f *flag.FlagSet) {}

func (r *remoteConfigCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	cmdr := subcommands.NewCommander(f, "remote-config")
	cmdr.Register(&remoteConfigGetCmd{c: r.c}, "")
	cmdr.Register(&remoteConfigSetCmd{c: r.c}, "")
	return cmdr.Execute(ctx)
}

type remoteConfigGetCmd struct {
	c *client
}

func (r *remoteConfigGetCmd) Name() string     { return "get" }
func (r *remoteConfigGetCmd) Synopsis() string { return "Configure a sabakan server." }
func (r *remoteConfigGetCmd) Usage() string {
	return "remote-config get -f <config-file.json>\n"
}
func (r *remoteConfigGetCmd) SetFlags(f *flag.FlagSet) {}

func (r *remoteConfigGetCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	conf, err := r.c.remoteConfigGet(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	e.Encode(conf)
	return 0
}

type remoteConfigSetCmd struct {
	c    *client
	file string
}

func (r *remoteConfigSetCmd) Name() string     { return "set" }
func (r *remoteConfigSetCmd) Synopsis() string { return "Configure a sabakan server." }
func (r *remoteConfigSetCmd) Usage() string {
	return "remote-config set -f <config-file.json>\n"
}
func (r *remoteConfigSetCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.file, "f", "", "settings file in JSON")
}

func (r *remoteConfigSetCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	file, err := os.Open(r.file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer file.Close()

	var conf sabakan.Config
	err = json.NewDecoder(file).Decode(&conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	r.c.remoteConfigSet(ctx, &conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
