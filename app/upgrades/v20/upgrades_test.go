package v20_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	providerkeeper "github.com/cosmos/interchain-security/v7/x/ccv/provider/keeper"
	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"
	"github.com/cosmos/interchain-security/v7/x/ccv/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/cosmos/gaia/v23/app/helpers"
	v20 "github.com/cosmos/gaia/v23/app/upgrades/v20"
)

func GetTestMsgConsumerAddition() providertypes.MsgConsumerAddition { //nolint:staticcheck
	return providertypes.MsgConsumerAddition{ //nolint:staticcheck
		ChainId:                           "chainid-1",
		InitialHeight:                     clienttypes.NewHeight(1, 1),
		GenesisHash:                       []byte(base64.StdEncoding.EncodeToString([]byte("gen_hash"))),
		BinaryHash:                        []byte(base64.StdEncoding.EncodeToString([]byte("bin_hash"))),
		SpawnTime:                         time.Now().UTC(),
		UnbondingPeriod:                   types.DefaultConsumerUnbondingPeriod,
		CcvTimeoutPeriod:                  types.DefaultCCVTimeoutPeriod,
		TransferTimeoutPeriod:             types.DefaultTransferTimeoutPeriod,
		ConsumerRedistributionFraction:    types.DefaultConsumerRedistributeFrac,
		BlocksPerDistributionTransmission: types.DefaultBlocksPerDistributionTransmission,
		HistoricalEntries:                 types.DefaultHistoricalEntries,
		DistributionTransmissionChannel:   "",
		Top_N:                             50,
		ValidatorsPowerCap:                0,
		ValidatorSetCap:                   0,
		Allowlist:                         nil,
		Denylist:                          nil,
		Authority:                         authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	}
}

func TestMigrateMsgConsumerAdditionWithNotPassedProposalAndInvalidParams(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	providerKeeper := gaiaApp.ProviderKeeper
	govKeeper := gaiaApp.GovKeeper

	// assert that when a not-passed proposal has invalid params, it gets deleted
	// create a sample message, so we can use it in the proposal
	messages := make([]*codectypes.Any, 1)
	messages[0] = &codectypes.Any{TypeUrl: "", Value: []byte{}}
	proposal := v1.Proposal{Messages: messages}
	err := govKeeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	// verify the proposal can be found
	_, err = govKeeper.Proposals.Get(ctx, 0)
	require.NoError(t, err)

	msgConsumerAddition := GetTestMsgConsumerAddition()
	msgConsumerAddition.Top_N = 13 // invalid param, not in [0]\union[50, 100]
	msgServer := providerkeeper.NewMsgServerImpl(&providerKeeper)
	err = v20.MigrateMsgConsumerAddition(ctx, msgServer,
		providerKeeper,
		*govKeeper,
		0,
		msgConsumerAddition,
		0)
	require.NoError(t, err)

	// verify that the proposal got deleted (we cannot find it)
	_, err = govKeeper.Proposals.Get(ctx, 0)
	require.ErrorContains(t, err, "not found")

	// (indirectly) verify that `CreateConsumer` was not called by checking that consumer id was not updated
	consumerID, found := providerKeeper.GetConsumerId(ctx)
	require.False(t, found)
	require.Equal(t, uint64(0), consumerID)
}

