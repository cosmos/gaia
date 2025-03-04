package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/suite"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

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

	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	"github.com/cosmos/solidity-ibc-eureka/abigen/ibcerc20"
	"github.com/cosmos/solidity-ibc-eureka/abigen/ics20transfer"
	"github.com/cosmos/solidity-ibc-eureka/abigen/ics26router"
	"github.com/cosmos/solidity-ibc-eureka/abigen/sp1ics07tendermint"

	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/chainconfig"
	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/e2esuite"
	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/ethereum"
	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/operator"
	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/relayer"
	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/testvalues"
	"github.com/srdtrk/solidity-ibc-eureka/e2e/v8/types/erc20"
	relayertypes "github.com/srdtrk/solidity-ibc-eureka/e2e/v8/types/relayer"
)

type MultichainTestSuite struct {
	e2esuite.TestSuite

	// The private key of a test account
	key *ecdsa.PrivateKey
	// The private key of the faucet account of interchaintest
	deployer *ecdsa.PrivateKey

	contractAddresses     ethereum.DeployedContracts
	chainBSP1Ics07Address string

	chainASP1Ics07Contract *sp1ics07tendermint.Contract
	chainBSP1Ics07Contract *sp1ics07tendermint.Contract
	ics26Contract          *ics26router.Contract
	ics20Contract          *ics20transfer.Contract
	erc20Contract          *erc20.Contract

	RelayerClient relayertypes.RelayerServiceClient

	SimdARelayerSubmitter ibc.Wallet
	SimdBRelayerSubmitter ibc.Wallet
	EthRelayerSubmitter   *ecdsa.PrivateKey
}

// TestWithMultichainTestSuite is the boilerplate code that allows the test suite to be run
func TestWithMultichainTestSuite(t *testing.T) {
	suite.Run(t, new(MultichainTestSuite))
}

