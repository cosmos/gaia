// Package wasm provides a custom wrapper around the CosmWasm wasmd module
// that temporarily blocks MsgStoreCode, MsgInstantiateContract,
// MsgInstantiateContract2, MsgStoreAndInstantiateContract, and
// MsgStoreAndMigrateContract at the MsgServer level. This ensures the block
// applies regardless of how the message was dispatched (direct tx, ICA,
// authz, wasm-emitted submessages, gov-executed proposals, or any future
// mechanism) — see msg_server.go and ante/wasm_disable_ante.go for the full
// rationale.
package wasm

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/CosmWasm/wasmd/x/wasm/exported"
	"github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/CosmWasm/wasmd/x/wasm/simulation"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// AppModule wraps the wasmd module to intercept RegisterServices and
// register our blocking MsgServer in place of the stock one.
type AppModule struct {
	wasm.AppModule
	keeper *keeper.Keeper

	// legacySubspace is used solely for migration of x/params managed parameters.
	legacySubspace exported.Subspace
}

// NewAppModule creates a new AppModule object that wraps the wasmd module
// with the store/instantiate block. Arguments are forwarded unchanged to
// wasm.NewAppModule.
func NewAppModule(
	cdc codec.Codec,
	k *keeper.Keeper,
	validatorSetSource keeper.ValidatorSetSource,
	ak wasmtypes.AccountKeeper,
	bk simulation.BankKeeper,
	router *baseapp.MsgServiceRouter,
	ss exported.Subspace,
) AppModule {
	return AppModule{
		AppModule:      wasm.NewAppModule(cdc, k, validatorSetSource, ak, bk, router, ss),
		keeper:         k,
		legacySubspace: ss,
	}
}

// RegisterServices overrides wasmd's RegisterServices to register our
// wrapped MsgServer; everything else (query server, migrations) mirrors
// wasmd's own AppModule.RegisterServices (x/wasm/module.go) exactly.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	wasmtypes.RegisterMsgServer(cfg.MsgServer(), NewMsgServerImpl(keeper.NewMsgServerImpl(am.keeper)))
	wasmtypes.RegisterQueryServer(cfg.QueryServer(), keeper.Querier(am.keeper))

	m := keeper.NewMigrator(*am.keeper, am.legacySubspace)
	if err := cfg.RegisterMigration(wasmtypes.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(err)
	}
	if err := cfg.RegisterMigration(wasmtypes.ModuleName, 2, m.Migrate2to3); err != nil {
		panic(err)
	}
	if err := cfg.RegisterMigration(wasmtypes.ModuleName, 3, m.Migrate3to4); err != nil {
		panic(err)
	}
}
