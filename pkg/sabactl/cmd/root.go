package cmd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/cybozu-go/sabakan/v2"
	"github.com/cybozu-go/sabakan/v2/client"
	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var (
	flagServer    string
	flagTLSServer string
	flagInsecure  bool
	httpApi       *client.Client
	httpsApi      *client.Client
)

const (
	// ExitSuccess represents no error.
	ExitSuccess = 0
	// ExitFailure represents general error.
	ExitFailure = 1
	// ExitUsageError represents bad usage of command.
	ExitUsageError = 2
	// ExitResponse4xx represents HTTP status 4xx.
	ExitResponse4xx = 4
	// ExitResponse5xx represents HTTP status 5xx.
	ExitResponse5xx = 5
	// ExitNotFound represents HTTP status 404.
	ExitNotFound = 14
	// ExitConflicted represents HTTP status 409.
	ExitConflicted = 19
)

// dummyRunFunc is used for subcommands which need not have Run or RunE.
func dummyRunFunc(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	return fmt.Errorf("unknown command %q for %q\n\nRun '%s --help' for usage", args[0], cmd.CommandPath(), cmd.CommandPath())
}

// isInvalidUsage is used to check for subcommand errors.
func isInvalidUsage(err error) bool {
	// "spf13/cobra" may return "unknown command" errors, so checking by a forward match of the error message.
	// ref: https://github.com/spf13/cobra/blob/v1.1.1/args.go#L22
	return strings.HasPrefix(err.Error(), "unknown command ")
}

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

		httpApi, err = client.NewClient(flagServer, &http.Client{})
		if err != nil {
			return err
		}
		httpsApi, err = client.NewClient(flagTLSServer, &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: flagInsecure,
				},
			},
		})
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
	case isInvalidUsage(err):
		code = ExitUsageError
	default:
		code = ExitFailure
	}
	os.Exit(code)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagServer, "server", "http://localhost:10080", "<Listen IP>:<Port number>")
	rootCmd.PersistentFlags().StringVar(&flagTLSServer, "tls-server", "https://localhost:10443", "<Listen IP>:<Port number>")
	rootCmd.PersistentFlags().BoolVar(&flagInsecure, "insecure", false, "Disable TLS verification")
}
