package v24

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/gaia/v23/app/upgrades"
	liquidtypes "github.com/cosmos/gaia/v23/x/liquid/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v24"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			liquidtypes.ModuleName,
		},
	},
}
