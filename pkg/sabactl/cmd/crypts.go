package cmd

import (
	"context"
	"errors"

	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var cryptsDeleteForce bool

var cryptsCmd = &cobra.Command{
	Use:   "crypts delete",
	Short: "manage disk encryption key",
	Long:  `Manage disk encryption key of the machines.`,
	RunE:  dummyRunFunc,
}

var cryptsDeleteCmd = &cobra.Command{
	Use:   "delete --force SERIAL",
	Short: "delete all encryption keys of a machine",
	Long:  `Delete all encryption keys of a machine.`,
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		if !cryptsDeleteForce {
			return errors.New("use --force explicitly")
		}
		well.Go(func(ctx context.Context) error {
			return api.CryptsDelete(ctx, args[0])
		})
		well.Stop()
		return well.Wait()

	},
}

func init() {
	cryptsDeleteCmd.Flags().BoolVar(&cryptsDeleteForce, "force", false, "forces the removal of the disk encryption key")

	cryptsCmd.AddCommand(cryptsDeleteCmd)
	rootCmd.AddCommand(cryptsCmd)
}
