package v29_0_0

import (
	store "cosmossdk.io/store/types"

	"github.com/cosmos/gaia/v29/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName      = "v29.0.0"
	providerStoreKey = "provider"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Deleted: []string{
			providerStoreKey,
		},
	},
}
