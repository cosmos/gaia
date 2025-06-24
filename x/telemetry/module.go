package telemetry

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v25/telemetry"
)

const (
	ModuleName = "telemetry"

	// valUpdateBlockRate determines how many blocks we wait before updating the validator status.
	valUpdateBlockRate = 20
)

var _ appmodule.HasPreBlocker = &Module{}

// Module is the module for the telemetry module.
// Its only responsibility is to report to the otel client if this node is a validator via PreBlocker.
type Module struct {
	oc *telemetry.OtelClient
	sk *stakingkeeper.Querier
}

func NewAppModule(sk *stakingkeeper.Querier, oc *telemetry.OtelClient) *Module {
	return &Module{
		oc: oc,
		sk: sk,
	}
}

func (m Module) Name() string {
	return ModuleName
}

func (m Module) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

func (m Module) RegisterInterfaces(_ types.InterfaceRegistry) {}

func (m Module) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}

func (m Module) IsOnePerModuleType() {}

func (m Module) IsAppModule() {}

func (m Module) PreBlock(ctx context.Context) (appmodule.ResponsePreBlock, error) {
	if m.oc.Enabled() {
		if sdk.UnwrapSDKContext(ctx).BlockHeight()%valUpdateBlockRate == 0 {
			addr := m.oc.GetValAddr()
			if addr != nil {
				val, err := m.sk.GetValidatorByConsAddr(ctx, sdk.ConsAddress(addr))
				if err == nil {
					isVal := val.GetStatus() == stakingtypes.Bonded
					m.oc.SetValidatorStatus(isVal)
				}
			}
		}
	}
	return sdk.ResponsePreBlock{}, nil
}
