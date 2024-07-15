//go:build upgrade_v8

package v8

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/cosmos/gaia/v19/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName         = "v8-Rho"
	GlobalFeeModuleName = "globalfee"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			GlobalFeeModuleName,
		},
	},
}
