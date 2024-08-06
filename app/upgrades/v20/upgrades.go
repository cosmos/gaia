package v20

import (
	"context"

	providerkeeper "github.com/cosmos/interchain-security/v5/x/ccv/provider/keeper"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/cosmos/gaia/v20/app/keepers"
)

// CreateUpgradeHandler returns an upgrade handler for Gaia v20.
// It performs module migrations, as well as the following tasks:
// - Initializes the MaxProviderConsensusValidators parameter in the provider module to 180.
// - Sets the ValidatorSetCap parameter in the provider module to 180 for all presently registered consumer chains.
// - Increases the MaxValidators parameter in the staking module to 200.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("Starting module migrations...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		InitializeMaxProviderConsensusParam(ctx, keepers.ProviderKeeper)
		InitializeMaxValidatorsForExistingConsumers(ctx, keepers.ProviderKeeper)
		SetMaxValidatorsTo200(ctx, *keepers.StakingKeeper)

		ctx.Logger().Info("Upgrade v20 complete")
		return vm, nil
	}
}

// InitializeMaxValidatorsForExistingConsumers initializes the max validators
// parameter for existing consumers to the MaxProviderConsensusValidators parameter.
// This is necessary to avoid those consumer chains having an excessive amount of validators.
func InitializeMaxValidatorsForExistingConsumers(ctx sdk.Context, providerKeeper providerkeeper.Keeper) {
	maxVals := providerKeeper.GetParams(ctx).MaxProviderConsensusValidators
	for _, chainID := range providerKeeper.GetAllRegisteredConsumerChainIDs(ctx) {
		providerKeeper.SetValidatorSetCap(ctx, chainID, uint32(maxVals))
	}
}

// InitializeMaxProviderConsensusParam initializes the MaxProviderConsensusValidators parameter.
// It is set to 180, which is the current number of validators participating in consensus on the Cosmos Hub.
// This parameter will be used to govern the number of validators participating in consensus on the Cosmos Hub,
// and takes over this role from the MaxValidators parameter in the staking module.
func InitializeMaxProviderConsensusParam(ctx sdk.Context, providerKeeper providerkeeper.Keeper) {
	params := providerKeeper.GetParams(ctx)
	if params.MaxProviderConsensusValidators == 0 {
		params.MaxProviderConsensusValidators = 180
		providerKeeper.SetParams(ctx, params)
	}
}

// SetMaxValidatorsTo200 sets the MaxValidators parameter in the staking module to 200,
// which is the current number of 180 plus 20.
// This is done in concert with the introduction of the inactive-validators feature
// in Interchain Security, after which the number of validators
// participating in consensus on the Cosmos Hub will be governed by the
// MaxProviderConsensusValidators parameter in the provider module.
func SetMaxValidatorsTo200(ctx sdk.Context, stakingKeeper stakingkeeper.Keeper) {
	params, err := stakingKeeper.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	params.MaxValidators = 200

	err = stakingKeeper.SetParams(ctx, params)
	if err != nil {
		panic(err)
	}
}
