package v8

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/gaia/v8/app/keepers"
	icacontrollertypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Fix export genesis error
		atomMetaData, found := keepers.BankKeeper.GetDenomMetaData(ctx, "uatom")
		if !found {
		}
		atomMetaData.Name = "Cosmos Hub Atom"
		atomMetaData.Symbol = "ATOM"

		controllerParams := icacontrollertypes.Params{
			ControllerEnabled: true,
		}

		// allowmessages = [*]
		hostParams := icahosttypes.Params{
			HostEnabled:   true,
			AllowMessages: []string{"*"},
		}

		keepers.ICAHostKeeper.SetParams(ctx, hostParams)
		keepers.ICAControllerKeeper.SetParams(ctx, controllerParams)

		//mauthModule, correctTypecast := mm.Modules[icamauth.ModuleName].(ica.AppModule)
		//if !correctTypecast {
		//	panic("mm.Modules[icamauth.ModuleName] is not of type ica.AppModule")
		//}
		//mauthModule.InitModule(ctx, controllerParams, hostParams)

		ctx.Logger().Info("start to run module migrations...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