func TestMigrateMsgConsumerAdditionWithNotPassedProposalAndValidParams(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	providerKeeper := gaiaApp.ProviderKeeper
	govKeeper := gaiaApp.GovKeeper

	// create a proposal with 2 messages and only update the second message (call `MigrateConsumerAddition` with
	// `indexOfMessageInProposal` being 1)
	messages := make([]*codectypes.Any, 2)
	messages[0] = &codectypes.Any{TypeUrl: "", Value: []byte{1, 2, 3}}
	messages[1] = &codectypes.Any{TypeUrl: "", Value: []byte{}}
	proposal := v1.Proposal{Messages: messages}
	err := govKeeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	msgConsumerAddition := GetTestMsgConsumerAddition()
	msgServer := providerkeeper.NewMsgServerImpl(&providerKeeper)
	err = v20.MigrateMsgConsumerAddition(ctx, msgServer,
		providerKeeper,
		*govKeeper,
		0,
		msgConsumerAddition,
		1)
	require.NoError(t, err)

	// (indirectly) verify that `CreateConsumer` was called by checking that consumer id was updated
	consumerID, found := providerKeeper.GetConsumerId(ctx)
	require.True(t, found)
	require.Equal(t, uint64(1), consumerID)
	consumerMetadata, err := providerKeeper.GetConsumerMetadata(ctx, "0")
	require.NoError(t, err)
	fmt.Println(consumerMetadata)
	require.Equal(t, msgConsumerAddition.ChainId, consumerMetadata.Name)

	proposal, err = govKeeper.Proposals.Get(ctx, 0)
	require.NoError(t, err)
	// first message was not updated
	require.Equal(t, messages[0].TypeUrl, proposal.Messages[0].TypeUrl)
	require.Equal(t, messages[0].Value, proposal.Messages[0].Value)

	// verify that the proposal's second message now contains a `MsgUpdateConsumer` message
	initParams, err := v20.CreateConsumerInitializationParameters(msgConsumerAddition)
	require.NoError(t, err)

	powerShapingParams, err := v20.CreatePowerShapingParameters(msgConsumerAddition.Top_N, msgConsumerAddition.ValidatorsPowerCap,
		msgConsumerAddition.ValidatorSetCap, msgConsumerAddition.Allowlist, msgConsumerAddition.Denylist, msgConsumerAddition.MinStake,
		msgConsumerAddition.AllowInactiveVals)
	require.NoError(t, err)

	expectedMsgUpdateConsumer := providertypes.MsgUpdateConsumer{
		Owner:                    govKeeper.GetAuthority(),
		ConsumerId:               "0",
		Metadata:                 nil,
		InitializationParameters: &initParams,
		PowerShapingParameters:   &powerShapingParams,
	}
	expectedMsgUpdateConsumerBytes, err := expectedMsgUpdateConsumer.Marshal()
	require.NoError(t, err)
	require.Equal(t, "/interchain_security.ccv.provider.v1.MsgUpdateConsumer", proposal.Messages[1].TypeUrl)
	require.Equal(t, expectedMsgUpdateConsumerBytes, proposal.Messages[1].Value)
}

func TestMigrateMsgConsumerAdditionWithPassedProposal(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	providerKeeper := gaiaApp.ProviderKeeper
	govKeeper := gaiaApp.GovKeeper

	// create a passed proposal with one message
	messages := make([]*codectypes.Any, 1)
	messages[0] = &codectypes.Any{TypeUrl: "", Value: []byte{1, 2, 3}}
	proposal := v1.Proposal{Messages: messages, Status: v1.ProposalStatus_PROPOSAL_STATUS_PASSED}
	err := govKeeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	msgConsumerAddition := GetTestMsgConsumerAddition()
	msgServer := providerkeeper.NewMsgServerImpl(&providerKeeper)
	err = v20.MigrateMsgConsumerAddition(ctx, msgServer,
		providerKeeper,
		*govKeeper,
		0,
		msgConsumerAddition,
		0)
	require.NoError(t, err)

	// (indirectly) verify that `CreateConsumer` was called by checking that consumer id was updated
	consumerID, found := providerKeeper.GetConsumerId(ctx)
	require.True(t, found)
	require.Equal(t, uint64(1), consumerID)
	consumerMetadata, err := providerKeeper.GetConsumerMetadata(ctx, "0")
	require.NoError(t, err)
	require.Equal(t, msgConsumerAddition.ChainId, consumerMetadata.Name)

	proposal, err = govKeeper.Proposals.Get(ctx, 0)
	require.NoError(t, err)
	// first message was not updated
	require.Equal(t, messages[0].TypeUrl, proposal.Messages[0].TypeUrl)
	require.Equal(t, messages[0].Value, proposal.Messages[0].Value)

	// verify that the proposal's second message now contains a `MsgUpdateConsumer` message
	initParams, err := v20.CreateConsumerInitializationParameters(msgConsumerAddition)
	require.NoError(t, err)

	powerShapingParams, err := v20.CreatePowerShapingParameters(msgConsumerAddition.Top_N, msgConsumerAddition.ValidatorsPowerCap,
		msgConsumerAddition.ValidatorSetCap, msgConsumerAddition.Allowlist, msgConsumerAddition.Denylist, msgConsumerAddition.MinStake,
		msgConsumerAddition.AllowInactiveVals)
	require.NoError(t, err)

	actualInitParams, err := providerKeeper.GetConsumerInitializationParameters(ctx, "0")
	require.NoError(t, err)
	actualPowerShapingParams, err := providerKeeper.GetConsumerPowerShapingParameters(ctx, "0")
	require.NoError(t, err)
	require.Equal(t, powerShapingParams, actualPowerShapingParams)
	require.Equal(t, initParams, actualInitParams)
}

