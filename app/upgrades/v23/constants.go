package v23

import (
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"

	"cosmossdk.io/store/types"

	"github.com/cosmos/gaia/v23/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v23"
	// RCUpgradeName defines the on-chain upgrade name specifically for the testnet RC upgrade.
	RCUpgradeName           = "23.0.0-rc3"
	IbcFeeStoreKey          = "feeibc"
	ClientUploaderAddress   = "cosmos1raa4kyx5ypz75qqk3566c6slx2mw3qzs5ps5du"
	IBCWasmStoreCodeTypeURL = "/ibc.lightclients.wasm.v1.MsgStoreCode"
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

var RCUpgrade = upgrades.Upgrade{
	UpgradeName:          RCUpgradeName,
	CreateUpgradeHandler: CreateRCUpgradeHandler,
	StoreUpgrades: types.StoreUpgrades{
		Added:   nil,
		Renamed: nil,
		Deleted: nil,
	},
}
