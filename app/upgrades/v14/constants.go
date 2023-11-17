package v14

import (
	"github.com/cosmos/gaia/v14/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v14"
	// Neutron and Stride consumer chains chain ID
	NeutronChainID = "neutron-1"
	StrideChainID  = "stride-1"
	// Neutron and Stride consumer chains equivocation evidence min height
	EquivocationEvidenceMinHeight = 30
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
