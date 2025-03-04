package main

import (
	"context"
	"encoding/hex"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	clienttypesv2 "github.com/cosmos/ibc-go/v10/modules/core/02-client/v2/types"
	channeltypesv2 "github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"
	commitmenttypes "github.com/cosmos/ibc-go/v10/modules/core/23-commitment/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v10/testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/e2esuite"
	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/relayer"
	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/testvalues"
	relayertypes "github.com/srdtrk/solidity-ibc-eureka/e2e/v8/types/relayer"
)

// CosmosRelayerTestSuite is a struct that holds the test suite for two Cosmos chains.
type CosmosRelayerTestSuite struct {
	e2esuite.TestSuite

	SimdA *cosmos.CosmosChain
	SimdB *cosmos.CosmosChain

	SimdASubmitter ibc.Wallet
	SimdBSubmitter ibc.Wallet

	RelayerClient relayertypes.RelayerServiceClient
}

// TestWithIbcEurekaTestSuite is the boilerplate code that allows the test suite to be run
func TestWithCosmosRelayerTestSuite(t *testing.T) {
	suite.Run(t, new(CosmosRelayerTestSuite))
}

// SetupSuite calls the underlying IbcEurekaTestSuite's SetupSuite method
// and deploys the IbcEureka contract
func (s *CosmosRelayerTestSuite) SetupSuite(ctx context.Context) {
	//chainconfig.DefaultChainSpecs = append(chainconfig.DefaultChainSpecs, chainconfig.IbcGoChainSpec("ibc-go-simd-2", "simd-2"))

	os.Setenv(testvalues.EnvKeyEthTestnetType, testvalues.EthTestnetTypeNone)

	s.TestSuite.SetupSuite(ctx)

	s.SimdA, s.SimdB = s.CosmosChains[0], s.CosmosChains[1]
	s.SimdASubmitter = s.CreateAndFundCosmosUser(ctx, s.SimdA)
	s.SimdBSubmitter = s.CreateAndFundCosmosUser(ctx, s.SimdB)

	var (
		relayerProcess *os.Process
		configInfo     relayer.CosmosToCosmosConfigInfo
	)
	s.Require().True(s.Run("Start Relayer", func() {
		err := os.Chdir("../..")
		s.Require().NoError(err)

		configInfo = relayer.CosmosToCosmosConfigInfo{
			ChainAID:    s.SimdA.Config().ChainID,
			ChainBID:    s.SimdB.Config().ChainID,
			ChainATmRPC: s.SimdA.GetHostRPCAddress(),
			ChainBTmRPC: s.SimdB.GetHostRPCAddress(),
			ChainAUser:  s.SimdASubmitter.FormattedAddress(),
			ChainBUser:  s.SimdBSubmitter.FormattedAddress(),
		}

		err = configInfo.GenerateCosmosToCosmosConfigFile(testvalues.RelayerConfigFilePath)
		s.Require().NoError(err)

		relayerProcess, err = relayer.StartRelayer(testvalues.RelayerConfigFilePath)
		s.Require().NoError(err)

		s.T().Cleanup(func() {
			os.Remove(testvalues.RelayerConfigFilePath)
		})
	}))

	s.T().Cleanup(func() {
		if relayerProcess != nil {
			err := relayerProcess.Kill()
			if err != nil {
				s.T().Logf("Failed to kill the relayer process: %v", err)
			}
		}
	})

	s.Require().True(s.Run("Create Relayer Client", func() {
		var err error
		s.RelayerClient, err = relayer.GetGRPCClient(relayer.DefaultRelayerGRPCAddress())
		s.Require().NoError(err)
	}))

	s.Require().True(s.Run("Create Light Client of Chain A on Chain B", func() {
		simdAHeader, err := s.FetchCosmosHeader(ctx, s.SimdA)
		s.Require().NoError(err)

		var (
			clientStateAny    *codectypes.Any
			consensusStateAny *codectypes.Any
		)
		s.Require().True(s.Run("Construct the client and consensus state", func() {
			tmConfig := ibctesting.NewTendermintConfig()
			revision := clienttypes.ParseChainID(simdAHeader.ChainID)
			height := clienttypes.NewHeight(revision, uint64(simdAHeader.Height))

			clientState := ibctm.NewClientState(
				simdAHeader.ChainID,
				tmConfig.TrustLevel, tmConfig.TrustingPeriod, tmConfig.UnbondingPeriod, tmConfig.MaxClockDrift,
				height, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath,
			)
			clientStateAny, err = codectypes.NewAnyWithValue(clientState)
			s.Require().NoError(err)

			consensusState := ibctm.NewConsensusState(simdAHeader.Time, commitmenttypes.NewMerkleRoot([]byte(ibctm.SentinelRoot)), simdAHeader.ValidatorsHash)
			consensusStateAny, err = codectypes.NewAnyWithValue(consensusState)
			s.Require().NoError(err)
		}))

		_, err = s.BroadcastMessages(ctx, s.SimdB, s.SimdBSubmitter, 200_000, &clienttypes.MsgCreateClient{
			ClientState:    clientStateAny,
			ConsensusState: consensusStateAny,
			Signer:         s.SimdBSubmitter.FormattedAddress(),
		})
		s.Require().NoError(err)
	}))

	s.Require().True(s.Run("Create Light Client of Chain B on Chain A", func() {
		simdBHeader, err := s.FetchCosmosHeader(ctx, s.SimdB)
		s.Require().NoError(err)

		var (
			clientStateAny    *codectypes.Any
			consensusStateAny *codectypes.Any
		)
		s.Require().True(s.Run("Construct the client and consensus state", func() {
			tmConfig := ibctesting.NewTendermintConfig()
			revision := clienttypes.ParseChainID(simdBHeader.ChainID)
			height := clienttypes.NewHeight(revision, uint64(simdBHeader.Height))

			clientState := ibctm.NewClientState(
				simdBHeader.ChainID,
				tmConfig.TrustLevel, tmConfig.TrustingPeriod, tmConfig.UnbondingPeriod, tmConfig.MaxClockDrift,
				height, commitmenttypes.GetSDKSpecs(), ibctesting.UpgradePath,
			)
			clientStateAny, err = codectypes.NewAnyWithValue(clientState)
			s.Require().NoError(err)

			consensusState := ibctm.NewConsensusState(simdBHeader.Time, commitmenttypes.NewMerkleRoot([]byte(ibctm.SentinelRoot)), simdBHeader.ValidatorsHash)
			consensusStateAny, err = codectypes.NewAnyWithValue(consensusState)
			s.Require().NoError(err)
		}))

		_, err = s.BroadcastMessages(ctx, s.SimdA, s.SimdASubmitter, 200_000, &clienttypes.MsgCreateClient{
			ClientState:    clientStateAny,
			ConsensusState: consensusStateAny,
			Signer:         s.SimdASubmitter.FormattedAddress(),
		})
		s.Require().NoError(err)
	}))

	s.Require().True(s.Run("Register counterparty on Chain A", func() {
		merklePathPrefix := [][]byte{[]byte(ibcexported.StoreKey), []byte("")}

		// We can do this because we know what the counterparty client ID will be
		_, err := s.BroadcastMessages(ctx, s.SimdA, s.SimdASubmitter, 200_000, &clienttypesv2.MsgRegisterCounterparty{
			ClientId:                 ibctesting.FirstClientID,
			CounterpartyClientId:     ibctesting.FirstClientID,
			CounterpartyMerklePrefix: merklePathPrefix,
			Signer:                   s.SimdASubmitter.FormattedAddress(),
		})
		s.Require().NoError(err)
	}))

	s.Require().True(s.Run("Register counterparty on Chain B", func() {
		merklePathPrefix := [][]byte{[]byte(ibcexported.StoreKey), []byte("")}

		_, err := s.BroadcastMessages(ctx, s.SimdB, s.SimdBSubmitter, 200_000, &clienttypesv2.MsgRegisterCounterparty{
			ClientId:                 ibctesting.FirstClientID,
			CounterpartyClientId:     ibctesting.FirstClientID,
			CounterpartyMerklePrefix: merklePathPrefix,
			Signer:                   s.SimdBSubmitter.FormattedAddress(),
		})
		s.Require().NoError(err)
	}))
}

