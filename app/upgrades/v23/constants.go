package v23

import (
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"

	"cosmossdk.io/store/types"

	"github.com/cosmos/gaia/v23/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v23"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: types.StoreUpgrades{
		Added:   nil,
		Renamed: nil,
		Deleted: []string{
			ibcfeetypes.StoreKey,
		},
	},
}
