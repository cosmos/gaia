package app

import (
	"fmt"

	"github.com/cosmos/gaia/v23/x/tokenfactory/app/upgrades"
	"github.com/cosmos/gaia/v23/x/tokenfactory/app/upgrades/noop"

	upgradetypes "cosmossdk.io/x/upgrade/types"
)

// Upgrades list of chain upgrades
var Upgrades = []upgrades.Upgrade{}

// RegisterUpgradeHandlers registers the chain upgrade handlers
func (app TokenFactoryApp) RegisterUpgradeHandlers() {
	if len(Upgrades) == 0 {
		// always have a unique upgrade registered for the current version to test in system tests
		Upgrades = append(Upgrades, noop.NewUpgrade(app.Version()))
	}

	keepers := upgrades.AppKeepers{AccountKeeper: app.AccountKeeper}
	// register all upgrade handlers
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(
				app.ModuleManager,
				app.configurator,
				&keepers,
			),
		)
	}

	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	// register store loader for current upgrade
	for _, upgrade := range Upgrades {
		if upgradeInfo.Name == upgrade.UpgradeName {
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrade.StoreUpgrades)) // nolint:gosec
			break
		}
	}
}
