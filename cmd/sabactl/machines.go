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

func (r machinesCmd) SetFlags(f *flag.FlagSet) {}

func (r machinesCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	cmdr := newCommander(f, "machines")
	cmdr.Register(machinesGetCommand(r.c), "")
	cmdr.Register(machinesCreateCommand(r.c), "")
	cmdr.Register(machinesRemoveCommand(r.c), "")
	return cmdr.Execute(ctx)
}

func machinesCommand(c *client.Client) subcommands.Command {
	return subcmd{
		machinesCmd{c},
		"machines",
		"manage machines",
		"machines ACTION ...",
	}
}

type machinesGetCmd struct {
	c     *client.Client
	query map[string]*string
}

var machinesGetQuery = map[string]string{
	"serial":     "Serial name",
	"datacenter": "Datacenter name",
	"rack":       "Rack name",
	"product":    "Product name (e.g. 'R630')",
	"ipv4":       "IPv4 address",
	"ipv6":       "IPv6 address",
	"bmc-type":   "BMC type",
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

func (r *machinesGetCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
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

func machinesGetCommand(c *client.Client) subcommands.Command {
	return subcmd{
		&machinesGetCmd{c: c},
		"get",
		"get machines information",
		"get",
	}
}

type machinesCreateCmd struct {
	c    *client.Client
	file string
}

func (r *machinesCreateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.file, "f", "", "machine file in JSON")
}

func (r *machinesCreateCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if r.file == "" {
		f.Usage()
		return client.ExitUsageError
	}
	file, err := os.Open(r.file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return client.ExitFailure
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

func machinesCreateCommand(c *client.Client) subcommands.Command {
	return subcmd{
		&machinesCreateCmd{c: c},
		"create",
		"create machines information",
		"create -f FILE",
	}
}

type machinesRemoveCmd struct {
	c *client.Client
}

func (r machinesRemoveCmd) SetFlags(f *flag.FlagSet) {}

func (r machinesRemoveCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if len(f.Args()) != 1 {
		f.Usage()
		return client.ExitUsageError
	}

	errorStatus := r.c.MachinesRemove(ctx, f.Args()[0])
	if errorStatus != nil {
		fmt.Fprintln(os.Stderr, errorStatus)
		return errorStatus.Code()
	}
	return client.ExitSuccess
}

func machinesRemoveCommand(c *client.Client) subcommands.Command {
	return subcmd{
		machinesRemoveCmd{c},
		"remove",
		"remove a machine information",
		"remove SERIAL",
	}
}
