package v8

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/controller/types"
	icahosttypes "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/host/types"

	"github.com/cosmos/gaia/v8/app/keepers"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Add atom name and symbol into the bank keeper
		atomMetaData, found := keepers.BankKeeper.GetDenomMetaData(ctx, "uatom")
		if !found {
			return nil, errors.New("atom denom not found")
		}
		atomMetaData.Name = "Cosmos Hub Atom"
		atomMetaData.Symbol = "ATOM"
		keepers.BankKeeper.SetDenomMetaData(ctx, atomMetaData)

		// Enable controller chain
		controllerParams := icacontrollertypes.Params{
			ControllerEnabled: true,
		}

		// Change hostParams allow_messages = [*] instead of whitelisting individual messages
		hostParams := icahosttypes.Params{
			HostEnabled:   true,
			AllowMessages: []string{"*"},
		}

		// Update params for host & controller keepers
		keepers.ICAHostKeeper.SetParams(ctx, hostParams)
		keepers.ICAControllerKeeper.SetParams(ctx, controllerParams)

		ctx.Logger().Info("start to run module migrations...")

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
