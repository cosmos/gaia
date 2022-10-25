package v8

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/cosmos/gaia/v8/app/keepers"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("start to run module migrations...")

		// Add atom name and symbol into the bank keeper
		atomMetaData, found := keepers.BankKeeper.GetDenomMetaData(ctx, "uatom")
		if !found {
			return nil, errors.New("atom denom not found")
		}
		atomMetaData.Name = "Cosmos Hub Atom"
		atomMetaData.Symbol = "ATOM"
		keepers.BankKeeper.SetDenomMetaData(ctx, atomMetaData)

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
