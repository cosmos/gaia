package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

// RegisterInvariants registers all liquidity invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "escrow-amount",
		LiquidityPoolsEscrowAmountInvariant(k))
}

// AllInvariants runs all invariants of the liquidity module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := LiquidityPoolsEscrowAmountInvariant(k)(ctx)
		return res, stop
	}
}

// LiquidityPoolsEscrowAmountInvariant checks that outstanding unwithdrawn fees are never negative.
func LiquidityPoolsEscrowAmountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		remainingCoins := sdk.NewCoins()
		batches := k.GetAllPoolBatches(ctx)
		for _, batch := range batches {
			swapMsgs := k.GetAllPoolBatchSwapMsgStatesNotToBeDeleted(ctx, batch)
			for _, msg := range swapMsgs {
				remainingCoins = remainingCoins.Add(msg.RemainingOfferCoin)
			}
			depositMsgs := k.GetAllPoolBatchDepositMsgStatesNotToBeDeleted(ctx, batch)
			for _, msg := range depositMsgs {
				remainingCoins = remainingCoins.Add(msg.Msg.DepositCoins...)
			}
			withdrawMsgs := k.GetAllPoolBatchWithdrawMsgStatesNotToBeDeleted(ctx, batch)
			for _, msg := range withdrawMsgs {
				remainingCoins = remainingCoins.Add(msg.Msg.PoolCoin)
			}
		}

		batchEscrowAcc := k.accountKeeper.GetModuleAddress(types.ModuleName)
		escrowAmt := k.bankKeeper.GetAllBalances(ctx, batchEscrowAcc)

		broken := !escrowAmt.IsAllGTE(remainingCoins)

		return sdk.FormatInvariant(types.ModuleName, "batch escrow amount invariant broken",
			"batch escrow amount LT batch remaining amount"), broken
	}
}

// These invariants cannot be registered via RegisterInvariants since the module uses per-block batch execution.
// We should approach adding these invariant checks inside actual logics of deposit / withdraw / swap.

var (
	BatchLogicInvariantCheckFlag = false // It is only used at the development stage, and is disabled at the product level.
	// For coin amounts less than coinAmountThreshold, a high errorRate does not mean
	// that the calculation logic has errors.
	// For example, if there were two X coins and three Y coins in the pool, and someone deposits
	// one X coin and one Y coin, it's an acceptable input.
	// But pool price would change from 2/3 to 3/4 so errorRate will report 1/8(=0.125),
	// meaning that the price has changed by 12.5%.
	// This happens with small coin amounts, so there should be a threshold for coin amounts
	// before we calculate the errorRate.
	errorRateThreshold  = sdk.NewDecWithPrec(5, 2) // 5%
	coinAmountThreshold = sdk.NewInt(20)           // If a decimal error occurs at a value less than 20, the error rate is over 5%.
)

func errorRate(expected, actual sdk.Dec) sdk.Dec {
	// To prevent divide-by-zero panics, return 1.0(=100%) as the error rate
	// when the expected value is 0.
	if expected.IsZero() {
		return sdk.OneDec()
	}
	return actual.Sub(expected).Quo(expected).Abs()
}

