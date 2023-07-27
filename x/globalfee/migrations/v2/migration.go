package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmos/gaia/v12/x/globalfee/types"
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
func MigrateStore(ctx sdk.Context, globalfeeSubspace paramtypes.Subspace) error {
	var oldGlobalMinGasPrices sdk.DecCoins
	globalfeeSubspace.Get(ctx, types.ParamStoreKeyMinGasPrices, &oldGlobalMinGasPrices)
	defaultParams := types.DefaultParams()
	params := types.Params{
		MinimumGasPrices:                oldGlobalMinGasPrices,
		BypassMinFeeMsgTypes:            defaultParams.BypassMinFeeMsgTypes,
		MaxTotalBypassMinFeeMsgGasUsage: defaultParams.MaxTotalBypassMinFeeMsgGasUsage,
	}

	if !globalfeeSubspace.HasKeyTable() {
		globalfeeSubspace = globalfeeSubspace.WithKeyTable(types.ParamKeyTable())
	}

	globalfeeSubspace.SetParamSet(ctx, &params)

	return nil
}
