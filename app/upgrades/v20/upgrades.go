package v20

import (
	"context"
	"fmt"

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
		msgServer := providerkeeper.NewMsgServerImpl(&providerKeeper)
		err = MigrateICSLegacyProposals(ctx, msgServer, keepers.ProviderKeeper, *keepers.GovKeeper)
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
func MigrateICSLegacyProposals(ctx sdk.Context, msgServer providertypes.MsgServer, providerKeeper providerkeeper.Keeper, govKeeper govkeeper.Keeper) error {
	return govKeeper.Proposals.Walk(ctx, nil, func(key uint64, proposal govtypes.Proposal) (stop bool, err error) {
		err = MigrateProposal(ctx, msgServer, providerKeeper, govKeeper, proposal)
		if err != nil {
			return true, errorsmod.Wrapf(err, "migrating proposal %d", key)
		}
		return false, nil
	})
}

// MigrateProposal migrates a proposal by converting legacy messages to new messages.
func MigrateProposal(
	ctx sdk.Context,
	msgServer providertypes.MsgServer,
	providerKeeper providerkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	proposal govtypes.Proposal,
) error {
	// ignore proposals that were rejected or failed
	if proposal.Status != govtypes.StatusDepositPeriod &&
		proposal.Status != govtypes.StatusVotingPeriod &&
		proposal.Status != govtypes.StatusPassed {
		return nil
	}

	// ignore proposals with more than one message as we are interested only
	// in legacy proposals (which have only one message)
	messages := proposal.GetMessages()
	if len(messages) != 1 {
		return nil
	}
	msg := messages[0]

	// ignore non-legacy proposals
	sdkLegacyMsg, isLegacyProposal := msg.GetCachedValue().(*govtypes.MsgExecLegacyContent)
	if !isLegacyProposal {
		return nil
	}
	content, err := govtypes.LegacyContentFromMessage(sdkLegacyMsg)
	if err != nil {
		return err
	}

	switch msg := content.(type) {
	case *providertypes.ConsumerAdditionProposal:
		if proposal.Status == govtypes.StatusPassed {
			// ConsumerAdditionProposal that passed -- they were added to the
			// list of pending consumer addition proposals, which was deleted during
			// the migration of the provider module
			if msg.SpawnTime.Before(ctx.BlockTime()) {
				// ignore proposals that already resulted in launched chains
				return nil
			}
			// create a new consumer chain with all the parameters
			metadata := metadataFromCAP(msg)
			initParams := initParamsFromCAP(msg)
			powerSharpingParams := powerShapingParamsFromCAP(msg)
			msgCreateConsumer := providertypes.MsgCreateConsumer{
				Signer:                   govKeeper.GetAuthority(),
				ChainId:                  msg.ChainId,
				Metadata:                 metadata,
				InitializationParameters: &initParams,
				PowerShapingParameters:   &powerSharpingParams,
			}
			resp, err := msgServer.CreateConsumer(ctx, &msgCreateConsumer)
			if err != nil {
				return err
			}
			ctx.Logger().Info(
				fmt.Sprintf(
					"Created consumer with ID(%s), chainID(%d), and spawnTime(%s) from proposal with ID(%d)",
					resp.ConsumerId, msg.ChainId, initParams.SpawnTime.String(), proposal.Id,
				),
			)
		} else {
			// ConsumerAdditionProposal that was submitted, but not yet passed

			// first, create a new consumer chain to get a consumer ID
			metadata := metadataFromCAP(msg)
			msgCreateConsumer := providertypes.MsgCreateConsumer{
				Signer:                   govKeeper.GetAuthority(),
				ChainId:                  msg.ChainId,
				Metadata:                 metadata,
				InitializationParameters: nil, // to be added to MsgUpdateConsumer
				PowerShapingParameters:   nil, // to be added to MsgUpdateConsumer
			}
			resp, err := msgServer.CreateConsumer(ctx, &msgCreateConsumer)
			if err != nil {
				return err
			}
			ctx.Logger().Info(
				fmt.Sprintf(
					"Created consumer with ID(%s), chainID(%d), and no spawnTime from proposal with ID(%d)",
					resp.ConsumerId, msg.ChainId, proposal.Id,
				),
			)

			// second, replace the message in the proposal with a MsgUpdateConsumer
			initParams := initParamsFromCAP(msg)
			powerSharpingParams := powerShapingParamsFromCAP(msg)
			msgUpdateConsumer := providertypes.MsgUpdateConsumer{
				Signer:                   govKeeper.GetAuthority(),
				ConsumerId:               resp.ConsumerId,
				Metadata:                 nil,
				InitializationParameters: &initParams,
				PowerShapingParameters:   &powerSharpingParams,
			}
			anyMsg, err := codec.NewAnyWithValue(&msgUpdateConsumer)
			if err != nil {
				return err
			}
			proposal.Messages[0] = anyMsg
			govKeeper.SetProposal(ctx, proposal) // TODO: check if we can do this
			ctx.Logger().Info(
				fmt.Sprintf(
					"Replaced proposal with ID(%d) with MsgUpdateConsumer - ID(%s), chainID(%d), spawnTime(%s)",
					proposal.Id, resp.ConsumerId, msg.ChainId, initParams.SpawnTime.String(),
				),
			)
		}

	case *providertypes.ConsumerRemovalProposal:
		anyMsg, err := MigrateLegacyConsumerRemoval(*msg, govKeeper.GetAuthority())
		if err != nil {
			return err
		}
		proposal.Messages[idx] = anyMsg

	case *providertypes.ConsumerModificationProposal:
		anyMsg, err := MigrateConsumerModificationProposal(*msg, govKeeper.GetAuthority())
		if err != nil {
			return err
		}
		proposal.Messages[idx] = anyMsg

	case *providertypes.ChangeRewardDenomsProposal:
		anyMsg, err := MigrateChangeRewardDenomsProposal(*msg, govKeeper.GetAuthority())
		if err != nil {
			return err
		}
		proposal.Messages[idx] = anyMsg
	}

	return govKeeper.SetProposal(ctx, proposal)
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

// metadataFromCAP returns ConsumerMetadata from a ConsumerAdditionProposal
func metadataFromCAP(
	prop *providertypes.ConsumerAdditionProposal,
) providertypes.ConsumerMetadata {
	return providertypes.ConsumerMetadata{
		Name:        prop.Title,
		Description: prop.Description,
		Metadata:    "",
	}
}

// initParamsFromCAP returns ConsumerInitializationParameters from
// a ConsumerAdditionProposal
func initParamsFromCAP(
	prop *providertypes.ConsumerAdditionProposal,
) providertypes.ConsumerInitializationParameters {
	return providertypes.ConsumerInitializationParameters{
		InitialHeight:                     prop.InitialHeight,
		GenesisHash:                       prop.GenesisHash,
		BinaryHash:                        prop.BinaryHash,
		SpawnTime:                         prop.SpawnTime,
		UnbondingPeriod:                   prop.UnbondingPeriod,
		CcvTimeoutPeriod:                  prop.CcvTimeoutPeriod,
		TransferTimeoutPeriod:             prop.TransferTimeoutPeriod,
		ConsumerRedistributionFraction:    prop.ConsumerRedistributionFraction,
		BlocksPerDistributionTransmission: prop.BlocksPerDistributionTransmission,
		HistoricalEntries:                 prop.HistoricalEntries,
		DistributionTransmissionChannel:   prop.DistributionTransmissionChannel,
	}
}

// initParamsFromCAP returns PowerShapingParameters from a ConsumerAdditionProposal
func powerShapingParamsFromCAP(
	prop *providertypes.ConsumerAdditionProposal,
) providertypes.PowerShapingParameters {
	return providertypes.PowerShapingParameters{
		Top_N:              prop.Top_N,
		ValidatorsPowerCap: prop.ValidatorsPowerCap,
		ValidatorSetCap:    prop.ValidatorSetCap,
		Allowlist:          prop.Allowlist,
		Denylist:           prop.Denylist,
		MinStake:           prop.MinStake,
		AllowInactiveVals:  prop.AllowInactiveVals,
	}
}
