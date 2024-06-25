package v16

import (
	ratelimittypes "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	ibcfeetypes "github.com/cosmos/ibc-go/v8/modules/apps/29-fee/types"

	store "cosmossdk.io/store/types"

	"github.com/cosmos/gaia/v18/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v16"

	RateLimitDenom         = "uatom"
	RateLimitDurationHours = 24
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			ratelimittypes.ModuleName,
			icacontrollertypes.SubModuleName,
			ibcfeetypes.ModuleName,
		},
	},
}
