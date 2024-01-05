package v15

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/cosmos/gaia/v15/app/keepers"
)

// adhere to prop 826 which sets the minimum commission rate to 5% for all validators
// https://www.mintscan.io/cosmos/proposals/826
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

		V15UpgradeHandler(ctx, keepers)

		ctx.Logger().Info("Upgrade v15 complete")
		return vm, err
	}
}

// V15UpgradeHandler sets the minimum commission rate staking parameter to 5%
// and updates the commission rate for all validators that have a commission rate less than 5%
//
// TODO: rename func name
// refactor in multiple function
func V15UpgradeHandler(ctx sdk.Context, keepers *keepers.AppKeepers) {
	params := keepers.StakingKeeper.GetParams(ctx)
	params.MinCommissionRate = sdk.NewDecWithPrec(5, 2)
	keepers.StakingKeeper.SetParams(ctx, params)

	for _, val := range keepers.StakingKeeper.GetAllValidators(ctx) {
		val := val
		if val.Commission.CommissionRates.Rate.LT(sdk.NewDecWithPrec(5, 2)) {
			// set the commmision rate to 5%
			val.Commission.CommissionRates.Rate = sdk.NewDecWithPrec(5, 2)
			// set the max rate to 5% if it is less than 5%
			if val.Commission.CommissionRates.MaxRate.LT(sdk.NewDecWithPrec(5, 2)) {
				val.Commission.CommissionRates.MaxRate = sdk.NewDecWithPrec(5, 2)
			}
			val.Commission.UpdateTime = ctx.BlockHeader().Time
			keepers.StakingKeeper.SetValidator(ctx, val)
		}
	}

	// TODO: call ValidatorSigningInfosFix
}

// UpgradeValidatorSigningInfos upgrades the validators signing infos for which the consensus address
// is missing using their store key, which contains the consensus address of the validator
// TODO: add more context
// , see https://github.com/cosmos/gaia/issues/1734.
func ValidatorSigningInfosFix(ctx sdk.Context, sk slashingkeeper.Keeper) {
	signingInfos := []slashingtypes.ValidatorSigningInfo{}

	sk.IterateValidatorSigningInfos(ctx, func(address sdk.ConsAddress, info types.ValidatorSigningInfo) (stop bool) {
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
