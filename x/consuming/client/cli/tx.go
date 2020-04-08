package cli

import (
	"fmt"
	"strconv"
	"strings"

	bandoracle "github.com/bandprotocol/chain/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/bandprotocol/band-consumer/x/consuming/types"
)

const (
	flagName     = "name"
	flagCalldata = "calldata"
	flagAskCount = "ask-count"
	flagMinCount = "min-count"
	flagChannel  = "channel"
)

// NewTxCmd returns the transaction commands for this module
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "consuming transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(NewRequestTxCmd())

	return txCmd
}

// NewRequestTxCmd implements the request command handler.
func NewRequestTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request [oracle-script-id] (-c [calldata]) (-r [requested-validator-count]) (-v [sufficient-validator-count]) (-x [expiration]) (-w [prepare-gas]) (-g [execute-gas])",
		Short: "Make a new data request via an existing oracle script",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Make a new request via an existing oracle script with the configuration flags.
Example:
$ %s tx consuming request 1 -c 1234abcdef -r 4 -v 3 -x 20 -w 50 -g 5000 --from mykey
$ %s tx consuming request 1 --calldata 1234abcdef --requested-validator-count 4 --sufficient-validator-count 3 --expiration 20 --prepare-gas 50 --execute-gas 5000 --from mykey
`,
				version.AppName, version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			int64OracleScriptID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			oracleScriptID := bandoracle.OracleScriptID(int64OracleScriptID)

			calldata, err := cmd.Flags().GetBytesHex(flagCalldata)
			if err != nil {
				return err
			}

			requestedValidatorCount, err := cmd.Flags().GetInt64(flagAskCount)
			if err != nil {
				return err
			}

			sufficientValidatorCount, err := cmd.Flags().GetInt64(flagMinCount)
			if err != nil {
				return err
			}

			channel, err := cmd.Flags().GetString(flagChannel)
			if err != nil {
				return err
			}

			msg := types.NewMsgRequestData(
				oracleScriptID,
				channel,
				calldata,
				requestedValidatorCount,
				sufficientValidatorCount,
				clientCtx.GetFromAddress(),
			)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().BytesHexP(flagCalldata, "c", nil, "Calldata used in calling the oracle script")
	cmd.Flags().Int64P(flagAskCount, "r", 0, "Number of top validators that need to report data for this request")
	cmd.MarkFlagRequired(flagAskCount)
	cmd.Flags().Int64P(flagMinCount, "v", 0, "Minimum number of reports sufficient to conclude the request's result")
	cmd.MarkFlagRequired(flagMinCount)
	cmd.Flags().String(flagChannel, "", "The channel id.")
	cmd.MarkFlagRequired(flagChannel)

	return cmd
}
