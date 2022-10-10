package v8

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	group "github.com/cosmos/cosmos-sdk/x/group"

	"github.com/cosmos/gaia/v8/app/upgrades"
	"github.com/cosmos/gaia/v8/x/globalfee"
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
			group.ModuleName,
			globalfee.ModuleName,
		},
	},
}
