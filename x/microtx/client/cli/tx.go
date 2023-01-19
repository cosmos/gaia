package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/althea-net/althea-chain/x/microtx/types"
)

// GetTxCmd bundles all the subcmds together so they appear under `gravity tx`
func GetTxCmd(storeKey string) *cobra.Command {
	// nolint: exhaustruct
	microtxTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "microtx transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	microtxTxCmd.AddCommand([]*cobra.Command{
		CmdName(),
	}...)

	return microtxTxCmd
}

// CmdName is an example CLI interface for submitting a Tx containing a MsgName
func CmdName() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "name [arg0]",
		Short: "short description",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			// cosmosAddr := cliCtx.GetFromAddress()

			// Make the message
			msg := types.NewMsgName(args[0])
			if err := msg.ValidateBasic(); err != nil {
				return sdkerrors.Wrap(err, "No Bobs allowed :P")
			}
			// Send it
			return tx.GenerateOrBroadcastTxCLI(cliCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
