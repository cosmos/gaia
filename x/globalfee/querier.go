package globalfee

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v12/x/globalfee/types"
)

var _ types.QueryServer = &GrpcQuerier{}

// ParamSource is a read only subset of paramtypes.Subspace
type ParamSource interface {
	Get(ctx sdk.Context, key []byte, ptr interface{})
	Has(ctx sdk.Context, key []byte) bool
}

type GrpcQuerier struct {
	paramSource ParamSource
}

func NewGrpcQuerier(paramSource ParamSource) GrpcQuerier {
	return GrpcQuerier{paramSource: paramSource}
}

// MinimumGasPrices return minimum gas prices
func (g GrpcQuerier) Params(stdCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	var minGasPrices sdk.DecCoins
	var bypassMinFeeMsgTypes []string
	var maxTotalBypassMinFeeMsgGasUsage uint64
	ctx := sdk.UnwrapSDKContext(stdCtx)

	// todo: if return err if not exist?
	if g.paramSource.Has(ctx, types.ParamStoreKeyMinGasPrices) {
		g.paramSource.Get(ctx, types.ParamStoreKeyMinGasPrices, &minGasPrices)
	}
	if g.paramSource.Has(ctx, types.ParamStoreKeyBypassMinFeeMsgTypes) {
		g.paramSource.Get(ctx, types.ParamStoreKeyBypassMinFeeMsgTypes, &bypassMinFeeMsgTypes)
	}
	if g.paramSource.Has(ctx, types.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage) {
		g.paramSource.Get(ctx, types.ParamStoreKeyMaxTotalBypassMinFeeMsgGasUsage, &maxTotalBypassMinFeeMsgGasUsage)
	}

	return &types.QueryParamsResponse{
		Params: types.Params{
			MinimumGasPrices:                minGasPrices,
			BypassMinFeeMsgTypes:            bypassMinFeeMsgTypes,
			MaxTotalBypassMinFeeMsgGasUsage: maxTotalBypassMinFeeMsgGasUsage,
		},
	}, nil
}
