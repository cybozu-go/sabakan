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

type machinesCmd struct {
	c *sabakan.Client
}

func (r *machinesCmd) Name() string     { return "machines" }
func (r *machinesCmd) Synopsis() string { return "manage machines." }
func (r *machinesCmd) Usage() string {
	return `Usage:
	machines get [options]
	machines update -f <machines-file.json>
`
}
func (r *machinesCmd) SetFlags(f *flag.FlagSet) {}

func (r *machinesCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	cmdr := subcommands.NewCommander(f, "machines")
	cmdr.Register(&machinesGetCmd{c: r.c}, "")
	cmdr.Register(&machinesUpdateCmd{c: r.c}, "")
	return cmdr.Execute(ctx)
}

type machinesGetCmd struct {
	c          *sabakan.Client
	serial     string
	datacenter string
	rack       string
	cluster    string
	product    string
	ipv4       string
	ipv6       string
}

func (r *machinesGetCmd) Name() string     { return "get" }
func (r *machinesGetCmd) Synopsis() string { return "get machine informations." }
func (r *machinesGetCmd) Usage() string {
	return "sabactl machines get [options]"
}
func (r *machinesGetCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.serial, "serial", "", "serial")
	f.StringVar(&r.datacenter, "datacenter", "", "datacenter")
	f.StringVar(&r.rack, "rack", "", "rack")
	f.StringVar(&r.cluster, "cluster", "", "cluster")
	f.StringVar(&r.product, "product", "", "product")
	f.StringVar(&r.ipv4, "ipv4", "", "ipv4 address")
	f.StringVar(&r.ipv6, "ipv6", "", "ipv6 address")
}

func (r *machinesGetCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	conf, err := r.c.MachinesGet(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	e.Encode(conf)
	return 0
}

type machinesUpdateCmd struct {
	c    *sabakan.Client
	file string
}

func (r *machinesUpdateCmd) Name() string     { return "update" }
func (r *machinesUpdateCmd) Synopsis() string { return "update machine informations." }
func (r *machinesUpdateCmd) Usage() string {
	return "machines update -f <machines-file.json>\n"
}
func (r *machinesUpdateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.file, "f", "", "settings file in JSON")
}

func (r *machinesUpdateCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	file, err := os.Open(r.file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer file.Close()

	var machines []sabakan.Machine
	err = json.NewDecoder(file).Decode(&machines)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	err = r.c.MachinesUpdate(ctx, machines)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