// MintingPoolCoinsInvariant checks the correct ratio of minting amount of pool coins.
func MintingPoolCoinsInvariant(poolCoinTotalSupply, mintPoolCoin, depositCoinA, depositCoinB, lastReserveCoinA, lastReserveCoinB, refundedCoinA, refundedCoinB sdk.Int) {
	if !refundedCoinA.IsZero() {
		depositCoinA = depositCoinA.Sub(refundedCoinA)
	}

	if !refundedCoinB.IsZero() {
		depositCoinB = depositCoinB.Sub(refundedCoinB)
	}

	poolCoinRatio := mintPoolCoin.ToDec().QuoInt(poolCoinTotalSupply)
	depositCoinARatio := depositCoinA.ToDec().QuoInt(lastReserveCoinA)
	depositCoinBRatio := depositCoinB.ToDec().QuoInt(lastReserveCoinB)
	expectedMintPoolCoinAmtBasedA := depositCoinARatio.MulInt(poolCoinTotalSupply).TruncateInt()
	expectedMintPoolCoinAmtBasedB := depositCoinBRatio.MulInt(poolCoinTotalSupply).TruncateInt()

	// NewPoolCoinAmount / LastPoolCoinSupply == AfterRefundedDepositCoinA / LastReserveCoinA
	// NewPoolCoinAmount / LastPoolCoinSupply == AfterRefundedDepositCoinA / LastReserveCoinB
	if depositCoinA.GTE(coinAmountThreshold) && depositCoinB.GTE(coinAmountThreshold) &&
		lastReserveCoinA.GTE(coinAmountThreshold) && lastReserveCoinB.GTE(coinAmountThreshold) &&
		mintPoolCoin.GTE(coinAmountThreshold) && poolCoinTotalSupply.GTE(coinAmountThreshold) {
		if errorRate(depositCoinARatio, poolCoinRatio).GT(errorRateThreshold) ||
			errorRate(depositCoinBRatio, poolCoinRatio).GT(errorRateThreshold) {
			panic("invariant check fails due to incorrect ratio of pool coins")
		}
	}

	if mintPoolCoin.GTE(coinAmountThreshold) &&
		(sdk.MaxInt(mintPoolCoin, expectedMintPoolCoinAmtBasedA).Sub(sdk.MinInt(mintPoolCoin, expectedMintPoolCoinAmtBasedA)).ToDec().QuoInt(mintPoolCoin).GT(errorRateThreshold) ||
			sdk.MaxInt(mintPoolCoin, expectedMintPoolCoinAmtBasedB).Sub(sdk.MinInt(mintPoolCoin, expectedMintPoolCoinAmtBasedA)).ToDec().QuoInt(mintPoolCoin).GT(errorRateThreshold)) {
		panic("invariant check fails due to incorrect amount of pool coins")
	}
}

// DepositInvariant checks after deposit amounts.
func DepositInvariant(lastReserveCoinA, lastReserveCoinB, depositCoinA, depositCoinB, afterReserveCoinA, afterReserveCoinB, refundedCoinA, refundedCoinB sdk.Int) {
	depositCoinA = depositCoinA.Sub(refundedCoinA)
	depositCoinB = depositCoinB.Sub(refundedCoinB)

	depositCoinRatio := depositCoinA.ToDec().Quo(depositCoinB.ToDec())
	lastReserveRatio := lastReserveCoinA.ToDec().Quo(lastReserveCoinB.ToDec())
	afterReserveRatio := afterReserveCoinA.ToDec().Quo(afterReserveCoinB.ToDec())

	// AfterDepositReserveCoinA = LastReserveCoinA + AfterRefundedDepositCoinA
	// AfterDepositReserveCoinB = LastReserveCoinB + AfterRefundedDepositCoinA
	if !afterReserveCoinA.Equal(lastReserveCoinA.Add(depositCoinA)) ||
		!afterReserveCoinB.Equal(lastReserveCoinB.Add(depositCoinB)) {
		panic("invariant check fails due to incorrect deposit amounts")
	}

	if depositCoinA.GTE(coinAmountThreshold) && depositCoinB.GTE(coinAmountThreshold) &&
		lastReserveCoinA.GTE(coinAmountThreshold) && lastReserveCoinB.GTE(coinAmountThreshold) {
		// AfterRefundedDepositCoinA / AfterRefundedDepositCoinA = LastReserveCoinA / LastReserveCoinB
		if errorRate(lastReserveRatio, depositCoinRatio).GT(errorRateThreshold) {
			panic("invariant check fails due to incorrect deposit ratio")
		}
		// LastReserveCoinA / LastReserveCoinB = AfterDepositReserveCoinA / AfterDepositReserveCoinB
		if errorRate(lastReserveRatio, afterReserveRatio).GT(errorRateThreshold) {
			panic("invariant check fails due to incorrect pool price ratio")
		}
	}
}

