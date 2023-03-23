package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ContainZeroCoins returns true if the given coins are empty or contain zero coins,
// Note that the coins denoms must be validated, see sdk.ValidateDenom
func ContainZeroCoins(coins sdk.Coins) bool {
	if len(coins) == 0 {
		return true
	}
	for _, coin := range coins {
		if coin.IsZero() {
			return true
		}
	}

	return false
}

// CombinedFeeRequirement returns the global fee and min_gas_price combined and sorted.
// Both globalFees and minGasPrices must be valid, but CombinedFeeRequirement
// does not validate them, so it may return 0denom.
func CombinedFeeRequirement(globalFees, minGasPrices sdk.Coins) sdk.Coins {
	// empty min_gas_price
	if len(minGasPrices) == 0 {
		return globalFees
	}
	// empty global fee is not possible if we set default global fee
	if len(globalFees) == 0 && len(minGasPrices) != 0 {
		return globalFees
	}

	// if min_gas_price denom is in globalfee, and the amount is higher than globalfee, add min_gas_price to allFees
	var allFees sdk.Coins
	for _, fee := range globalFees {
		// min_gas_price denom in global fee
		ok, c := Find(minGasPrices, fee.Denom)
		if ok && c.Amount.GT(fee.Amount) {
			allFees = append(allFees, c)
		} else {
			allFees = append(allFees, fee)
		}
	}

	return allFees.Sort()
}

// Find replaces the functionality of Coins.Find from SDK v0.46.x
func Find(coins sdk.Coins, denom string) (bool, sdk.Coin) {
	switch len(coins) {
	case 0:
		return false, sdk.Coin{}

	case 1:
		coin := coins[0]
		if coin.Denom == denom {
			return true, coin
		}
		return false, sdk.Coin{}

	default:
		midIdx := len(coins) / 2 // 2:1, 3:1, 4:2
		coin := coins[midIdx]
		switch {
		case denom < coin.Denom:
			return Find(coins[:midIdx], denom)
		case denom == coin.Denom:
			return true, coin
		default:
			return Find(coins[midIdx+1:], denom)
		}
	}
}

// RemovingZeroDenomCoins return feeCoins with removing coins whose denom is zero coin's denom in globalfees
func RemovingZeroDenomCoins(feeCoins sdk.Coins, zeroGlobalFeesDenom map[string]bool) sdk.Coins {
	feeCoinsNoZeroDenomCoins := []sdk.Coin{}
	for _, fc := range feeCoins {
		if _, found := zeroGlobalFeesDenom[fc.Denom]; !found {
			feeCoinsNoZeroDenomCoins = append(feeCoinsNoZeroDenomCoins, fc)
		}
	}

	return feeCoinsNoZeroDenomCoins
}

// splitGlobalFees returns the sorted nonzero coins  and zero denoms in globalfee
func splitGlobalFees(globalfees sdk.Coins) (sdk.Coins, map[string]bool) {

	requiredGlobalFeesNonZero := sdk.Coins{}
	requiredGlobalFeesZeroDenom := map[string]bool{}

	for _, gf := range globalfees {

		if gf.IsZero() {
			requiredGlobalFeesZeroDenom[gf.Denom] = true
		} else {
			requiredGlobalFeesNonZero = append(requiredGlobalFeesNonZero, gf)
		}
	}

	return requiredGlobalFeesNonZero.Sort(), requiredGlobalFeesZeroDenom
}
