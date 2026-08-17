package v29_0_0

import (
	"github.com/cosmos/gaia/v29/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v29.0.0"

	// providerStoreKey is the deprecated ICS "provider" kv-store. Its
	// contents are wiped inside CreateUpgradeHandler (a plain KV iterate+
	// delete over ctx.KVStore) rather than via StoreUpgrades.Deleted:
	// StoreUpgrades.Deleted only recognizes a deletion at the exact restart
	// transitioning into the upgrade height, and permanently excludes the
	// store from CommitInfo from that commit onward. Any later restart that
	// still mounts this key (see app/keepers/keys.go) would then hit a
	// "version mismatch" LoadVersion error. Wiping contents as an ordinary
	// state mutation keeps "provider" appearing (now empty) in CommitInfo at
	// every future height, so no future restart/export/state-sync can ever
	// mismatch, and the keys.go mount can be dropped safely in any later
	// release with no urgency.
	providerStoreKey = "provider"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
