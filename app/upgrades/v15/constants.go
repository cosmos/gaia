package v15

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"

	"github.com/cosmos/gaia/v15/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v15"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			// v47 modules
			crisistypes.ModuleName,
			consensustypes.ModuleName,
		},
	},
}
