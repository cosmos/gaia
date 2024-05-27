package v18

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/cosmos/gaia/v18/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v18"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			wasmtypes.ModuleName,
		},
	},
}