func (s *MultichainTestSuite) SetupSuite(ctx context.Context, proofType operator.SupportedProofType) {
	chainconfig.DefaultChainSpecs = append(chainconfig.DefaultChainSpecs, chainconfig.IbcGoChainSpec("ibc-go-simd-2", "simd-2"))

	s.TestSuite.SetupSuite(ctx)

	eth, simdA, simdB := s.EthChain, s.CosmosChains[0], s.CosmosChains[1]

	var prover string
	s.Require().True(s.Run("Set up environment", func() {
		err := os.Chdir("../..")
		s.Require().NoError(err)

		s.key, err = eth.CreateAndFundUser()
		s.Require().NoError(err)

		s.EthRelayerSubmitter, err = eth.CreateAndFundUser()
		s.Require().NoError(err)

		operatorKey, err := eth.CreateAndFundUser()
		s.Require().NoError(err)

		s.deployer, err = eth.CreateAndFundUser()
		s.Require().NoError(err)

		s.SimdARelayerSubmitter = s.CreateAndFundCosmosUser(ctx, simdA)
		s.SimdBRelayerSubmitter = s.CreateAndFundCosmosUser(ctx, simdB)

		prover = os.Getenv(testvalues.EnvKeySp1Prover)
		switch prover {
		case "", testvalues.EnvValueSp1Prover_Mock:
			s.T().Logf("Using mock prover")
			prover = testvalues.EnvValueSp1Prover_Mock
			os.Setenv(testvalues.EnvKeySp1Prover, testvalues.EnvValueSp1Prover_Mock)
			os.Setenv(testvalues.EnvKeyVerifier, testvalues.EnvValueVerifier_Mock)

			s.Require().Empty(
				os.Getenv(testvalues.EnvKeyGenerateSolidityFixtures),
				"Fixtures are not supported for mock prover",
			)
		case testvalues.EnvValueSp1Prover_Network:
			s.Require().Empty(
				os.Getenv(testvalues.EnvKeyVerifier),
				fmt.Sprintf("%s should not be set when using the network prover in e2e tests.", testvalues.EnvKeyVerifier),
			)
		default:
			s.Require().Fail("invalid prover type: %s", prover)
		}

		os.Setenv(testvalues.EnvKeyRustLog, testvalues.EnvValueRustLog_Info)
		os.Setenv(testvalues.EnvKeyEthRPC, eth.RPC)
		os.Setenv(testvalues.EnvKeySp1Prover, prover)
		os.Setenv(testvalues.EnvKeyOperatorPrivateKey, hex.EncodeToString(crypto.FromECDSA(operatorKey)))
	}))

	s.Require().True(s.Run("Deploy ethereum contracts with SimdA client", func() {
		os.Setenv(testvalues.EnvKeyTendermintRPC, simdA.GetHostRPCAddress())

		args := append([]string{
			"--trust-level", testvalues.DefaultTrustLevel.String(),
			"--trusting-period", strconv.Itoa(testvalues.DefaultTrustPeriod),
			"-o", testvalues.Sp1GenesisFilePath,
		}, proofType.ToOperatorArgs()...)
		s.Require().NoError(operator.RunGenesis(args...))

		var (
			stdout []byte
			err    error
		)
		switch prover {
		case testvalues.EnvValueSp1Prover_Mock:
			stdout, err = eth.ForgeScript(s.deployer, testvalues.E2EDeployScriptPath)
			s.Require().NoError(err)
		case testvalues.EnvValueSp1Prover_Network:
			// make sure that the NETWORK_PRIVATE_KEY is set.
			s.Require().NotEmpty(os.Getenv(testvalues.EnvKeyNetworkPrivateKey))

			stdout, err = eth.ForgeScript(s.deployer, testvalues.E2EDeployScriptPath)
			s.Require().NoError(err)
		default:
			s.Require().Fail("invalid prover type: %s", prover)
		}

		s.contractAddresses, err = ethereum.GetEthContractsFromDeployOutput(string(stdout))
		s.Require().NoError(err)
		s.chainASP1Ics07Contract, err = sp1ics07tendermint.NewContract(ethcommon.HexToAddress(s.contractAddresses.Ics07Tendermint), eth.RPCClient)
		s.Require().NoError(err)
		s.ics26Contract, err = ics26router.NewContract(ethcommon.HexToAddress(s.contractAddresses.Ics26Router), eth.RPCClient)
		s.Require().NoError(err)
		s.ics20Contract, err = ics20transfer.NewContract(ethcommon.HexToAddress(s.contractAddresses.Ics20Transfer), eth.RPCClient)
		s.Require().NoError(err)
		s.erc20Contract, err = erc20.NewContract(ethcommon.HexToAddress(s.contractAddresses.Erc20), eth.RPCClient)
		s.Require().NoError(err)
	}))

	s.Require().True(s.Run("Deploy SimdB light client on ethereum", func() {
		os.Setenv(testvalues.EnvKeyTendermintRPC, simdB.GetHostRPCAddress())

		args := append([]string{
			"--trust-level", testvalues.DefaultTrustLevel.String(),
			"--trusting-period", strconv.Itoa(testvalues.DefaultTrustPeriod),
			"-o", testvalues.Sp1GenesisFilePath,
		}, proofType.ToOperatorArgs()...)
		s.Require().NoError(operator.RunGenesis(args...))

		s.T().Cleanup(func() {
			_ = os.Remove(testvalues.Sp1GenesisFilePath)
		})

		var (
			stdout []byte
			err    error
		)
		switch prover {
		case testvalues.EnvValueSp1Prover_Mock:
			stdout, err = eth.ForgeScript(s.deployer, testvalues.SP1ICS07DeployScriptPath, "--json")
			s.Require().NoError(err)
		case testvalues.EnvValueSp1Prover_Network:
			// make sure that the NETWORK_PRIVATE_KEY is set.
			s.Require().NotEmpty(os.Getenv(testvalues.EnvKeyNetworkPrivateKey))

			stdout, err = eth.ForgeScript(s.deployer, testvalues.SP1ICS07DeployScriptPath, "--json")
			s.Require().NoError(err)
		default:
			s.Require().Fail("invalid prover type: %s", prover)
		}

		s.chainBSP1Ics07Address, err = ethereum.GetOnlySp1Ics07AddressFromStdout(string(stdout))
		s.Require().NoError(err)
		s.Require().NotEmpty(s.chainBSP1Ics07Address)
		s.Require().True(ethcommon.IsHexAddress(s.chainBSP1Ics07Address))

		s.chainBSP1Ics07Contract, err = sp1ics07tendermint.NewContract(ethcommon.HexToAddress(s.chainBSP1Ics07Address), eth.RPCClient)
		s.Require().NoError(err)
	}))

	s.Require().True(s.Run("Fund address with ERC20", func() {
		tx, err := s.erc20Contract.Transfer(s.GetTransactOpts(eth.Faucet, eth), crypto.PubkeyToAddress(s.key.PublicKey), testvalues.StartingERC20Balance)
		s.Require().NoError(err)

		_, err = eth.GetTxReciept(ctx, tx.Hash()) // wait for the tx to be mined
		s.Require().NoError(err)
	}))

	s.Require().True(s.Run("Add ethereum light client on SimdA", func() {
		s.CreateEthereumLightClient(ctx, simdA, s.SimdARelayerSubmitter, s.contractAddresses.Ics26Router)
	}))

	s.Require().True(s.Run("Add simdA client and counterparty on EVM", func() {
		counterpartyInfo := ics26router.IICS02ClientMsgsCounterpartyInfo{
			ClientId:     testvalues.FirstWasmClientID,
			MerklePrefix: [][]byte{[]byte(ibcexported.StoreKey), []byte("")},
		}
		lightClientAddress := ethcommon.HexToAddress(s.contractAddresses.Ics07Tendermint)
		tx, err := s.ics26Contract.AddClient(s.GetTransactOpts(s.deployer, eth), counterpartyInfo, lightClientAddress)
		s.Require().NoError(err)

		receipt, err := eth.GetTxReciept(ctx, tx.Hash())
		s.Require().NoError(err)

		event, err := e2esuite.GetEvmEvent(receipt, s.ics26Contract.ParseICS02ClientAdded)
		s.Require().NoError(err)
		s.Require().Equal(testvalues.FirstUniversalClientID, event.ClientId)
		s.Require().Equal(testvalues.FirstWasmClientID, event.CounterpartyInfo.ClientId)
	}))

	s.Require().True(s.Run("Add ethereum light client on SimdB", func() {
		s.CreateEthereumLightClient(ctx, simdB, s.SimdBRelayerSubmitter, s.contractAddresses.Ics26Router)
	}))

	s.Require().True(s.Run("Add simdB client and counterparty on EVM", func() {
		counterpartyInfo := ics26router.IICS02ClientMsgsCounterpartyInfo{
			ClientId:     testvalues.FirstWasmClientID,
			MerklePrefix: [][]byte{[]byte(ibcexported.StoreKey), []byte("")},
		}
		lightClientAddress := ethcommon.HexToAddress(s.chainBSP1Ics07Address)
		tx, err := s.ics26Contract.AddClient(s.GetTransactOpts(s.deployer, eth), counterpartyInfo, lightClientAddress)
		s.Require().NoError(err)

		receipt, err := eth.GetTxReciept(ctx, tx.Hash())
		s.Require().NoError(err)

		event, err := e2esuite.GetEvmEvent(receipt, s.ics26Contract.ParseICS02ClientAdded)
		s.Require().NoError(err)
		s.Require().Equal(testvalues.SecondUniversalClientID, event.ClientId)
		s.Require().Equal(testvalues.FirstWasmClientID, event.CounterpartyInfo.ClientId)
	}))

	s.Require().True(s.Run("Register counterparty on SimdA", func() {
		merklePathPrefix := [][]byte{[]byte("")}

		_, err := s.BroadcastMessages(ctx, simdA, s.SimdARelayerSubmitter, 200_000, &clienttypesv2.MsgRegisterCounterparty{
			ClientId:                 testvalues.FirstWasmClientID,
			CounterpartyClientId:     testvalues.FirstUniversalClientID,
			CounterpartyMerklePrefix: merklePathPrefix,
			Signer:                   s.SimdARelayerSubmitter.FormattedAddress(),
		})
		s.Require().NoError(err)
	}))

	s.Require().True(s.Run("Register counterparty on SimdB", func() {
		merklePathPrefix := [][]byte{[]byte("")}

		_, err := s.BroadcastMessages(ctx, simdB, s.SimdBRelayerSubmitter, 200_000, &clienttypesv2.MsgRegisterCounterparty{
			ClientId:                 testvalues.FirstWasmClientID,
			CounterpartyClientId:     testvalues.SecondUniversalClientID,
			CounterpartyMerklePrefix: merklePathPrefix,
			Signer:                   s.SimdBRelayerSubmitter.FormattedAddress(),
		})
		s.Require().NoError(err)
	}))

	s.Require().True(s.Run("Create Light Client of Chain A on Chain B", func() {
		simdAHeader, err := s.FetchCosmosHeader(ctx, simdA)
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

		_, err = s.BroadcastMessages(ctx, simdB, s.SimdBRelayerSubmitter, 2_000_000, &clienttypes.MsgCreateClient{
			ClientState:    clientStateAny,
			ConsensusState: consensusStateAny,
			Signer:         s.SimdBRelayerSubmitter.FormattedAddress(),
		})
		s.Require().NoError(err)
	}))

	s.Require().True(s.Run("Create Light Client of Chain B on Chain A", func() {
		simdBHeader, err := s.FetchCosmosHeader(ctx, simdB)
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

		_, err = s.BroadcastMessages(ctx, simdA, s.SimdARelayerSubmitter, 2_000_000, &clienttypes.MsgCreateClient{
			ClientState:    clientStateAny,
			ConsensusState: consensusStateAny,
			Signer:         s.SimdARelayerSubmitter.FormattedAddress(),
		})
		s.Require().NoError(err)
	}))

	s.Require().True(s.Run("Create Channel and register counterparty on Chain A", func() {
		merklePathPrefix := [][]byte{[]byte(ibcexported.StoreKey), []byte("")}

		// We can do this because we know what the counterparty channel ID will be
		_, err := s.BroadcastMessages(ctx, simdA, s.SimdARelayerSubmitter, 200_000, &clienttypesv2.MsgRegisterCounterparty{
			ClientId:                 ibctesting.SecondClientID,
			CounterpartyClientId:     ibctesting.SecondClientID,
			CounterpartyMerklePrefix: merklePathPrefix,
			Signer:                   s.SimdARelayerSubmitter.FormattedAddress(),
		})
		s.Require().NoError(err)
	}))

	s.Require().True(s.Run("Create Channel and register counterparty on Chain B", func() {
		merklePathPrefix := [][]byte{[]byte(ibcexported.StoreKey), []byte("")}

		_, err := s.BroadcastMessages(ctx, simdB, s.SimdBRelayerSubmitter, 200_000, &clienttypesv2.MsgRegisterCounterparty{
			ClientId:                 ibctesting.SecondClientID,
			CounterpartyClientId:     ibctesting.SecondClientID,
			CounterpartyMerklePrefix: merklePathPrefix,
			Signer:                   s.SimdBRelayerSubmitter.FormattedAddress(),
		})
		s.Require().NoError(err)
	}))

	var relayerProcess *os.Process
	var configInfo relayer.MultichainConfigInfo
	s.Require().True(s.Run("Start Relayer", func() {
		beaconAPI := ""
		// The BeaconAPIClient is nil when the testnet is `pow`
		if eth.BeaconAPIClient != nil {
			beaconAPI = eth.BeaconAPIClient.GetBeaconAPIURL()
		}

		var sp1Config string
		switch prover {
		case testvalues.EnvValueSp1Prover_Mock:
			sp1Config = testvalues.EnvValueSp1Prover_Mock
		case testvalues.EnvValueSp1Prover_Network:
			sp1Config = "env"
		default:
			s.Require().Fail("Unsupported prover type: %s", prover)
		}

		configInfo = relayer.MultichainConfigInfo{
			ChainAID:            simdA.Config().ChainID,
			ChainBID:            simdB.Config().ChainID,
			EthChainID:          eth.ChainID.String(),
			ChainATmRPC:         simdA.GetHostRPCAddress(),
			ChainASignerAddress: s.SimdARelayerSubmitter.FormattedAddress(),
			ChainBTmRPC:         simdB.GetHostRPCAddress(),
			ChainBSignerAddress: s.SimdBRelayerSubmitter.FormattedAddress(),
			ICS26Address:        s.contractAddresses.Ics26Router,
			EthRPC:              eth.RPC,
			BeaconAPI:           beaconAPI,
			SP1Config:           sp1Config,
			MockWasmClient:      os.Getenv(testvalues.EnvKeyEthTestnetType) == testvalues.EthTestnetTypePoW,
		}

		err := configInfo.GenerateMultichainConfigFile(testvalues.RelayerConfigFilePath)
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

	s.Require().True(s.Run("Create Relayer Clients", func() {
		var err error
		s.RelayerClient, err = relayer.GetGRPCClient(relayer.DefaultRelayerGRPCAddress())
		s.Require().NoError(err)
	}))
}

func (s *MultichainTestSuite) TestDeploy_Groth16() {
	ctx := context.Background()
	proofType := operator.ProofTypeGroth16

	s.SetupSuite(ctx, proofType)

	eth, simdA, simdB := s.EthChain, s.CosmosChains[0], s.CosmosChains[1]

	s.Require().True(s.Run("Verify SimdA SP1 Client", func() {
		clientState, err := s.chainASP1Ics07Contract.ClientState(nil)
		s.Require().NoError(err)

		stakingParams, err := simdA.StakingQueryParams(ctx)
		s.Require().NoError(err)

		s.Require().Equal(simdA.Config().ChainID, clientState.ChainId)
		s.Require().Equal(uint8(testvalues.DefaultTrustLevel.Numerator), clientState.TrustLevel.Numerator)
		s.Require().Equal(uint8(testvalues.DefaultTrustLevel.Denominator), clientState.TrustLevel.Denominator)
		s.Require().Equal(uint32(testvalues.DefaultTrustPeriod), clientState.TrustingPeriod)
		s.Require().Equal(uint32(stakingParams.UnbondingTime.Seconds()), clientState.UnbondingPeriod)
		s.Require().False(clientState.IsFrozen)
		s.Require().Equal(uint32(1), clientState.LatestHeight.RevisionNumber)
		s.Require().Greater(clientState.LatestHeight.RevisionHeight, uint32(0))
	}))

	s.Require().True(s.Run("Verify SimdB SP1 Client", func() {
		clientState, err := s.chainBSP1Ics07Contract.ClientState(nil)
		s.Require().NoError(err)

		stakingParams, err := simdB.StakingQueryParams(ctx)
		s.Require().NoError(err)

		s.Require().Equal(simdB.Config().ChainID, clientState.ChainId)
		s.Require().Equal(uint8(testvalues.DefaultTrustLevel.Numerator), clientState.TrustLevel.Numerator)
		s.Require().Equal(uint8(testvalues.DefaultTrustLevel.Denominator), clientState.TrustLevel.Denominator)
		s.Require().Equal(uint32(testvalues.DefaultTrustPeriod), clientState.TrustingPeriod)
		s.Require().Equal(uint32(stakingParams.UnbondingTime.Seconds()), clientState.UnbondingPeriod)
		s.Require().False(clientState.IsFrozen)
		s.Require().Equal(uint32(2), clientState.LatestHeight.RevisionNumber)
		s.Require().Greater(clientState.LatestHeight.RevisionHeight, uint32(0))
	}))

	s.Require().True(s.Run("Verify ICS02 Client", func() {
		clientAddress, err := s.ics26Contract.GetClient(nil, testvalues.FirstUniversalClientID)
		s.Require().NoError(err)
		s.Require().Equal(s.contractAddresses.Ics07Tendermint, strings.ToLower(clientAddress.Hex()))

		counterpartyInfo, err := s.ics26Contract.GetCounterparty(nil, testvalues.FirstUniversalClientID)
		s.Require().NoError(err)
		s.Require().Equal(testvalues.FirstWasmClientID, counterpartyInfo.ClientId)

		clientAddress, err = s.ics26Contract.GetClient(nil, testvalues.SecondUniversalClientID)
		s.Require().NoError(err)
		s.Require().Equal(s.chainBSP1Ics07Address, strings.ToLower(clientAddress.Hex()))

		counterpartyInfo, err = s.ics26Contract.GetCounterparty(nil, testvalues.SecondUniversalClientID)
		s.Require().NoError(err)
		s.Require().Equal(testvalues.FirstWasmClientID, counterpartyInfo.ClientId)
	}))

	s.Require().True(s.Run("Verify ICS26 Router", func() {
		hasRole, err := s.ics26Contract.HasRole(nil, testvalues.PortCustomizerRole, crypto.PubkeyToAddress(s.deployer.PublicKey))
		s.Require().NoError(err)
		s.Require().True(hasRole)

		transferAddress, err := s.ics26Contract.GetIBCApp(nil, transfertypes.PortID)
		s.Require().NoError(err)
		s.Require().Equal(s.contractAddresses.Ics20Transfer, strings.ToLower(transferAddress.Hex()))
	}))

	s.Require().True(s.Run("Verify ERC20 Genesis", func() {
		userBalance, err := s.erc20Contract.BalanceOf(nil, crypto.PubkeyToAddress(s.key.PublicKey))
		s.Require().NoError(err)
		s.Require().Equal(testvalues.StartingERC20Balance, userBalance)
	}))

	s.Require().True(s.Run("Verify ethereum light client for SimdA", func() {
		_, err := e2esuite.GRPCQuery[clienttypes.QueryClientStateResponse](ctx, simdA, &clienttypes.QueryClientStateRequest{
			ClientId: testvalues.FirstWasmClientID,
		})
		s.Require().NoError(err)

		// TODO: https://github.com/cosmos/ibc-go/issues/7875
		// channelResp, err := e2esuite.GRPCQuery[channeltypesv2.QueryChannelResponse](ctx, simdA, &channeltypesv2.QueryChannelRequest{
		// 	ChannelId: ibctesting.FirstChannelID,
		// })
		// s.Require().NoError(err)
		// s.Require().Equal(testvalues.FirstWasmClientID, channelResp.Channel.ClientId)
		// s.Require().Equal(testvalues.FirstUniversalClientID, channelResp.Channel.CounterpartyChannelId)
	}))

	s.Require().True(s.Run("Verify ethereum light client for SimdB", func() {
		_, err := e2esuite.GRPCQuery[clienttypes.QueryClientStateResponse](ctx, simdB, &clienttypes.QueryClientStateRequest{
			ClientId: testvalues.FirstWasmClientID,
		})
		s.Require().NoError(err)

		// TODO: https://github.com/cosmos/ibc-go/issues/7875
		// channelResp, err := e2esuite.GRPCQuery[channeltypesv2.QueryChannelResponse](ctx, simdB, &channeltypesv2.QueryChannelRequest{
		// 	ChannelId: ibctesting.FirstChannelID,
		// })
		// s.Require().NoError(err)
		// s.Require().Equal(testvalues.FirstWasmClientID, channelResp.Channel.ClientId)
		// s.Require().Equal(ibctesting.SecondClientID, channelResp.Channel.CounterpartyChannelId)
	}))

	s.Require().True(s.Run("Verify Light Client of Chain A on Chain B", func() {
		clientStateResp, err := e2esuite.GRPCQuery[clienttypes.QueryClientStateResponse](ctx, simdB, &clienttypes.QueryClientStateRequest{
			ClientId: ibctesting.SecondClientID,
		})
		s.Require().NoError(err)
		s.Require().NotZero(clientStateResp.ClientState.Value)

		var clientState ibctm.ClientState
		err = proto.Unmarshal(clientStateResp.ClientState.Value, &clientState)
		s.Require().NoError(err)
		s.Require().Equal(simdA.Config().ChainID, clientState.ChainId)
	}))

	s.Require().True(s.Run("Verify Light Client of Chain B on Chain A", func() {
		clientStateResp, err := e2esuite.GRPCQuery[clienttypes.QueryClientStateResponse](ctx, simdA, &clienttypes.QueryClientStateRequest{
			ClientId: ibctesting.SecondClientID,
		})
		s.Require().NoError(err)
		s.Require().NotZero(clientStateResp.ClientState.Value)

		var clientState ibctm.ClientState
		err = proto.Unmarshal(clientStateResp.ClientState.Value, &clientState)
		s.Require().NoError(err)
		s.Require().Equal(simdB.Config().ChainID, clientState.ChainId)
	}))

	time.Sleep(5 * time.Second) // wait for the relayer to start

	s.Require().True(s.Run("Verify SimdA to Eth Relayer Info", func() {
		info, err := s.RelayerClient.Info(context.Background(), &relayertypes.InfoRequest{
			SrcChain: simdA.Config().ChainID,
			DstChain: eth.ChainID.String(),
		})
		s.Require().NoError(err)
		s.Require().NotNil(info)
		s.Require().Equal(simdA.Config().ChainID, info.SourceChain.ChainId)
		s.Require().Equal(eth.ChainID.String(), info.TargetChain.ChainId)
	}))

	s.Require().True(s.Run("Verify Eth to SimdA Relayer Info", func() {
		info, err := s.RelayerClient.Info(context.Background(), &relayertypes.InfoRequest{
			SrcChain: eth.ChainID.String(),
			DstChain: simdA.Config().ChainID,
		})
		s.Require().NoError(err)
		s.Require().NotNil(info)
		s.Require().Equal(eth.ChainID.String(), info.SourceChain.ChainId)
		s.Require().Equal(simdA.Config().ChainID, info.TargetChain.ChainId)
	}))

	s.Require().True(s.Run("Verify SimdB to Eth Relayer Info", func() {
		info, err := s.RelayerClient.Info(context.Background(), &relayertypes.InfoRequest{
			SrcChain: simdB.Config().ChainID,
			DstChain: eth.ChainID.String(),
		})
		s.Require().NoError(err)
		s.Require().NotNil(info)
		s.Require().Equal(simdB.Config().ChainID, info.SourceChain.ChainId)
		s.Require().Equal(eth.ChainID.String(), info.TargetChain.ChainId)
	}))

	s.Require().True(s.Run("Verify Eth to SimdB Relayer Info", func() {
		info, err := s.RelayerClient.Info(context.Background(), &relayertypes.InfoRequest{
			SrcChain: eth.ChainID.String(),
			DstChain: simdB.Config().ChainID,
		})
		s.Require().NoError(err)
		s.Require().NotNil(info)
		s.Require().Equal(eth.ChainID.String(), info.SourceChain.ChainId)
		s.Require().Equal(simdB.Config().ChainID, info.TargetChain.ChainId)
	}))

	s.Require().True(s.Run("Verify Chain A to Chain B Relayer Info", func() {
		info, err := s.RelayerClient.Info(context.Background(), &relayertypes.InfoRequest{
			SrcChain: simdA.Config().ChainID,
			DstChain: simdB.Config().ChainID,
		})
		s.Require().NoError(err)
		s.Require().NotNil(info)
		s.Require().Equal(simdA.Config().ChainID, info.SourceChain.ChainId)
		s.Require().Equal(simdB.Config().ChainID, info.TargetChain.ChainId)
	}))

	s.Require().True(s.Run("Verify Chain B to Chain A Relayer Info", func() {
		info, err := s.RelayerClient.Info(context.Background(), &relayertypes.InfoRequest{
			SrcChain: simdB.Config().ChainID,
			DstChain: simdA.Config().ChainID,
		})
		s.Require().NoError(err)
		s.Require().NotNil(info)
		s.Require().Equal(simdB.Config().ChainID, info.SourceChain.ChainId)
		s.Require().Equal(simdA.Config().ChainID, info.TargetChain.ChainId)
	}))
}

func (s *MultichainTestSuite) TestTransferCosmosToEthToCosmos_Groth16() {
	ctx := context.Background()
	proofType := operator.ProofTypeGroth16

	s.SetupSuite(ctx, proofType)

	eth, simdA, simdB := s.EthChain, s.CosmosChains[0], s.CosmosChains[1]

	ics26Address := ethcommon.HexToAddress(s.contractAddresses.Ics26Router)
	ics20Address := ethcommon.HexToAddress(s.contractAddresses.Ics20Transfer)
	transferAmount := big.NewInt(testvalues.TransferAmount)
	ethereumUserAddress := crypto.PubkeyToAddress(s.key.PublicKey)
	simdAUser, simdBUser := s.CosmosUsers[0], s.CosmosUsers[1]

	var simdASendTxHash []byte
	s.Require().True(s.Run("Send transfer on SimdA chain", func() {
		timeout := uint64(time.Now().Add(30 * time.Minute).Unix())
		transferCoin := sdk.NewCoin(simdA.Config().Denom, sdkmath.NewIntFromBigInt(transferAmount))

		transferPayload := transfertypes.FungibleTokenPacketData{
			Denom:    transferCoin.Denom,
			Amount:   transferCoin.Amount.String(),
			Sender:   simdAUser.FormattedAddress(),
			Receiver: strings.ToLower(ethereumUserAddress.Hex()),
			Memo:     "",
		}
		encodedPayload, err := transfertypes.EncodeABIFungibleTokenPacketData(&transferPayload)
		s.Require().NoError(err)

		payload := channeltypesv2.Payload{
			SourcePort:      transfertypes.PortID,
			DestinationPort: transfertypes.PortID,
			Version:         transfertypes.V1,
			Encoding:        transfertypes.EncodingABI,
			Value:           encodedPayload,
		}
		msgSendPacket := channeltypesv2.MsgSendPacket{
			SourceClient:     testvalues.FirstWasmClientID,
			TimeoutTimestamp: timeout,
			Payloads: []channeltypesv2.Payload{
				payload,
			},
			Signer: simdAUser.FormattedAddress(),
		}

		resp, err := s.BroadcastMessages(ctx, simdA, simdAUser, 200_000, &msgSendPacket)
		s.Require().NoError(err)
		s.Require().NotEmpty(resp.TxHash)

		simdASendTxHash, err = hex.DecodeString(resp.TxHash)
		s.Require().NoError(err)

		s.Require().True(s.Run("Verify balances on Cosmos chain", func() {
			// Check the balance of UserB
			resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, simdA, &banktypes.QueryBalanceRequest{
				Address: simdAUser.FormattedAddress(),
				Denom:   transferCoin.Denom,
			})
			s.Require().NoError(err)
			s.Require().NotNil(resp.Balance)
			s.Require().Equal(testvalues.InitialBalance-testvalues.TransferAmount, resp.Balance.Amount.Int64())
		}))
	}))

	var (
		ibcERC20        *ibcerc20.Contract
		ibcERC20Address ethcommon.Address
	)
	s.Require().True(s.Run("Receive packet on Ethereum", func() {
		var recvRelayTx []byte
		s.Require().True(s.Run("Retrieve relay tx", func() {
			resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
				SrcChain:       simdA.Config().ChainID,
				DstChain:       eth.ChainID.String(),
				SourceTxIds:    [][]byte{simdASendTxHash},
				TargetClientId: testvalues.FirstUniversalClientID,
			})
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.Tx)
			s.Require().Equal(resp.Address, ics26Address.String())

			recvRelayTx = resp.Tx
		}))

		var packet ics26router.IICS26RouterMsgsPacket
		s.Require().True(s.Run("Submit relay tx", func() {
			receipt, err := eth.BroadcastTx(ctx, s.EthRelayerSubmitter, 5_000_000, ics26Address, recvRelayTx)
			s.Require().NoError(err)
			s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status, fmt.Sprintf("Tx failed: %+v", receipt))

			ethReceiveAckEvent, err := e2esuite.GetEvmEvent(receipt, s.ics26Contract.ParseWriteAcknowledgement)
			s.Require().NoError(err)

			packet = ethReceiveAckEvent.Packet
			// ackTxHash = receipt.TxHash.Bytes()
			// NOTE: ackTxHash is not used in the test since acking the packet is not necessary
		}))

		// Recreate the full denom path
		transferCoin := sdk.NewCoin(simdA.Config().Denom, sdkmath.NewIntFromBigInt(transferAmount))
		denomOnEthereum := transfertypes.NewDenom(transferCoin.Denom, transfertypes.NewHop(packet.Payloads[0].DestPort, packet.DestClient))

		var err error
		ibcERC20Address, err = s.ics20Contract.IbcERC20Contract(nil, denomOnEthereum.Path())
		s.Require().NoError(err)

		ibcERC20, err = ibcerc20.NewContract(ibcERC20Address, eth.RPCClient)
		s.Require().NoError(err)

		actualFullDenom, err := ibcERC20.FullDenomPath(nil)
		s.Require().NoError(err)
		s.Require().Equal(denomOnEthereum.Path(), actualFullDenom)

		s.True(s.Run("Verify balances on Ethereum", func() {
			// User balance on Ethereum
			userBalance, err := ibcERC20.BalanceOf(nil, ethereumUserAddress)
			s.Require().NoError(err)
			s.Require().Equal(transferAmount, userBalance)

			// ICS20 contract balance on Ethereum
			ics20TransferBalance, err := ibcERC20.BalanceOf(nil, ics20Address)
			s.Require().NoError(err)
			s.Require().Zero(ics20TransferBalance.Int64())
		}))
	}))

	s.Require().True(s.Run("Approve the ICS20Transfer.sol contract to spend the erc20 tokens", func() {
		tx, err := ibcERC20.Approve(s.GetTransactOpts(s.key, eth), ics20Address, transferAmount)
		s.Require().NoError(err)

		receipt, err := eth.GetTxReciept(ctx, tx.Hash())
		s.Require().NoError(err)
		s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status)

		allowance, err := ibcERC20.Allowance(nil, ethereumUserAddress, ics20Address)
		s.Require().NoError(err)
		s.Require().Equal(transferAmount, allowance)
	}))

	var ethSendTxHash []byte
	s.Require().True(s.Run("Transfer tokens from Ethereum to SimdB", func() {
		timeout := uint64(time.Now().Add(30 * time.Minute).Unix())
		msgSendPacket := ics20transfer.IICS20TransferMsgsSendTransferMsg{
			Denom:            ibcERC20Address,
			Amount:           transferAmount,
			Receiver:         simdBUser.FormattedAddress(),
			SourceClient:     testvalues.SecondUniversalClientID,
			DestPort:         transfertypes.PortID,
			TimeoutTimestamp: timeout,
			Memo:             "",
		}

		tx, err := s.ics20Contract.SendTransfer(s.GetTransactOpts(s.key, eth), msgSendPacket)
		s.Require().NoError(err)

		receipt, err := eth.GetTxReciept(ctx, tx.Hash())
		s.Require().NoError(err)
		s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status)

		ethSendTxHash = tx.Hash().Bytes()

		s.True(s.Run("Verify balances on Ethereum", func() {
			userBalance, err := ibcERC20.BalanceOf(nil, ethereumUserAddress)
			s.Require().NoError(err)
			s.Require().Zero(userBalance.Int64())

			// the whole balance should have been burned
			ics20TransferBalance, err := ibcERC20.BalanceOf(nil, ics20Address)
			s.Require().NoError(err)
			s.Require().Zero(ics20TransferBalance.Int64())
		}))
	}))

	s.Require().True(s.Run("Receive packet on SimdB", func() {
		var relayTxBodyBz []byte
		s.Require().True(s.Run("Retrieve relay tx", func() {
			resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
				SrcChain:       eth.ChainID.String(),
				DstChain:       simdB.Config().ChainID,
				SourceTxIds:    [][]byte{ethSendTxHash},
				TargetClientId: testvalues.FirstWasmClientID,
			})
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.Tx)
			s.Require().Empty(resp.Address)

			relayTxBodyBz = resp.Tx
		}))

		s.Require().True(s.Run("Broadcast relay tx", func() {
			_ = s.BroadcastSdkTxBody(ctx, simdB, s.SimdBRelayerSubmitter, 2_000_000, relayTxBodyBz)
		}))

		s.Require().True(s.Run("Verify balances on Cosmos chain", func() {
			finalDenom := transfertypes.NewDenom(
				simdA.Config().Denom,
				transfertypes.NewHop(transfertypes.PortID, testvalues.FirstWasmClientID),
				transfertypes.NewHop(transfertypes.PortID, testvalues.FirstUniversalClientID),
			)

			// Check the balance of UserB
			resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, simdB, &banktypes.QueryBalanceRequest{
				Address: simdBUser.FormattedAddress(),
				Denom:   finalDenom.IBCDenom(),
			})
			s.Require().NoError(err)
			s.Require().NotNil(resp.Balance)
			s.Require().Equal(testvalues.TransferAmount, resp.Balance.Amount.Int64())
			s.Require().Equal(finalDenom.IBCDenom(), resp.Balance.Denom)
		}))
	}))
}

