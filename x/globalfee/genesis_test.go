package globalfee

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/gaia/v8/x/globalfee/types"
)

func TestDefaultGenesis(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	gotJson := AppModuleBasic{}.DefaultGenesis(encCfg.Codec)
	assert.JSONEq(t, `{"params":{"minimum_gas_prices":[]}}`, string(gotJson), string(gotJson))
}

func TestValidateGenesis(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	specs := map[string]struct {
		src    string
		expErr bool
	}{
		"all good": {
			src: `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"}]}}`,
		},
		"empty minimum": {
			src: `{"params":{"minimum_gas_prices":[]}}`,
		},
		"minimum not set": {
			src: `{"params":{}}`,
		},
		"zero amount not allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"0"}]}}`,
			expErr: true,
		},
		"duplicate denoms not allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"},{"denom":"ALX", "amount":"2"}]}}`,
			expErr: true,
		},
		"negative amounts not allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"-1"}]}}`,
			expErr: true,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := AppModuleBasic{}.ValidateGenesis(encCfg.Codec, nil, []byte(spec.src))
			if spec.expErr {
				require.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
		})
	}
}

func TestInitExportGenesis(t *testing.T) {
	specs := map[string]struct {
		src string
		exp types.GenesisState
	}{
		"single fee": {
			src: `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"}]}}`,
			exp: types.GenesisState{types.Params{MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.NewInt(1)))}},
		},
		"multiple fee options": {
			src: `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"}, {"denom":"BLX", "amount":"0.001"}]}}`,
			exp: types.GenesisState{types.Params{MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.NewInt(1)),
				sdk.NewDecCoinFromDec("BLX", sdk.NewDecWithPrec(1, 3)))}},
		},
		"no fee set": {
			src: `{"params":{}}`,
			exp: types.GenesisState{types.Params{MinimumGasPrices: sdk.DecCoins{}}},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, encCfg, subspace := setupTestStore(t)
			m := NewAppModule(subspace)
			m.InitGenesis(ctx, encCfg.Codec, []byte(spec.src))
			gotJson := m.ExportGenesis(ctx, encCfg.Codec)
			var got types.GenesisState
			require.NoError(t, encCfg.Codec.UnmarshalJSON(gotJson, &got))
			assert.Equal(t, spec.exp, got, string(gotJson))
		})
	}
}
