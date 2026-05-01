package v28_0_0

import (
	"context"
	"fmt"
	"math"

	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v28/app/keepers"
)

// CreateUpgradeHandler returns an upgrade handler for Gaia v28.0.0.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("Starting upgrade", "name", UpgradeName)

		// 1. Read the provider module store
		providerKey := keepers.GetKey(providerStoreKey)
		if providerKey == nil {
			return vm, fmt.Errorf("provider store key not found")
		}
		providerStore := ctx.KVStore(providerKey)

		// 2. Read max_provider_consensus_validators from the provider KV store.
		// The provider module stores its params at key 0xFF in its own module store.
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

		// Validate maxVals is a positive value within uint32 range.
		// Rejecting zero prevents the handler from setting max_validators=0,
		// which would unbond every validator and halt the chain.
		if maxVals <= 0 || maxVals > math.MaxUint32 {
			return vm, fmt.Errorf("invalid max_provider_consensus_validators value: %d (must be between 1 and %d)", maxVals, uint32(math.MaxUint32))
		}

		// 3. Set staking max_validators to the former max_provider_consensus_validators,
		// but only if it is lower than the current value. This ensures the upgrade
		// reduces (or preserves) the active validator set size and never inflates it.
		stakingParams, err := keepers.StakingKeeper.GetParams(ctx)
		if err != nil {
			return vm, fmt.Errorf("failed to get staking params: %w", err)
		}
		if uint32(maxVals) < stakingParams.MaxValidators {
			stakingParams.MaxValidators = uint32(maxVals)
			if err := keepers.StakingKeeper.SetParams(ctx, stakingParams); err != nil {
				return vm, fmt.Errorf("failed to set staking params: %w", err)
			}
			ctx.Logger().Info("Set staking max_validators", "value", maxVals)

			if err := trimBGroupValidators(ctx, keepers, uint32(maxVals)); err != nil {
				return vm, fmt.Errorf("failed to trim B-group validators: %w", err)
			}
		} else {
			ctx.Logger().Info("Skipping max_validators update: provider value is not lower than current",
				"provider_value", maxVals, "current_value", stakingParams.MaxValidators)
		}

		// 4. Transfer consumer rewards pool balance to community pool.
		rewardsPoolAddr := authtypes.NewModuleAddress("consumer_rewards_pool")
		balances := keepers.BankKeeper.GetAllBalances(ctx, rewardsPoolAddr)
		if !balances.IsZero() {
			if err := keepers.DistrKeeper.FundCommunityPool(ctx, balances, rewardsPoolAddr); err != nil {
				return vm, fmt.Errorf("failed to fund community pool from consumer rewards pool: %w", err)
			}
			ctx.Logger().Info("Transferred consumer rewards pool balance to community pool", "amount", balances)
		}

		// 5. Delete all pending VSCPackets from the provider store.
		// These are keyed by 0x11 + consumerID and will never be sent since
		// all consumer chains are removed in this upgrade.
		deleted := deleteProviderPendingVSCs(providerStore)
		ctx.Logger().Info("Deleted pending VSC entries from provider store", "count", deleted)

		// 6. Close all open IBC channels on the provider port.
		// Attempt ChanCloseInit first: a channel_close_init event is emitted,
		// and relayers can propagate ChanCloseConfirm to the counterparty.
		// Fall back to SetChannel if ChanCloseInit fails: the client or connection
		// backing the channel is no longer in a valid state for the normal close
		// handshake (e.g., expired client on a stale consumer chain), in which case
		// we still need to mark the channel CLOSED on the Gaia side even though the
		// counterparty will not be notified automatically.
		channels := keepers.IBCKeeper.ChannelKeeper.GetAllChannels(ctx)
		for _, ch := range channels {
			if ch.PortId != providerModuleName {
				continue
			}
			if ch.State == channeltypes.CLOSED {
				continue
			}

			// Try the normal close path first.
			chanCloseErr := keepers.IBCKeeper.ChannelKeeper.ChanCloseInit(ctx, ch.PortId, ch.ChannelId)
			if chanCloseErr == nil {
				ctx.Logger().Info("Closed IBC channel on provider port via ChanCloseInit",
					"channel", ch.ChannelId)
				continue
			}
			ctx.Logger().Info("ChanCloseInit failed on provider port channel; falling back to direct close",
				"channel", ch.ChannelId, "error", chanCloseErr.Error())

			// Fallback: direct SetChannel write. No close event is emitted; the
			// counterparty must be closed manually or allowed to go stale.
			channel, found := keepers.IBCKeeper.ChannelKeeper.GetChannel(ctx, ch.PortId, ch.ChannelId)
			if !found {
				continue
			}
			channel.State = channeltypes.CLOSED
			keepers.IBCKeeper.ChannelKeeper.SetChannel(ctx, ch.PortId, ch.ChannelId, channel)
			ctx.Logger().Info("Force-closed IBC channel on provider port via direct SetChannel",
				"channel", ch.ChannelId)
		}

		ctx.Logger().Info("Starting module migrations...")
		vm, err = mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "running module migrations")
		}

		ctx.Logger().Info("Upgrade complete", "name", UpgradeName)
		return vm, nil
	}
}

