package simulation

// DONTCOVER

import (
	"math/rand"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/cosmos/gaia/v9/x/liquidity/keeper"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

// create simulated accounts due to gas usage overflow issue.
// Read this issue: https://github.com/tendermint/liquidity/issues/349
var randomAccounts []simtypes.Account

// mintCoins mints and send coins to the simulated account.
func mintCoins(ctx sdk.Context, r *rand.Rand, bk types.BankKeeper, acc simtypes.Account, denoms []string) error {
	var mintCoins, sendCoins sdk.Coins
	for _, denom := range denoms {
		mintAmt := sdk.NewInt(int64(simtypes.RandIntBetween(r, 1e15, 1e16)))
		sendAmt := sdk.NewInt(int64(simtypes.RandIntBetween(r, 1e13, 1e14)))
		mintCoins = mintCoins.Add(sdk.NewCoin(denom, mintAmt))
		sendCoins = sendCoins.Add(sdk.NewCoin(denom, sendAmt))
	}

	feeCoin := int64(simtypes.RandIntBetween(r, 1e13, 1e14))
	mintCoins = mintCoins.Add(sdk.NewInt64Coin(sdk.DefaultBondDenom, feeCoin))
	sendCoins = sendCoins.Add(sdk.NewInt64Coin(sdk.DefaultBondDenom, feeCoin))

	err := bk.MintCoins(ctx, types.ModuleName, mintCoins)
	if err != nil {
		return err
	}

	err = bk.SendCoinsFromModuleToAccount(ctx, types.ModuleName, acc.Address, sendCoins)
	if err != nil {
		return err
	}

	return nil
}

// randomLiquidity returns a random liquidity pool.
func randomLiquidity(r *rand.Rand, k keeper.Keeper, ctx sdk.Context) (pool types.Pool, ok bool) {
	pools := k.GetAllPools(ctx)
	if len(pools) == 0 {
		return types.Pool{}, false
	}

	i := r.Intn(len(pools))

	return pools[i], true
}

// randomDepositCoin returns deposit amount between more than minimum deposit amount and less than 1e9.
func randomDepositCoin(r *rand.Rand, minInitDepositAmount sdk.Int, denom string) sdk.Coin {
	amount := int64(simtypes.RandIntBetween(r, int(minInitDepositAmount.Int64()+1), 1e8))
	return sdk.NewInt64Coin(denom, amount)
}

// randomWithdrawCoin returns random withdraw amount.
func randomWithdrawCoin(r *rand.Rand, denom string, balance sdk.Int) sdk.Coin {
	// prevent panic from RandIntBetween
	if balance.Quo(sdk.NewInt(10)).Int64() <= 1 {
		return sdk.NewInt64Coin(denom, 1)
	}

	amount := int64(simtypes.RandIntBetween(r, 1, int(balance.Quo(sdk.NewInt(10)).Int64())))
	return sdk.NewInt64Coin(denom, amount)
}

// randomOfferCoin returns random offer amount of coin.
func randomOfferCoin(r *rand.Rand, k keeper.Keeper, ctx sdk.Context, pool types.Pool, denom string) sdk.Coin {
	params := k.GetParams(ctx)
	reserveCoinAmt := k.GetReserveCoins(ctx, pool).AmountOf(denom)
	maximumOrderableAmt := reserveCoinAmt.ToDec().Mul(params.MaxOrderAmountRatio).TruncateInt()
	amt := int64(simtypes.RandIntBetween(r, 1, int(maximumOrderableAmt.Int64())))
	return sdk.NewInt64Coin(denom, amt)
}

// randomOrderPrice returns random order price that is sufficient for matchable swap.
func randomOrderPrice(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 1e2)), 2)
}

// randomFees returns a random amount of bond denom fee and
// if the user doesn't have enough funds for paying fees, it returns empty coins.
func randomFees(r *rand.Rand, spendableCoins sdk.Coins) (sdk.Coins, error) {
	if spendableCoins.Empty() {
		return nil, nil
	}

	if spendableCoins.AmountOf(sdk.DefaultBondDenom).Equal(sdk.ZeroInt()) {
		return nil, nil
	}

	amt, err := simtypes.RandPositiveInt(r, spendableCoins.AmountOf(sdk.DefaultBondDenom))
	if err != nil {
		return nil, err
	}

	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amt))

	return fees, nil
}

// randomDenoms returns randomly generated two different denoms that has a length anywhere between 4 and 6.
func randomDenoms(r *rand.Rand) (string, string) {
	denomA := simtypes.RandStringOfLength(r, simtypes.RandIntBetween(r, 4, 6))
	denomB := simtypes.RandStringOfLength(r, simtypes.RandIntBetween(r, 4, 6))
	denomA, denomB = types.AlphabeticalDenomPair(strings.ToLower(denomA), strings.ToLower(denomB))
	return denomA, denomB
}
