package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

// DeleteAndInitPoolBatches resets batch msg states that were previously executed
// and deletes msg states that were marked to be deleted.
func (k Keeper) DeleteAndInitPoolBatches(ctx sdk.Context) {
	k.IterateAllPoolBatches(ctx, func(poolBatch types.PoolBatch) bool {
		// Re-initialize the executed batch.
		if poolBatch.Executed {
			// On the other hand, BatchDeposit, BatchWithdraw, is all handled by the endblock if there is no error.
			// If there are BatchMsgs left, reset the Executed, Succeeded flag so that it can be executed in the next batch.
			depositMsgs := k.GetAllRemainingPoolBatchDepositMsgStates(ctx, poolBatch)
			if len(depositMsgs) > 0 {
				for _, msg := range depositMsgs {
					msg.Executed = false
					msg.Succeeded = false
				}
				k.SetPoolBatchDepositMsgStatesByPointer(ctx, poolBatch.PoolId, depositMsgs)
			}

			withdrawMsgs := k.GetAllRemainingPoolBatchWithdrawMsgStates(ctx, poolBatch)
			if len(withdrawMsgs) > 0 {
				for _, msg := range withdrawMsgs {
					msg.Executed = false
					msg.Succeeded = false
				}
				k.SetPoolBatchWithdrawMsgStatesByPointer(ctx, poolBatch.PoolId, withdrawMsgs)
			}

			height := ctx.BlockHeight()

			// In the case of remaining swap msg states, those are either fractionally matched
			// or has not yet been expired.
			swapMsgs := k.GetAllRemainingPoolBatchSwapMsgStates(ctx, poolBatch)
			if len(swapMsgs) > 0 {
				for _, msg := range swapMsgs {
					if height > msg.OrderExpiryHeight {
						msg.ToBeDeleted = true
					} else {
						msg.Executed = false
						msg.Succeeded = false
					}
				}
				k.SetPoolBatchSwapMsgStatesByPointer(ctx, poolBatch.PoolId, swapMsgs)
			}

			// Delete all batch msg states that are ready to be deleted.
			k.DeleteAllReadyPoolBatchDepositMsgStates(ctx, poolBatch)
			k.DeleteAllReadyPoolBatchWithdrawMsgStates(ctx, poolBatch)
			k.DeleteAllReadyPoolBatchSwapMsgStates(ctx, poolBatch)

			if err := k.InitNextPoolBatch(ctx, poolBatch); err != nil {
				panic(err)
			}
		}
		return false
	})
}

// InitNextPoolBatch re-initializes the batch and increases the batch index.
func (k Keeper) InitNextPoolBatch(ctx sdk.Context, poolBatch types.PoolBatch) error {
	if !poolBatch.Executed {
		return types.ErrBatchNotExecuted
	}

	poolBatch.Index++
	poolBatch.BeginHeight = ctx.BlockHeight()
	poolBatch.Executed = false

	k.SetPoolBatch(ctx, poolBatch)
	return nil
}

// ExecutePoolBatches executes the accumulated msgs in the batch.
// The order is (1)swap, (2)deposit, (3)withdraw.
func (k Keeper) ExecutePoolBatches(ctx sdk.Context) {
	params := k.GetParams(ctx)
	logger := k.Logger(ctx)

	k.IterateAllPoolBatches(ctx, func(poolBatch types.PoolBatch) bool {
		if !poolBatch.Executed && ctx.BlockHeight()%int64(params.UnitBatchHeight) == 0 {
			executedMsgCount, err := k.SwapExecution(ctx, poolBatch)
			if err != nil {
				panic(err)
			}

			k.IterateAllPoolBatchDepositMsgStates(ctx, poolBatch, func(batchMsg types.DepositMsgState) bool {
				if batchMsg.Executed || batchMsg.ToBeDeleted || batchMsg.Succeeded {
					return false
				}
				executedMsgCount++
				if err := k.ExecuteDeposit(ctx, batchMsg, poolBatch); err != nil {
					logger.Error("deposit failed",
						"poolID", poolBatch.PoolId,
						"batchIndex", poolBatch.Index,
						"msgIndex", batchMsg.MsgIndex,
						"depositor", batchMsg.Msg.GetDepositor(),
						"error", err)
					if err := k.RefundDeposit(ctx, batchMsg, poolBatch); err != nil {
						panic(err)
					}
				}
				return false
			})

			k.IterateAllPoolBatchWithdrawMsgStates(ctx, poolBatch, func(batchMsg types.WithdrawMsgState) bool {
				if batchMsg.Executed || batchMsg.ToBeDeleted || batchMsg.Succeeded {
					return false
				}
				executedMsgCount++
				if err := k.ExecuteWithdrawal(ctx, batchMsg, poolBatch); err != nil {
					logger.Error("withdraw failed",
						"poolID", poolBatch.PoolId,
						"batchIndex", poolBatch.Index,
						"msgIndex", batchMsg.MsgIndex,
						"withdrawer", batchMsg.Msg.GetWithdrawer(),
						"error", err)
					if err := k.RefundWithdrawal(ctx, batchMsg, poolBatch); err != nil {
						panic(err)
					}
				}
				return false
			})

			// Mark the batch as executed when any msgs were executed.
			if executedMsgCount > 0 {
				poolBatch.Executed = true
				k.SetPoolBatch(ctx, poolBatch)
			}
		}
		return false
	})
}

// HoldEscrow sends coins to the module account for an escrow.
func (k Keeper) HoldEscrow(ctx sdk.Context, depositor sdk.AccAddress, depositCoins sdk.Coins) error {
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleName, depositCoins)
	return err
}

