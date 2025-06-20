package amiavalidator

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gaia/v25/telemetry"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"cosmossdk.io/core/appmodule"
)

const ModuleName = "amiavalidator"

var _ appmodule.HasPreBlocker = &Module{}

// Module is the module for the amiavalidator module.
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

func (m Module) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {}

func (m Module) RegisterInterfaces(registry types.InterfaceRegistry) {}

func (m Module) RegisterGRPCGatewayRoutes(c client.Context, mux *runtime.ServeMux) {}

func (m Module) IsOnePerModuleType() {}

func (m Module) IsAppModule() {}

func (m Module) PreBlock(ctx context.Context) (appmodule.ResponsePreBlock, error) {
	if m.oc.Enabled() {
		// TODO(technicallyty): maybe we can just update it every 20 blocks, so we don't have to query the val every block?
		if sdk.UnwrapSDKContext(ctx).BlockHeight()%20 == 0 {
			addr := m.oc.GetValAddr()
			val, err := m.sk.GetValidatorByConsAddr(ctx, sdk.ConsAddress(addr))
			if err == nil {
				isVal := val.GetStatus() == stakingtypes.Bonded
				m.oc.SetValidatorStatus(isVal)
			}
		}

	}
	return sdk.ResponsePreBlock{}, nil
}
