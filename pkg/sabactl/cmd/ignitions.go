package cmd

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/cybozu-go/sabakan/client"
	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var (
	ignitionsSetFile string
	ignitionsSetMeta map[string]string
)

var ignitionsCmd = &cobra.Command{
	Use:   "ignitions",
	Short: "manage ignitions",
	Long:  `Get and update ignitions in sabakan.`,
}

var ignitionsGetCmd = &cobra.Command{
	Use:   "get ROLE",
	Short: "get IDs of the ignitions",
	Long:  `Get IDs of the ignitions registered for ROLE from sabakan.`,
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		role := args[0]
		well.Go(func(ctx context.Context) error {
			index, err := api.IgnitionsGet(ctx, role)
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(index)
		})
		well.Stop()
		return well.Wait()
	},
}

var ignitionsSetCmd = &cobra.Command{
	Use:   "set ROLE ID",
	Short: "add a new ignition or update current ignition",
	Long:  `Add a new ignition or update current ignition by ROLE and ID.`,
	Args:  cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		role, id := args[0], args[1]

		var buf bytes.Buffer
		err := client.AssembleIgnitionTemplate(ignitionsSetFile, &buf)
		if err != nil {
			return err
		}

		well.Go(func(ctx context.Context) error {
			return api.IgnitionsSet(ctx, role, id, &buf, ignitionsSetMeta)
		})
		well.Stop()
		return well.Wait()
	},
}

var ignitionsCatCmd = &cobra.Command{
	Use:   "cat ROLE ID",
	Short: "show the ignition template",
	Long:  `Show the ignition template for the ID and ROLE.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		role, id := args[0], args[1]
		well.Go(func(ctx context.Context) error {
			return api.IgnitionsCat(ctx, role, id, cmd.OutOrStdout())
		})
		well.Stop()
		return well.Wait()
	},
}

var ignitionsDeleteCmd = &cobra.Command{
	Use:   "delete ROLE ID",
	Short: "delete an ignition template",
	Long:  `Delete an ignition template for the ID and ROLE.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		role, id := args[0], args[1]
		well.Go(func(ctx context.Context) error {
			return api.IgnitionsDelete(ctx, role, id)
		})
		well.Stop()
		return well.Wait()
	},
}

func init() {
	ignitionsSetCmd.Flags().StringVarP(&ignitionsSetFile, "file", "f", "", "ignition entry point in yaml")
	ignitionsSetCmd.Flags().StringToStringVar(&ignitionsSetMeta, "meta", nil, "additional metadata for the ignitions as <KEY1>=<VALUE1>,<KEY2>=<VALUE2>,...")
	ignitionsSetCmd.MarkFlagRequired("file")

	ignitionsCmd.AddCommand(ignitionsGetCmd)
	ignitionsCmd.AddCommand(ignitionsSetCmd)
	ignitionsCmd.AddCommand(ignitionsCatCmd)
	ignitionsCmd.AddCommand(ignitionsDeleteCmd)
	rootCmd.AddCommand(ignitionsCmd)
}
