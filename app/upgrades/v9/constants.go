//go:build upgrade_v9

package v9

import (
	store "cosmossdk.io/store/types"
	"github.com/cosmos/gaia/v18/app/upgrades"
	ccvprovider "github.com/cosmos/interchain-security/v5/x/ccv/provider/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v9-Lambda"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			ccvprovider.ModuleName,
		},
	},
}
