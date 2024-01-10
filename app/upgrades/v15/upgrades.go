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
// * set the MinCommissionRate param of the staking module to %5
// * update the slashing module SigningInfos records with empty consensus address
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

		MigrateMinCommissionRate(ctx, *keepers.StakingKeeper)
		MigrateSigningInfos(ctx, keepers.SlashingKeeper)

		ctx.Logger().Info("Upgrade v15 complete")
		return vm, err
	}
}

// MigrateMinCommissionRate adheres to prop 826 https://www.mintscan.io/cosmos/proposals/826
// by setting the minimum commission rate staking parameter to 5%
// and updating the commission rate for all validators that have a commission rate less than 5%
func MigrateMinCommissionRate(ctx sdk.Context, sk stakingkeeper.Keeper) {
	params := sk.GetParams(ctx)
	params.MinCommissionRate = sdk.NewDecWithPrec(5, 2)
	if err := sk.SetParams(ctx, params); err != nil {
		panic(err)
	}

	for _, val := range sk.GetAllValidators(ctx) {
		val := val
		if val.Commission.CommissionRates.Rate.LT(sdk.NewDecWithPrec(5, 2)) {
			// set the commmision rate to 5%
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

// MigrateSigningInfos updates validators signing infos for which the consensus address
// is missing using their store key, which contains the consensus address of the validator,
// see https://github.com/cosmos/gaia/issues/1734.
func MigrateSigningInfos(ctx sdk.Context, sk slashingkeeper.Keeper) {
	signingInfos := []slashingtypes.ValidatorSigningInfo{}

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