// deleteProviderPendingVSCs iterates all pending VSCPacket entries in the
// provider KV store (prefix 0x11) and deletes them. It returns the number of
// entries removed. No proto decoding is needed — we delete by key only.
func deleteProviderPendingVSCs(providerStore storetypes.KVStore) int {
	prefix := providerPendingVSCsKeyPrefix
	iter := storetypes.KVStorePrefixIterator(providerStore, prefix)
	var keys [][]byte
	for ; iter.Valid(); iter.Next() {
		key := make([]byte, len(iter.Key()))
		copy(key, iter.Key())
		keys = append(keys, key)
	}
	iter.Close()
	for _, k := range keys {
		providerStore.Delete(k)
	}
	return len(keys)
}

// trimBGroupValidators removes LastValidatorPower entries for validators that
// fall outside the new maxValidators cap (the "B group" — validators bonded in
// staking but excluded from the CometBFT consensus set by the ICS provider).
//
// This is necessary after lowering staking.params.max_validators: without it,
// GetLastValidators (called by TrackHistoricalInfo in BeginBlocker) panics
// because it finds more LastValidatorPower entries than maxValidators allows.
//
// For each evicted validator that is currently Bonded, BeginUnbondingValidator
// is called to transition it to Unbonding and queue it for completion. The
// corresponding token transfer (BondedPool → NotBondedPool) is batched into a
// single SendCoinsFromModuleToModule call, mirroring what ARVSU does at the end
// of its loop. The EndBlocker's ApplyAndReturnValidatorSetUpdates will then see
// a consistent state and handle these validators correctly going forward.
//
// Emitting ABCI zero-power updates for these validators is intentionally
// skipped: B-group validators were already excluded from CometBFT's validator
// set by the ICS provider, so no consensus-engine update is required.
func trimBGroupValidators(ctx sdk.Context, keepers *keepers.AppKeepers, maxValidators uint32) error {
	// Build the set of operator address bytes for the top-N validators by power.
	topNAddrs := make(map[string]struct{}, maxValidators)
	powerIter, err := keepers.StakingKeeper.ValidatorsPowerStoreIterator(ctx)
	if err != nil {
		return fmt.Errorf("failed to get validators power store iterator: %w", err)
	}
	powerCount := 0
	for ; powerIter.Valid() && powerCount < int(maxValidators); powerIter.Next() {
		topNAddrs[string(powerIter.Value())] = struct{}{}
		powerCount++
	}
	powerIter.Close()

	// Collect validators present in LastValidatorPower that are not in the top-N.
	var toEvict []sdk.ValAddress
	if err = keepers.StakingKeeper.IterateLastValidatorPowers(ctx, func(addr sdk.ValAddress, _ int64) bool {
		if _, ok := topNAddrs[string(addr)]; !ok {
			toEvict = append(toEvict, addr)
		}
		return false
	}); err != nil {
		return fmt.Errorf("failed to iterate last validator powers: %w", err)
	}

	bondDenom, err := keepers.StakingKeeper.BondDenom(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bond denom: %w", err)
	}
	totalToTransfer := sdkmath.ZeroInt()
	for _, addr := range toEvict {
		val, err := keepers.StakingKeeper.GetValidator(ctx, addr)
		if err != nil {
			return fmt.Errorf("failed to get validator %s: %w", addr, err)
		}
		if val.IsBonded() {
			tokens := val.Tokens
			if _, err = keepers.StakingKeeper.BeginUnbondingValidator(ctx, val); err != nil {
				return fmt.Errorf("failed to begin unbonding validator %s: %w", addr, err)
			}
			totalToTransfer = totalToTransfer.Add(tokens)
			ctx.Logger().Info("Began unbonding B-group validator",
				"operator", val.OperatorAddress,
				"tokens", tokens,
				"moniker", val.Description.Moniker)
		}
		if err = keepers.StakingKeeper.DeleteLastValidatorPower(ctx, addr); err != nil {
			return fmt.Errorf("failed to delete last validator power for %s: %w", addr, err)
		}
	}
	if totalToTransfer.IsPositive() {
		coins := sdk.NewCoins(sdk.NewCoin(bondDenom, totalToTransfer))
		if err = keepers.BankKeeper.SendCoinsFromModuleToModule(ctx,
			stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, coins); err != nil {
			return fmt.Errorf("failed to transfer tokens from bonded to not-bonded pool: %w", err)
		}
	}
	ctx.Logger().Info("Trimmed B-group validators from LastValidatorPower",
		"evicted", len(toEvict), "tokens_unbonded", totalToTransfer)
	return nil
}
