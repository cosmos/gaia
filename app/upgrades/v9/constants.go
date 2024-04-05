//go:build upgrade_v9

package v9

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	ccvprovider "github.com/cosmos/interchain-security/v4/x/ccv/provider/types"

	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/cosmos/gaia/v16/app/upgrades"
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
