package cmd

import (
	"net/http"
	"os"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/client"
	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var (
	flagServer string

	api *client.Client
)

const (
	// ExitSuccess represents no error.
	ExitSuccess = 0
	// ExitFailure represents general error.
	ExitFailure = 1
	// ExitUsageError represents bad usage of command.
	ExitUsageError = 2
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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "sabactl",
	Short:   "command-line interface to control sabakan",
	Long:    `sabactl is a command-line interface to control sabakan.`,
	Version: sabakan.Version,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// without this, each subcommand's RunE would display usage text.
		cmd.SilenceUsage = true

		err := well.LogConfig{}.Apply()
		if err != nil {
			return err
		}

		api, err = client.NewClient(flagServer, &http.Client{})
		if err != nil {
			return err
		}

		return nil
	},
}

// Execute executes sabactl
func Execute() {
	err := rootCmd.Execute()
	if err == nil {
		return
	}

	var code int
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
		code = ExitFailure
	}
	os.Exit(code)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagServer, "server", "http://localhost:10080", "<Listen IP>:<Port number>")
}
