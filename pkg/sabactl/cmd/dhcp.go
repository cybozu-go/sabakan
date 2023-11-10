package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/cybozu-go/sabakan/v3"
	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var dhcpConfigFile string

var dhcpCmd = &cobra.Command{
	Use:   "dhcp",
	Short: "manage DHCP configurations",
	Long:  `Get and set DHCP configurations in sabakan.`,
	RunE:  dummyRunFunc,
}

var dhcpGetCmd = &cobra.Command{
	Use:   "get",
	Short: "dump current DHCP configuration",
	Long:  `Dump current DHCP configuration in sabakan.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		well.Go(func(ctx context.Context) error {
			conf, err := httpApi.DHCPConfigGet(ctx)
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
var dhcpSetCmd = &cobra.Command{
	Use:   "set -f FILE",
	Short: "update DHCP configuration",
	Long:  `Update DHCP configuration from FILE.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(dhcpConfigFile)
		if err != nil {
			return err
		}
		defer f.Close()

		var conf sabakan.DHCPConfig
		err = json.NewDecoder(f).Decode(&conf)
		if err != nil {
			return err
		}

		well.Go(func(ctx context.Context) error {
			return httpApi.DHCPConfigSet(ctx, &conf)
		})
		well.Stop()
		return well.Wait()
	},
}

func init() {
	dhcpSetCmd.Flags().StringVarP(&dhcpConfigFile, "file", "f", "", "DHCP configuration in json")
	dhcpSetCmd.MarkFlagRequired("file")

	dhcpCmd.AddCommand(dhcpGetCmd)
	dhcpCmd.AddCommand(dhcpSetCmd)
	rootCmd.AddCommand(dhcpCmd)
}
