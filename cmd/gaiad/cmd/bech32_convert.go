package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	addressutil "github.com/cosmos/gaia/v23/pkg/address"
)

var flagBech32Prefix = "prefix"

// AddBech32ConvertCommand returns bech32-convert cobra Command.
func AddBech32ConvertCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bech32-convert [address]",
		Short: "Convert any bech32 string to the cosmos prefix",
		Long: `Convert any bech32 string to the cosmos prefix

Example:
	gaiad debug bech32-convert akash1a6zlyvpnksx8wr6wz8wemur2xe8zyh0ytz6d88

	gaiad debug bech32-convert stride1673f0t8p893rqyqe420mgwwz92ac4qv6synvx2 --prefix osmo
	`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bech32prefix, err := cmd.Flags().GetString(flagBech32Prefix)
			if err != nil {
				return err
			}

			address := args[0]
			convertedAddress, err := addressutil.ConvertBech32Prefix(address, bech32prefix)
			if err != nil {
				return fmt.Errorf("conversation failed: %s", err)
			}

			cmd.Println(convertedAddress)

			return nil
		},
	}

	cmd.Flags().StringP(flagBech32Prefix, "p", "cosmos", "Bech32 Prefix to encode to")

	return cmd
}

// addDebugCommands injects custom debug commands into another command as children.
func addDebugCommands(cmd *cobra.Command) *cobra.Command {
	cmd.AddCommand(AddBech32ConvertCommand())
	return cmd
}
