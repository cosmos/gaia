package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmos/gaia/v9/x/globalfee/types"
)

// MigrateStore performs in-place params migrations of
// BypassMinFeeMsgTypes and MaxTotalBypassMinFeeMsgGasUsage
// from app.toml to globalfee params.
// The migration includes:
// Add bypass-min-fee-msg-types params that are set
// ["/ibc.core.channel.v1.MsgRecvPacket",
// "/ibc.core.channel.v1.MsgAcknowledgement",
// "/ibc.core.client.v1.MsgUpdateClient",
// "/ibc.core.channel.v1.MsgTimeout",
// "/ibc.core.channel.v1.MsgTimeoutOnClose"] as default and
// add MaxTotalBypassMinFeeMsgGasUsage that is set 1_000_000 as default.
// todo check name, migrateStore or migrateParams?
func MigrateStore(ctx sdk.Context, globalfeeSubspace paramtypes.Subspace) error {
	var globalMinGasPrices sdk.DecCoins

	if globalfeeSubspace.Has(ctx, types.ParamStoreKeyMinGasPrices) {
		globalfeeSubspace.Get(ctx, types.ParamStoreKeyMinGasPrices, &globalMinGasPrices)
	} else {
		// todo return err
		return nil
	}

	var params types.Params

	defaultParams := types.DefaultParams()
	params.MinimumGasPrices = globalMinGasPrices
	params.BypassMinFeeMsgTypes = defaultParams.BypassMinFeeMsgTypes
	params.MaxTotalBypassMinFeeMsgGasUsage = defaultParams.MaxTotalBypassMinFeeMsgGasUsage

	globalfeeSubspace.SetParamSet(ctx, &params)

	return nil
}
