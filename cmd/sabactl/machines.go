package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/client"
	"github.com/google/subcommands"
)

type machinesCmd struct {
	c *client.Client
}

func (r *machinesCmd) Name() string     { return "machines" }
func (r *machinesCmd) Synopsis() string { return "manage machines." }
func (r *machinesCmd) Usage() string {
	return `Usage:
	machines get [options]
	machines create -f <machines-file.json>
`
}
func (r *machinesCmd) SetFlags(f *flag.FlagSet) {}

func (r *machinesCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	cmdr := subcommands.NewCommander(f, "machines")
	cmdr.Register(&machinesGetCmd{c: r.c}, "")
	cmdr.Register(&machinesCreateCmd{c: r.c}, "")
	return cmdr.Execute(ctx)
}

type machinesGetCmd struct {
	c     *client.Client
	query map[string]*string
}

func (r *machinesGetCmd) Name() string     { return "get" }
func (r *machinesGetCmd) Synopsis() string { return "get machines information." }
func (r *machinesGetCmd) Usage() string {
	return "sabactl machines get [options]\n"
}

var machinesGetQuery = map[string]string{
	"serial":     "Serial name",
	"datacenter": "Datacenter name",
	"rack":       "Rack name",
	"product":    "Product name (e.g. 'R630')",
	"ipv4":       "IPv4 address",
	"ipv6":       "IPv6 address",
}

func (r *machinesGetCmd) SetFlags(f *flag.FlagSet) {
	r.query = map[string]*string{}
	for k, v := range machinesGetQuery {
		r.query[k] = f.String(k, "", v)
	}
}

func (r *machinesGetCmd) getParams() map[string]string {
	var params = map[string]string{}
	for k := range machinesGetQuery {
		params[k] = *r.query[k]
	}
	return params
}

func (r *machinesGetCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	machines, err := r.c.MachinesGet(ctx, r.getParams())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err.Code()
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	e.Encode(machines)
	return client.ExitSuccess
}

type machinesCreateCmd struct {
	c    *client.Client
	file string
}

func (r *machinesCreateCmd) Name() string     { return "create" }
func (r *machinesCreateCmd) Synopsis() string { return "create machines information." }
func (r *machinesCreateCmd) Usage() string {
	return "machines create -f <machines-file.json>\n"
}
func (r *machinesCreateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.file, "f", "", "machine file in JSON")
}

func (r *machinesCreateCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	file, err := os.Open(r.file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return client.ExitError
	}
	defer file.Close()

	var machines []sabakan.Machine
	err = json.NewDecoder(file).Decode(&machines)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return client.ExitInvalidParams
	}

	errorStatus := r.c.MachinesCreate(ctx, machines)
	if errorStatus != nil {
		fmt.Fprintln(os.Stderr, errorStatus)
		return errorStatus.Code()
	}
	return client.ExitSuccess
}

type machinesDeleteCmd struct {
	c      *client.Client
	serial string
}

func (r *machinesDeleteCmd) Name() string     { return "delete" }
func (r *machinesDeleteCmd) Synopsis() string { return "delete machine information." }
func (r *machinesDeleteCmd) Usage() string {
	return "machines delete <machine-serial>\n"
}
func (r *machinesDeleteCmd) SetFlags(f *flag.FlagSet) {}

func (r *machinesDeleteCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if len(f.Args()) != 1 {
		return client.ExitBadUsage
	}

	errorStatus := r.c.MachinesDelete(ctx, f.Args()[0])
	if errorStatus != nil {
		fmt.Fprintln(os.Stderr, errorStatus)
		return errorStatus.Code()
	}
	return client.ExitSuccess
}
