package v8

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/cosmos/gaia/v9/app/upgrades"
	"github.com/cosmos/gaia/v9/x/globalfee"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v8-Rho"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			globalfee.ModuleName,
		},
	},
}
