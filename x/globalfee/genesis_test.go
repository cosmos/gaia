package globalfee

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmos/gaia/v12/x/globalfee/types"
)

func TestDefaultGenesis(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	gotJSON := AppModuleBasic{}.DefaultGenesis(encCfg.Marshaler)
	assert.JSONEq(t,
		`{"params":{"minimum_gas_prices":[],"bypass_min_fee_msg_types":["/ibc.core.channel.v1.MsgRecvPacket","/ibc.core.channel.v1.MsgAcknowledgement","/ibc.core.client.v1.MsgUpdateClient","/ibc.core.channel.v1.MsgTimeout","/ibc.core.channel.v1.MsgTimeoutOnClose"], "max_total_bypass_min_fee_msg_gas_usage":"1000000"}}`,
		string(gotJSON), string(gotJSON))
}

func TestValidateGenesis(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	specs := map[string]struct {
		src    string
		expErr bool
	}{
		"all good": {
			src:    `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"}], "bypass_min_fee_msg_types":["/ibc.core.channel.v1.MsgRecvPacket"]}}`,
			expErr: false,
		},
		"empty minimum": {
			src:    `{"params":{"minimum_gas_prices":[], "bypass_min_fee_msg_types":[]}}`,
			expErr: false,
		},
		"minimum and bypass not set": {
			src:    `{"params":{}}`,
			expErr: false,
		},
		"minimum not set": {
			src:    `{"params":{"bypass_min_fee_msg_types":[]}}`,
			expErr: false,
		},
		"bypass not set": {
			src:    `{"params":{"minimum_gas_prices":[]}}`,
			expErr: false,
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
		"empty bypass msg types not allowed": {
			src:    `{"params":{"bypass_min_fee_msg_types":[""]}}`,
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
			src: `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"}], "bypass_min_fee_msg_types":["/ibc.core.channel.v1.MsgRecvPacket"]}}`,
			exp: types.GenesisState{
				Params: types.Params{
					MinimumGasPrices:     sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.NewInt(1))),
					BypassMinFeeMsgTypes: []string{"/ibc.core.channel.v1.MsgRecvPacket"},
				},
			},
		},
		"multiple fee options": {
			src: `{"params":{"minimum_gas_prices":[{"denom":"ALX", "amount":"1"}, {"denom":"BLX", "amount":"0.001"}], "bypass_min_fee_msg_types":["/ibc.core.channel.v1.MsgRecvPacket","/ibc.core.channel.v1.MsgTimeoutOnClose"]}}`,
			exp: types.GenesisState{
				Params: types.Params{
					MinimumGasPrices: sdk.NewDecCoins(sdk.NewDecCoin("ALX", sdk.NewInt(1)),
						sdk.NewDecCoinFromDec("BLX", sdk.NewDecWithPrec(1, 3))),
					BypassMinFeeMsgTypes: []string{"/ibc.core.channel.v1.MsgRecvPacket", "/ibc.core.channel.v1.MsgTimeoutOnClose"},
				},
			},
		},
		"no fee set": {
			src: `{"params":{}}`,
			exp: types.GenesisState{
				Params: types.Params{
					MinimumGasPrices:     sdk.DecCoins{},
					BypassMinFeeMsgTypes: []string{},
				},
			},
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
	t.Helper()
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
