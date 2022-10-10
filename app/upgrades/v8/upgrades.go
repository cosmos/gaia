package v8

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v8/app/keepers"
	"github.com/cosmos/gaia/v8/x/globalfee"
	globalfeetypes "github.com/cosmos/gaia/v8/x/globalfee/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// globalfee params
		globalFeeParams := globalfeetypes.DefaultParams()
		globalFeeModule, err := mm.Modules[globalfee.ModuleName].(globalfee.AppModule)
		if err {
			panic("mm.Modules[globalfee.ModuleName] is not of type globalfee.AppModule")
		}
		globalFeeModule.InitModule(ctx, globalFeeParams)
		// override versions for _new_ modules as to not skip InitGenesis...only for existing module
		//vm[group.ModuleName] = 0

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