// BurningPoolCoinsInvariant checks the correct burning amount of pool coins.
func BurningPoolCoinsInvariant(burnedPoolCoin, withdrawCoinA, withdrawCoinB, reserveCoinA, reserveCoinB, lastPoolCoinSupply sdk.Int, withdrawFeeCoins sdk.Coins) {
	burningPoolCoinRatio := burnedPoolCoin.ToDec().Quo(lastPoolCoinSupply.ToDec())
	if burningPoolCoinRatio.Equal(sdk.OneDec()) {
		return
	}

	withdrawCoinARatio := withdrawCoinA.Add(withdrawFeeCoins[0].Amount).ToDec().Quo(reserveCoinA.ToDec())
	withdrawCoinBRatio := withdrawCoinB.Add(withdrawFeeCoins[1].Amount).ToDec().Quo(reserveCoinB.ToDec())

	// BurnedPoolCoinAmount / LastPoolCoinSupply >= (WithdrawCoinA+WithdrawFeeCoinA) / LastReserveCoinA
	// BurnedPoolCoinAmount / LastPoolCoinSupply >= (WithdrawCoinB+WithdrawFeeCoinB) / LastReserveCoinB
	if withdrawCoinARatio.GT(burningPoolCoinRatio) || withdrawCoinBRatio.GT(burningPoolCoinRatio) {
		panic("invariant check fails due to incorrect ratio of burning pool coins")
	}

	expectedBurningPoolCoinBasedA := lastPoolCoinSupply.ToDec().MulTruncate(withdrawCoinARatio).TruncateInt()
	expectedBurningPoolCoinBasedB := lastPoolCoinSupply.ToDec().MulTruncate(withdrawCoinBRatio).TruncateInt()

	if burnedPoolCoin.GTE(coinAmountThreshold) &&
		(sdk.MaxInt(burnedPoolCoin, expectedBurningPoolCoinBasedA).Sub(sdk.MinInt(burnedPoolCoin, expectedBurningPoolCoinBasedA)).ToDec().QuoInt(burnedPoolCoin).GT(errorRateThreshold) ||
			sdk.MaxInt(burnedPoolCoin, expectedBurningPoolCoinBasedB).Sub(sdk.MinInt(burnedPoolCoin, expectedBurningPoolCoinBasedB)).ToDec().QuoInt(burnedPoolCoin).GT(errorRateThreshold)) {
		panic("invariant check fails due to incorrect amount of burning pool coins")
	}
}

// WithdrawReserveCoinsInvariant checks the after withdraw amounts.
func WithdrawReserveCoinsInvariant(withdrawCoinA, withdrawCoinB, reserveCoinA, reserveCoinB,
	afterReserveCoinA, afterReserveCoinB, afterPoolCoinTotalSupply, lastPoolCoinSupply, burnedPoolCoin sdk.Int,
) {
	// AfterWithdrawReserveCoinA = LastReserveCoinA - WithdrawCoinA
	if !afterReserveCoinA.Equal(reserveCoinA.Sub(withdrawCoinA)) {
		panic("invariant check fails due to incorrect withdraw coin A amount")
	}

	// AfterWithdrawReserveCoinB = LastReserveCoinB - WithdrawCoinB
	if !afterReserveCoinB.Equal(reserveCoinB.Sub(withdrawCoinB)) {
		panic("invariant check fails due to incorrect withdraw coin B amount")
	}

	// AfterWithdrawPoolCoinSupply = LastPoolCoinSupply - BurnedPoolCoinAmount
	if !afterPoolCoinTotalSupply.Equal(lastPoolCoinSupply.Sub(burnedPoolCoin)) {
		panic("invariant check fails due to incorrect total supply")
	}
}

// WithdrawAmountInvariant checks the correct ratio of withdraw coin amounts.
func WithdrawAmountInvariant(withdrawCoinA, withdrawCoinB, reserveCoinA, reserveCoinB, burnedPoolCoin, poolCoinSupply sdk.Int, withdrawFeeRate sdk.Dec) {
	ratio := burnedPoolCoin.ToDec().Quo(poolCoinSupply.ToDec()).Mul(sdk.OneDec().Sub(withdrawFeeRate))
	idealWithdrawCoinA := reserveCoinA.ToDec().Mul(ratio)
	idealWithdrawCoinB := reserveCoinB.ToDec().Mul(ratio)
	diffA := idealWithdrawCoinA.Sub(withdrawCoinA.ToDec()).Abs()
	diffB := idealWithdrawCoinB.Sub(withdrawCoinB.ToDec()).Abs()
	if !burnedPoolCoin.Equal(poolCoinSupply) {
		if diffA.GTE(sdk.OneDec()) {
			panic(fmt.Sprintf("withdraw coin amount %v differs too much from %v", withdrawCoinA, idealWithdrawCoinA))
		}
		if diffB.GTE(sdk.OneDec()) {
			panic(fmt.Sprintf("withdraw coin amount %v differs too much from %v", withdrawCoinB, idealWithdrawCoinB))
		}
	}
}

