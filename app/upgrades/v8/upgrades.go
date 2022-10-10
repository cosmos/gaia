package v8

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v8/app/keepers"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Fix export genesis error
		banktypes.DefaultGenesisState().DenomMetadata = append(banktypes.DefaultGenesisState().DenomMetadata,
			banktypes.Metadata{
				Name:   "Cosmos Hub Atom",
				Symbol: "Atom",
			})
		ctx.Logger().Info(authtypes.DefaultGenesisState().Params.String())

		//controllerParams := icacontrollertypes.Params{}
		//// allowmessages = [*]
		//hostParams := icahosttypes.Params{
		//	HostEnabled:   true,
		//	AllowMessages: []string{"*"},
		//}
		//
		//mauthModule, correctTypecast := mm.Modules[icamauth.ModuleName].(ica.AppModule)
		//if !correctTypecast {
		//	panic("mm.Modules[icamauth.ModuleName] is not of type ica.AppModule")
		//}
		//mauthModule.InitModule(ctx, controllerParams, hostParams)

		ctx.Logger().Info("start to run module migrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