func TestMigrateMsgConsumerAdditionWithPassedProposalOfAnAlreadyHandleChain(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	providerKeeper := gaiaApp.ProviderKeeper
	govKeeper := gaiaApp.GovKeeper

	// create a passed proposal with one message
	messages := make([]*codectypes.Any, 1)
	messages[0] = &codectypes.Any{TypeUrl: "", Value: []byte{1, 2, 3}}
	proposal := v1.Proposal{Messages: messages, Status: v1.ProposalStatus_PROPOSAL_STATUS_PASSED}
	err := govKeeper.SetProposal(ctx, proposal)
	require.NoError(t, err)

	msgConsumerAddition := GetTestMsgConsumerAddition()

	// the chain is already handled and launched
	providerKeeper.FetchAndIncrementConsumerId(ctx)
	providerKeeper.SetConsumerPhase(ctx, "0", providertypes.CONSUMER_PHASE_LAUNCHED)
	providerKeeper.SetConsumerChainId(ctx, "0", msgConsumerAddition.ChainId)

	msgServer := providerkeeper.NewMsgServerImpl(&providerKeeper)
	err = v20.MigrateMsgConsumerAddition(ctx, msgServer,
		providerKeeper,
		*govKeeper,
		0,
		msgConsumerAddition,
		0)
	require.NoError(t, err)

	// (indirectly) verify that `CreateConsumer` was not called by checking there are no consumer metadata
	_, err = providerKeeper.GetConsumerMetadata(ctx, "0")
	require.Error(t, err)
}

func TestSetICSConsumerMetadata(t *testing.T) {
	gaiaApp := helpers.Setup(t)
	ctx := gaiaApp.NewUncachedContext(true, tmproto.Header{})

	pk := gaiaApp.ProviderKeeper

	// Add consumer chains
	neutronConsumerID := pk.FetchAndIncrementConsumerId(ctx)
	pk.SetConsumerChainId(ctx, neutronConsumerID, "neutron-1")
	pk.SetConsumerPhase(ctx, neutronConsumerID, providertypes.CONSUMER_PHASE_LAUNCHED)
	strideConsumerID := pk.FetchAndIncrementConsumerId(ctx)
	pk.SetConsumerChainId(ctx, strideConsumerID, "stride-1")
	pk.SetConsumerPhase(ctx, strideConsumerID, providertypes.CONSUMER_PHASE_LAUNCHED)

	err := v20.SetICSConsumerMetadata(ctx, pk)
	require.NoError(t, err)

	metadata, err := pk.GetConsumerMetadata(ctx, neutronConsumerID)
	require.NoError(t, err)
	require.Equal(t, "Neutron", metadata.Name)
	expectedMetadataField := map[string]string{
		"phase":          "mainnet",
		"forge_json_url": "https://raw.githubusercontent.com/neutron-org/neutron/main/forge.json",
	}
	metadataField := map[string]string{}
	err = json.Unmarshal([]byte(metadata.Metadata), &metadataField)
	require.NoError(t, err)
	require.Equal(t, expectedMetadataField, metadataField)

	metadata, err = pk.GetConsumerMetadata(ctx, strideConsumerID)
	require.NoError(t, err)
	require.Equal(t, "Stride", metadata.Name)
	expectedMetadataField = map[string]string{
		"phase":          "mainnet",
		"forge_json_url": "https://raw.githubusercontent.com/Stride-Labs/stride/main/forge.json",
	}
	metadataField = map[string]string{}
	err = json.Unmarshal([]byte(metadata.Metadata), &metadataField)
	require.NoError(t, err)
	require.Equal(t, expectedMetadataField, metadataField)
}
