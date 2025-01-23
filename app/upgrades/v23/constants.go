package v23

import (
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/gaia/v23/app/upgrades"
	lsmtypes "github.com/cosmos/gaia/v23/x/lsm/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v23"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			lsmtypes.ModuleName,
		},
	},
}
