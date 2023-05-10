package cli

// DONTCOVER
// client is excluded from test coverage in the poc phase milestone 1 and will be included in milestone 2 with completeness

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

// GetTxCmd returns a root CLI command handler for all x/liquidity transaction commands.
func GetTxCmd() *cobra.Command {
	liquidityTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Liquidity transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	liquidityTxCmd.AddCommand(
		NewCreatePoolCmd(),
		NewDepositWithinBatchCmd(),
		NewWithdrawWithinBatchCmd(),
		NewSwapWithinBatchCmd(),
	)

	return liquidityTxCmd
}

// Create new liquidity pool with the specified pool type and deposit coins.
func NewCreatePoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [pool-type] [deposit-coins]",
		Args:  cobra.ExactArgs(2),
		Short: "Create liquidity pool and deposit coins",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create liquidity pool and deposit coins.

Example:
$ %s tx %s create-pool 1 1000000000uatom,50000000000uusd --from mykey

This example creates a liquidity pool of pool-type 1 (two coins) and deposits 1000000000uatom and 50000000000uusd.
New liquidity pools can be created only for coin combinations that do not already exist in the network.

[pool-type]: The id of the liquidity pool-type. The only supported pool type is 1
[deposit-coins]: The amount of coins to deposit to the liquidity pool. The number of deposit coins must be 2 in pool type 1.
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			poolCreator := clientCtx.GetFromAddress()

			// Get pool type index
			poolTypeID, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("pool-type %s not a valid uint, input a valid unsigned 32-bit integer for pool-type", args[0])
			}

			// Get deposit coins
			depositCoins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			err = depositCoins.Validate()
			if err != nil {
				return err
			}

			if poolTypeID != 1 {
				return types.ErrPoolTypeNotExists
			}

			if depositCoins.Len() != 2 {
				return fmt.Errorf("the number of deposit coins must be two in pool-type 1")
			}

			msg := types.NewMsgCreatePool(poolCreator, uint32(poolTypeID), depositCoins)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Deposit coins to the specified liquidity pool.
func NewDepositWithinBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [pool-id] [deposit-coins]",
		Args:  cobra.ExactArgs(2),
		Short: "Deposit coins to a liquidity pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Deposit coins a liquidity pool.

This deposit request is not processed immediately since it is accumulated in the liquidity pool batch.
All requests in a batch are treated equally and executed at the same swap price.

Example:
$ %s tx %s deposit 1 100000000uatom,5000000000uusd --from mykey

This example request deposits 100000000uatom and 5000000000uusd to pool-id 1.
Deposits must be the same coin denoms as the reserve coins.

[pool-id]: The pool id of the liquidity pool
[deposit-coins]: The amount of coins to deposit to the liquidity pool
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			depositor := clientCtx.GetFromAddress()

			// Get pool type index
			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("pool-id %s not a valid uint, input a valid unsigned 32-bit integer for pool-id", args[0])
			}

			// Get deposit coins
			depositCoins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			err = depositCoins.Validate()
			if err != nil {
				return err
			}

			if depositCoins.Len() != 2 {
				return fmt.Errorf("the number of deposit coins must be two in the pool-type 1")
			}

			msg := types.NewMsgDepositWithinBatch(depositor, poolID, depositCoins)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Withdraw pool coin from the specified liquidity pool.
func NewWithdrawWithinBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [pool-id] [pool-coin]",
		Args:  cobra.ExactArgs(2),
		Short: "Withdraw pool coin from the specified liquidity pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw pool coin from the specified liquidity pool.

This swap request is not processed immediately since it is accumulated in the liquidity pool batch.
All requests in a batch are treated equally and executed at the same swap price.

Example:
$ %s tx %s withdraw 1 10000pool96EF6EA6E5AC828ED87E8D07E7AE2A8180570ADD212117B2DA6F0B75D17A6295 --from mykey

