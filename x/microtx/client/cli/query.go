package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/althea-net/althea-chain/x/microtx/types"
)

const (
	FlagOrder     = "order"
	FlagClaimType = "claim-type"
	FlagNonce     = "nonce"
	FlagEthHeight = "eth-height"
	FlagUseV1Key  = "use-v1-key"
)

// GetQueryCmd bundles all the query subcmds together so they appear under `gravity query` or `gravity q`
func GetQueryCmd() *cobra.Command {
	// nolint: exhaustruct
	microtxQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the gravity module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	microtxQueryCmd.AddCommand([]*cobra.Command{
		CmdQueryData(),
	}...)

	return microtxQueryCmd
}

// CmdQueryData is an example CLI interface for users to query the data endpoint
func CmdQueryData() *cobra.Command {
	// nolint: exhaustruct
	cmd := &cobra.Command{
		Use:   "data [arg0]",
		Short: "Query data",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryData{Field: args[0]}

			res, err := queryClient.Data(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
