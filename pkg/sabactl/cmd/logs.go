package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cybozu-go/sabakan/v2"
	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var (
	newline = []byte("\n")

	logsJSON bool
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

var logsCmd = &cobra.Command{
	Use:   "logs [--json] [START_DATE] [END_DATE}",
	Short: "retrieve logs",
	Long: `If START_DATE is given, and END_DATE is NOT given, logs
of START_DATE are retrieved.

If both of START_DATE and END_DATE are given, logs between them
are retrieved.`,
	Args: cobra.MaximumNArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		var since, until time.Time
		switch len(args) {
		case 0:
			// pass
		case 2:
			t, err := time.Parse("20060102", args[1])
			if err != nil {
				return err
			}
			until = t
			fallthrough
		case 1:
			t, err := time.Parse("20060102", args[0])
			if err != nil {
				return err
			}
			since = t
			if until.IsZero() {
				until = t.Add(24 * time.Hour)
			}
		}
		well.Go(func(ctx context.Context) error {
			w := cmd.OutOrStdout()
			if !logsJSON {
				w = &logPrinter{w: w}
			}
			return httpApi.LogsGet(ctx, since, until, w)
		})
		well.Stop()
		return well.Wait()
	},
}

func init() {
	logsCmd.Flags().BoolVar(&logsJSON, "json", false, "show logs in JSON")

	rootCmd.AddCommand(logsCmd)
}
