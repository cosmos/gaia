package v20

import (
	"context"
	"fmt"
	"github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"time"

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
	// NewMaxValidators will be set to 200 (up from 180),
	// to allow the first 20 inactive validators
	// to participate on consumer chains.
	NewMaxValidators = 200
	// NewMaxProviderConsensusValidators will be set to 180,
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

		ctx.Logger().Info("Migrating ICS proposals...")
		msgServer := providerkeeper.NewMsgServerImpl(&keepers.ProviderKeeper)
		err = MigrateICSProposals(ctx, msgServer, keepers.ProviderKeeper, *keepers.GovKeeper)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "migrating ICS proposals during migration")
		}

		ctx.Logger().Info("Setting ICS consumers metadata...")
		err = SetICSConsumerMetadata(ctx, keepers.ProviderKeeper)
		if err != nil {
			return vm, errorsmod.Wrapf(err, "setting ICS consumers metadata during migration")
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

// MigrateICSProposals migrates ICS legacy proposals
func MigrateICSProposals(ctx sdk.Context, msgServer providertypes.MsgServer, providerKeeper providerkeeper.Keeper, govKeeper govkeeper.Keeper) error {
	proposals := []govtypes.Proposal{}
	err := govKeeper.Proposals.Walk(ctx, nil, func(key uint64, proposal govtypes.Proposal) (stop bool, err error) {
		proposals = append(proposals, proposal)
		return false, nil // go through the entire collection
	})
	if err != nil {
		return errorsmod.Wrapf(err, "iterating through proposals")
	}
	for _, proposal := range proposals {
		err := MigrateICSLegacyProposal(ctx, msgServer, providerKeeper, govKeeper, proposal)
		if err != nil {
			return errorsmod.Wrapf(err, "migrating legacy proposal %d", proposal.Id)
		}

		err = MigrateICSProposal(ctx, msgServer, providerKeeper, govKeeper, proposal)
		if err != nil {
			return errorsmod.Wrapf(err, "migrating proposal %d", proposal.Id)
		}
	}
	return nil
}

func ConsumerAdditionProposalToMsgConsumerAddition(proposal providertypes.ConsumerAdditionProposal, authority string) providertypes.MsgConsumerAddition {
	return providertypes.MsgConsumerAddition{
		ChainId:                           proposal.ChainId,
		InitialHeight:                     proposal.InitialHeight,
		GenesisHash:                       proposal.GenesisHash,
		BinaryHash:                        proposal.BinaryHash,
		SpawnTime:                         proposal.SpawnTime,
		UnbondingPeriod:                   proposal.UnbondingPeriod,
		CcvTimeoutPeriod:                  proposal.CcvTimeoutPeriod,
		TransferTimeoutPeriod:             proposal.TransferTimeoutPeriod,
		ConsumerRedistributionFraction:    proposal.ConsumerRedistributionFraction,
		BlocksPerDistributionTransmission: proposal.BlocksPerDistributionTransmission,
		HistoricalEntries:                 proposal.HistoricalEntries,
		DistributionTransmissionChannel:   proposal.DistributionTransmissionChannel,
		Top_N:                             proposal.Top_N,
		ValidatorsPowerCap:                proposal.ValidatorsPowerCap,
		ValidatorSetCap:                   proposal.ValidatorSetCap,
		Allowlist:                         proposal.Allowlist,
		Denylist:                          proposal.Denylist,
		Authority:                         authority,
		MinStake:                          proposal.MinStake,
		AllowInactiveVals:                 proposal.AllowInactiveVals,
	}
}

func ConsumerRemovalProposalToMsgConsumerRemoval(proposal providertypes.ConsumerRemovalProposal, authority string) providertypes.MsgConsumerRemoval {
	return providertypes.MsgConsumerRemoval{
		ChainId:   proposal.ChainId,
		StopTime:  proposal.StopTime,
		Authority: authority,
	}
}

func ConsumerModificationProposalToMsgConsumerModification(proposal providertypes.ConsumerModificationProposal, authority string) providertypes.MsgConsumerModification {
	return providertypes.MsgConsumerModification{
		Title:              proposal.Title,
		Description:        proposal.Description,
		ChainId:            proposal.ChainId,
		Top_N:              proposal.Top_N,
		ValidatorsPowerCap: proposal.ValidatorsPowerCap,
		ValidatorSetCap:    proposal.ValidatorSetCap,
		Allowlist:          proposal.Allowlist,
		Denylist:           proposal.Denylist,
		Authority:          authority,
		MinStake:           proposal.MinStake,
		AllowInactiveVals:  proposal.AllowInactiveVals,
	}
}

func ChangeRewardDenomsProposalToMsgChangeRewardDenoms(proposal providertypes.ChangeRewardDenomsProposal, authority string) providertypes.MsgChangeRewardDenoms {
	return providertypes.MsgChangeRewardDenoms{
		DenomsToAdd:    proposal.DenomsToAdd,
		DenomsToRemove: proposal.DenomsToRemove,
		Authority:      authority,
	}
}

func MigrateICSProposal(
	ctx sdk.Context,
	msgServer providertypes.MsgServer,
	providerKeeper providerkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	proposal govtypes.Proposal) error {
	// ignore proposals that were rejected or failed
	if proposal.Status != govtypes.StatusDepositPeriod &&
		proposal.Status != govtypes.StatusVotingPeriod &&
		proposal.Status != govtypes.StatusPassed {
		return nil
	}

	proposalMessages := proposal.GetMessages()
	for index, proposalMsg := range proposalMessages {
		switch msg := proposalMsg.GetCachedValue().(type) {
		case *providertypes.MsgConsumerAddition:
			err := MigrateMsgConsumerAddition(
				ctx,
				msgServer,
				providerKeeper,
				govKeeper,
				proposal.Id,
				*msg,
				index,
			)
			if err != nil {
				return err
			}
		case *providertypes.MsgConsumerRemoval:
			err := MigrateMsgConsumerRemoval(
				ctx,
				msgServer,
				providerKeeper,
				govKeeper,
				proposal.Id,
				*msg,
				index,
			)
			if err != nil {
				return err
			}
		case *providertypes.MsgConsumerModification:
			err := MigrateMsgConsumerModification(
				ctx,
				providerKeeper,
				govKeeper,
				proposal.Id,
				*msg,
				index,
			)
			if err != nil {
				return err
			}
		case *providertypes.MsgChangeRewardDenoms:
			err := MigrateMsgChangeRewardDenoms(
				ctx,
				govKeeper,
				proposal.Id,
				*msg,
				index,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func MigrateICSLegacyProposal(
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
		return MigrateMsgConsumerAddition(ctx,
			msgServer,
			providerKeeper,
			govKeeper,
			proposal.Id,
			ConsumerAdditionProposalToMsgConsumerAddition(*msg, "authority"),
			0)

	case *providertypes.ConsumerRemovalProposal:
		return MigrateMsgConsumerRemoval(
			ctx,
			msgServer,
			providerKeeper,
			govKeeper,
			proposal.Id,
			ConsumerRemovalProposalToMsgConsumerRemoval(*msg, "authority"),
			0,
		)

	case *providertypes.ConsumerModificationProposal:
		return MigrateMsgConsumerModification(
			ctx,
			providerKeeper,
			govKeeper,
			proposal.Id,
			ConsumerModificationProposalToMsgConsumerModification(*msg, "authority"),
			0,
		)

	case *providertypes.ChangeRewardDenomsProposal:
		return MigrateMsgChangeRewardDenoms(
			ctx,
			govKeeper,
			proposal.Id,
			ChangeRewardDenomsProposalToMsgChangeRewardDenoms(*msg, "authority"),
			0,
		)
	}

	return nil
}

func MigrateMsgConsumerAddition(
	ctx sdk.Context,
	msgServer providertypes.MsgServer,
	providerKeeper providerkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	proposalId uint64,
	msg providertypes.MsgConsumerAddition,
	indexOfMessageInProposal int,
) error {
	proposal, err := govKeeper.Proposals.Get(ctx, proposalId)
	if err != nil {
		return err
	}
	if proposal.Status == govtypes.StatusPassed {
		// MsgConsumerAddition that passed
		for _, consumerId := range providerKeeper.GetAllActiveConsumerIds(ctx) {
			chainId, err := providerKeeper.GetConsumerChainId(ctx, consumerId)
			if err != nil {
				return err // this means something is wrong with the provider state
			}
			if chainId == msg.ChainId {
				// this proposal was already handled in a previous block
				ctx.Logger().Info(
					fmt.Sprintf(
						"Proposal with ID(%d) was skipped as it was already handled - consumerId(%s), chainId(%s), spawnTime(%s)",
						proposal.Id, consumerId, msg.ChainId, msg.SpawnTime.String(),
					),
				)
				return nil
			}
		}

		// This proposal would have been handled in a future block.
		// If the proposal is invalid, just ignore it.
		// Otherwise, call CreateConsumer, which will schedule the consumer
		// chain to be launched at msg.SpawnTime.

		// create a new consumer chain with all the parameters
		metadata := CreateConsumerMetadata(fmt.Sprintf("Chain with chain id %s", msg.ChainId), "TBA")

		initParams, err := CreateConsumerInitializationParameters(msg.InitialHeight, msg.GenesisHash, msg.BinaryHash,
			msg.SpawnTime, msg.UnbondingPeriod, msg.CcvTimeoutPeriod, msg.TransferTimeoutPeriod, msg.ConsumerRedistributionFraction,
			msg.BlocksPerDistributionTransmission, msg.HistoricalEntries, msg.DistributionTransmissionChannel)
		if err != nil {
			// invalid init params -- ignore proposal
			ctx.Logger().Error(
				fmt.Sprintf(
					"Proposal with ID(%d) was skipped as the init params are invalid, chainID(%s), spawnTime(%s): %s",
					proposal.Id, msg.ChainId, msg.SpawnTime.String(), err.Error(),
				),
			)
			return nil
		}

		powerShapingParams, err := CreatePowerShapingParameters(msg.Top_N, msg.ValidatorsPowerCap, msg.ValidatorSetCap,
			msg.Allowlist, msg.Denylist, msg.MinStake, msg.AllowInactiveVals)
		if err != nil {
			// invalid power shaping params -- ignore proposal
			ctx.Logger().Error(
				fmt.Sprintf(
					"Proposal with ID(%d) was skipped as the power shaping params are invalid, chainID(%s), spawnTime(%s): %s",
					proposal.Id, msg.ChainId, msg.SpawnTime.String(), err.Error(),
				),
			)
			return nil
		}
		// first, create an Opt-In consumer chain
		msgCreateConsumer := providertypes.MsgCreateConsumer{
			Signer:                   govKeeper.GetAuthority(),
			ChainId:                  msg.ChainId,
			Metadata:                 metadata,
			InitializationParameters: nil,
			PowerShapingParameters:   nil,
		}
		resp, err := msgServer.CreateConsumer(ctx, &msgCreateConsumer)
		if err != nil {
			return err
		}
		// second, update the consumer chain to be TopN
		msgUpdateConsumer := providertypes.MsgUpdateConsumer{
			Signer:                   govKeeper.GetAuthority(),
			ConsumerId:               resp.ConsumerId,
			Metadata:                 nil,
			InitializationParameters: &initParams,
			PowerShapingParameters:   &powerShapingParams,
		}
		_, err = msgServer.UpdateConsumer(ctx, &msgUpdateConsumer)
		if err != nil {
			return err
		}
		ctx.Logger().Info(
			fmt.Sprintf(
				"Created consumer with ID(%s), chainID(%s), and spawnTime(%s) from proposal with ID(%d)",
				resp.ConsumerId, msg.ChainId, initParams.SpawnTime.String(), proposal.Id,
			),
		)
	} else {
		// MsgConsumerAddition that was submitted, but not yet passed.
		// If the proposal is invalid, remove it.
		// Otherwise, create a new consumer chain (MsgCreateConsumer), and
		// replace the proposal's content with a MsgUpdateConsumer

		metadata := CreateConsumerMetadata(fmt.Sprintf("Chain with chain id %s", msg.ChainId), "TBA")

		initParams, err := CreateConsumerInitializationParameters(msg.InitialHeight, msg.GenesisHash, msg.BinaryHash,
			msg.SpawnTime, msg.UnbondingPeriod, msg.CcvTimeoutPeriod, msg.TransferTimeoutPeriod, msg.ConsumerRedistributionFraction,
			msg.BlocksPerDistributionTransmission, msg.HistoricalEntries, msg.DistributionTransmissionChannel)
		if err != nil {
			// invalid init params -- delete proposal
			if err := govKeeper.DeleteProposal(ctx, proposal.Id); err != nil {
				return err
			}
			ctx.Logger().Error(
				fmt.Sprintf(
					"Proposal with ID(%d) was deleted as the init params are invalid, chainID(%s), spawnTime(%s): %s",
					proposal.Id, msg.ChainId, msg.SpawnTime.String(), err.Error(),
				),
			)
			return nil
		}

		powerShapingParams, err := CreatePowerShapingParameters(msg.Top_N, msg.ValidatorsPowerCap, msg.ValidatorSetCap,
			msg.Allowlist, msg.Denylist, msg.MinStake, msg.AllowInactiveVals)
		if err != nil {
			// invalid power shaping params -- delete proposal
			if err := govKeeper.DeleteProposal(ctx, proposal.Id); err != nil {
				return err
			}
			ctx.Logger().Error(
				fmt.Sprintf(
					"Proposal with ID(%d) was deleted as the power shaping params are invalid, chainID(%s), spawnTime(%s): %s",
					proposal.Id, msg.ChainId, msg.SpawnTime.String(), err.Error(),
				),
			)
			return nil
		}

		// first, create a new consumer chain to get a consumer ID
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
				"Created consumer with ID(%s), chainID(%s), and no spawnTime from proposal with ID(%d)",
				resp.ConsumerId, msg.ChainId, proposal.Id,
			),
		)

		// second, replace the message in the proposal with a MsgUpdateConsumer
		msgUpdateConsumer := providertypes.MsgUpdateConsumer{
			Signer:                   govKeeper.GetAuthority(),
			ConsumerId:               resp.ConsumerId,
			Metadata:                 nil,
			InitializationParameters: &initParams,
			PowerShapingParameters:   &powerShapingParams,
		}
		anyMsg, err := codec.NewAnyWithValue(&msgUpdateConsumer)
		if err != nil {
			return err
		}
		proposal.Messages[indexOfMessageInProposal] = anyMsg
		if err := govKeeper.SetProposal(ctx, proposal); err != nil {
			return err
		}
		ctx.Logger().Info(
			fmt.Sprintf(
				"Replaced proposal with ID(%d) with MsgUpdateConsumer - consumerID(%s), chainID(%s), spawnTime(%s)",
				proposal.Id, resp.ConsumerId, msg.ChainId, initParams.SpawnTime.String(),
			),
		)
	}
	return nil
}

func CreateConsumerMetadata(title string, description string) providertypes.ConsumerMetadata {
	metadata := providertypes.ConsumerMetadata{
		Name:        title,
		Description: description,
		Metadata:    "TBA",
	}
	err := providertypes.ValidateConsumerMetadata(metadata)
	if err != nil {
		metadata.Name = providertypes.TruncateString(metadata.Name, providertypes.MaxNameLength)
		metadata.Description = providertypes.TruncateString(metadata.Description, providertypes.MaxDescriptionLength)
	}
	return metadata
}

func CreateConsumerInitializationParameters(
	initialHeight types.Height, genesisHash []byte, binaryHash []byte, spawnTime time.Time, unbondingPeriod time.Duration,
	ccvTimeoutPeriod time.Duration, transferTimeoutPeriod time.Duration, consumerRedistributionFraction string,
	blocksPerDistributionTransmission int64, historicalEntries int64, distributionTransmissionChannel string) (providertypes.ConsumerInitializationParameters, error) {
	initParams := providertypes.ConsumerInitializationParameters{
		InitialHeight:                     initialHeight,
		GenesisHash:                       genesisHash,
		BinaryHash:                        binaryHash,
		SpawnTime:                         spawnTime,
		UnbondingPeriod:                   unbondingPeriod,
		CcvTimeoutPeriod:                  ccvTimeoutPeriod,
		TransferTimeoutPeriod:             transferTimeoutPeriod,
		ConsumerRedistributionFraction:    consumerRedistributionFraction,
		BlocksPerDistributionTransmission: blocksPerDistributionTransmission,
		HistoricalEntries:                 historicalEntries,
		DistributionTransmissionChannel:   distributionTransmissionChannel,
	}
	err := providertypes.ValidateInitializationParameters(initParams)
	return initParams, err
}

func CreatePowerShapingParameters(
	topN uint32, validatorsPowerCap uint32, validatorSetCap uint32, allowlist []string, denylist []string,
	minStake uint64, allowInactiveVals bool,
) (providertypes.PowerShapingParameters, error) {
	powerShapingParams := providertypes.PowerShapingParameters{
		Top_N:              topN,
		ValidatorsPowerCap: validatorsPowerCap,
		ValidatorSetCap:    validatorSetCap,
		Allowlist:          allowlist,
		Denylist:           denylist,
		MinStake:           minStake,
		AllowInactiveVals:  allowInactiveVals,
	}
	err := providertypes.ValidatePowerShapingParameters(powerShapingParams)
	return powerShapingParams, err
}

func MigrateMsgConsumerRemoval(
	ctx sdk.Context,
	msgServer providertypes.MsgServer,
	providerKeeper providerkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	proposalId uint64,
	msg providertypes.MsgConsumerRemoval,
	indexOfMessageInProposal int,
) error {
	proposal, err := govKeeper.Proposals.Get(ctx, proposalId)
	if err != nil {
		return err
	}
	// identify the consumer chain
	rmConsumerID := ""
	for _, consumerID := range providerKeeper.GetAllActiveConsumerIds(ctx) {
		chainID, err := providerKeeper.GetConsumerChainId(ctx, consumerID)
		if err != nil {
			return err // this means is something wrong with the provider state
		}
		if chainID == msg.ChainId {
			rmConsumerID = consumerID
			break
		}
	}
	if rmConsumerID == "" {
		// ignore proposal as there is no consumer with that chain ID
		ctx.Logger().Info(
			fmt.Sprintf(
				"Proposal with ID(%d) was skipped as there is no consumer with chainID(%s)",
				proposal.Id, msg.ChainId,
			),
		)
		if proposal.Status != govtypes.StatusPassed {
			// if the proposal didn't pass yet, then just remove it
			if err := govKeeper.DeleteProposal(ctx, proposal.Id); err != nil {
				return err
			}
			ctx.Logger().Info(
				fmt.Sprintf(
					"Proposal with ID(%d) was deleted -- chainID(%s)",
					proposal.Id, msg.ChainId,
				),
			)
		}
		return nil
	}

	msgRemoveConsumer := providertypes.MsgRemoveConsumer{
		ConsumerId: rmConsumerID,
		Signer:     govKeeper.GetAuthority(),
	}

	if proposal.Status == govtypes.StatusPassed {
		// ConsumerRemovalProposal that passed -- it was added to the
		// list of pending consumer removal proposals, which was deleted during
		// the migration of the provider module
		_, err := msgServer.RemoveConsumer(ctx, &msgRemoveConsumer)
		if err != nil {
			ctx.Logger().Error(
				fmt.Sprintf(
					"Could not remove consumer with ID(%s), chainID(%s), and stopTime(%s) as per proposal with ID(%d)",
					rmConsumerID, msg.ChainId, msg.StopTime.String(), proposal.Id,
				),
			)
			return nil // do not stop the migration because of this
		}
		ctx.Logger().Info(
			fmt.Sprintf(
				"Consumer with ID(%s), chainID(%s) will stop at stopTime(%s) as per proposal with ID(%d)",
				rmConsumerID, msg.ChainId, msg.StopTime.String(), proposal.Id,
			),
		)
	} else {
		// MsgConsumerRemoval that was submitted, but not yet passed

		// replace the message in the proposal with a MsgRemoveConsumer
		anyMsg, err := codec.NewAnyWithValue(&msgRemoveConsumer)
		if err != nil {
			return err
		}

		proposal.Messages[indexOfMessageInProposal] = anyMsg
		if err := govKeeper.SetProposal(ctx, proposal); err != nil {
			return err
		}
		ctx.Logger().Info(
			fmt.Sprintf(
				"Replaced proposal with ID(%d) with MsgRemoveConsumer - consumerID(%s), chainID(%s), spawnTime(%s)",
				proposal.Id, rmConsumerID, msg.ChainId, msg.StopTime.String(),
			),
		)
	}
	return nil
}

func MigrateMsgConsumerModification(
	ctx sdk.Context,
	providerKeeper providerkeeper.Keeper,
	govKeeper govkeeper.Keeper,
	proposalId uint64,
	msg providertypes.MsgConsumerModification,
	indexOfMessageInProposal int,
) error {
	proposal, err := govKeeper.Proposals.Get(ctx, proposalId)
	if err != nil {
		return err
	}
	if proposal.Status == govtypes.StatusPassed {
		// proposal that passed -- it was already handled in
		// a previous block since these proposals are handled immediately
		ctx.Logger().Info(
			fmt.Sprintf(
				"Proposal with ID(%d) was skipped as it was already handled - chainID(%s)",
				proposal.Id, msg.ChainId,
			),
		)
		return nil
	}

	// proposal that was submitted, but not yet passed
	modifyConsumerID := ""
	for _, consumerID := range providerKeeper.GetAllActiveConsumerIds(ctx) {
		chainID, err := providerKeeper.GetConsumerChainId(ctx, consumerID)
		if err != nil {
			return err // this means is something wrong with the provider state
		}
		if chainID == msg.ChainId {
			modifyConsumerID = consumerID
			break
		}
	}
	if modifyConsumerID == "" {
		// delete proposal as there is no consumer with that chain ID
		if err := govKeeper.DeleteProposal(ctx, proposal.Id); err != nil {
			return err
		}
		ctx.Logger().Info(
			fmt.Sprintf(
				"Proposal with ID(%d) was deleted - chainID(%s)",
				proposal.Id, msg.ChainId,
			),
		)
		return nil
	}

	// replace the message in the proposal with a MsgUpdateConsumer
	powerShapingParams, err := CreatePowerShapingParameters(msg.Top_N, msg.ValidatorsPowerCap, msg.ValidatorSetCap,
		msg.Allowlist, msg.Denylist, msg.MinStake, msg.AllowInactiveVals)
	if err != nil {
		// invalid power shaping params -- delete proposal
		if err := govKeeper.DeleteProposal(ctx, proposal.Id); err != nil {
			return err
		}
		ctx.Logger().Error(
			fmt.Sprintf(
				"Proposal with ID(%d) was deleted as the power shaping params are invalid, consumerID(%s), chainID(%s): %s",
				proposal.Id, modifyConsumerID, msg.ChainId, err.Error(),
			),
		)
		return nil
	}
	msgUpdateConsumer := providertypes.MsgUpdateConsumer{
		Signer:                   govKeeper.GetAuthority(),
		ConsumerId:               modifyConsumerID,
		Metadata:                 nil,
		InitializationParameters: nil,
		PowerShapingParameters:   &powerShapingParams,
	}
	anyMsg, err := codec.NewAnyWithValue(&msgUpdateConsumer)
	if err != nil {
		return err
	}
	proposal.Messages[indexOfMessageInProposal] = anyMsg
	if err := govKeeper.SetProposal(ctx, proposal); err != nil {
		return err
	}
	ctx.Logger().Info(
		fmt.Sprintf(
			"Replaced proposal with ID(%d) with MsgUpdateConsumer - consumerID(%s), chainID(%s)",
			proposal.Id, modifyConsumerID, msg.ChainId,
		),
	)
	return nil
}

func MigrateMsgChangeRewardDenoms(
	ctx sdk.Context,
	govKeeper govkeeper.Keeper,
	proposalId uint64,
	msg providertypes.MsgChangeRewardDenoms,
	indexOfMessageInProposal int,
) error {
	proposal, err := govKeeper.Proposals.Get(ctx, proposalId)
	if err != nil {
		return err
	}
	if proposal.Status == govtypes.StatusPassed {
		// ChangeRewardDenomsProposal that passed -- it was already handled in
		// a previous block since these proposals are handled immediately
		ctx.Logger().Info(
			fmt.Sprintf("Proposal with ID(%d) was skipped as it was already handled", proposal.Id),
		)
	} else {
		// ChangeRewardDenomsProposal that was submitted, but not yet passed

		// replace the message in the proposal with a MsgChangeRewardDenoms
		msgChangeRewardDenoms := providertypes.MsgChangeRewardDenoms{
			Authority:      govKeeper.GetAuthority(),
			DenomsToAdd:    msg.DenomsToAdd,
			DenomsToRemove: msg.DenomsToRemove,
		}
		if err := msgChangeRewardDenoms.ValidateBasic(); err != nil {
			// this should not happen if the original ChangeRewardDenomsProposal
			// was well formed
			if err := govKeeper.DeleteProposal(ctx, proposal.Id); err != nil {
				return err
			}
			ctx.Logger().Error(
				fmt.Sprintf(
					"Proposal with ID(%d) was deleted as it failed validation: %s",
					proposal.Id, err.Error(),
				),
			)
			return nil
		}
		anyMsg, err := codec.NewAnyWithValue(&msgChangeRewardDenoms)
		if err != nil {
			return err
		}
		proposal.Messages[indexOfMessageInProposal] = anyMsg
		if err := govKeeper.SetProposal(ctx, proposal); err != nil {
			return err
		}
		ctx.Logger().Info(
			fmt.Sprintf("Replaced proposal with ID(%d) with MsgChangeRewardDenoms", proposal.Id),
		)
	}
	return nil
}

// SetICSConsumerMetadata sets the metadata for launched consumer chains
func SetICSConsumerMetadata(ctx sdk.Context, providerKeeper providerkeeper.Keeper) error {
	for _, consumerID := range providerKeeper.GetAllActiveConsumerIds(ctx) {
		phase := providerKeeper.GetConsumerPhase(ctx, consumerID)
		if phase != providertypes.ConsumerPhase_CONSUMER_PHASE_LAUNCHED {
			continue
		}
		chainID, err := providerKeeper.GetConsumerChainId(ctx, consumerID)
		if err != nil {
			ctx.Logger().Error(
				fmt.Sprintf("cannot get chain ID for consumer chain, consumerID(%s)", consumerID),
			)
			continue
		}

		if chainID == "stride-1" {
			metadata := providertypes.ConsumerMetadata{
				Name: "Stride",
				Description: "The Stride blockchain has a single purpose: to provide the best liquid staking service for chains in the Cosmos ecosystem. " +
					"Stride protocol currently provides liquid staking for seven Cosmos chains, and has over 80%% of Cosmos ecosystem liquid staking market share.\n" +
					"With Stride's stTokens integrated into major DeFi apps across the Cosmos, a rapidly growing TVL, and more IBC traffic than almost any other Cosmos chain - the Stride blockchain has clearly achieved product market fit.\n" +
					"Stride's top priority is security; it always has been and always will be. " +
					"Like the Cosmos Hub, Stride is a highly secure minimalist blockchain, with no smart contracts and no other apps beside the core liquid staking protocol. " +
					"The Stride codebase has been fully audited by numerous security firms, and receives continuous auditing from Informal Systems. " +
					"And the Stride blockchain is protected by IBC rate-limiting.",
				Metadata: "https://github.com/Stride-Labs/stride",
			}
			err = providerKeeper.SetConsumerMetadata(ctx, consumerID, metadata)
			if err != nil {
				ctx.Logger().Error(
					fmt.Sprintf("cannot set consumer metadata, consumerID(%s), chainID(%s): %s", consumerID, chainID, err.Error()),
				)
				continue
			}
		} else if chainID == "neutron-1" {
			metadata := providertypes.ConsumerMetadata{
				Name: "Neutron",
				Description: "Neutron is the only blockchain network specifically designed to support Integrated Applications. " +
					"By granting contracts the power of an appchain, Neutron minimises development overhead, facilitates improved mechanism design, lowers risk and enhances the scalability of decentralised applications.\n" +
					"Unlike typical smart contracts, Integrated Applications have access to components of the blockchain that exist outside of their virtual machine: " +
					"begin and end block automation, mempools, transaction fee mechanisms, consensus votes, interchain transactions and queries, and more.\n" +
					"Integrated Applications can customise their network's blockspace to provide gasless onboarding to newly acquired users and do away with network selectors to onboard deposits from any connected blockchain in a single click. " +
					"They can deploy and manage capital and integrations across multiple chains, maximising network effects and the ubiquity of their denominations.\n" +
					"These features allow Integrated Applications to establish stronger moats around their technology and business model, while providing a competitive edge that standard applications lack. " +
					"This makes them inherently more attractive and competitive, as they operate on an enhanced platform offering higher performance and broader reach compared to traditional applications.",
				Metadata: "https://github.com/neutron-org/neutron",
			}
			err = providerKeeper.SetConsumerMetadata(ctx, consumerID, metadata)
			if err != nil {
				ctx.Logger().Error(
					fmt.Sprintf("cannot set consumer metadata, consumerID(%s), chainID(%s): %s", consumerID, chainID, err.Error()),
				)
				continue
			}
		}
	}
	return nil
}
