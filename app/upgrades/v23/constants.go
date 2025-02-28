package v23

import (
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"

	"cosmossdk.io/store/types"

	"github.com/cosmos/gaia/v23/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName    = "v23"
	IbcFeeStoreKey = "feeibc"
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
			IbcFeeStoreKey,
		},
	},
}
