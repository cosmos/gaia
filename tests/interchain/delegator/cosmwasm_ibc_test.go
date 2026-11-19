package delegator_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/gaia/v29/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v29/tests/interchain/delegator"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v10/modules/core/23-commitment/types"
	commitmenttypesv2 "github.com/cosmos/ibc-go/v10/modules/core/23-commitment/types/v2"
	host "github.com/cosmos/ibc-go/v10/modules/core/24-host"
	ibctm "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/stretchr/testify/suite"
)

const (
	// grpcQueryProxyContractFile forwards {"grpc":{"path","data"}} to the chain's
	// wasm Grpc query plugin and returns the raw protobuf response.
	grpcQueryProxyContractFile = "testdata/grpc_query_proxy.wasm"

	verifyMembershipPath = "/ibc.core.client.v1.Query/VerifyMembership"
	consensusStatePath   = "/ibc.core.client.v1.Query/ConsensusState"
	clientStatusPath     = "/ibc.core.client.v1.Query/ClientStatus"
	clientStatePath      = "/ibc.core.client.v1.Query/ClientState"
)

type CosmWasmIBCSuite struct {
	*delegator.Suite
	Counterparty *chainsuite.Chain
	ProxyAddr    string
	// ClientID is the light client of Counterparty tracked on the Hub.
	ClientID string
	// CounterpartyClientID is the light client of the Hub tracked on Counterparty,
	// whose client state is used as the key/value to prove.
	CounterpartyClientID string
}

func (s *CosmWasmIBCSuite) SetupSuite() {
	s.Suite.SetupSuite()

	counterparty, err := s.Chain.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, chainsuite.DefaultChainSpec(s.Env))
	s.Require().NoError(err)
	s.Counterparty = counterparty

	s.ClientID = s.clientIDFor(s.Chain, s.Counterparty)
	s.CounterpartyClientID = s.clientIDFor(s.Counterparty, s.Chain)

	codeID, err := s.Chain.StoreContract(s.GetContext(), s.DelegatorWallet.FormattedAddress(), grpcQueryProxyContractFile)
	s.Require().NoError(err)
	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet.FormattedAddress(),
		"wasm", "instantiate", codeID, "{}", "--label", "grpc-query-proxy", "--no-admin",
	)
	s.Require().NoError(err)
	contractAddrJSON, err := s.Chain.QueryJSON(s.GetContext(), "contracts.@reverse.0", "wasm", "list-contract-by-code", codeID)
	s.Require().NoError(err)
	s.ProxyAddr = contractAddrJSON.String()
}

func (s *CosmWasmIBCSuite) TestContractCanVerifyMembership() {
	req := s.counterpartyClientStateProof()

	res := &clienttypes.QueryVerifyMembershipResponse{}
	s.Require().NoError(s.contractGrpcQuery(verifyMembershipPath, req, res))
	s.Require().True(res.Success)
}

func (s *CosmWasmIBCSuite) TestContractVerifyMembershipRejectsWrongValue() {
	req := s.counterpartyClientStateProof()
	req.Value = append(req.Value, 0x00)

	res := &clienttypes.QueryVerifyMembershipResponse{}
	s.Require().NoError(s.contractGrpcQuery(verifyMembershipPath, req, res))
	s.Require().False(res.Success)
}

func (s *CosmWasmIBCSuite) TestContractCanQueryConsensusState() {
	height := s.latestClientHeight()

	res := &clienttypes.QueryConsensusStateResponse{}
	s.Require().NoError(s.contractGrpcQuery(consensusStatePath, &clienttypes.QueryConsensusStateRequest{
		ClientId:       s.ClientID,
		RevisionNumber: height.RevisionNumber,
		RevisionHeight: height.RevisionHeight,
	}, res))

	s.Require().Equal("/ibc.lightclients.tendermint.v1.ConsensusState", res.ConsensusState.TypeUrl)
	consensusState := &ibctm.ConsensusState{}
	s.Require().NoError(consensusState.Unmarshal(res.ConsensusState.Value))

	// The trusted root must be the app hash committed in the counterparty's block header.
	blockHeight := int64(height.RevisionHeight)
	block, err := s.counterpartyRPC().Block(s.GetContext(), &blockHeight)
	s.Require().NoError(err)
	s.Require().Equal([]byte(block.Block.AppHash), consensusState.Root.Hash)
}

func (s *CosmWasmIBCSuite) TestContractCanQueryClientStatus() {
	res := &clienttypes.QueryClientStatusResponse{}
	s.Require().NoError(s.contractGrpcQuery(clientStatusPath, &clienttypes.QueryClientStatusRequest{ClientId: s.ClientID}, res))
	s.Require().Equal("Active", res.Status)
}

