package v12

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v14/app/upgrades"
)

const (
	// UpgradeName defines the on-chain upgrade name.
	UpgradeName = "v12"
)

var (
	// The ValidatorBondFactor dictates the cap on the liquid shares
	// for a validator - determined as a multiple to their validator bond
	// (e.g. ValidatorBondShares = 1000, BondFactor = 250 -> LiquidSharesCap: 250,000)
	ValidatorBondFactor = sdk.NewDec(250)
	// GlobalLiquidStakingCap represents a cap on the portion of stake that
	// comes from liquid staking providers for a specific validator
	ValidatorLiquidStakingCap = sdk.MustNewDecFromStr("0.5") // 50%
	// GlobalLiquidStakingCap represents the percentage cap on
	// the portion of a chain's total stake can be liquid
	GlobalLiquidStakingCap = sdk.MustNewDecFromStr("0.25") // 25%
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
}
