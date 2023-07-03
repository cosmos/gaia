package v9

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	ccvprovider "github.com/cosmos/interchain-security/v2/x/ccv/provider/types"

	"github.com/cosmos/gaia/v11/app/upgrades"
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
