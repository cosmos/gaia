package v2_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	v2 "github.com/cosmos/gaia/v11/x/globalfee/migrations/v2"
	globalfeetypes "github.com/cosmos/gaia/v11/x/globalfee/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
)

func TestMigrateStore(t *testing.T) {
	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)

	storeKey := sdk.NewKVStoreKey(paramtypes.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey("mem_key")

	stateStore.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, sdk.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	require.NoError(t, stateStore.LoadLatestVersion())

	// Create new empty subspace
	newSubspace := paramtypes.NewSubspace(cdc,
		codec.NewLegacyAmino(),
		storeKey,
		memStoreKey,
		paramtypes.ModuleName,
	)

	// register the subspace with the v11 paramKeyTable
	newSubspace = newSubspace.WithKeyTable(globalfeetypes.ParamKeyTable())

	// check MinGasPrices isn't set
	_, ok := getMinGasPrice(newSubspace, ctx)
	require.Equal(t, ok, false)

	// set a minGasPrice different that default value
	minGasPrices := sdk.NewDecCoins(sdk.NewDecCoin("uatom", sdk.OneInt()))
	newSubspace.Set(ctx, globalfeetypes.ParamStoreKeyMinGasPrices, minGasPrices)
	require.False(t, minGasPrices.IsEqual(globalfeetypes.DefaultMinGasPrices))

	// check that the new parameters aren't set
	_, ok = getBypassMsgTypes(newSubspace, ctx)
	require.Equal(t, ok, false)
	_, ok = getMaxTotalBypassMinFeeMsgGasUsage(newSubspace, ctx)
	require.Equal(t, ok, false)

	// run global fee migration
	err := v2.MigrateStore(ctx, newSubspace)
	require.NoError(t, err)

	newMinGasPrices, _ := getMinGasPrice(newSubspace, ctx)
	bypassMsgTypes, _ := getBypassMsgTypes(newSubspace, ctx)
	maxGas, _ := getMaxTotalBypassMinFeeMsgGasUsage(newSubspace, ctx)

	require.Equal(t, bypassMsgTypes, globalfeetypes.DefaultBypassMinFeeMsgTypes)
	require.Equal(t, maxGas, globalfeetypes.DefaultmaxTotalBypassMinFeeMsgGasUsage)
	require.Equal(t, minGasPrices, newMinGasPrices)
}

func getBypassMsgTypes(globalfeeSubspace paramtypes.Subspace, ctx sdk.Context) ([]string, bool) {
	bypassMsgs := []string{}
	if globalfeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyBypassMinFeeMsgTypes) {
		globalfeeSubspace.Get(ctx, globalfeetypes.ParamStoreKeyBypassMinFeeMsgTypes, &bypassMsgs)
	} else {
		return bypassMsgs, false
	}

	return bypassMsgs, true
}

func getMaxTotalBypassMinFeeMsgGasUsage(globalfeeSubspace paramtypes.Subspace, ctx sdk.Context) (uint64, bool) {
	var maxTotalBypassMinFeeMsgGasUsage uint64
	if globalfeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage) {
		globalfeeSubspace.Get(ctx, globalfeetypes.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage, &maxTotalBypassMinFeeMsgGasUsage)
	} else {
		return maxTotalBypassMinFeeMsgGasUsage, false
	}

	return maxTotalBypassMinFeeMsgGasUsage, true
}

func getMinGasPrice(globalfeeSubspace paramtypes.Subspace, ctx sdk.Context) (sdk.DecCoins, bool) {
	var globalMinGasPrices sdk.DecCoins
	if globalfeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMinGasPrices) {
		globalfeeSubspace.Get(ctx, globalfeetypes.ParamStoreKeyMinGasPrices, &globalMinGasPrices)
	} else {
		return globalMinGasPrices, false
	}

	return globalMinGasPrices, true
}
