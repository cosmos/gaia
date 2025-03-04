package e2esuite

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	ethcommon "github.com/ethereum/go-ethereum/common"

	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/v10/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v10/testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	"github.com/cosmos/solidity-ibc-eureka/abigen/ics26router"

	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/testvalues"
	ethereumtypes "github.com/srdtrk/solidity-ibc-eureka/e2e/v8/types/ethereum"
)

func (s *TestSuite) CreateEthereumLightClient(ctx context.Context, cosmosChain *cosmos.CosmosChain, simdRelayerUser ibc.Wallet, ibcContractAddress string) {
	switch s.ethTestnetType {
	case testvalues.EthTestnetTypePoW:
		s.createDummyLightClient(ctx, cosmosChain, simdRelayerUser)
	case testvalues.EthTestnetTypePoS:
		s.createEthereumLightClient(ctx, cosmosChain, simdRelayerUser, ibcContractAddress)
	default:
		panic(fmt.Sprintf("Unrecognized Ethereum testnet type: %v", s.ethTestnetType))
	}
}

func (s *TestSuite) createEthereumLightClient(
	ctx context.Context,
	cosmosChain *cosmos.CosmosChain,
	simdRelayerUser ibc.Wallet,
	ibcContractAddress string,
) {
	eth := s.EthChain

	file, err := os.Open("e2e/interchaintestv8/wasm/cw_ics08_wasm_eth.wasm.gz")
	s.Require().NoError(err)

	etheruemClientChecksum := s.PushNewWasmClientProposal(ctx, cosmosChain, simdRelayerUser, file)
	s.Require().NotEmpty(etheruemClientChecksum, "checksum was empty but should not have been")

	genesis, err := eth.BeaconAPIClient.GetGenesis()
	s.Require().NoError(err)
	spec, err := eth.BeaconAPIClient.GetSpec()
	s.Require().NoError(err)

	executionHeight, err := eth.BeaconAPIClient.GetExecutionHeight("finalized")
	s.Require().NoError(err)
	executionNumberHex := fmt.Sprintf("0x%x", executionHeight)

	ethClientState := ethereumtypes.ClientState{
		ChainID:                      eth.ChainID.Uint64(),
		GenesisValidatorsRoot:        ethcommon.Bytes2Hex(genesis.GenesisValidatorsRoot[:]),
		MinSyncCommitteeParticipants: 32,
		GenesisTime:                  uint64(genesis.GenesisTime.Unix()),
		ForkParameters:               spec.ToForkParameters(),
		SecondsPerSlot:               uint64(spec.SecondsPerSlot.Seconds()),
		SlotsPerEpoch:                spec.SlotsPerEpoch,
		EpochsPerSyncCommitteePeriod: spec.EpochsPerSyncCommitteePeriod,
		LatestSlot:                   executionHeight,
		IsFrozen:                     false,
		IbcCommitmentSlot:            testvalues.IbcCommitmentSlotHex,
		IbcContractAddress:           ibcContractAddress,
	}

	ethClientStateBz, err := json.Marshal(&ethClientState)
	s.Require().NoError(err)
	wasmClientChecksum, err := hex.DecodeString(etheruemClientChecksum)
	s.Require().NoError(err)
	latestHeightSlot := clienttypes.Height{
		RevisionNumber: 0,
		RevisionHeight: executionHeight,
	}
	clientState := ibcwasmtypes.ClientState{
		Data:         ethClientStateBz,
		Checksum:     wasmClientChecksum,
		LatestHeight: latestHeightSlot,
	}
	clientStateAny, err := clienttypes.PackClientState(&clientState)
	s.Require().NoError(err)

	proofOfIBCContract, err := eth.EthAPI.GetProof(ibcContractAddress, []string{ics26router.IbcStoreStorageSlot}, executionNumberHex)
	s.Require().NoError(err)

	header, err := eth.BeaconAPIClient.GetHeader(strconv.Itoa(int(executionHeight)))
	s.Require().NoError(err)
	bootstrap, err := eth.BeaconAPIClient.GetBootstrap(header.Root)
	s.Require().NoError(err)

	if bootstrap.Data.Header.Beacon.Slot != executionHeight {
		s.Require().Fail(fmt.Sprintf("creating client: expected exec height %d, to equal boostrap slot %d", executionHeight, bootstrap.Data.Header.Beacon.Slot))
	}

	unixTimestamp := bootstrap.Data.Header.Execution.Timestamp

	currentPeriod := executionHeight / spec.Period()
	clientUpdates, err := eth.BeaconAPIClient.GetLightClientUpdates(currentPeriod, 1)
	s.Require().NoError(err)
	s.Require().NotEmpty(clientUpdates)

	ethConsensusState := ethereumtypes.ConsensusState{
		Slot:                 bootstrap.Data.Header.Beacon.Slot,
		StateRoot:            bootstrap.Data.Header.Execution.StateRoot,
		StorageRoot:          proofOfIBCContract.StorageHash,
		Timestamp:            unixTimestamp,
		CurrentSyncCommittee: bootstrap.Data.CurrentSyncCommittee.AggregatePubkey,
		NextSyncCommittee:    clientUpdates[0].Data.NextSyncCommittee.AggregatePubkey,
	}

	ethConsensusStateBz, err := json.Marshal(&ethConsensusState)
	s.Require().NoError(err)
	consensusState := ibcwasmtypes.ConsensusState{
		Data: ethConsensusStateBz,
	}
	consensusStateAny, err := clienttypes.PackConsensusState(&consensusState)
	s.Require().NoError(err)

	res, err := s.BroadcastMessages(ctx, cosmosChain, simdRelayerUser, 200_000, &clienttypes.MsgCreateClient{
		ClientState:    clientStateAny,
		ConsensusState: consensusStateAny,
		Signer:         simdRelayerUser.FormattedAddress(),
	})
	s.Require().NoError(err)

	ethereumLightClientID, err := ibctesting.ParseClientIDFromEvents(res.Events)
	s.Require().NoError(err)
	s.Require().Equal(testvalues.FirstWasmClientID, ethereumLightClientID)
}

