package v2_test

import (
	"fmt"
	v2 "github.com/cosmos/gaia/v9/x/globalfee/migrations/v2"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/x/params/types"
	gaiahelpers "github.com/cosmos/gaia/v9/app/helpers" // todo this is v9 when other packages gose to v10
	"github.com/cosmos/gaia/v9/x/globalfee"
	globalfeetypes "github.com/cosmos/gaia/v9/x/globalfee/types"
	"github.com/stretchr/testify/require"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestMigrateStore(t *testing.T) {
	appV9 := gaiahelpers.Setup(t)
	ctx := appV9.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})

	globalfeeSubspace := appV9.GetSubspace(globalfee.ModuleName)

	// todo: add this check back when the module is v10
	//_, ok := getBypassMsgTypes(globalfeeSubspace, ctx)
	//require.Equal(t, ok, false)
	//_, ok = getMaxTotalBypassMinFeeMsgGasUsage(globalfeeSubspace, ctx)
	//require.Equal(t, ok, false)

	err := v2.MigrateStore(ctx, globalfeeSubspace)
	require.NoError(t, err)

	bypassMsgTypes, _ := getBypassMsgTypes(globalfeeSubspace, ctx)
	maxGas, _ := getMaxTotalBypassMinFeeMsgGasUsage(globalfeeSubspace, ctx)

	require.Equal(t, bypassMsgTypes, globalfeetypes.DefaultBypassMinFeeMsgTypes)
	require.Equal(t, maxGas, globalfeetypes.DefaultmaxTotalBypassMinFeeMsgGasUsage)
	require.Equal(t, sdk.DecCoins{}, globalfeetypes.DefaultMinGasPrices)
}

func getBypassMsgTypes(globalfeeSubspace sdktypes.Subspace, ctx sdk.Context) ([]string, bool) {
	bypassMsgs := []string{}
	if globalfeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyBypassMinFeeMsgTypes) {
		globalfeeSubspace.Get(ctx, globalfeetypes.ParamStoreKeyBypassMinFeeMsgTypes, &bypassMsgs)
	} else {
		return bypassMsgs, false
	}

	return bypassMsgs, true
}

func getMaxTotalBypassMinFeeMsgGasUsage(globalfeeSubspace sdktypes.Subspace, ctx sdk.Context) (uint64, bool) {
	var maxTotalBypassMinFeeMsgGasUsage uint64
	if globalfeeSubspace.Has(ctx, globalfeetypes.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage) {
		globalfeeSubspace.Get(ctx, globalfeetypes.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage, &maxTotalBypassMinFeeMsgGasUsage)
	} else {
		return maxTotalBypassMinFeeMsgGasUsage, false
	}

	return maxTotalBypassMinFeeMsgGasUsage, true
}