This example request withdraws 10000 pool coin from the specified liquidity pool.
The appropriate pool coin must be requested from the specified pool.

[pool-id]: The pool id of the liquidity pool
[pool-coin]: The amount of pool coin to withdraw from the liquidity pool
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			withdrawer := clientCtx.GetFromAddress()

			// Get pool type index
			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("pool-id %s not a valid uint, input a valid unsigned 32-bit integer for pool-id", args[0])
			}

			// Get pool coin of the target pool
			poolCoin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			err = poolCoin.Validate()
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawWithinBatch(withdrawer, poolID, poolCoin)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// Swap offer coin with demand coin from the specified liquidity pool with the given order price.
func NewSwapWithinBatchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap [pool-id] [swap-type] [offer-coin] [demand-coin-denom] [order-price] [swap-fee-rate]",
		Args:  cobra.ExactArgs(6),
		Short: "Swap offer coin with demand coin from the liquidity pool with the given order price",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Swap offer coin with demand coin from the liquidity pool with the given order price.

This swap request is not processed immediately since it is accumulated in the liquidity pool batch.
All requests in a batch are treated equally and executed at the same swap price.
The order of swap requests is ignored since the universal swap price is calculated in every batch to prevent front running.

The requested swap is executed with a swap price that is calculated from the given swap price function of the pool, the other swap requests, and the liquidity pool coin reserve status.
Swap orders are executed only when the execution swap price is equal to or greater than the submitted order price of the swap order.

Example:
$ %s tx %s swap 1 1 50000000uusd uatom 0.019 0.003 --from mykey

For this example, imagine that an existing liquidity pool has with 1000000000uatom and 50000000000uusd.
This example request swaps 50000000uusd for at least 950000uatom with the order price of 0.019 and swap fee rate of 0.003.
A sufficient balance of half of the swap-fee-rate of the offer coin is required to reserve the offer coin fee.

The order price is the exchange ratio of X/Y, where X is the amount of the first coin and Y is the amount of the second coin when their denoms are sorted alphabetically.
Increasing order price reduces the possibility for your request to be processed and results in buying uatom at a lower price than the pool price.

For explicit calculations, The swap fee rate must be the value that set as liquidity parameter in the current network.
The only supported swap-type is 1. For the detailed swap algorithm, see https://github.com/cosmos/gaia/v9

[pool-id]: The pool id of the liquidity pool 
[swap-type]: The swap type of the swap message. The only supported swap type is 1 (instant swap).
[offer-coin]: The amount of offer coin to swap 
[demand-coin-denom]: The denomination of the coin to exchange with offer coin 
[order-price]: The limit order price for the swap order. The price is the exchange ratio of X/Y where X is the amount of the first coin and Y is the amount of the second coin when their denoms are sorted alphabetically 
[swap-fee-rate]: The swap fee rate to pay for swap that is proportional to swap amount. The swap fee rate must be the value that set as liquidity parameter in the current network.
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			swapRequester := clientCtx.GetFromAddress()

			// Get pool id
			poolID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("pool-id %s not a valid uint, input a valid unsigned 32-bit integer for pool-id", args[0])
			}

			// Get swap type
			swapTypeID, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("swap-type %s not a valid uint, input a valid unsigned 32-bit integer for swap-type", args[2])
			}

			if swapTypeID != 1 {
				return types.ErrSwapTypeNotExists
			}

			// Get offer coin
			offerCoin, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			err = offerCoin.Validate()
			if err != nil {
				return err
			}

			err = sdk.ValidateDenom(args[3])
			if err != nil {
				return err
			}

			orderPrice, err := sdk.NewDecFromStr(args[4])
			if err != nil {
				return err
			}

			swapFeeRate, err := sdk.NewDecFromStr(args[5])
			if err != nil {
				return err
			}

			msg := types.NewMsgSwapWithinBatch(swapRequester, poolID, uint32(swapTypeID), offerCoin, args[3], orderPrice, swapFeeRate)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
