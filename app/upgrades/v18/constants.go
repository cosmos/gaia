package v18

import (
	feemarkettypes "github.com/skip-mev/feemarket/x/feemarket/types"

	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/cosmos/gaia/v18/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName         = "v18"
	GlobalFeeModuleName = "globalfee"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			feemarkettypes.ModuleName,
		},
		Deleted: []string{
			GlobalFeeModuleName,
		},
	},
}
