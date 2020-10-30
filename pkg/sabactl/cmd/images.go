package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var imagesOS string

var imagesCmd = &cobra.Command{
	Use:   "images",
	Short: "manage OS images",
	Long:  `Manage OS registered in sabakan.`,
	RunE:  dummyRunFunc,
}

var imagesIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "get registered images",
	Long:  `Get registered images from sabakan.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		well.Go(func(ctx context.Context) error {
			index, err := api.ImagesIndex(ctx, imagesOS)
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

var imagesUploadCmd = &cobra.Command{
	Use:   "upload ID KERNEL INITRD",
	Short: "add a new image or update current image",
	Long:  `Add a new image or update current image.`,
	Args:  cobra.ExactArgs(3),

	RunE: func(cmd *cobra.Command, args []string) error {
		id, kernelPath, initrdPath := args[0], args[1], args[2]
		kernelInfo, err := os.Stat(kernelPath)
		if err != nil {
			return err
		}
		initrdInfo, err := os.Stat(initrdPath)
		if err != nil {
			return err
		}
		kernel, err := os.Open(kernelPath)
		if err != nil {
			return err
		}
		defer kernel.Close()
		initrd, err := os.Open(initrdPath)
		if err != nil {
			return err
		}
		defer initrd.Close()

		well.Go(func(ctx context.Context) error {
			return api.ImagesUpload(ctx, imagesOS, id, kernel, kernelInfo.Size(), initrd, initrdInfo.Size())
		})
		well.Stop()
		return well.Wait()
	},
}

var imagesDeleteCmd = &cobra.Command{
	Use:   "delete ID",
	Short: "delete the image",
	Long:  `Delete the image of the ID.`,
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		well.Go(func(ctx context.Context) error {
			return api.ImagesDelete(ctx, imagesOS, id)
		})
		well.Stop()
		return well.Wait()
	},
}

func init() {
	imagesCmd.Flags().StringVar(&imagesOS, "os", "coreos", "OS identifier")

	imagesCmd.AddCommand(imagesIndexCmd)
	imagesCmd.AddCommand(imagesUploadCmd)
	imagesCmd.AddCommand(imagesDeleteCmd)
	rootCmd.AddCommand(imagesCmd)
}
