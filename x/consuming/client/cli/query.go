package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/bandprotocol/band-consumer/x/consuming/types"
)

// GetQueryCmd returns
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the consuming module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(GetCmdReadResult())

	return queryCmd
}

// GetCmdReadResult queries request by reqID
func GetCmdReadResult() *cobra.Command {
	return &cobra.Command{
		Use:  "result",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Revisit query
			// clientCtx, err := client.GetClientQueryContext(cmd)
			// if err != nil {
			// 	return err
			// }
			// queryClient := types.NewQueryClient(clientCtx)

			// reqID := args[0]

			// res, _, err := cliCtx.QueryWithData(
			// 	fmt.Sprintf("custom/%s/result/%s", queryRoute, reqID),
			// 	nil,
			// )
			// if err != nil {
			// 	fmt.Printf("read request fail - %s \n", reqID)
			// 	return nil
			// }
			// return cliCtx.PrintOutput(res)
			return nil
		},
	}
}
