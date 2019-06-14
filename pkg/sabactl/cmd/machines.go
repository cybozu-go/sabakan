package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cybozu-go/sabakan/v2"
	"github.com/cybozu-go/well"
	"github.com/spf13/cobra"
)

var (
	machinesGetParams  = make(map[string]*string)
	machinesCreateFile string
)

var machinesCmd = &cobra.Command{
	Use:   "machines action",
	Short: "manage machines",
	Long:  `Manage machines registered in sabakan.`,
}

var machinesGetCmd = &cobra.Command{
	Use:   "get [options]",
	Short: "get machines from sabakan",
	Long:  `Get machines from sabakan.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		params := make(map[string]string)
		for k, v := range machinesGetParams {
			params[k] = *v
		}
		well.Go(func(ctx context.Context) error {
			ms, err := api.MachinesGet(ctx, params)
			if err != nil {
				return err
			}
			e := json.NewEncoder(cmd.OutOrStdout())
			e.SetIndent("", "  ")
			return e.Encode(ms)
		})
		well.Stop()
		return well.Wait()
	},
}

var machinesCreateCmd = &cobra.Command{
	Use:   "create -f FILE",
	Short: "create a new machines",
	Long:  `Create a new machines to sabakan.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		well.Go(func(ctx context.Context) error {
			f, err := os.Open(machinesCreateFile)
			if err != nil {
				return err
			}
			defer f.Close()

			var specs []*sabakan.MachineSpec
			err = json.NewDecoder(f).Decode(&specs)
			if err != nil {
				return err
			}
			return api.MachinesCreate(ctx, specs)
		})
		well.Stop()
		return well.Wait()
	},
}

var machinesRemoveCmd = &cobra.Command{
	Use:   "remove SERIAL",
	Short: "remove registered machine",
	Long:  `Remove registered machine by SERIAL from sabakan.`,
	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		serial := args[0]
		well.Go(func(ctx context.Context) error {
			return api.MachinesRemove(ctx, serial)
		})
		well.Stop()
		return well.Wait()
	},
}

var machinesGetStateCmd = &cobra.Command{
	Use:   "get-state SERIAL",
	Short: "get current state of the machine",
	Long:  `Get current state of the machine by SERIAL from sabakan.`,

	Args: cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		serial := args[0]
		well.Go(func(ctx context.Context) error {
			state, err := api.MachinesGetState(ctx, serial)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), state.String())
			return nil
		})
		well.Stop()
		return well.Wait()
	},
}

var machinesSetStateCmd = &cobra.Command{
	Use:   "set-state SERIAL STATE",
	Short: "update current state of the machine",
	Long: `Update current state of the machine by SERIAL to STATE.
STATE can be one of:
    uninitialized  The machine is not yet initialized.
    healthy        The machine has no problems.
    unhealthy      The machine has some problems.
    unreachable    The machine does not communicate with others.
    updating       The machine is updating.
    retiring       The machine should soon be retired/repaired.
    retired        The machine's disk encryption keys were deleted.
	`,
	Args: cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		serial, state := args[0], strings.ToLower(args[1])
		well.Go(func(ctx context.Context) error {
			return api.MachinesSetState(ctx, serial, state)
		})
		well.Stop()
		return well.Wait()
	},
}

var machinesGetLabelCmd = &cobra.Command{
	Use:   "get-label SERIAL NAME",
	Short: "get the label value of the machine",
	Long:  `Get the value of the label named NAME of the machine.`,
	Args:  cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		serial, label := args[0], args[1]
		well.Go(func(ctx context.Context) error {
			machines, err := api.MachinesGet(ctx, map[string]string{
				"serial": serial,
			})
			if err != nil {
				return err
			}
			if len(machines) == 0 {
				return errors.New("machine not found")
			} else if len(machines) > 1 {
				return errors.New("too many machines found")
			}

			value, ok := machines[0].Spec.Labels[label]
			if !ok {
				return errors.New("label not found")
			}
			fmt.Fprintln(cmd.OutOrStdout(), value)
			return nil
		})
		well.Stop()
		return well.Wait()
	},
}

var machinesSetLabelCmd = &cobra.Command{
	Use:   "set-label SERIAL NAME VALUE",
	Short: "add or update a label for the machine",
	Long:  `Add or update a label of "NAME: VALUE" for the machine.`,
	Args:  cobra.ExactArgs(3),

	RunE: func(cmd *cobra.Command, args []string) error {
		serial, label, value := args[0], args[1], args[2]
		well.Go(func(ctx context.Context) error {
			return api.MachinesSetLabel(ctx, serial, label, value)
		})
		well.Stop()
		return well.Wait()
	},
}

var machinesRemoveLabelCmd = &cobra.Command{
	Use:   "remove-label SERIAL NAME",
	Short: "remove a label from the machine",
	Long:  `Remove a label named NAME from the machine.`,
	Args:  cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		serial, label := args[0], args[1]
		well.Go(func(ctx context.Context) error {
			return api.MachinesRemoveLabel(ctx, serial, label)
		})
		well.Stop()
		return well.Wait()
	},
}

var machinesSetRetireDateCmd = &cobra.Command{
	Use:   "set-retire-date SERIAL YYYY-MM-DD",
	Short: "set the retire date of the machine",
	Long:  `Set the retire date of the machine by SERIAL.`,
	Args:  cobra.ExactArgs(2),

	RunE: func(cmd *cobra.Command, args []string) error {
		serial := args[0]
		date, err := time.Parse("2006-01-02", args[1])
		if err != nil {
			return err
		}
		well.Go(func(ctx context.Context) error {
			return api.MachinesSetRetireDate(ctx, serial, date)
		})
		well.Stop()
		return well.Wait()
	},
}

func init() {
	getOpts := map[string]string{
		"serial":   "Serial name",
		"rack":     "Rack name",
		"role":     "Role name",
		"labels":   "Label name and value (--labels key=val,...)",
		"ipv4":     "IPv4 address",
		"ipv6":     "IPv6 address",
		"bmc-type": "BMC type",
		"state":    "State",
	}
	for k, v := range getOpts {
		val := new(string)
		machinesGetParams[k] = val
		machinesGetCmd.Flags().StringVar(val, k, "", v)
	}
	machinesCreateCmd.Flags().StringVarP(&machinesCreateFile, "file", "f", "", "machiens in json")
	machinesCreateCmd.MarkFlagRequired("file")

	machinesCmd.AddCommand(machinesGetCmd)
	machinesCmd.AddCommand(machinesCreateCmd)
	machinesCmd.AddCommand(machinesRemoveCmd)
	machinesCmd.AddCommand(machinesGetStateCmd)
	machinesCmd.AddCommand(machinesSetStateCmd)
	machinesCmd.AddCommand(machinesGetLabelCmd)
	machinesCmd.AddCommand(machinesSetLabelCmd)
	machinesCmd.AddCommand(machinesRemoveLabelCmd)
	machinesCmd.AddCommand(machinesSetRetireDateCmd)
	rootCmd.AddCommand(machinesCmd)
}