func (s *TestSuite) createDummyLightClient(ctx context.Context, cosmosChain *cosmos.CosmosChain, simdRelayerUser ibc.Wallet) {
	eth := s.EthChain

	file, err := os.Open("e2e/interchaintestv8/wasm/wasm_dummy_light_client.wasm.gz")
	s.Require().NoError(err)

	dummyClientChecksum := s.PushNewWasmClientProposal(ctx, cosmosChain, simdRelayerUser, file)
	s.Require().NotEmpty(dummyClientChecksum, "checksum was empty but should not have been")

	_, ethHeight, err := eth.EthAPI.GetBlockNumber()
	s.Require().NoError(err)

	wasmClientChecksum, err := hex.DecodeString(dummyClientChecksum)
	s.Require().NoError(err)
	latestHeight := clienttypes.Height{
		RevisionNumber: 0,
		RevisionHeight: ethHeight,
	}
	s.Require().NoError(err)
	clientState := ibcwasmtypes.ClientState{
		Data:         []byte("doesnt matter"),
		Checksum:     wasmClientChecksum,
		LatestHeight: latestHeight,
	}
	clientStateAny, err := clienttypes.PackClientState(&clientState)
	s.Require().NoError(err)

	consensusState := ibcwasmtypes.ConsensusState{
		Data: []byte("doesnt matter"),
	}
	consensusStateAny, err := clienttypes.PackConsensusState(&consensusState)
	s.Require().NoError(err)

	res, err := s.BroadcastMessages(ctx, cosmosChain, simdRelayerUser, 200_000, &clienttypes.MsgCreateClient{
		ClientState:    clientStateAny,
		ConsensusState: consensusStateAny,
		Signer:         simdRelayerUser.FormattedAddress(),
	})
	s.Require().NoError(err)

	ethereumLightClientID, err := ibctesting.ParseClientIDFromEvents(res.Events)
	s.Require().NoError(err)
	s.Require().Equal(testvalues.FirstWasmClientID, ethereumLightClientID)
}
