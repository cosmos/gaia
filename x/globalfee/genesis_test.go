package globalfee

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/gaia/v9/x/globalfee/types"
)

func TestDefaultGenesis(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	gotJSON := AppModuleBasic{}.DefaultGenesis(encCfg.Marshaler)
	assert.JSONEq(t, `{"params":{"minimum_gas_prices":[]}}`, string(gotJSON), string(gotJSON))
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
		"zero amount allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"0"}]}}`,
			expErr: false,
		},
		"duplicate denoms not allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"},{"denom":"ALX", "amount":"2"}]}}`,
			expErr: true,
		},
		"negative amounts not allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"-1"}]}}`,
			expErr: true,
		},
		"denom must be sorted": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ZLX", "amount":"1"},{"denom":"ALX", "amount":"2"}]}}`,
			expErr: true,
		},
		"sorted denoms is allowed": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"},{"denom":"ZLX", "amount":"2"}]}}`,
			expErr: false,
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			gotErr := AppModuleBasic{}.ValidateGenesis(encCfg.Marshaler, nil, []byte(spec.src))
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
			exp: types.GenesisState{Params: types.Params{MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.NewInt(1)))}},
		},
		"multiple fee options": {
			src: `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"}, {"denom":"BLX", "amount":"0.001"}]}}`,
			exp: types.GenesisState{Params: types.Params{MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.NewInt(1)),
				sdk.NewDecCoinFromDec("BLX", sdk.NewDecWithPrec(1, 3)))}},
		},
		"no fee set": {
			src: `{"params":{}}`,
			exp: types.GenesisState{Params: types.Params{MinimumGasPrices: sdk.DecCoins{}}},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			ctx, encCfg, subspace := setupTestStore(t)
			m := NewAppModule(subspace)
			m.InitGenesis(ctx, encCfg.Marshaler, []byte(spec.src))
			gotJSON := m.ExportGenesis(ctx, encCfg.Marshaler)
			var got types.GenesisState
			require.NoError(t, encCfg.Marshaler.UnmarshalJSON(gotJSON, &got))
			assert.Equal(t, spec.exp, got, string(gotJSON))
		})
	}
}

func setupTestStore(t *testing.T) (sdk.Context, simappparams.EncodingConfig, paramstypes.Subspace) {
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	encCfg := simapp.MakeTestEncodingConfig()
	keyParams := sdk.NewKVStoreKey(paramstypes.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(paramstypes.TStoreKey)
	ms.MountStoreWithDB(keyParams, storetypes.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, storetypes.StoreTypeTransient, db)
	require.NoError(t, ms.LoadLatestVersion())

	paramsKeeper := paramskeeper.NewKeeper(encCfg.Marshaler, encCfg.Amino, keyParams, tkeyParams)

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	subspace := paramsKeeper.Subspace(ModuleName).WithKeyTable(types.ParamKeyTable())
	return ctx, encCfg, subspace
}
