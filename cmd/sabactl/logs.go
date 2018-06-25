package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/client"
	"github.com/google/subcommands"
)

var (
	newline = []byte("\n")
)

func ppLog(line []byte, w io.Writer) error {
	a := new(sabakan.AuditLog)
	err := json.Unmarshal(line, a)
	if err != nil {
		return err
	}

	ts := a.Timestamp.Format("Jan 02 15:04:05.000")
	detail := strings.Replace(a.Detail, "\n", " ", -1)
	if len(detail) > 20 {
		detail = detail[:20]
	}
	_, err = fmt.Fprintf(w, "%s %s@%s %s/%s %s detail:%s\n",
		ts, a.User, a.IP, string(a.Category), a.Instance, a.Action, detail)
	return err
}

type logPrinter struct {
	w   io.Writer
	buf bytes.Buffer
}

func (lp *logPrinter) Write(data []byte) (int, error) {
	newlines := bytes.Count(data, newline)
	n, err := lp.buf.Write(data)
	if err != nil {
		return n, err
	}

	for i := 0; i < newlines; i++ {
		line, err := lp.buf.ReadBytes('\n')
		if err != nil {
			return n, err
		}
		err = ppLog(line, lp.w)
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

type logsCmd struct {
	showJSON bool
}

func (r *logsCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&r.showJSON, "json", false, "show logs in JSON")
}

func (r *logsCmd) Execute(ctx context.Context, f *flag.FlagSet) subcommands.ExitStatus {
	var since, until time.Time
	switch f.NArg() {
	case 0:
		// pass
	case 2:
		t, err := time.Parse("20060102", f.Arg(1))
		if err != nil {
			return handleError(err)
		}
		until = t
		fallthrough
	case 1:
		t, err := time.Parse("20060102", f.Arg(0))
		if err != nil {
			return handleError(err)
		}
		since = t
		if until.IsZero() {
			until = t.Add(24 * time.Hour)
		}
	default:
		f.Usage()
		return subcommands.ExitUsageError
	}

	var w io.Writer = os.Stdout
	if !r.showJSON {
		w = &logPrinter{w: w}
	}

	err := client.LogsGet(ctx, since, until, w)
	return handleError(err)
}

func logsCommand() subcommands.Command {
	return subcmd{
		&logsCmd{},
		"logs",
		"retrieve logs",
		`logs [-json] [START_DATE] [END_DATE]

If START_DATE is given, and END_DATE is NOT given, logs
of START_DATE are retrieved.

If both of START_DATE and END_DATE are given, logs between them
are retrieved.`,
	}
}
