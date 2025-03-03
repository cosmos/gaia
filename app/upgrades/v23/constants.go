package v23

import (
	_ "embed"

	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"

	"cosmossdk.io/store/types"

	"github.com/cosmos/gaia/v23/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName    = "v23"
	IbcFeeStoreKey = "feeibc"

	ExpectedEthLightClientChecksum = "f82549f5bc8adaef18e5ce4f5b68269947343742c938dac322faf1583319172c"
)

//go:embed cw_ics08_wasm_eth.wasm.gz
var ethWasmLightClient []byte

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: types.StoreUpgrades{
		Added: []string{
			ibcwasmtypes.StoreKey,
		},
		Renamed: nil,
		Deleted: []string{
			IbcFeeStoreKey,
		},
	},
}
