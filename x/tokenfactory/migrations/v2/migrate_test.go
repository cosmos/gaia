package v2_test

import (
	"testing"

	"github.com/cosmos/gaia/v23/x/tokenfactory"
	"github.com/cosmos/gaia/v23/x/tokenfactory/exported"
	v2 "github.com/cosmos/gaia/v23/x/tokenfactory/migrations/v2"
	"github.com/cosmos/gaia/v23/x/tokenfactory/types"
	"github.com/stretchr/testify/require"

	sdkstore "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

type mockSubspace struct {
	ps types.Params
}

func newMockSubspace(ps types.Params) mockSubspace {
	return mockSubspace{ps: ps}
}

func (ms mockSubspace) GetParamSet(_ sdk.Context, ps exported.ParamSet) {
	*ps.(*types.Params) = ms.ps
}

func TestMigrate(t *testing.T) {
	// x/param conversion
	encCfg := moduletestutil.MakeTestEncodingConfig(tokenfactory.AppModuleBasic{})
	cdc := encCfg.Codec

	storeKey := sdkstore.NewKVStoreKey(v2.ModuleName)
	tKey := sdkstore.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	legacySubspace := newMockSubspace(types.Params{
		DenomCreationFee:        nil,
		DenomCreationGasConsume: 2_000_000,
	})
	require.NoError(t, v2.Migrate(ctx, store, legacySubspace, cdc))

	var res types.Params
	bz := store.Get(v2.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &res))
	require.Equal(t, legacySubspace.ps, res)
}
