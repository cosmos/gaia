package types

import (
	"cosmossdk.io/math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestParamsEqual(t *testing.T) {
	p := DefaultParams()
	require.EqualValues(t, p.MinimumGasPrices, sdk.DecCoins{})
}

func Test_validateParams(t *testing.T) {
	tests := map[string]struct {
		coins     interface{} // not sdk.DeCoins, but Decoins defined in glboalfee
		expectErr bool
	}{
		"DefaultParams, pass": {
			DefaultParams().MinimumGasPrices,
			false,
		},
		"DecCoins conversion fails, fail": {
			sdk.Coins{sdk.NewCoin("photon", math.OneInt())},
			true,
		},
		//"coin denom does not match the denom reg expr, fail": {
		//	sdk.DecCoins{
		//		sdk.NewDecCoin("**!", math.OneInt()),
		//	},
		//	true,
		//},
		//"coin amount is negtive, fail": {
		//	sdk.DecCoins{
		//		sdk.NewDecCoin("photon", math.NewInt(-1)),
		//	},
		//	true,
		//},
		"coins amounts are zero, pass": {
			sdk.DecCoins{
				sdk.NewDecCoin("atom", math.ZeroInt()),
				sdk.NewDecCoin("photon", math.ZeroInt()),
			},
			false,
		},
		"duplicate coins denoms, fail": {
			sdk.DecCoins{
				sdk.NewDecCoin("photon", math.OneInt()),
				sdk.NewDecCoin("photon", math.OneInt()),
			},
			true,
		},
		"coins are not sorted by denom alphabetically, fail": {
			sdk.DecCoins{
				sdk.NewDecCoin("photon", math.OneInt()),
				sdk.NewDecCoin("atom", math.OneInt()),
			},
			true,
		},
	}

	for name, test := range tests {
		t.Log(name)
		err := validateMinimumGasPrices(test.coins)

		require.Equal(t, err == nil, !test.expectErr)
	}
}
