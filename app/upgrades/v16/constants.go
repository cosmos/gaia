package v16

import (
	"github.com/cosmos/gaia/v16/app/upgrades"

	ratelimittypes "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v16"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			ratelimittypes.ModuleName,
		},
	},
}
