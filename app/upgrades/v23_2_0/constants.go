package v23_2_0 //nolint:revive

import (
	"github.com/cosmos/gaia/v24/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName      = "v23.2.0"
	HexChecksum      = "b92e9904aab2292916507f0db04b7ab6d024c2fdb57a9d52e6725f69b2e684c1"
	MigrateMsgBase64 = "eyJtaWdyYXRpb24iOnsidXBkYXRlX2ZvcmtfcGFyYW1ldGVycyI6eyJnZW5lc2lzX2ZvcmtfdmVyc2lvbiI6IjB4MDAwMDAwMDAiLCJnZW5lc2lzX3Nsb3QiOjAsImFsdGFpciI6eyJ2ZXJzaW9uIjoiMHgwMTAwMDAwMCIsImVwb2NoIjo3NDI0MH0sImJlbGxhdHJpeCI6eyJ2ZXJzaW9uIjoiMHgwMjAwMDAwMCIsImVwb2NoIjoxNDQ4OTZ9LCJjYXBlbGxhIjp7InZlcnNpb24iOiIweDAzMDAwMDAwIiwiZXBvY2giOjE5NDA0OH0sImRlbmViIjp7InZlcnNpb24iOiIweDA0MDAwMDAwIiwiZXBvY2giOjI2OTU2OH0sImVsZWN0cmEiOnsidmVyc2lvbiI6IjB4MDUwMDAwMDAiLCJlcG9jaCI6MzY0MDMyfX19fQ"
	SignerAccount    = "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn"
	ClientID         = "08-wasm-1369"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
