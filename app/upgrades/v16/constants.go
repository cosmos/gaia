package v16

import (
	ratelimittypes "github.com/Stride-Labs/ibc-rate-limiting/ratelimit/types"

	icacontrollertypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/types"

	store "github.com/cosmos/cosmos-sdk/store/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/gaia/v16/app/upgrades"
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
			wasmtypes.ModuleName,
		},
	},
}
