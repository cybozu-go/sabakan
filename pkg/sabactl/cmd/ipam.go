package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/cybozu-go/sabakan/v2"
	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var ipamConfigFile string

var ipamCmd = &cobra.Command{
	Use:   "ipam",
	Short: "managa IPAM configurations",
	Long:  `Get and set IPAM configurations.`,
	RunE:  dummyRunFunc,
}

var ipamGetCmd = &cobra.Command{
	Use:   "get",
	Short: "dump current IPAM configuration",
	Long:  `Dump current IPAM configuration in sabakan.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		well.Go(func(ctx context.Context) error {
			conf, err := api.IPAMConfigGet(ctx)
			if err != nil {
				return err
			}
			e := json.NewEncoder(cmd.OutOrStdout())
			e.SetIndent("", "  ")
			return e.Encode(conf)
		})
		well.Stop()
		return well.Wait()
	},
}

var ipamSetCmd = &cobra.Command{
	Use:   "set -f FILE",
	Short: "update IPAM configuration",
	Long:  `Update IPAM configuration from FILE.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(ipamConfigFile)
		if err != nil {
			return err
		}
		defer f.Close()

		var conf sabakan.IPAMConfig
		err = json.NewDecoder(f).Decode(&conf)
		if err != nil {
			return err
		}

		well.Go(func(ctx context.Context) error {
			return api.IPAMConfigSet(ctx, &conf)
		})
		well.Stop()
		return well.Wait()
	},
}

func init() {
	ipamSetCmd.Flags().StringVarP(&ipamConfigFile, "file", "f", "", "IPAM configuration in json")
	ipamSetCmd.MarkFlagRequired("file")

	ipamCmd.AddCommand(ipamGetCmd)
	ipamCmd.AddCommand(ipamSetCmd)
	rootCmd.AddCommand(ipamCmd)
}
