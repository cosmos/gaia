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
// if globalfee is empty, CombinedFeeRequirement return sdk.Coins{}
func CombinedFeeRequirement(globalFees, minGasPrices sdk.Coins) sdk.Coins {
	// empty min_gas_price
	if len(minGasPrices) == 0 {
		return globalFees
	}
	// empty global fee is not possible if we set default global fee
	if len(globalFees) == 0 && len(minGasPrices) != 0 {
		return sdk.Coins{}
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

// splitCoinsByDenoms returns the given coins split in two whether
// their demon is or isn't found in the given denom map.
func splitCoinsByDenoms(feeCoins sdk.Coins, denomMap map[string]bool) (sdk.Coins, sdk.Coins) {
	feeCoinsNonZeroDenom, feeCoinsZeroDenom := sdk.Coins{}, sdk.Coins{}

	for _, fc := range feeCoins {
		_, found := denomMap[fc.Denom]
		if found {
			feeCoinsZeroDenom = append(feeCoinsZeroDenom, fc)
		} else {
			feeCoinsNonZeroDenom = append(feeCoinsNonZeroDenom, fc)
		}
	}

	return feeCoinsNonZeroDenom.Sort(), feeCoinsZeroDenom.Sort()
}

// getNonZeroFees returns the given fees nonzero coins
// and a map storing the zero coins's denoms
func getNonZeroFees(fees sdk.Coins) (sdk.Coins, map[string]bool) {
	requiredFeesNonZero := sdk.Coins{}
	requiredFeesZeroDenom := map[string]bool{}

	for _, gf := range fees {
		if gf.IsZero() {
			requiredFeesZeroDenom[gf.Denom] = true
		} else {
			requiredFeesNonZero = append(requiredFeesNonZero, gf)
		}
	}

	return requiredFeesNonZero.Sort(), requiredFeesZeroDenom
}