// ImmutablePoolPriceAfterWithdrawInvariant checks the immutable pool price after withdrawing coins.
func ImmutablePoolPriceAfterWithdrawInvariant(reserveCoinA, reserveCoinB, withdrawCoinA, withdrawCoinB, afterReserveCoinA, afterReserveCoinB sdk.Int) {
	// TestReinitializePool tests a scenario where after reserve coins are zero
	if !afterReserveCoinA.IsZero() && !afterReserveCoinB.IsZero() {
		reserveCoinA = reserveCoinA.Sub(withdrawCoinA)
		reserveCoinB = reserveCoinB.Sub(withdrawCoinB)

		reserveCoinRatio := reserveCoinA.ToDec().Quo(reserveCoinB.ToDec())
		afterReserveCoinRatio := afterReserveCoinA.ToDec().Quo(afterReserveCoinB.ToDec())

		// LastReserveCoinA / LastReserveCoinB = AfterWithdrawReserveCoinA / AfterWithdrawReserveCoinB
		if reserveCoinA.GTE(coinAmountThreshold) && reserveCoinB.GTE(coinAmountThreshold) &&
			withdrawCoinA.GTE(coinAmountThreshold) && withdrawCoinB.GTE(coinAmountThreshold) &&
			errorRate(reserveCoinRatio, afterReserveCoinRatio).GT(errorRateThreshold) {
			panic("invariant check fails due to incorrect pool price ratio")
		}
	}
}

// SwapMatchingInvariants checks swap matching results of both X to Y and Y to X cases.
func SwapMatchingInvariants(xToY, yToX []*types.SwapMsgState, matchResultXtoY, matchResultYtoX []types.MatchResult) {
	beforeMatchingXtoYLen := len(xToY)
	beforeMatchingYtoXLen := len(yToX)
	afterMatchingXtoYLen := len(matchResultXtoY)
	afterMatchingYtoXLen := len(matchResultYtoX)

	notMatchedXtoYLen := beforeMatchingXtoYLen - afterMatchingXtoYLen
	notMatchedYtoXLen := beforeMatchingYtoXLen - afterMatchingYtoXLen

	if notMatchedXtoYLen != types.CountNotMatchedMsgs(xToY) {
		panic("invariant check fails due to invalid xToY match length")
	}

	if notMatchedYtoXLen != types.CountNotMatchedMsgs(yToX) {
		panic("invariant check fails due to invalid yToX match length")
	}
}

// SwapPriceInvariants checks swap price invariants.
func SwapPriceInvariants(matchResultXtoY, matchResultYtoX []types.MatchResult, _, _, poolXDelta2, poolYDelta2 sdk.Dec, result types.BatchResult) {
	invariantCheckX := sdk.ZeroDec()
	invariantCheckY := sdk.ZeroDec()

	for _, m := range matchResultXtoY {
		invariantCheckX = invariantCheckX.Sub(m.TransactedCoinAmt)
		invariantCheckY = invariantCheckY.Add(m.ExchangedDemandCoinAmt)
	}

	for _, m := range matchResultYtoX {
		invariantCheckY = invariantCheckY.Sub(m.TransactedCoinAmt)
		invariantCheckX = invariantCheckX.Add(m.ExchangedDemandCoinAmt)
	}

	invariantCheckX = invariantCheckX.Add(poolXDelta2)
	invariantCheckY = invariantCheckY.Add(poolYDelta2)

	if !invariantCheckX.IsZero() && !invariantCheckY.IsZero() {
		panic(fmt.Errorf("invariant check fails due to invalid swap price: %s", invariantCheckX.String()))
	}

	validitySwapPrice := types.CheckSwapPrice(matchResultXtoY, matchResultYtoX, result.SwapPrice)
	if !validitySwapPrice {
		panic("invariant check fails due to invalid swap price")
	}
}