// If batch messages have expired or have not been processed, coins that were deposited with this function are refunded to the escrow.
func (k Keeper) ReleaseEscrow(ctx sdk.Context, withdrawer sdk.AccAddress, withdrawCoins sdk.Coins) error {
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, withdrawer, withdrawCoins)

	return err
}

// Generate inputs and outputs to treat escrow refunds atomically.
func (k Keeper) ReleaseEscrowForMultiSend(withdrawer sdk.AccAddress, withdrawCoins sdk.Coins) (
	banktypes.Input, banktypes.Output, error,
) {
	var input banktypes.Input
	var output banktypes.Output

	input = banktypes.NewInput(k.accountKeeper.GetModuleAddress(types.ModuleName), withdrawCoins)
	output = banktypes.NewOutput(withdrawer, withdrawCoins)

	if err := banktypes.ValidateInputsOutputs([]banktypes.Input{input}, []banktypes.Output{output}); err != nil {
		return banktypes.Input{}, banktypes.Output{}, err
	}

	return input, output, nil
}

// In order to deal with the batch at the same time, the coins of msgs are deposited in escrow.
func (k Keeper) DepositWithinBatch(ctx sdk.Context, msg *types.MsgDepositWithinBatch) (types.DepositMsgState, error) {
	if err := k.ValidateMsgDepositWithinBatch(ctx, *msg); err != nil {
		return types.DepositMsgState{}, err
	}

	poolBatch, found := k.GetPoolBatch(ctx, msg.PoolId)
	if !found {
		return types.DepositMsgState{}, types.ErrPoolBatchNotExists
	}

	if poolBatch.BeginHeight == 0 {
		poolBatch.BeginHeight = ctx.BlockHeight()
	}

	msgState := types.DepositMsgState{
		MsgHeight: ctx.BlockHeight(),
		MsgIndex:  poolBatch.DepositMsgIndex,
		Msg:       msg,
	}

	if err := k.HoldEscrow(ctx, msg.GetDepositor(), msg.DepositCoins); err != nil {
		return types.DepositMsgState{}, err
	}

	poolBatch.DepositMsgIndex++
	k.SetPoolBatch(ctx, poolBatch)
	k.SetPoolBatchDepositMsgState(ctx, poolBatch.PoolId, msgState)

	return msgState, nil
}

// In order to deal with the batch at the same time, the coins of msgs are deposited in escrow.
func (k Keeper) WithdrawWithinBatch(ctx sdk.Context, msg *types.MsgWithdrawWithinBatch) (types.WithdrawMsgState, error) {
	if err := k.ValidateMsgWithdrawWithinBatch(ctx, *msg); err != nil {
		return types.WithdrawMsgState{}, err
	}

	poolBatch, found := k.GetPoolBatch(ctx, msg.PoolId)
	if !found {
		return types.WithdrawMsgState{}, types.ErrPoolBatchNotExists
	}

	if poolBatch.BeginHeight == 0 {
		poolBatch.BeginHeight = ctx.BlockHeight()
	}

	batchPoolMsg := types.WithdrawMsgState{
		MsgHeight: ctx.BlockHeight(),
		MsgIndex:  poolBatch.WithdrawMsgIndex,
		Msg:       msg,
	}

	if err := k.HoldEscrow(ctx, msg.GetWithdrawer(), sdk.NewCoins(msg.PoolCoin)); err != nil {
		return types.WithdrawMsgState{}, err
	}

	poolBatch.WithdrawMsgIndex++
	k.SetPoolBatch(ctx, poolBatch)
	k.SetPoolBatchWithdrawMsgState(ctx, poolBatch.PoolId, batchPoolMsg)

	return batchPoolMsg, nil
}

// In order to deal with the batch at the same time, the coins of msgs are deposited in escrow.
func (k Keeper) SwapWithinBatch(ctx sdk.Context, msg *types.MsgSwapWithinBatch, orderExpirySpanHeight int64) (*types.SwapMsgState, error) {
	pool, found := k.GetPool(ctx, msg.PoolId)
	if !found {
		return nil, types.ErrPoolNotExists
	}
	if k.IsDepletedPool(ctx, pool) {
		return nil, types.ErrDepletedPool
	}
	if err := k.ValidateMsgSwapWithinBatch(ctx, *msg, pool); err != nil {
		return nil, err
	}
	poolBatch, found := k.GetPoolBatch(ctx, msg.PoolId)
	if !found {
		return nil, types.ErrPoolBatchNotExists
	}

	if poolBatch.BeginHeight == 0 {
		poolBatch.BeginHeight = ctx.BlockHeight()
	}

	currentHeight := ctx.BlockHeight()

	if orderExpirySpanHeight == 0 {
		params := k.GetParams(ctx)
		u := int64(params.UnitBatchHeight)
		orderExpirySpanHeight = (u - currentHeight%u) % u
	}

	batchPoolMsg := types.SwapMsgState{
		MsgHeight:            currentHeight,
		MsgIndex:             poolBatch.SwapMsgIndex,
		Executed:             false,
		Succeeded:            false,
		ToBeDeleted:          false,
		OrderExpiryHeight:    currentHeight + orderExpirySpanHeight,
		ExchangedOfferCoin:   sdk.NewCoin(msg.OfferCoin.Denom, sdk.ZeroInt()),
		RemainingOfferCoin:   msg.OfferCoin,
		ReservedOfferCoinFee: msg.OfferCoinFee,
		Msg:                  msg,
	}

	if err := k.HoldEscrow(ctx, msg.GetSwapRequester(), sdk.NewCoins(msg.OfferCoin.Add(msg.OfferCoinFee))); err != nil {
		return nil, err
	}

	poolBatch.SwapMsgIndex++
	k.SetPoolBatch(ctx, poolBatch)
	k.SetPoolBatchSwapMsgState(ctx, poolBatch.PoolId, batchPoolMsg)

	return &batchPoolMsg, nil
}
