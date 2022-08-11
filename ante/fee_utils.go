package ante

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gaia/v8/x/globalfee/types"
	tmstrings "github.com/tendermint/tendermint/libs/strings"
)

// ParamStoreKeyMinGasPrices type require coins sorted. getGlobalFee will also return sorted coins (might return 0denom if globalMinGasPrice is 0)
func (mfd BypassMinFeeDecorator) getGlobalFee(ctx sdk.Context, feeTx sdk.FeeTx) sdk.Coins {
	var globalMinGasPrices sdk.DecCoins
	if mfd.GlobalMinFee.Has(ctx, types.ParamStoreKeyMinGasPrices) {
		mfd.GlobalMinFee.Get(ctx, types.ParamStoreKeyMinGasPrices, &globalMinGasPrices)
	}
	// global fee is empty set, set global fee to 0uatom
	if len(globalMinGasPrices) == 0 {
		globalMinGasPrices = DefaultZeroGlobalFee()
	}
	requiredGlobalFees := make(sdk.Coins, len(globalMinGasPrices))
	// Determine the required fees by multiplying each required minimum gas
	// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
	glDec := sdk.NewDec(int64(feeTx.GetGas()))
	for i, gp := range globalMinGasPrices {
		fee := gp.Amount.Mul(glDec)
		requiredGlobalFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
	}

	return requiredGlobalFees.Sort()
}

// getMinGasPrice will also return sorted coins
func getMinGasPrice(ctx sdk.Context, feeTx sdk.FeeTx) sdk.Coins {
	minGasPrices := ctx.MinGasPrices()
	gas := feeTx.GetGas()
	// special case: if minGasPrices=[], requiredFees=[]
	requiredFees := make(sdk.Coins, len(minGasPrices))
	// if not all coins are zero, check fee with min_gas_price
	if !minGasPrices.IsZero() {
		// Determine the required fees by multiplying each required minimum gas
		// price by the gas limit, where fee = ceil(minGasPrice * gasLimit).
		glDec := sdk.NewDec(int64(gas))
		for i, gp := range minGasPrices {
			fee := gp.Amount.Mul(glDec)
			requiredFees[i] = sdk.NewCoin(gp.Denom, fee.Ceil().RoundInt())
		}
	}

	return requiredFees.Sort()
}

func (mfd BypassMinFeeDecorator) bypassMinFeeMsgs(msgs []sdk.Msg) bool {
	for _, msg := range msgs {
		if tmstrings.StringInSlice(sdk.MsgTypeURL(msg), mfd.BypassMinFeeMsgTypes) {
			continue
		}
		return false
	}

	return true
}

// DenomsSubsetOfIncludingZero and IsAnyGTEIncludingZero are similar to DenomsSubsetOf and IsAnyGTE in sdk. Since we allow zero coins in global fee(zero coins means the chain does not want to set a global fee but still want to define the fee's denom)
//
// overwrite DenomsSubsetOfIncludingZero from sdk, to allow zero amt coins in superset. e.g. 1stake is DenomsSubsetOfIncludingZero 0stake. [] is the DenomsSubsetOfIncludingZero of [0stake] but not [1stake].
// DenomsSubsetOfIncludingZero returns true if coins's denoms is subset of coinsB's denoms.
// if coins is empty set, empty set is any sets' subset
func DenomsSubsetOfIncludingZero(coins, coinsB sdk.Coins) bool {
	// more denoms in B than in receiver
	if len(coins) > len(coinsB) {
		return false
	}
	// coins=[], coinsB=[0stake]
	// let all len(coins) == 0 pass and reject later at IsAnyGTEIncludingZero
	if len(coins) == 0 && containZeroCoins(coinsB) {
		return true
	}
	// coins=1stake, coinsB=[0stake,1uatom]
	for _, coin := range coins {
		err := sdk.ValidateDenom(coin.Denom)
		if err != nil {
			panic(err)
		}

		if ok, _ := coinsB.Find(coin.Denom); !ok {
			return false
		}
	}

	return true
}

// overwrite the IsAnyGTEIncludingZero from sdk to allow zero coins in coins and coinsB.
// IsAnyGTEIncludingZero returns true if coins contain at least one denom that is present at a greater or equal amount in coinsB; it returns false otherwise.
// if CoinsB is emptyset, no coins sets are IsAnyGTEIncludingZero coinsB unless coins is also empty set.
// NOTE: IsAnyGTEIncludingZero operates under the invariant that both coin sets are sorted by denoms.
// contract !!!! coins must be DenomsSubsetOfIncludingZero of coinsB
func IsAnyGTEIncludingZero(coins, coinsB sdk.Coins) bool {
	// no set is empty set's subset except empty set
	// this is different from sdk, sdk return false for coinsB empty
	if len(coinsB) == 0 && len(coins) == 0 {
		return true
	}
	// nothing is gte empty coins
	if len(coinsB) == 0 && len(coins) != 0 {
		return false
	}
	// if feecoins empty (len(coins)==0 && len(coinsB) != 0 ), and globalfee has one denom of amt zero, return true
	if len(coins) == 0 {
		return containZeroCoins(coinsB)
	}

	//  len(coinsB) != 0 && len(coins) != 0
	// special case: coins=1stake, coinsB=[2stake,0uatom], fail
	for _, coin := range coins {
		// not find coin in CoinsB
		if ok, _ := coinsB.Find(coin.Denom); ok {
			// find coin in coinsB, and if the amt == 0, mean either coin=0denom or coinsB=0denom...both true
			amt := coinsB.AmountOf(coin.Denom)
			if coin.Amount.GTE(amt) {
				return true
			}
		}
	}

	return false
}

// return true if coinsB is empty or contains zero coins,
// CoinsB must be validate coins !!!
func containZeroCoins(coinsB sdk.Coins) bool {
	if len(coinsB) == 0 {
		return true
	}
	for _, coin := range coinsB {
		if coin.IsZero() {
			return true
		}
	}

	return false
}

// CombinedFeeRequirement will combine the global fee and min_gas_price. Both globalFees and minGasPrices must be valid, but CombinedFeeRequirement does not validate them so it may return 0denom.
func CombinedFeeRequirement(globalFees, minGasPrices sdk.Coins) sdk.Coins {
	// empty min_gas_price
	if len(minGasPrices) == 0 {
		return globalFees
	}
	// empty global fee is not possible if we set default global fee
	if len(globalFees) == 0 && len(minGasPrices) != 0 {
		return globalFees
	}

	// if find min_gas_price denom in globalfee, and amt is higher than globalfee, add it
	var allFees sdk.Coins
	for _, fee := range globalFees {
		// min_gas_price denom in global fee
		ok, c := minGasPrices.Find(fee.Denom)
		if ok {
			if c.Amount.GT(fee.Amount) {
				allFees = append(allFees, c)
			} else {
				allFees = append(allFees, fee)
			}
		} else {
			allFees = append(allFees, fee)
		}
	}

	return allFees.Sort()
}

// getTxPriority returns a naive tx priority based on the amount of the smallest denomination of the fee
// provided in a transaction.
func GetTxPriority(fee sdk.Coins) int64 {
	var priority int64
	for _, c := range fee {
		p := int64(math.MaxInt64)
		if c.Amount.IsInt64() {
			p = c.Amount.Int64()
		}
		if priority == 0 || p < priority {
			priority = p
		}
	}

	return priority
}
