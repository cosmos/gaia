package v8

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/x/group"

	"github.com/cosmos/gaia/v8/app/upgrades"
	"github.com/cosmos/gaia/v8/x/globalfee"
	icamauth "github.com/cosmos/gaia/v8/x/icamauth/types"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v8-Rho"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			globalfee.ModuleName,
			group.ModuleName,
			icamauth.ModuleName,
		},
	},
}
