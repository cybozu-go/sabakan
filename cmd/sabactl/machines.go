package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/cybozu-go/sabakan"
	"github.com/google/subcommands"
)

type machinesCmd struct{}

func (r machinesCmd) SetFlags(f *flag.FlagSet) {}

func (r machinesCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	cmdr := newCommander(f, "machines")
	cmdr.Register(machinesGetCommand(), "")
	cmdr.Register(machinesCreateCommand(), "")
	cmdr.Register(machinesRemoveCommand(), "")
	cmdr.Register(machinesSetStateCommand(), "")
	cmdr.Register(machinesGetStateCommand(), "")
	cmdr.Register(machinesSetRetireDateCommand(), "")
	return cmdr.Execute(ctx)
}

func machinesCommand() subcommands.Command {
	return subcmd{
		machinesCmd{},
		"machines",
		"manage machines",
		"machines ACTION ...",
	}
}

type machinesGetCmd struct {
	query map[string]*string
}

var machinesGetQuery = map[string]string{
	"serial":   "Serial name",
	"rack":     "Rack name",
	"role":     "Role name",
	"labels":   "Label name and value (-labels key=val,...)",
	"ipv4":     "IPv4 address",
	"ipv6":     "IPv6 address",
	"bmc-type": "BMC type",
	"state":    "State",
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
	machines, err := api.MachinesGet(ctx, r.getParams())
	if err != nil {
		return handleError(err)
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	e.Encode(machines)
	return ExitSuccess
}

func machinesGetCommand() subcommands.Command {
	return subcmd{
		&machinesGetCmd{},
		"get",
		"get machines information",
		"get",
	}
}

type machinesCreateCmd struct {
	file string
}

func (r *machinesCreateCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&r.file, "f", "", "machine file in JSON")
}

func (r *machinesCreateCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if r.file == "" {
		f.Usage()
		return ExitUsageError
	}
	file, err := os.Open(r.file)
	if err != nil {
		return handleError(err)
	}
	defer file.Close()

	var specs []*sabakan.MachineSpec
	err = json.NewDecoder(file).Decode(&specs)
	if err != nil {
		return handleError(err)
	}

	errorStatus := api.MachinesCreate(ctx, specs)
	return handleError(errorStatus)
}

func machinesCreateCommand() subcommands.Command {
	return subcmd{
		&machinesCreateCmd{},
		"create",
		"create machines information",
		"create -f FILE",
	}
}

type machinesRemoveCmd struct{}

func (r machinesRemoveCmd) SetFlags(f *flag.FlagSet) {}

func (r machinesRemoveCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if f.NArg() != 1 {
		f.Usage()
		return ExitUsageError
	}

	errorStatus := api.MachinesRemove(ctx, f.Arg(0))
	return handleError(errorStatus)
}

func machinesRemoveCommand() subcommands.Command {
	return subcmd{
		machinesRemoveCmd{},
		"remove",
		"remove a machine information",
		"remove SERIAL",
	}
}

type machinesSetStateCmd struct{}

func (r machinesSetStateCmd) SetFlags(f *flag.FlagSet) {}

func (r machinesSetStateCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if f.NArg() != 2 {
		f.Usage()
		return ExitUsageError
	}

	serial := f.Arg(0)
	state := strings.ToLower(f.Arg(1))

	errorStatus := api.MachinesSetState(ctx, serial, state)
	return handleError(errorStatus)
}

func machinesSetStateCommand() subcommands.Command {
	return subcmd{
		machinesSetStateCmd{},
		"set-state",
		"set the state of the machine",
		`Usage: sabactl machines set-state SERIAL STATE

STATE can be one of:
    uninitialized  The machine is not yet initialized.
    healthy        The machine has no problems.
    unhealthy      The machine has some problems.
    unreachable    The machine does not communicate with others.
    updating       The machine is updating.
    retiring       The machine should soon be retired/repaired.
`,
	}
}

type machinesGetStateCmd struct{}

func (r machinesGetStateCmd) SetFlags(f *flag.FlagSet) {}

func (r machinesGetStateCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if f.NArg() != 1 {
		f.Usage()
		return ExitUsageError
	}

	state, errorStatus := api.MachinesGetState(ctx, f.Arg(0))
	if errorStatus != nil {
		return handleError(errorStatus)
	}
	_, err := os.Stdout.WriteString(state.String())
	return handleError(err)
}

func machinesGetStateCommand() subcommands.Command {
	return subcmd{
		machinesGetStateCmd{},
		"get-state",
		"get the state of the machine",
		"get-state SERIAL",
	}
}

type machinesSetRetireDateCmd struct{}

func (r machinesSetRetireDateCmd) SetFlags(f *flag.FlagSet) {}

func (r machinesSetRetireDateCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	if f.NArg() != 2 {
		f.Usage()
		return ExitUsageError
	}

	serial := f.Arg(0)
	date, err := time.Parse("2006-01-02", f.Arg(1))
	if err != nil {
		return handleError(err)
	}

	errorStatus := api.MachinesSetRetireDate(ctx, serial, date)
	return handleError(errorStatus)
}

func machinesSetRetireDateCommand() subcommands.Command {
	return subcmd{
		machinesSetRetireDateCmd{},
		"set-retire-date",
		"set the retire date of the machine",
		`Usage: sabactl machines set-retire-date SERIAL YYYY-MM-DD`,
	}
}
