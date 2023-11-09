package v15

import (
	"github.com/cosmos/gaia/v15/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v15"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
