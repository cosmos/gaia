package v29_4_0

import (
	"github.com/cosmos/gaia/v29/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v29.4.0"

	// RefundRecipient is ch_user from prop 1056.
	RefundRecipient = "cosmos1dd25c4sshelrpfs0433apg24c5phrhk8m96c4n"

	// RefundAmount is 1,221,120 ATOM in uatom.
	RefundAmount = 1_221_120_000_000

	// RefundUnlockTime is 2028-10-01 00:00 UTC. The refund can be staked but
	// not transferred until then.
	RefundUnlockTime = 1853971200
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
