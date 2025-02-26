package v23

import (
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"

	"cosmossdk.io/store/types"

	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"

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
		Added: []string{
			ibcwasmtypes.StoreKey,
		},
		Renamed: nil,
		Deleted: []string{
			ibcfeetypes.StoreKey,
		},
	},
}
