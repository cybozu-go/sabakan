package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/cybozu-go/sabakan/client"
	"github.com/google/subcommands"
)

const (
	// ExitSuccess represents no error.
	ExitSuccess subcommands.ExitStatus = subcommands.ExitSuccess
	// ExitFailure represents general error.
	ExitFailure = subcommands.ExitFailure
	// ExitUsageError represents bad usage of command.
	ExitUsageError = subcommands.ExitUsageError
	// ExitInvalidParams represents invalid input parameters for command.
	ExitInvalidParams = 3
	// ExitResponse4xx represents HTTP status 4xx.
	ExitResponse4xx = 4
	// ExitResponse5xx represents HTTP status 5xx.
	ExitResponse5xx = 5
	// ExitNotFound represents HTTP status 404.
	ExitNotFound = 14
	// ExitConflicted represents HTTP status 409.
	ExitConflicted = 19
)

// sub is the interface for newCommand
type sub interface {
	SetFlags(f *flag.FlagSet)
	Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus
}

type subcmd struct {
	sub
	name     string
	synopsis string
	usage    string
}

func (c subcmd) Name() string {
	return c.name
}

func (c subcmd) Synopsis() string {
	return c.synopsis
}

func (c subcmd) Usage() string {
	return c.usage + "\nFlags:\n"
}

func (c subcmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	return c.sub.Execute(ctx, f)
}

// handleError returns subcommands.ExitSuccess if err is nil.
// If err is non-nil, it returns subcommands.ExitFailure.
func handleError(err error) subcommands.ExitStatus {
	if err == nil {
		return subcommands.ExitSuccess
	}

	var code subcommands.ExitStatus
	switch {
	case client.IsNotFound(err):
		code = ExitNotFound
	case client.IsConflict(err):
		code = ExitConflicted
	case client.Is4xx(err):
		code = ExitResponse4xx
	case client.Is5xx(err):
		code = ExitResponse5xx
	default:
		code = subcommands.ExitFailure
	}

	fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
	return code
}

// newCommander creates a subcommands.Commander for nested sub commands.
// This registers "flags" and "help" sub commands for the new commander.
func newCommander(f *flag.FlagSet, name string) *subcommands.Commander {
	name = fmt.Sprintf("%s %s", path.Base(os.Args[0]), name)
	c := subcommands.NewCommander(f, name)
	c.Register(c.FlagsCommand(), "misc")
	c.Register(c.HelpCommand(), "misc")
	return c
}
