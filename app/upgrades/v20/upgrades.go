package v20

import (
	"context"

	providerkeeper "github.com/cosmos/interchain-security/v5/x/ccv/provider/keeper"
	providertypes "github.com/cosmos/interchain-security/v5/x/ccv/provider/types"

	errorsmod "cosmossdk.io/errors"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	codec "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/cosmos/gaia/v20/app/keepers"
)

// Constants for the new parameters in the v20 upgrade.
const (
	// MaxValidators will be set to 200 (up from 180),
	// to allow the first 20 inactive validators
	// to participate on consumer chains.
	NewMaxValidators = 200
	// MaxProviderConsensusValidators will be set to 180,
	// to preserve the behaviour of only the first 180
	// validators participating in consensus on the Cosmos Hub.
	NewMaxProviderConsensusValidators = 180
)

// CreateUpgradeHandler returns an upgrade handler for Gaia v20.
// It performs module migrations, as well as the following tasks:
// - Initializes the MaxProviderConsensusValidators parameter in the provider module to 180.
// - Increases the MaxValidators parameter in the staking module to 200.
// - Initializes the last provider consensus validator set in the provider module
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
			return vm, errorsmod.Wrapf(err, "running module migrations")
		}

		ctx.Logger().Info("Initializing MaxProviderConsensusValidators parameter...")
		InitializeMaxProviderConsensusParam(ctx, keepers.ProviderKeeper)

		ctx.Logger().Info("Setting MaxValidators parameter...")
		err = SetMaxValidators(ctx, *keepers.StakingKeeper)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "setting MaxValidators during migration")
		}

		ctx.Logger().Info("Initializing LastProviderConsensusValidatorSet...")
		err = InitializeLastProviderConsensusValidatorSet(ctx, keepers.ProviderKeeper, *keepers.StakingKeeper)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "initializing LastProviderConsensusValSet during migration")
		}

		ctx.Logger().Info("Migrating ICS legacy proposals...")
		err = MigrateICSLegacyProposals(ctx, *keepers.GovKeeper)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "migrating ICS legacy proposals during migration")
		}

		ctx.Logger().Info("Upgrade v20 complete")
		return vm, nil
	}
}

// InitializeMaxProviderConsensusParam initializes the MaxProviderConsensusValidators parameter.
// It is set to 180, which is the current number of validators participating in consensus on the Cosmos Hub.
// This parameter will be used to govern the number of validators participating in consensus on the Cosmos Hub,
// and takes over this role from the MaxValidators parameter in the staking module.
func InitializeMaxProviderConsensusParam(ctx sdk.Context, providerKeeper providerkeeper.Keeper) {
	params := providerKeeper.GetParams(ctx)
	params.MaxProviderConsensusValidators = NewMaxProviderConsensusValidators
	providerKeeper.SetParams(ctx, params)
}

// SetMaxValidators sets the MaxValidators parameter in the staking module to 200,
// which is the current number of 180 plus 20.
// This is done in concert with the introduction of the inactive-validators feature
// in Interchain Security, after which the number of validators
// participating in consensus on the Cosmos Hub will be governed by the
// MaxProviderConsensusValidators parameter in the provider module.
func SetMaxValidators(ctx sdk.Context, stakingKeeper stakingkeeper.Keeper) error {
	params, err := stakingKeeper.GetParams(ctx)
	if err != nil {
		return err
	}

	params.MaxValidators = NewMaxValidators

	return stakingKeeper.SetParams(ctx, params)
}

// InitializeLastProviderConsensusValidatorSet initializes the last provider consensus validator set
// by setting it to the first 180 validators from the current validator set of the staking module.
func InitializeLastProviderConsensusValidatorSet(
	ctx sdk.Context, providerKeeper providerkeeper.Keeper, stakingKeeper stakingkeeper.Keeper,
) error {
	vals, err := stakingKeeper.GetBondedValidatorsByPower(ctx)
	if err != nil {
		return err
	}

	// cut the validator set to the first 180 validators
	if len(vals) > NewMaxProviderConsensusValidators {
		vals = vals[:NewMaxProviderConsensusValidators]
	}

	// create consensus validators for the staking validators
	lastValidators := []providertypes.ConsensusValidator{}
	for _, val := range vals {
		consensusVal, err := providerKeeper.CreateProviderConsensusValidator(ctx, val)
		if err != nil {
			return err
		}

		lastValidators = append(lastValidators, consensusVal)
	}

	providerKeeper.SetLastProviderConsensusValSet(ctx, lastValidators)
	return nil
}

// MigrateICSLegacyProposals migrates ICS legacy proposals
func MigrateICSLegacyProposals(ctx sdk.Context, govKeeper govkeeper.Keeper) error {
	return govKeeper.Proposals.Walk(ctx, nil, func(key uint64, proposal govtypes.Proposal) (stop bool, err error) {
		err = MigrateProposal(ctx, govKeeper, proposal)
		if err != nil {
			return true, errorsmod.Wrapf(err, "migrating proposal %d", key)
		}
		return false, nil
	})
}