func (s *MultichainTestSuite) TestTransferEthToCosmosToCosmos_Groth16() {
	ctx := context.Background()
	proofType := operator.ProofTypeGroth16

	s.SetupSuite(ctx, proofType)

	eth, simdA, simdB := s.EthChain, s.CosmosChains[0], s.CosmosChains[1]

	ics20Address := ethcommon.HexToAddress(s.contractAddresses.Ics20Transfer)
	erc20Address := ethcommon.HexToAddress(s.contractAddresses.Erc20)

	transferAmount := big.NewInt(testvalues.TransferAmount)
	ethereumUserAddress := crypto.PubkeyToAddress(s.key.PublicKey)
	simdAUser, simdBUser := s.CosmosUsers[0], s.CosmosUsers[1]

	s.Require().True(s.Run("Approve the ICS20Transfer.sol contract to spend the erc20 tokens", func() {
		tx, err := s.erc20Contract.Approve(s.GetTransactOpts(s.key, eth), ics20Address, transferAmount)
		s.Require().NoError(err)

		receipt, err := eth.GetTxReciept(ctx, tx.Hash())
		s.Require().NoError(err)
		s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status)

		allowance, err := s.erc20Contract.Allowance(nil, ethereumUserAddress, ics20Address)
		s.Require().NoError(err)
		s.Require().Equal(transferAmount, allowance)
	}))

	var (
		ethSendTxHash []byte
		escrowAddress ethcommon.Address
	)
	s.Require().True(s.Run("Send from Ethereum to SimdA", func() {
		timeout := uint64(time.Now().Add(30 * time.Minute).Unix())

		msgSendPacket := ics20transfer.IICS20TransferMsgsSendTransferMsg{
			Denom:            erc20Address,
			Amount:           transferAmount,
			Receiver:         simdAUser.FormattedAddress(),
			SourceClient:     testvalues.FirstUniversalClientID,
			DestPort:         transfertypes.PortID,
			TimeoutTimestamp: timeout,
			Memo:             "",
		}

		tx, err := s.ics20Contract.SendTransfer(s.GetTransactOpts(s.key, eth), msgSendPacket)
		s.Require().NoError(err)
		receipt, err := eth.GetTxReciept(ctx, tx.Hash())
		s.Require().NoError(err)
		s.Require().Equal(ethtypes.ReceiptStatusSuccessful, receipt.Status)

		ethSendTxHash = tx.Hash().Bytes()

		s.True(s.Run("Verify balances on Ethereum", func() {
			// User balance on Ethereum
			userBalance, err := s.erc20Contract.BalanceOf(nil, ethereumUserAddress)
			s.Require().NoError(err)
			s.Require().Equal(new(big.Int).Sub(testvalues.StartingERC20Balance, transferAmount), userBalance)

			// Get the escrow contract address
			escrowAddress, err = s.ics20Contract.GetEscrow(nil, testvalues.FirstUniversalClientID)
			s.Require().NoError(err)

			// ICS20 contract balance on Ethereum
			escrowBalance, err := s.erc20Contract.BalanceOf(nil, escrowAddress)
			s.Require().NoError(err)
			s.Require().Equal(transferAmount, escrowBalance)
		}))
	}))

	s.Require().True(s.Run("Receive packets on SimdA", func() {
		var relayTxBodyBz []byte
		s.Require().True(s.Run("Retrieve relay tx", func() {
			resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
				SrcChain:       eth.ChainID.String(),
				DstChain:       simdA.Config().ChainID,
				SourceTxIds:    [][]byte{ethSendTxHash},
				TargetClientId: testvalues.FirstWasmClientID,
			})
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.Tx)
			s.Require().Empty(resp.Address)

			relayTxBodyBz = resp.Tx
		}))

		s.Require().True(s.Run("Broadcast relay tx", func() {
			_ = s.BroadcastSdkTxBody(ctx, simdA, s.SimdARelayerSubmitter, 2_000_000, relayTxBodyBz)
			// NOTE: We don't need to check the response since we don't need to acknowledge the packet
		}))

		s.Require().True(s.Run("Verify balances on Cosmos chain", func() {
			denomOnSimdA := transfertypes.NewDenom(s.contractAddresses.Erc20, transfertypes.NewHop(transfertypes.PortID, testvalues.FirstWasmClientID))

			// User balance on Cosmos chain
			resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, simdA, &banktypes.QueryBalanceRequest{
				Address: simdAUser.FormattedAddress(),
				Denom:   denomOnSimdA.IBCDenom(),
			})
			s.Require().NoError(err)
			s.Require().NotNil(resp.Balance)
			s.Require().Equal(transferAmount, resp.Balance.Amount.BigInt())
			s.Require().Equal(denomOnSimdA.IBCDenom(), resp.Balance.Denom)
		}))
	}))

	var simdASendTxHash []byte
	s.Require().True(s.Run("Send from SimdA to SimdB", func() {
		denomOnSimdA := transfertypes.NewDenom(s.contractAddresses.Erc20, transfertypes.NewHop(transfertypes.PortID, testvalues.FirstWasmClientID))
		timeout := uint64(time.Now().Add(30 * time.Minute).Unix())

		transferPayload := transfertypes.FungibleTokenPacketData{
			// Denom:    denomOnSimdA.IBCDenom(),
			// BUG: Allowing user to choose the above is a bug in ibc-go
			// https://github.com/cosmos/ibc-go/issues/7848
			Denom:    denomOnSimdA.Path(),
			Amount:   transferAmount.String(),
			Sender:   simdAUser.FormattedAddress(),
			Receiver: simdBUser.FormattedAddress(),
			Memo:     "",
		}
		encodedPayload, err := transfertypes.EncodeABIFungibleTokenPacketData(&transferPayload)
		s.Require().NoError(err)

		payload := channeltypesv2.Payload{
			SourcePort:      transfertypes.PortID,
			DestinationPort: transfertypes.PortID,
			Version:         transfertypes.V1,
			Encoding:        transfertypes.EncodingABI,
			Value:           encodedPayload,
		}

		resp, err := s.BroadcastMessages(ctx, simdA, simdAUser, 2_000_000, &channeltypesv2.MsgSendPacket{
			SourceClient:     ibctesting.SecondClientID,
			TimeoutTimestamp: timeout,
			Payloads: []channeltypesv2.Payload{
				payload,
			},
			Signer: simdAUser.FormattedAddress(),
		})
		s.Require().NoError(err)
		s.Require().NotEmpty(resp.TxHash)

		simdASendTxHash, err = hex.DecodeString(resp.TxHash)
		s.Require().NoError(err)
	}))

	s.Require().True(s.Run("Receive packet on SimdB", func() {
		var txBodyBz []byte
		s.Require().True(s.Run("Retrieve relay tx to SimdB", func() {
			resp, err := s.RelayerClient.RelayByTx(context.Background(), &relayertypes.RelayByTxRequest{
				SrcChain:       simdA.Config().ChainID,
				DstChain:       simdB.Config().ChainID,
				SourceTxIds:    [][]byte{simdASendTxHash},
				TargetClientId: ibctesting.SecondClientID,
			})
			s.Require().NoError(err)
			s.Require().NotEmpty(resp.Tx)
			s.Require().Empty(resp.Address)

			txBodyBz = resp.Tx
		}))

		s.Require().True(s.Run("Broadcast relay tx on SimdB", func() {
			_ = s.BroadcastSdkTxBody(ctx, simdB, s.SimdBRelayerSubmitter, 2_000_000, txBodyBz)
			// NOTE: We don't need to check the response since we don't need to acknowledge the packet
		}))

		s.Require().True(s.Run("Verify balances on Cosmos chain", func() {
			finalDenom := transfertypes.NewDenom(
				s.contractAddresses.Erc20,
				transfertypes.NewHop(transfertypes.PortID, ibctesting.SecondClientID),
				transfertypes.NewHop(transfertypes.PortID, testvalues.FirstWasmClientID),
			)

			// User balance on Cosmos chain
			resp, err := e2esuite.GRPCQuery[banktypes.QueryBalanceResponse](ctx, simdB, &banktypes.QueryBalanceRequest{
				Address: simdBUser.FormattedAddress(),
				Denom:   finalDenom.IBCDenom(),
			})
			s.Require().NoError(err)
			s.Require().NotNil(resp.Balance)
			s.Require().Equal(transferAmount, resp.Balance.Amount.BigInt())
			s.Require().Equal(finalDenom.IBCDenom(), resp.Balance.Denom)
		}))
	}))
}
