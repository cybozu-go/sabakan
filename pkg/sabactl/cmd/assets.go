package cmd

import (
	"context"
	"encoding/json"

	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var assetsUploadMeta map[string]string

var assetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "manage assets",
	Long:  `Get, add, update and delete assets in sabakan.`,
	RunE:  dummyRunFunc,
}

var assetsIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "get index of assets",
	Long:  `Get index of assets registered in sabakan.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		well.Go(func(ctx context.Context) error {
			index, err := httpApi.AssetsIndex(ctx)
			if err != nil {
				return err
			}
			e := json.NewEncoder(cmd.OutOrStdout())
			e.SetIndent("", "  ")
			return e.Encode(index)
		})
		well.Stop()
		return well.Wait()
	},
}

var assetsInfoCmd = &cobra.Command{
	Use:   "info NAME",
	Short: "get metadata of the asset",
	Long:  `Get metadata of the asset registered in sabakan.`,
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		well.Go(func(ctx context.Context) error {
			info, err := httpApi.AssetsInfo(ctx, name)
			if err != nil {
				return err
			}

			e := json.NewEncoder(cmd.OutOrStdout())
			e.SetIndent("", "  ")
			return e.Encode(info)
		})
		well.Stop()
		return well.Wait()
	},
}

var assetsUploadCmd = &cobra.Command{
	Use:   "upload NAME FILE",
	Short: "add a new asset or update current asset",
	Long:  `Add a new asset or update current asset to sabakan.`,
	Args:  cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		name, path := args[0], args[1]
		well.Go(func(ctx context.Context) error {
			st, err := httpApi.AssetsUpload(ctx, name, path, assetsUploadMeta)
			if err != nil {
				return err
			}

			e := json.NewEncoder(cmd.OutOrStdout())
			e.SetIndent("", "  ")
			return e.Encode(st)
		})
		well.Stop()
		return well.Wait()
	},
}

var assetsDeleteCmd = &cobra.Command{
	Use:   "delete NAME",
	Short: "delete the registered asset",
	Long:  `Delete the registered asset from sabakan.`,
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		well.Go(func(ctx context.Context) error {
			return httpApi.AssetsDelete(ctx, name)
		})
		well.Stop()
		return well.Wait()
	},
}

func init() {
	assetsUploadCmd.Flags().StringToStringVar(&assetsUploadMeta, "meta", nil, "Additional metadata for the assets as <KEY1>=<VALUE1>,<KEY2>=<VALUE2>,...")

	assetsCmd.AddCommand(assetsIndexCmd)
	assetsCmd.AddCommand(assetsInfoCmd)
	assetsCmd.AddCommand(assetsUploadCmd)
	assetsCmd.AddCommand(assetsDeleteCmd)
	rootCmd.AddCommand(assetsCmd)
}
