package cmd

import (
	"context"
	"fmt"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var kernelParamsOS string

var kernelParamsCmd = &cobra.Command{
	Use:   "kernel-params",
	Short: "manage kernel parameters",
	Long:  `Manage kernel parameters in sabakan.`,
}

var kernelParamsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "get current kernel parameters",
	Long:  `Get current kernel parameters configured in sabakan.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		well.Go(func(ctx context.Context) error {
			params, err := api.KernelParamsGet(ctx, kernelParamsOS)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), params)
			return nil
		})
		well.Stop()
		return well.Wait()
	},
}
var kernelParamsSetCmd = &cobra.Command{
	Use:   "set PARAMS",
	Short: "set new kernel parameters",
	Long:  `Set new kernel parameters to sabakan.`,
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		params := args[0]
		well.Go(func(ctx context.Context) error {
			return api.KernelParamsSet(ctx, kernelParamsOS, sabakan.KernelParams(params))
		})
		well.Stop()
		return well.Wait()
	},
}

func init() {
	kernelParamsCmd.Flags().StringVar(&kernelParamsOS, "os", "coreos", "OS identifier")

	kernelParamsCmd.AddCommand(kernelParamsGetCmd)
	kernelParamsCmd.AddCommand(kernelParamsSetCmd)
	rootCmd.AddCommand(kernelParamsCmd)
}
