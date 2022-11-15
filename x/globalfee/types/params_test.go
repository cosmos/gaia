package types

import (
	"cosmossdk.io/math"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestDefaultParams(t *testing.T) {
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
		t.Run(name, func(t *testing.T) {
			err := validateMinimumGasPrices(test.coins)
			if test.expectErr {
				require.Error(t, err)	
				return
			}
			require.NoError(t, err)
		})
	}
}
