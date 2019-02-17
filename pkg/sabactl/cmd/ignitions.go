package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/cybozu-go/sabakan/v2/client"
	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var ignSetOpts struct {
	filename string
	metafile string
	json     bool
}

var ignitionsCmd = &cobra.Command{
	Use:   "ignitions",
	Short: "manage ignitions",
	Long:  `Get and update ignitions in sabakan.`,
}

var ignitionsGetCmd = &cobra.Command{
	Use:   "get ROLE [ID]",
	Short: "get an ignition template / template IDs",
	Long: `If ID is not given, this command retrieves available template IDs for ROLE.
If ID is given, this command outputs the ignition template in JSON format.`,
	Args: cobra.RangeArgs(1, 2),

	RunE: func(cmd *cobra.Command, args []string) error {
		role := args[0]
		well.Go(func(ctx context.Context) error {
			if len(args) == 1 {
				ids, err := api.IgnitionsListIDs(ctx, role)
				if err != nil {
					return err
				}
				return json.NewEncoder(cmd.OutOrStdout()).Encode(ids)
			}

			tmpl, err := api.IgnitionsGet(ctx, role, args[1])
			if err != nil {
				return err
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(tmpl)
		})
		well.Stop()
		return well.Wait()
	},
}

var ignitionsSetCmd = &cobra.Command{
	Use:   "set -f FILENAME ROLE ID",
	Short: "add a new ignition template",
	Long: `Add a new ignition template for ROLE.
ID must be a valid version string conforming to semver.

If --json is specified, raw JSON template is read from FILENAME.
If not, FILENAME should be a YAML file like this:

	version: 2.3
	include: ../common/common.yml
	passwd: passwd.yml
	files:
	  - /etc/hostname
	networkd:
	  - 10-eth0.network
	systemd:
	  - name: foo.service
		enable: true
`,
	Args: cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		role, id := args[0], args[1]

		var tmpl *client.IgnitionTemplate

		if ignSetOpts.json {
			f, err := os.Open(ignSetOpts.filename)
			if err != nil {
				return err
			}
			defer f.Close()

			rawTemplate := &client.IgnitionTemplate{}
			err = json.NewDecoder(f).Decode(&rawTemplate)
			if err != nil {
				return err
			}
			tmpl = rawTemplate
		} else {
			var metadata map[string]interface{}
			if ignSetOpts.metafile != "" {
				f, err := os.Open(ignSetOpts.metafile)
				if err != nil {
					return err
				}
				defer f.Close()

				err = json.NewDecoder(f).Decode(&metadata)
				if err != nil {
					return err
				}
			}
			var err error
			tmpl, err = client.BuildIgnitionTemplate(ignSetOpts.filename, metadata)
			if err != nil {
				return err
			}
		}

		well.Go(func(ctx context.Context) error {
			return api.IgnitionsSet(ctx, role, id, tmpl)
		})
		well.Stop()
		return well.Wait()
	},
}

var ignitionsDeleteCmd = &cobra.Command{
	Use:   "delete ROLE ID",
	Short: "delete an ignition template",
	Long:  `Delete an ignition template for the ID and ROLE.`,
	Args:  cobra.ExactArgs(2),

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
	ignitionsSetCmd.Flags().StringVarP(&ignSetOpts.filename, "file", "f", "", "ignition template filename")
	ignitionsSetCmd.Flags().StringVar(&ignSetOpts.metafile, "meta", "", "JSON file containing meta data")
	ignitionsSetCmd.Flags().BoolVar(&ignSetOpts.json, "json", false, "read raw JSON template")
	ignitionsSetCmd.MarkFlagRequired("file")

	ignitionsCmd.AddCommand(ignitionsGetCmd)
	ignitionsCmd.AddCommand(ignitionsSetCmd)
	ignitionsCmd.AddCommand(ignitionsDeleteCmd)
	rootCmd.AddCommand(ignitionsCmd)
}
