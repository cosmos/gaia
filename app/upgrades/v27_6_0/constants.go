package v27_6_0

import (
	"github.com/cosmos/gaia/v27/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v27.6.0"

	// MainnetChainID defines the chain ID of the Cosmos Hub mainnet. The IBC
	// client recovery performed by this upgrade only applies there.
	MainnetChainID = "cosmoshub-4"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