// TestRelayer is a test that runs the relayer
func (s *CosmosRelayerTestSuite) TestRelayerInfo() {
	ctx := context.Background()
	s.SetupSuite(ctx)

	s.Require().True(s.Run("Verify Chain A to Chain B Relayer Info", func() {
		info, err := s.RelayerClient.Info(context.Background(), &relayertypes.InfoRequest{
			SrcChain: s.SimdA.Config().ChainID,
			DstChain: s.SimdB.Config().ChainID,
		})
		s.Require().NoError(err)
		s.Require().NotNil(info)
		s.Require().Equal(s.SimdA.Config().ChainID, info.SourceChain.ChainId)
		s.Require().Equal(s.SimdB.Config().ChainID, info.TargetChain.ChainId)
	}))

	s.Require().True(s.Run("Verify Chain B to Chain A Relayer Info", func() {
		info, err := s.RelayerClient.Info(context.Background(), &relayertypes.InfoRequest{
			SrcChain: s.SimdB.Config().ChainID,
			DstChain: s.SimdA.Config().ChainID,
		})
		s.Require().NoError(err)
		s.Require().NotNil(info)
		s.Require().Equal(s.SimdB.Config().ChainID, info.SourceChain.ChainId)
		s.Require().Equal(s.SimdA.Config().ChainID, info.TargetChain.ChainId)
	}))
}

