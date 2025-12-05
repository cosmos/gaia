package v26_0_0

import (
	tokenfactorytypes "github.com/cosmos/tokenfactory/x/tokenfactory/types"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/gaia/v26/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v26.0.0"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{tokenfactorytypes.ModuleName},
	},
}
