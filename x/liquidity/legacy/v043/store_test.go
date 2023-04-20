package v043_test

import (
	"bytes"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v042liquidity "github.com/cosmos/gaia/v9/x/liquidity/legacy/v042"
	v043liquidity "github.com/cosmos/gaia/v9/x/liquidity/legacy/v043"
	"github.com/cosmos/gaia/v9/x/liquidity/types"
)

func TestStoreMigration(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	liquidityKey := sdk.NewKVStoreKey(v042liquidity.ModuleName)
	ctx := testutil.DefaultContext(liquidityKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(liquidityKey)

	_, _, reserveAcc1 := testdata.KeyTestPubAddr()
	_, _, reserveAcc2 := testdata.KeyTestPubAddr()

	// Use dummy value for all keys.
	value := encCfg.Marshaler.MustMarshal(&gogotypes.UInt64Value{Value: 1})

	testCases := []struct {
		name   string
		oldKey []byte
		newKey []byte
	}{
		{
			"reserveAcc1",
			v042liquidity.GetPoolByReserveAccIndexKey(reserveAcc1),
			types.GetPoolByReserveAccIndexKey(reserveAcc1),
		},
		{
			"reserveAcc2",
			v042liquidity.GetPoolByReserveAccIndexKey(reserveAcc2),
			types.GetPoolByReserveAccIndexKey(reserveAcc2),
		},
		{
			"poolBatchIndexKeyPrefix1",
			v042liquidity.GetPoolBatchIndexKey(1),
			nil,
		},
		{
			"poolBatchIndexKeyPrefix2",
			v042liquidity.GetPoolBatchIndexKey(2),
			nil,
		},
	}

	// Set all the old keys to the store
	for _, tc := range testCases {
		store.Set(tc.oldKey, value)
	}

	// Run migrations.
	err := v043liquidity.MigrateStore(ctx, liquidityKey)
	require.NoError(t, err)

	// Make sure the new keys are set and old keys are deleted.
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if !bytes.Equal(tc.oldKey, tc.newKey) {
				require.Nil(t, store.Get(tc.oldKey))
			}
			if tc.newKey != nil {
				require.Equal(t, value, store.Get(tc.newKey))
			}
		})
	}
}
