package v12

import (
	"github.com/cosmos/gaia/v12/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v12"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
