package v23

import (
	"cosmossdk.io/store/types"

	"github.com/cosmos/gaia/v23/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v23"
	StoreKey    = "feeibc"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: types.StoreUpgrades{
		Added:   nil,
		Renamed: nil,
		Deleted: []string{
			StoreKey,
		},
	},
}