// MigrateProposal migrates a proposal by converting legacy messages to new messages.
func MigrateProposal(ctx sdk.Context, govKeeper govkeeper.Keeper, proposal govtypes.Proposal) error {
	for idx, msg := range proposal.GetMessages() {
		sdkLegacyMsg, isLegacyProposal := msg.GetCachedValue().(*govtypes.MsgExecLegacyContent)
		if !isLegacyProposal {
			continue
		}
		content, err := govtypes.LegacyContentFromMessage(sdkLegacyMsg)
		if err != nil {
			continue
		}

		msgAdd, ok := content.(*providertypes.ConsumerAdditionProposal)
		if ok {
			anyMsg, err := MigrateLegacyConsumerAddition(*msgAdd, govKeeper.GetAuthority())
			if err != nil {
				return err
			}
			proposal.Messages[idx] = anyMsg
			continue // skip the rest of the loop
		}

		msgRemove, ok := content.(*providertypes.ConsumerRemovalProposal)
		if ok {
			anyMsg, err := MigrateLegacyConsumerRemoval(*msgRemove, govKeeper.GetAuthority())
			if err != nil {
				return err
			}
			proposal.Messages[idx] = anyMsg
			continue // skip the rest of the loop
		}

		msgMod, ok := content.(*providertypes.ConsumerModificationProposal)
		if ok {
			anyMsg, err := MigrateConsumerModificationProposal(*msgMod, govKeeper.GetAuthority())
			if err != nil {
				return err
			}
			proposal.Messages[idx] = anyMsg
			continue // skip the rest of the loop
		}

		msgChangeRewardDenoms, ok := content.(*providertypes.ChangeRewardDenomsProposal)
		if ok {
			anyMsg, err := MigrateChangeRewardDenomsProposal(*msgChangeRewardDenoms, govKeeper.GetAuthority())
			if err != nil {
				return err
			}
			proposal.Messages[idx] = anyMsg
			continue // skip the rest of the loop
		}
	}
	return govKeeper.SetProposal(ctx, proposal)
}

// MigrateLegacyConsumerAddition converts a ConsumerAdditionProposal to a MsgConsumerAdditionProposal
// and returns it as `Any` suitable to replace the legacy message.
// `authority` contains the signer address
func MigrateLegacyConsumerAddition(msg providertypes.ConsumerAdditionProposal, authority string) (*codec.Any, error) {
	sdkMsg := providertypes.MsgConsumerAddition{
		ChainId:                           msg.ChainId,
		InitialHeight:                     msg.InitialHeight,
		GenesisHash:                       msg.GenesisHash,
		BinaryHash:                        msg.BinaryHash,
		SpawnTime:                         msg.SpawnTime,
		UnbondingPeriod:                   msg.UnbondingPeriod,
		CcvTimeoutPeriod:                  msg.CcvTimeoutPeriod,
		TransferTimeoutPeriod:             msg.TransferTimeoutPeriod,
		ConsumerRedistributionFraction:    msg.ConsumerRedistributionFraction,
		BlocksPerDistributionTransmission: msg.BlocksPerDistributionTransmission,
		HistoricalEntries:                 msg.HistoricalEntries,
		DistributionTransmissionChannel:   msg.DistributionTransmissionChannel,
		Top_N:                             msg.Top_N,
		ValidatorsPowerCap:                msg.ValidatorsPowerCap,
		ValidatorSetCap:                   msg.ValidatorSetCap,
		Allowlist:                         msg.Allowlist,
		Denylist:                          msg.Denylist,
		Authority:                         authority,
		MinStake:                          msg.MinStake,
		AllowInactiveVals:                 msg.AllowInactiveVals,
	}
	return codec.NewAnyWithValue(&sdkMsg)
}

// MigrateLegacyConsumerRemoval converts a ConsumerRemovalProposal to a MsgConsumerRemovalProposal
// and returns it as `Any` suitable to replace the legacy message.
// `authority` contains the signer address
func MigrateLegacyConsumerRemoval(msg providertypes.ConsumerRemovalProposal, authority string) (*codec.Any, error) {
	sdkMsg := providertypes.MsgConsumerRemoval{
		ChainId:   msg.ChainId,
		StopTime:  msg.StopTime,
		Authority: authority,
	}
	return codec.NewAnyWithValue(&sdkMsg)
}

// MigrateConsumerModificationProposal converts a ConsumerModificationProposal to a MsgConsumerModificationProposal
// and returns it as `Any` suitable to replace the legacy message.
// `authority` contains the signer address
func MigrateConsumerModificationProposal(msg providertypes.ConsumerModificationProposal, authority string) (*codec.Any, error) {
	sdkMsg := providertypes.MsgConsumerModification{
		ChainId:   msg.ChainId,
		Allowlist: msg.Allowlist,
		Denylist:  msg.Denylist,
		Authority: authority,
	}
	return codec.NewAnyWithValue(&sdkMsg)
}

// MigrateChangeRewardDenomsProposal converts a ChangeRewardDenomsProposal to a MigrateChangeRewardDenomsProposal
// and returns it as `Any` suitable to replace the legacy message.
// `authority` contains the signer address
func MigrateChangeRewardDenomsProposal(msg providertypes.ChangeRewardDenomsProposal, authority string) (*codec.Any, error) {
	sdkMsg := providertypes.MsgChangeRewardDenoms{
		DenomsToAdd:    msg.GetDenomsToAdd(),
		DenomsToRemove: msg.GetDenomsToRemove(),
		Authority:      authority,
	}
	return codec.NewAnyWithValue(&sdkMsg)
}
