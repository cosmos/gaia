package v15

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/cosmos/gaia/v15/app/keepers"
)

// CreateUpgradeHandler returns a upgrade handler for Gaia v15
// which executes the following migrations:
//   - adhere to prop 826 which sets the minimum commission rate to 5% for all validators,
//     see https://www.mintscan.io/cosmos/proposals/826
//   - update the slashing module SigningInfos for which the consensus address is empty,
//     see https://github.com/cosmos/gaia/issues/1734.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx.Logger().Info("Starting module migrations...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		UpgradeMinCommissionRate(ctx, *keepers.StakingKeeper)
		UpgradeSigningInfos(ctx, keepers.SlashingKeeper)

		ctx.Logger().Info("Upgrade v15 complete")
		return vm, err
	}
}

// UpgradeMinCommissionRate sets the minimum commission rate staking parameter to 5%
// and updates the commission rate for all validators that have a commission rate less than 5%
func UpgradeMinCommissionRate(ctx sdk.Context, sk stakingkeeper.Keeper) {
	params := sk.GetParams(ctx)
	params.MinCommissionRate = sdk.NewDecWithPrec(5, 2)
	err := sk.SetParams(ctx, params)
	if err != nil {
		panic(err)
	}

	for _, val := range sk.GetAllValidators(ctx) {
		if val.Commission.CommissionRates.Rate.LT(sdk.NewDecWithPrec(5, 2)) {
			// set the commission rate to 5%
			val.Commission.CommissionRates.Rate = sdk.NewDecWithPrec(5, 2)
			// set the max rate to 5% if it is less than 5%
			if val.Commission.CommissionRates.MaxRate.LT(sdk.NewDecWithPrec(5, 2)) {
				val.Commission.CommissionRates.MaxRate = sdk.NewDecWithPrec(5, 2)
			}
			val.Commission.UpdateTime = ctx.BlockHeader().Time
			sk.SetValidator(ctx, val)
		}
	}
}

// UpgradeSigningInfos updates the signing infos of validators for which
// the consensus address is missing
func UpgradeSigningInfos(ctx sdk.Context, sk slashingkeeper.Keeper) {
	signingInfos := []slashingtypes.ValidatorSigningInfo{}

	// update consensus address in signing info
	// using the store key of validators
	sk.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info slashingtypes.ValidatorSigningInfo) (stop bool) {
		if info.Address == "" {
			info.Address = address.String()
			signingInfos = append(signingInfos, info)
		}

		return false
	})

	for _, si := range signingInfos {
		addr, err := sdk.ConsAddressFromBech32(si.Address)
		if err != nil {
			panic(err)
		}
		sk.SetValidatorSigningInfo(ctx, addr, si)
	}
}
