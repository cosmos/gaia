package v28_0_0

import (
	"context"
	"fmt"
	"math"

	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/gaia/v28/app/keepers"
)

// providerStoreKey is the store key used by the ICS provider module.
// Hardcoded to avoid importing providertypes in production app wiring.
const providerStoreKey = "provider"

// CreateUpgradeHandler returns an upgrade handler for Gaia v28.0.0.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("Starting module migrations...")

		// 1. Read max_provider_consensus_validators from the provider KV store.
		// The provider module stores its params at key 0xFF in its own module store.
		providerKey := keepers.GetKey(providerStoreKey)
		if providerKey == nil {
			return vm, fmt.Errorf("provider store key not found")
		}
		providerStore := ctx.KVStore(providerKey)
		paramsBz := providerStore.Get(providerParametersKey)
		if paramsBz == nil {
			return vm, fmt.Errorf("provider params not found in store")
		}
		var providerParams icsProviderParams
		if err := proto.Unmarshal(paramsBz, &providerParams); err != nil {
			return vm, fmt.Errorf("failed to unmarshal provider params: %w", err)
		}
		maxVals := providerParams.MaxProviderConsensusValidators
		ctx.Logger().Info("Read provider max_provider_consensus_validators", "value", maxVals)

		// Validate maxVals is within uint32 range
		if maxVals < 0 || maxVals > math.MaxUint32 {
			return vm, fmt.Errorf("invalid max_provider_consensus_validators value: %d (must be between 0 and %d)", maxVals, uint32(math.MaxUint32))
		}

		// 2. Set staking max_validators to the former max_provider_consensus_validators,
		// but only if it is lower than the current value. This ensures the upgrade
		// reduces (or preserves) the active validator set size and never inflates it.
		stakingParams, err := keepers.StakingKeeper.GetParams(ctx)
		if err != nil {
			return vm, fmt.Errorf("failed to get staking params: %w", err)
		}
		updatedMaxValidators := false
		if uint32(maxVals) < stakingParams.MaxValidators {
			stakingParams.MaxValidators = uint32(maxVals)
			if err := keepers.StakingKeeper.SetParams(ctx, stakingParams); err != nil {
				return vm, fmt.Errorf("failed to set staking params: %w", err)
			}
			updatedMaxValidators = true
			ctx.Logger().Info("Set staking max_validators", "value", maxVals)
		} else {
			ctx.Logger().Info("Skipping max_validators update: provider value is not lower than current",
				"provider_value", maxVals, "current_value", stakingParams.MaxValidators)
		}

		// 3. Transfer consumer rewards pool balance to community pool.
		rewardsPoolAddr := authtypes.NewModuleAddress("consumer_rewards_pool")
		balances := keepers.BankKeeper.GetAllBalances(ctx, rewardsPoolAddr)
		if !balances.IsZero() {
			if err := keepers.DistrKeeper.FundCommunityPool(ctx, balances, rewardsPoolAddr); err != nil {
				return vm, fmt.Errorf("failed to fund community pool from consumer rewards pool: %w", err)
			}
			ctx.Logger().Info("Transferred consumer rewards pool balance to community pool", "amount", balances)
		}

		// 4. Force-close all open IBC channels on the provider port.
		channels := keepers.IBCKeeper.ChannelKeeper.GetAllChannels(ctx)
		for _, ch := range channels {
			if ch.PortId != providerModuleName {
				continue
			}
			if ch.State == channeltypes.CLOSED {
				continue
			}
			channel, found := keepers.IBCKeeper.ChannelKeeper.GetChannel(ctx, ch.PortId, ch.ChannelId)
			if !found {
				continue
			}
			channel.State = channeltypes.CLOSED
			keepers.IBCKeeper.ChannelKeeper.SetChannel(ctx, ch.PortId, ch.ChannelId, channel)
			ctx.Logger().Info("Force-closed IBC channel on provider port", "channel", ch.ChannelId)
		}

		// 5. Apply validator set updates using the new max_validators parameter.
		// This bonds the top N validators and begins unbonding those beyond the cutoff.
		// Only necessary if max_validators was actually lowered.
		if updatedMaxValidators {
			if _, err := keepers.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx); err != nil {
				return vm, fmt.Errorf("failed to apply validator set updates: %w", err)
			}
			ctx.Logger().Info("Applied validator set updates with new max_validators")
		}

		vm, err = mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "running module migrations")
		}

		ctx.Logger().Info("Upgrade v28.0.0 complete")
		return vm, nil
	}
}
