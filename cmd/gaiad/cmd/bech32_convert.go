package cmd

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/spf13/cobra"
)

var flagBech32Prefix = "prefix"

// AddBech32ConvertCommand returns bech32-convert cobra Command.
func AddBech32ConvertCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bech32-convert [bech32 string]",
		Short: "Convert any bech32 string to the cosmos prefix",
		Long: `Convert any bech32 string to the cosmos prefix

Example:
	gaiad debug bech32-convert akash1a6zlyvpnksx8wr6wz8wemur2xe8zyh0ytz6d88
	`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bech32prefix, err := cmd.Flags().GetString(flagBech32Prefix)
			if err != nil {
				return err
			}

			_, bz, err := bech32.DecodeAndConvert(args[0])
			if err != nil {
				return fmt.Errorf("cannot decode address", err)
			}

			bech32Addr, err := bech32.ConvertAndEncode(bech32prefix, bz)
			if err != nil {
				return fmt.Errorf("cannot convert address", err)
			}

			cmd.Println(bech32Addr)

			return nil
		},
	}

	cmd.Flags().StringP(flagBech32Prefix, "p", "cosmos", "Bech32 Prefix to encode to")

	return cmd
}

// injectConvertBech32Command injects bech32-convert command into another command as a child.
func injectConvertBech32Command(cmd *cobra.Command) *cobra.Command {
	cmd.AddCommand(AddBech32ConvertCommand())
	return cmd
}
