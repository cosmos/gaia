package bindings

import (
	"context"
	"fmt"

	bindingstypes "github.com/strangelove-ventures/tokenfactory/x/tokenfactory/bindings/types"
	tokenfactorykeeper "github.com/strangelove-ventures/tokenfactory/x/tokenfactory/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

type QueryPlugin struct {
	bankKeeper         bankkeeper.Keeper
	tokenFactoryKeeper *tokenfactorykeeper.Keeper
}

// NewQueryPlugin returns a reference to a new QueryPlugin.
func NewQueryPlugin(b bankkeeper.Keeper, tfk *tokenfactorykeeper.Keeper) *QueryPlugin {
	return &QueryPlugin{
		bankKeeper:         b,
		tokenFactoryKeeper: tfk,
	}
}

// GetDenomAdmin is a query to get denom admin.
func (qp QueryPlugin) GetDenomAdmin(ctx context.Context, denom string) (*bindingstypes.AdminResponse, error) {
	metadata, err := qp.tokenFactoryKeeper.GetAuthorityMetadata(sdk.UnwrapSDKContext(ctx), denom)
	if err != nil {
		return nil, fmt.Errorf("failed to get admin for denom: %s", denom)
	}
	return &bindingstypes.AdminResponse{Admin: metadata.Admin}, nil
}

func (qp QueryPlugin) GetDenomsByCreator(ctx context.Context, creator string) (*bindingstypes.DenomsByCreatorResponse, error) {
	// TODO: validate creator address
	denoms := qp.tokenFactoryKeeper.GetDenomsFromCreator(sdk.UnwrapSDKContext(ctx), creator)
	return &bindingstypes.DenomsByCreatorResponse{Denoms: denoms}, nil
}

func (qp QueryPlugin) GetMetadata(ctx context.Context, denom string) (*bindingstypes.MetadataResponse, error) {
	metadata, found := qp.bankKeeper.GetDenomMetaData(ctx, denom)
	var parsed *bindingstypes.Metadata
	if found {
		parsed = SdkMetadataToWasm(metadata)
	}
	return &bindingstypes.MetadataResponse{Metadata: parsed}, nil
}

func (qp QueryPlugin) GetParams(ctx context.Context) (*bindingstypes.ParamsResponse, error) {
	params := qp.tokenFactoryKeeper.GetParams(sdk.UnwrapSDKContext(ctx))
	return &bindingstypes.ParamsResponse{
		Params: bindingstypes.Params{
			DenomCreationFee: ConvertSdkCoinsToWasmCoins(params.DenomCreationFee),
		},
	}, nil
}
