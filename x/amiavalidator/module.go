package amiavalidator

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	types2 "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gaia/v25/telemetry"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"cosmossdk.io/core/appmodule"
)

const ModuleName = "amiavalidator"

var _ appmodule.HasPreBlocker = &Module{}

type Module struct {
	oc *telemetry.OtelClient
	sk *stakingkeeper.Querier
}

func (m Module) Name() string {
	return ModuleName
}

func (m Module) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {}

func (m Module) RegisterInterfaces(registry types.InterfaceRegistry) {}

func (m Module) RegisterGRPCGatewayRoutes(c client.Context, mux *runtime.ServeMux) {}

func NewAppModule(sk *stakingkeeper.Querier, oc *telemetry.OtelClient) *Module {
	return &Module{
		oc: oc,
		sk: sk,
	}
}

func (m Module) IsOnePerModuleType() {}

func (m Module) IsAppModule() {}

func (m Module) PreBlock(ctx context.Context) (appmodule.ResponsePreBlock, error) {
	addr := m.oc.GetValAddr()
	val, err := m.sk.GetValidatorByConsAddr(ctx, sdk.ConsAddress(addr))
	if err == nil {
		isVal := val.GetStatus() == types2.Bonded
		m.oc.SetValidatorStatus(isVal)
	}
	return sdk.ResponsePreBlock{}, nil
}
