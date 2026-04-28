package ratelimit

import (
	modulev1 "github.com/cosmos/ibc-apps/modules/rate-limiting/v11/api/ratelimit/module/v1"
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v11/keeper"
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v11/types"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"

	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appmodule.Register(&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config       *modulev1.Module
	Cdc          codec.Codec
	StoreService store.KVStoreService
	Subspace     paramstypes.Subspace
	BankKeeper   types.BankKeeper
}

type ModuleOutputs struct {
	depinject.Out

	Keeper *keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	if in.Config.Authority == "" {
		panic("authority for x/ratelimit module must be set")
	}

	authority := authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	k := keeper.NewKeeper(
		in.Cdc,
		in.StoreService,
		in.Subspace,
		authority.String(),
		in.BankKeeper,
		nil,
		nil,
		nil,
	)
	m := NewAppModule(in.Cdc, *k)

	return ModuleOutputs{Keeper: k, Module: m}
}