func (s *CosmosRelayerTestSuite) TestICS20RecvAndAckPacket() {
	ctx := context.Background()
	s.ICS20RecvAndAckPacketTest(ctx, 1)
}

func (s *CosmosRelayerTestSuite) Test_10_ICS20RecvAndAckPacket() {
	ctx := context.Background()
	s.ICS20RecvAndAckPacketTest(ctx, 10)
}

func (s *CosmosRelayerTestSuite) ICS20RecvAndAckPacketTest(ctx context.Context, numOfTransfers int) {
	s.Require().Greater(numOfTransfers, 0)

	s.SetupSuite(ctx)

	simdAUser, simdBUser := s.CosmosUsers[0], s.CosmosUsers[1]
	transferAmount := big.NewInt(testvalues.TransferAmount)
	totalTransferAmount := testvalues.TransferAmount * int64(numOfTransfers)

	var txHashes [][]byte
	s.Require().True(s.Run("Send transfers on Chain A", func() {
		for i := 0; i < numOfTransfers; i++ {
			timeout := uint64(time.Now().Add(30 * time.Minute).Unix())
			transferCoin := sdk.NewCoin(s.SimdA.Config().Denom, sdkmath.NewIntFromBigInt(transferAmount))

			transferPayload := transfertypes.FungibleTokenPacketData{
				Denom:    transferCoin.Denom,
				Amount:   transferCoin.Amount.String(),
				Sender:   simdAUser.FormattedAddress(),
				Receiver: simdBUser.FormattedAddress(),
				Memo:     "",
			}

			payload := channeltypesv2.Payload{
				SourcePort:      transfertypes.PortID,
				DestinationPort: transfertypes.PortID,
				Version:         transfertypes.V1,
				Encoding:        transfertypes.EncodingJSON,
				Value:           transferPayload.GetBytes(),
			}
			msgSendPacket := channeltypesv2.MsgSendPacket{
				SourceClient:     ibctesting.FirstClientID,
				TimeoutTimestamp: timeout,
				Payloads: []channeltypesv2.Payload{
					payload,
				},
				Signer: simdAUser.FormattedAddress(),
			}

			resp, err := s.BroadcastMessages(ctx, s.SimdA, simdAUser, 200_000, &msgSendPacket)
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.TxHash)

			txHash, err := hex.DecodeString(resp.TxHash)
			s.Require().NoError(err)
			s.Require().NotEmpty(txHash)

			txHashes = append(txHashes, txHash)
		}

		s.Require().True(s.Run("Verify balances on Chain A", func() {
			resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, s.SimdA, &banktypes.QueryBalanceRequest{
				Address: simdAUser.FormattedAddress(),
				Denom:   s.SimdA.Config().Denom,
			})
			s.Require().NoError(err)
			s.Require().NotNil(resp.Balance)
			s.Require().Equal(testvalues.InitialBalance-totalTransferAmount, resp.Balance.Amount.Int64())
		}))
	}))

	var ackTxHash []byte
	s.Require().True(s.Run("Receive packets on Chain B", func() {
		var txBodyBz []byte
		s.Require().True(s.Run("Retrieve relay tx", func() {
			resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
				SrcChain:       s.SimdA.Config().ChainID,
				DstChain:       s.SimdB.Config().ChainID,
				SourceTxIds:    txHashes,
				TargetClientId: ibctesting.FirstClientID,
			})
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.Tx)
			s.Require().Empty(resp.Address)

			txBodyBz = resp.Tx
		}))

		s.Require().True(s.Run("Broadcast relay tx", func() {
			resp := s.BroadcastSdkTxBody(ctx, s.SimdB, s.SimdBSubmitter, 2_000_000, txBodyBz)

			var err error
			ackTxHash, err = hex.DecodeString(resp.TxHash)
			s.Require().NoError(err)
			s.Require().NotEmpty(ackTxHash)

			s.Require().True(s.Run("Verify balances on Chain B", func() {
				ibcDenom := transfertypes.NewDenom(s.SimdA.Config().Denom, transfertypes.NewHop(transfertypes.PortID, ibctesting.FirstClientID)).IBCDenom()
				// User balance on Cosmos chain
				resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, s.SimdB, &banktypes.QueryBalanceRequest{
					Address: simdBUser.FormattedAddress(),
					Denom:   ibcDenom,
				})
				s.Require().NoError(err)
				s.Require().NotNil(resp.Balance)
				s.Require().Equal(totalTransferAmount, resp.Balance.Amount.Int64())
				s.Require().Equal(ibcDenom, resp.Balance.Denom)
			}))
		}))
	}))

	s.Require().True(s.Run("Acknowledge packets on Chain A", func() {
		s.Require().True(s.Run("Verify commitments exists", func() {
			for i := 0; i < numOfTransfers; i++ {
				resp, err := e2esuite.GRPCQuery[channeltypesv2.QueryPacketCommitmentResponse](ctx, s.SimdA, &channeltypesv2.QueryPacketCommitmentRequest{
					ClientId: ibctesting.FirstClientID,
					Sequence: uint64(i) + 1,
				})
				s.Require().NoError(err)
				s.Require().NotEmpty(resp.Commitment)
			}
		}))

		var ackTxBodyBz []byte
		s.Require().True(s.Run("Retrieve ack tx to Chain A", func() {
			resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
				SrcChain:       s.SimdB.Config().ChainID,
				DstChain:       s.SimdA.Config().ChainID,
				SourceTxIds:    [][]byte{ackTxHash},
				TargetClientId: ibctesting.FirstClientID,
			})
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.Tx)
			s.Require().Empty(resp.Address)

			ackTxBodyBz = resp.Tx
		}))

		s.Require().True(s.Run("Broadcast ack tx on Chain A", func() {
			_ = s.BroadcastSdkTxBody(ctx, s.SimdA, s.SimdASubmitter, 2_000_000, ackTxBodyBz)
		}))

		s.Require().True(s.Run("Verify commitments removed", func() {
			for i := 0; i < numOfTransfers; i++ {
				_, err := e2esuite.GRPCQuery[channeltypesv2.QueryPacketCommitmentResponse](ctx, s.SimdA, &channeltypesv2.QueryPacketCommitmentRequest{
					ClientId: ibctesting.FirstClientID,
					Sequence: uint64(i) + 1,
				})
				s.Require().ErrorContains(err, "packet commitment hash not found")
			}
		}))
	}))
}

