// Package gov provides a custom wrapper around the SDK gov module
// that adds vote validation at the MsgServer level. This ensures
// that all governance vote messages (from user transactions, ICA,
// wasm, authz, or any future mechanism) are validated for stake
// requirements before being processed.
package gov

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// AppModule wraps the SDK gov module to add custom MsgServer
// that validates voter stake for all vote messages.
type AppModule struct {
	gov.AppModule
	keeper        *govkeeper.Keeper
	accountKeeper govtypes.AccountKeeper
	stakingKeeper *stakingkeeper.Keeper

	// legacySubspace is used solely for migration of x/params managed parameters
	legacySubspace govtypes.ParamSubspace
}

// NewAppModule creates a new AppModule object that wraps the SDK gov module
// with additional vote validation.
func NewAppModule(
	cdc codec.Codec,
	keeper *govkeeper.Keeper,
	ak govtypes.AccountKeeper,
	bk govtypes.BankKeeper,
	stakingKeeper *stakingkeeper.Keeper,
	ss govtypes.ParamSubspace,
) AppModule {
	return AppModule{
		AppModule:      gov.NewAppModule(cdc, keeper, ak, bk, ss),
		keeper:         keeper,
		accountKeeper:  ak,
		stakingKeeper:  stakingKeeper,
		legacySubspace: ss,
	}
}

// RegisterServices registers module services with custom MsgServer implementations
// that include vote validation.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// Create our decorated MsgServer that validates votes
	msgServer := NewMsgServerImpl(am.keeper, am.stakingKeeper)

	// Register the decorated v1 MsgServer
	govv1.RegisterMsgServer(cfg.MsgServer(), msgServer)

	// Register the legacy v1beta1 MsgServer - it delegates to our decorated v1 server
	govv1beta1.RegisterMsgServer(cfg.MsgServer(), govkeeper.NewLegacyMsgServerImpl(
		am.accountKeeper.GetModuleAddress(govtypes.ModuleName).String(),
		msgServer,
	))

	// Register query servers normally
	legacyQueryServer := govkeeper.NewLegacyQueryServer(am.keeper)
	govv1beta1.RegisterQueryServer(cfg.QueryServer(), legacyQueryServer)
	govv1.RegisterQueryServer(cfg.QueryServer(), govkeeper.NewQueryServer(am.keeper))

	// Handle migrations (same as SDK gov module)
	m := govkeeper.NewMigrator(am.keeper, am.legacySubspace)
	if err := cfg.RegisterMigration(govtypes.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/gov from version 1 to 2: %v", err))
	}

	if err := cfg.RegisterMigration(govtypes.ModuleName, 2, m.Migrate2to3); err != nil {
		panic(fmt.Sprintf("failed to migrate x/gov from version 2 to 3: %v", err))
	}

	if err := cfg.RegisterMigration(govtypes.ModuleName, 3, m.Migrate3to4); err != nil {
		panic(fmt.Sprintf("failed to migrate x/gov from version 3 to 4: %v", err))
	}

	if err := cfg.RegisterMigration(govtypes.ModuleName, 4, m.Migrate4to5); err != nil {
		panic(fmt.Sprintf("failed to migrate x/gov from version 4 to 5: %v", err))
	}
}
