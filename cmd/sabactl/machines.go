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
	cmdr.Register(&machinesAddCmd{c: r.c}, "")
	cmdr.Register(&machinesUpdateCmd{c: r.c}, "")
	return cmdr.Execute(ctx)
}

//
// machines get
//
type machinesGetCmd struct {
	c          *sabakan.Client
	query      map[string]*string
	serial     string
	datacenter string
	rack       string
	cluster    string
	product    string
	ipv4       string
	ipv6       string
}

func (r *machinesGetCmd) Name() string     { return "get" }
func (r *machinesGetCmd) Synopsis() string { return "get machines information." }
func (r *machinesGetCmd) Usage() string {
	return "sabactl machines get [options]"
}
func (r *machinesGetCmd) SetFlags(f *flag.FlagSet) {
	r.query = map[string]*string{}
	for _, k := range []string{"serial", "datacenter", "rack", "cluster", "product", "ipv4", "ipv6"} {
		r.query[k] = f.String(k, "", k)
	}
}

func (r *machinesGetCmd) getParams() map[string]string {
	var params = map[string]string{}
	for k, v := range r.query {
		if len(*v) > 0 {
			params[k] = *v
		}
	}
	return params
}

func (r *machinesGetCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	machines, err := r.c.MachinesGet(ctx, r.getParams())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	e.Encode(machines)
	return 0
}

//
// machines add
//
type machinesAddCmd struct {
	c    *sabakan.Client
	file string
}

func (r *machinesAddCmd) Name() string     { return "add" }
func (r *machinesAddCmd) Synopsis() string { return "add machines information." }
func (r *machinesAddCmd) Usage() string {
	return "machines add -f <machines-file.json>\n"
}
func (r *machinesAddCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.file, "f", "", "machine file in JSON")
}

func (r *machinesAddCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
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

	err = r.c.MachinesAdd(ctx, machines)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

//
// machines update
//
type machinesUpdateCmd struct {
	c    *sabakan.Client
	file string
}

func (r *machinesUpdateCmd) Name() string     { return "update" }
func (r *machinesUpdateCmd) Synopsis() string { return "update machines information." }
func (r *machinesUpdateCmd) Usage() string {
	return "machines update -f <machines-file.json>\n"
}
func (r *machinesUpdateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.file, "f", "", "machine file in JSON")
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