func (s *CosmosRelayerTestSuite) TestICS20TimeoutPacket() {
	ctx := context.Background()
	s.ICS20TimeoutPacketTest(ctx, 1)
}

func (s *CosmosRelayerTestSuite) Test_10_ICS20TimeoutPacket() {
	ctx := context.Background()
	s.ICS20TimeoutPacketTest(ctx, 10)
}

func (s *CosmosRelayerTestSuite) ICS20TimeoutPacketTest(ctx context.Context, numOfTransfers int) {
	s.Require().Greater(numOfTransfers, 0)

	s.SetupSuite(ctx)

	simdAUser, simdBUser := s.CosmosUsers[0], s.CosmosUsers[1]
	transferAmount := big.NewInt(testvalues.TransferAmount)
	totalTransferAmount := testvalues.TransferAmount * int64(numOfTransfers)

	var txHashes [][]byte
	s.Require().True(s.Run("Send transfers on Chain A", func() {
		for i := 0; i < numOfTransfers; i++ {
			timeout := uint64(time.Now().Add(30 * time.Second).Unix())
			transferCoin := sdk.NewCoin(s.SimdA.Config().Denom, sdkmath.NewIntFromBigInt(transferAmount))

			transferPayload := transfertypes.FungibleTokenPacketData{
				Denom:    transferCoin.Denom,
				Amount:   transferCoin.Amount.String(),
				Sender:   simdAUser.FormattedAddress(),
				Receiver: simdBUser.FormattedAddress(),
				Memo:     "",
			}

			payload := channeltypesv2.Payload{
				SourcePort:      transfertypes.PortID,
				DestinationPort: transfertypes.PortID,
				Version:         transfertypes.V1,
				Encoding:        transfertypes.EncodingJSON,
				Value:           transferPayload.GetBytes(),
			}
			msgSendPacket := channeltypesv2.MsgSendPacket{
				SourceClient:     ibctesting.FirstClientID,
				TimeoutTimestamp: timeout,
				Payloads: []channeltypesv2.Payload{
					payload,
				},
				Signer: simdAUser.FormattedAddress(),
			}

			resp, err := s.BroadcastMessages(ctx, s.SimdA, simdAUser, 200_000, &msgSendPacket)
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.TxHash)

			txHash, err := hex.DecodeString(resp.TxHash)
			s.Require().NoError(err)
			s.Require().NotEmpty(txHash)

			txHashes = append(txHashes, txHash)
		}

		s.Require().True(s.Run("Verify balances on Chain A", func() {
			resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, s.SimdA, &banktypes.QueryBalanceRequest{
				Address: simdAUser.FormattedAddress(),
				Denom:   s.SimdA.Config().Denom,
			})
			s.Require().NoError(err)
			s.Require().NotNil(resp.Balance)
			s.Require().Equal(testvalues.InitialBalance-totalTransferAmount, resp.Balance.Amount.Int64())
		}))
	}))

	// Wait until timeout
	time.Sleep(30 * time.Second)

	s.Require().True(s.Run("Timeout packet on Chain A", func() {
		var timeoutTxBodyBz []byte
		s.Require().True(s.Run("Retrieve timeout tx", func() {
			resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
				SrcChain:       s.SimdB.Config().ChainID,
				DstChain:       s.SimdA.Config().ChainID,
				TimeoutTxIds:   txHashes,
				TargetClientId: ibctesting.FirstClientID,
			})
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.Tx)
			s.Require().Empty(resp.Address)

			timeoutTxBodyBz = resp.Tx
		}))

		s.Require().True(s.Run("Broadcast timeout tx", func() {
			_ = s.BroadcastSdkTxBody(ctx, s.SimdA, s.SimdASubmitter, 2_000_000, timeoutTxBodyBz)
		}))

		s.Require().True(s.Run("Verify balances on Chain A", func() {
			resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, s.SimdA, &banktypes.QueryBalanceRequest{
				Address: simdAUser.FormattedAddress(),
				Denom:   s.SimdA.Config().Denom,
			})
			s.Require().NoError(err)
			s.Require().NotNil(resp.Balance)
			s.Require().Equal(testvalues.InitialBalance, resp.Balance.Amount.Int64())
		}))
	}))
}
