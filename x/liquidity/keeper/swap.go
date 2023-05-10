package keeper

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

// Execute Swap of the pool batch, Collect swap messages in batch for transact the same price for each batch and run them on endblock.
func (k Keeper) SwapExecution(ctx sdk.Context, poolBatch types.PoolBatch) (uint64, error) {
	// get all swap message batch states that are not executed, not succeeded, and not to be deleted.
	swapMsgStates := k.GetAllNotProcessedPoolBatchSwapMsgStates(ctx, poolBatch)
	if len(swapMsgStates) == 0 {
		return 0, nil
	}

	pool, found := k.GetPool(ctx, poolBatch.PoolId)
	if !found {
		return 0, types.ErrPoolNotExists
	}

	if k.IsDepletedPool(ctx, pool) {
		return 0, types.ErrDepletedPool
	}

	currentHeight := ctx.BlockHeight()
	// set executed states of all messages to true
	executedMsgCount := uint64(0)
	var swapMsgStatesNotToBeDeleted []*types.SwapMsgState
	for _, sms := range swapMsgStates {
		sms.Executed = true
		executedMsgCount++
		if currentHeight > sms.OrderExpiryHeight {
			sms.ToBeDeleted = true
		}
		if err := k.ValidateMsgSwapWithinBatch(ctx, *sms.Msg, pool); err != nil {
			sms.ToBeDeleted = true
		}
		if !sms.ToBeDeleted {
			swapMsgStatesNotToBeDeleted = append(swapMsgStatesNotToBeDeleted, sms)
		} else {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeSwapTransacted,
					sdk.NewAttribute(types.AttributeValuePoolId, strconv.FormatUint(pool.Id, 10)),
					sdk.NewAttribute(types.AttributeValueBatchIndex, strconv.FormatUint(poolBatch.Index, 10)),
					sdk.NewAttribute(types.AttributeValueMsgIndex, strconv.FormatUint(sms.MsgIndex, 10)),
					sdk.NewAttribute(types.AttributeValueSwapRequester, sms.Msg.GetSwapRequester().String()),
					sdk.NewAttribute(types.AttributeValueSwapTypeId, strconv.FormatUint(uint64(sms.Msg.SwapTypeId), 10)),
					sdk.NewAttribute(types.AttributeValueOfferCoinDenom, sms.Msg.OfferCoin.Denom),
					sdk.NewAttribute(types.AttributeValueOfferCoinAmount, sms.Msg.OfferCoin.Amount.String()),
					sdk.NewAttribute(types.AttributeValueDemandCoinDenom, sms.Msg.DemandCoinDenom),
					sdk.NewAttribute(types.AttributeValueOrderPrice, sms.Msg.OrderPrice.String()),
					sdk.NewAttribute(types.AttributeValueRemainingOfferCoinAmount, sms.RemainingOfferCoin.Amount.String()),
					sdk.NewAttribute(types.AttributeValueExchangedOfferCoinAmount, sms.ExchangedOfferCoin.Amount.String()),
					sdk.NewAttribute(types.AttributeValueReservedOfferCoinFeeAmount, sms.ReservedOfferCoinFee.Amount.String()),
					sdk.NewAttribute(types.AttributeValueOrderExpiryHeight, strconv.FormatInt(sms.OrderExpiryHeight, 10)),
					sdk.NewAttribute(types.AttributeValueSuccess, types.Failure),
				))
		}
	}
	k.SetPoolBatchSwapMsgStatesByPointer(ctx, pool.Id, swapMsgStates)
	swapMsgStates = swapMsgStatesNotToBeDeleted

	types.ValidateStateAndExpireOrders(swapMsgStates, currentHeight, false)

	// get reserve coins from the liquidity pool and calculate the current pool price (p = x / y)
	reserveCoins := k.GetReserveCoins(ctx, pool)

	X := reserveCoins[0].Amount.ToDec()
	Y := reserveCoins[1].Amount.ToDec()
	currentPoolPrice := X.Quo(Y)
	denomX := reserveCoins[0].Denom
	denomY := reserveCoins[1].Denom

	// make orderMap, orderbook by sort orderMap
	orderMap, xToY, yToX := types.MakeOrderMap(swapMsgStates, denomX, denomY, false)
	orderBook := orderMap.SortOrderBook()

	// check orderbook validity and compute batchResult(direction, swapPrice, ..)
	result, found := orderBook.Match(X, Y)

	if !found || X.Quo(Y).IsZero() {
		err := k.RefundSwaps(ctx, pool, swapMsgStates)
		return executedMsgCount, err
	}

	// find order match, calculate pool delta with the total x, y amounts for the invariant check
	var matchResultXtoY, matchResultYtoX []types.MatchResult

	poolXDelta := sdk.ZeroDec()
	poolYDelta := sdk.ZeroDec()

	if result.MatchType != types.NoMatch {
		var poolXDeltaXtoY, poolXDeltaYtoX, poolYDeltaYtoX, poolYDeltaXtoY sdk.Dec
		matchResultXtoY, poolXDeltaXtoY, poolYDeltaXtoY = types.FindOrderMatch(types.DirectionXtoY, xToY, result.EX, result.SwapPrice, currentHeight)
		matchResultYtoX, poolXDeltaYtoX, poolYDeltaYtoX = types.FindOrderMatch(types.DirectionYtoX, yToX, result.EY, result.SwapPrice, currentHeight)
		poolXDelta = poolXDeltaXtoY.Add(poolXDeltaYtoX)
		poolYDelta = poolYDeltaXtoY.Add(poolYDeltaYtoX)
	}

	xToY, yToX, X, Y, poolXDelta2, poolYDelta2 := types.UpdateSwapMsgStates(X, Y, xToY, yToX, matchResultXtoY, matchResultYtoX)

	lastPrice := X.Quo(Y)

	if BatchLogicInvariantCheckFlag {
		SwapMatchingInvariants(xToY, yToX, matchResultXtoY, matchResultYtoX)
		SwapPriceInvariants(matchResultXtoY, matchResultYtoX, poolXDelta, poolYDelta, poolXDelta2, poolYDelta2, result)
	}

	types.ValidateStateAndExpireOrders(xToY, currentHeight, false)
	types.ValidateStateAndExpireOrders(yToX, currentHeight, false)

	orderMapExecuted, _, _ := types.MakeOrderMap(append(xToY, yToX...), denomX, denomY, true)
	orderBookExecuted := orderMapExecuted.SortOrderBook()
	if !orderBookExecuted.Validate(lastPrice) {
		return executedMsgCount, types.ErrOrderBookInvalidity
	}

	types.ValidateStateAndExpireOrders(xToY, currentHeight, true)
	types.ValidateStateAndExpireOrders(yToX, currentHeight, true)

	// make index map for match result
	matchResultMap := make(map[uint64]types.MatchResult)
	for _, match := range append(matchResultXtoY, matchResultYtoX...) {
		if _, ok := matchResultMap[match.SwapMsgState.MsgIndex]; ok {
			return executedMsgCount, fmt.Errorf("duplicate match order")
		}
		matchResultMap[match.SwapMsgState.MsgIndex] = match
	}

	if BatchLogicInvariantCheckFlag {
		SwapPriceDirectionInvariants(currentPoolPrice, result)
		SwapMsgStatesInvariants(matchResultXtoY, matchResultYtoX, matchResultMap, swapMsgStates, xToY, yToX)
		SwapOrdersExecutionStateInvariants(matchResultMap, swapMsgStates, result, denomX)
	}

	// execute transact, refund, expire, send coins with escrow, update state by TransactAndRefundSwapLiquidityPool
	if err := k.TransactAndRefundSwapLiquidityPool(ctx, swapMsgStates, matchResultMap, pool, result); err != nil {
		return executedMsgCount, err
	}

	return executedMsgCount, nil
}
