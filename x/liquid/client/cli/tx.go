package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"

	"github.com/cosmos/gaia/v27/x/liquid/types"
)

// NewTxCmd returns a root CLI command handler for all x/liquid transaction commands.
func NewTxCmd(valAddrCodec, ac address.Codec) *cobra.Command {
	liquidTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Liquid transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	liquidTxCmd.AddCommand(
		NewTokenizeSharesCmd(valAddrCodec, ac),
		NewRedeemTokensCmd(),
		NewTransferTokenizeShareRecordCmd(ac),
		NewDisableTokenizeShares(),
		NewEnableTokenizeShares(),
		NewWithdrawTokenizeShareRecordRewardCmd(ac),
		NewWithdrawAllTokenizeShareRecordRewardCmd(ac),
	)

	return liquidTxCmd
}

// NewTokenizeSharesCmd defines a command for tokenizing shares from a validator.
func NewTokenizeSharesCmd(valAddrCodec, ac address.Codec) *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()

	cmd := &cobra.Command{
		Use:   "tokenize-share [validator-addr] [amount] [rewardOwner]",
		Short: "Tokenize delegation to share tokens",
		Args:  cobra.ExactArgs(3),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Tokenize delegation to share tokens.

Example:
$ %s tx liquid tokenize-share %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 100stake %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj --from mykey
`,
				version.AppName, bech32PrefixValAddr, bech32PrefixAccAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			delAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			valAddr := args[0]
			if _, err = valAddrCodec.StringToBytes(valAddr); err != nil {
				return err
			}

			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			_, err = ac.StringToBytes(args[2])
			if err != nil {
				return err
			}

			msg := &types.MsgTokenizeShares{
				DelegatorAddress:    delAddr,
				ValidatorAddress:    valAddr,
				Amount:              amount,
				TokenizedShareOwner: args[2],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewRedeemTokensCmd defines a command for redeeming tokens from a validator for shares.
func NewRedeemTokensCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redeem-tokens [amount]",
		Short: "Redeem specified amount of share tokens to delegation",
		Args:  cobra.ExactArgs(1),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Redeem specified amount of share tokens to delegation.

Example:
$ %s tx liquid redeem-tokens 100sharetoken --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			delAddr := clientCtx.GetFromAddress()

			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return err
			}

			msg := &types.MsgRedeemTokensForShares{
				DelegatorAddress: delAddr.String(),
				Amount:           amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewTransferTokenizeShareRecordCmd defines a command to transfer ownership of TokenizeShareRecord
func NewTransferTokenizeShareRecordCmd(ac address.Codec) *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()

	cmd := &cobra.Command{
		Use:   "transfer-tokenize-share-record [record-id] [new-owner]",
		Short: "Transfer ownership of TokenizeShareRecord",
		Args:  cobra.ExactArgs(2),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Transfer ownership of TokenizeShareRecord.

Example:
$ %s tx liquid transfer-tokenize-share-record 1 %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj --from mykey
`,
				version.AppName, bech32PrefixAccAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			recordID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			ownerAddr := args[1]
			_, err = ac.StringToBytes(ownerAddr)
			if err != nil {
				return err
			}

			msg := &types.MsgTransferTokenizeShareRecord{
				Sender:                clientCtx.GetFromAddress().String(),
				TokenizeShareRecordId: recordID,
				NewOwner:              ownerAddr,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewDisableTokenizeShares defines a command to disable tokenization for an address
func NewDisableTokenizeShares() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable-tokenize-shares",
		Short: "Disable tokenization of shares",
		Args:  cobra.ExactArgs(0),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Disables the tokenization of shares for an address. The account
must explicitly re-enable if they wish to tokenize again, at which point they must wait 
the chain's unbonding period. 

Example:
$ %s tx liquid disable-tokenize-shares --from mykey
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgDisableTokenizeShares{
				DelegatorAddress: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewEnableTokenizeShares defines a command to re-enable tokenization for an address
func NewEnableTokenizeShares() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable-tokenize-shares",
		Short: "Enable tokenization of shares",
		Args:  cobra.ExactArgs(0),
		Long: strings.TrimSpace(
			fmt.Sprintf(`Enables the tokenization of shares for an address after 
it had been disable. This transaction queues the enablement of tokenization, but
the address must wait 1 unbonding period from the time of this transaction before
tokenization is permitted.

Example:
$ %s tx liquid enable-tokenize-shares --from mykey
`, version.AppName),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgEnableTokenizeShares{
				DelegatorAddress: clientCtx.GetFromAddress().String(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// WithdrawAllTokenizeShareRecordReward defines a method to withdraw reward for all owning TokenizeShareRecord
func NewWithdrawAllTokenizeShareRecordRewardCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-all-tokenize-share-rewards",
		Args:  cobra.ExactArgs(0),
		Short: "Withdraw reward for all owning TokenizeShareRecord",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw reward for all owned TokenizeShareRecord
Example:
$ %s tx distribution withdraw-all-tokenize-share-rewards --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ownerAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawAllTokenizeShareRecordReward(ownerAddr)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// WithdrawTokenizeShareRecordReward defines a method to withdraw reward for an owning TokenizeShareRecord
func NewWithdrawTokenizeShareRecordRewardCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-tokenize-share-rewards [record-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Withdraw reward for an owning TokenizeShareRecord",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw reward for an owned TokenizeShareRecord
Example:
$ %s tx distribution withdraw-tokenize-share-rewards 1 --from mykey
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			ownerAddr, err := ac.BytesToString(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			recordID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawTokenizeShareRecordReward(ownerAddr, recordID)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
