package simulation

import (
	"math/rand"

	appparams "github.com/cosmos/gaia/v23/x/tokenfactory/app/params"
	"github.com/cosmos/gaia/v23/x/tokenfactory/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func RandDenomCreationFeeParam(r *rand.Rand) sdk.Coins {
	amount := r.Int63n(10_000_000)
	return sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, sdkmath.NewInt(amount)))
}

func RandomizedGenState(simstate *module.SimulationState) {
	tfGenesis := types.DefaultGenesis()

	_, err := simstate.Cdc.MarshalJSON(tfGenesis)
	if err != nil {
		panic(err)
	}

	simstate.GenState[types.ModuleName] = simstate.Cdc.MustMarshalJSON(tfGenesis)
}