// SwapPriceDirectionInvariants checks whether the calculated swap price is increased, decreased, or stayed from the last pool price.
func SwapPriceDirectionInvariants(currentPoolPrice sdk.Dec, batchResult types.BatchResult) {
	switch batchResult.PriceDirection {
	case types.Increasing:
		if !batchResult.SwapPrice.GT(currentPoolPrice) {
			panic("invariant check fails due to incorrect price direction")
		}
	case types.Decreasing:
		if !batchResult.SwapPrice.LT(currentPoolPrice) {
			panic("invariant check fails due to incorrect price direction")
		}
	case types.Staying:
		if !batchResult.SwapPrice.Equal(currentPoolPrice) {
			panic("invariant check fails due to incorrect price direction")
		}
	}
}

// SwapMsgStatesInvariants checks swap match result states invariants.
func SwapMsgStatesInvariants(matchResultXtoY, matchResultYtoX []types.MatchResult, matchResultMap map[uint64]types.MatchResult,
	swapMsgStates []*types.SwapMsgState, xToY, yToX []*types.SwapMsgState,
) {
	if len(matchResultXtoY)+len(matchResultYtoX) != len(matchResultMap) {
		panic("invalid length of match result")
	}

	for k, v := range matchResultMap {
		if k != v.SwapMsgState.MsgIndex {
			panic("broken map consistency")
		}
	}

	for _, sms := range swapMsgStates {
		for _, smsXtoY := range xToY {
			if sms.MsgIndex == smsXtoY.MsgIndex {
				if *(sms) != *(smsXtoY) || sms != smsXtoY {
					panic("swap message state not matched")
				} else {
					break
				}
			}
		}

		for _, smsYtoX := range yToX {
			if sms.MsgIndex == smsYtoX.MsgIndex {
				if *(sms) != *(smsYtoX) || sms != smsYtoX {
					panic("swap message state not matched")
				} else {
					break
				}
			}
		}

		if msgAfter, ok := matchResultMap[sms.MsgIndex]; ok {
			if sms.MsgIndex == msgAfter.SwapMsgState.MsgIndex {
				if *(sms) != *(msgAfter.SwapMsgState) || sms != msgAfter.SwapMsgState {
					panic("batch message not matched")
				}
			} else {
				panic("fail msg pointer consistency")
			}
		}
	}
}

// SwapOrdersExecutionStateInvariants checks all executed orders have order price which is not "executable" or not "unexecutable".
func SwapOrdersExecutionStateInvariants(matchResultMap map[uint64]types.MatchResult, swapMsgStates []*types.SwapMsgState,
	batchResult types.BatchResult, denomX string,
) {
	for _, sms := range swapMsgStates {
		if _, ok := matchResultMap[sms.MsgIndex]; ok {
			if !sms.Executed || !sms.Succeeded {
				panic("swap msg state consistency error, matched but not succeeded")
			}

			if sms.Msg.OfferCoin.Denom == denomX {
				// buy orders having equal or higher order price than found swapPrice
				if !sms.Msg.OrderPrice.GTE(batchResult.SwapPrice) {
					panic("execution validity failed, executed but unexecutable")
				}
			} else {
				// sell orders having equal or lower order price than found swapPrice
				if !sms.Msg.OrderPrice.LTE(batchResult.SwapPrice) {
					panic("execution validity failed, executed but unexecutable")
				}
			}
		} else {
			// check whether every unexecuted orders have order price which is not "executable"
			if sms.Executed && sms.Succeeded {
				panic("sms consistency error, not matched but succeeded")
			}

			if sms.Msg.OfferCoin.Denom == denomX {
				// buy orders having equal or lower order price than found swapPrice
				if !sms.Msg.OrderPrice.LTE(batchResult.SwapPrice) {
					panic("execution validity failed, unexecuted but executable")
				}
			} else {
				// sell orders having equal or higher order price than found swapPrice
				if !sms.Msg.OrderPrice.GTE(batchResult.SwapPrice) {
					panic("execution validity failed, unexecuted but executable")
				}
			}
		}
	}
}