func (s *CosmWasmIBCSuite) TestContractCannotQueryNonAllowlistedPath() {
	err := s.contractGrpcQuery(clientStatePath, &clienttypes.QueryClientStateRequest{ClientId: s.ClientID}, &clienttypes.QueryClientStateResponse{})
	s.Require().Error(err)
}

// counterpartyClientStateProof builds a VerifyMembership request proving the client
// state stored under CounterpartyClientID on the counterparty, against the consensus
// state at the latest height of the Hub's client.
func (s *CosmWasmIBCSuite) counterpartyClientStateProof() *clienttypes.QueryVerifyMembershipRequest {
	height := s.latestClientHeight()
	key := host.FullClientStateKey(s.CounterpartyClientID)

	// A proof queried at height H verifies against the app hash in the header of block H+1.
	queryHeight := int64(height.RevisionHeight) - 1
	abciRes, err := s.counterpartyRPC().ABCIQueryWithOptions(s.GetContext(), "store/ibc/key", key,
		rpcclient.ABCIQueryOptions{Height: queryHeight, Prove: true})
	s.Require().NoError(err)
	s.Require().True(abciRes.Response.IsOK(), abciRes.Response.Log)
	s.Require().Equal(queryHeight, abciRes.Response.Height)
	s.Require().NotEmpty(abciRes.Response.Value)

	merkleProof, err := commitmenttypes.ConvertProofs(abciRes.Response.ProofOps)
	s.Require().NoError(err)
	proof, err := merkleProof.Marshal()
	s.Require().NoError(err)

	return &clienttypes.QueryVerifyMembershipRequest{
		ClientId:    s.ClientID,
		Proof:       proof,
		ProofHeight: height,
		MerklePath:  commitmenttypesv2.NewMerklePath([]byte("ibc"), key),
		Value:       abciRes.Response.Value,
	}
}

func (s *CosmWasmIBCSuite) latestClientHeight() clienttypes.Height {
	latest, err := s.Chain.QueryJSON(s.GetContext(), "client_state.latest_height", "ibc", "client", "state", s.ClientID)
	s.Require().NoError(err)
	return clienttypes.NewHeight(latest.Get("revision_number").Uint(), latest.Get("revision_height").Uint())
}

func (s *CosmWasmIBCSuite) clientIDFor(chain, counterparty *chainsuite.Chain) string {
	clients, err := s.Relayer.GetClients(s.GetContext(), chainsuite.GetRelayerExecReporter(s.GetContext()), chain.Config().ChainID)
	s.Require().NoError(err)
	for _, c := range clients {
		if c.ClientState.ChainID == counterparty.Config().ChainID {
			return c.ClientID
		}
	}
	s.FailNow(fmt.Sprintf("no client of %s found on %s", counterparty.Config().ChainID, chain.Config().ChainID))
	return ""
}

func (s *CosmWasmIBCSuite) counterpartyRPC() *rpchttp.HTTP {
	client, err := rpchttp.New(s.Counterparty.GetHostRPCAddress(), "/websocket")
	s.Require().NoError(err)
	return client
}

// contractGrpcQuery sends req to path through the proxy contract and decodes the response into res.
func (s *CosmWasmIBCSuite) contractGrpcQuery(path string, req, res interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
},
) error {
	reqBz, err := req.Marshal()
	s.Require().NoError(err)

	queryMsg := fmt.Sprintf(`{"grpc":{"path":%q,"data":%q}}`, path, base64.StdEncoding.EncodeToString(reqBz))
	data, err := s.Chain.QueryJSON(s.GetContext(), "data", "wasm", "contract-state", "smart", s.ProxyAddr, queryMsg)
	if err != nil {
		return err
	}
	resBz, err := base64.StdEncoding.DecodeString(data.String())
	s.Require().NoError(err)
	return res.Unmarshal(resBz)
}

func TestCosmWasmIBC(t *testing.T) {
	// Wasm code upload/instantiation is permissionless on the Cosmos Hub.
	wasmGenesis := append(chainsuite.DefaultGenesis(),
		cosmos.NewGenesisKV("app_state.wasm.params.code_upload_access.permission", "Everybody"),
		cosmos.NewGenesisKV("app_state.wasm.params.instantiate_default_permission", "Everybody"),
	)
	chainSpec := chainsuite.DefaultChainSpec(chainsuite.GetEnvironment())
	chainSpec.ChainConfig.ModifyGenesis = cosmos.ModifyGenesis(wasmGenesis)

	s := &CosmWasmIBCSuite{
		Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			ChainSpec:      chainSpec,
			UpgradeOnSetup: true,
			CreateRelayer:  true,
		})},
	}
	suite.Run(t, s)
}
