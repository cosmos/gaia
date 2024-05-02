package globalfee

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/cosmos/gaia/v17/x/globalfee/client/cli"
	"github.com/cosmos/gaia/v17/x/globalfee/keeper"
	"github.com/cosmos/gaia/v17/x/globalfee/types"
)

var (
	_ module.AppModuleBasic   = AppModuleBasic{}
	_ module.AppModuleGenesis = AppModule{}
	_ module.AppModule        = AppModule{}
)

// AppModuleBasic defines the basic application module used by the wasm module.
type AppModuleBasic struct{}

func (a AppModuleBasic) Name() string {
	return types.ModuleName
}

func (a AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&types.GenesisState{
		Params: types.DefaultParams(),
	})
}

func (a AppModuleBasic) ValidateGenesis(marshaler codec.JSONCodec, _ client.TxEncodingConfig, message json.RawMessage) error {
	var data types.GenesisState
	err := marshaler.UnmarshalJSON(message, &data)
	if err != nil {
		return err
	}
	if err := data.Params.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "params")
	}
	return nil
}

func (a AppModuleBasic) RegisterInterfaces(_ codectypes.InterfaceRegistry) {
}

func (a AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {
}

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
	if err != nil {
		// same behavior as in cosmos-sdk
		panic(err)
	}
}

func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

func (a AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {
}

type AppModule struct {
	AppModuleBasic
	paramSpace paramstypes.Subspace
}

// NewAppModule constructor
func NewAppModule(paramSpace paramstypes.Subspace) *AppModule {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return &AppModule{paramSpace: paramSpace}
}

func (a AppModule) InitGenesis(ctx sdk.Context, marshaler codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	marshaler.MustUnmarshalJSON(message, &genesisState)

	a.paramSpace.SetParamSet(ctx, &genesisState.Params)
	return nil
}

func (a AppModule) ExportGenesis(ctx sdk.Context, marshaler codec.JSONCodec) json.RawMessage {
	var genState types.GenesisState
	a.paramSpace.GetParamSet(ctx, &genState.Params)
	return marshaler.MustMarshalJSON(&genState)
}

func (a AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {
}

func (a AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), NewGrpcQuerier(a.paramSpace))

	m := keeper.NewMigrator(a.paramSpace)
	if err := cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/globalfee from version 1 to 2: %v", err))
	}
}

func (a AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {
}

func (a AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return nil
}

// ConsensusVersion is a sequence number for state-breaking change of the
// module. It should be incremented on each consensus-breaking change
// introduced by the module. To avoid wrong/empty versions, the initial version
// should be set to 1.
func (a AppModule) ConsensusVersion() uint64 {
	return 2
}
