package delegator_test

import (
	"fmt"
	"testing"

<<<<<<< HEAD
	"github.com/cosmos/gaia/v27/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v27/tests/interchain/delegator"
=======
	"github.com/cosmos/gaia/v28/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v28/tests/interchain/delegator"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
>>>>>>> ad3637a (!feat: Update wasm configuration (#4082))
	"github.com/stretchr/testify/suite"
)

type CosmWasmSuite struct {
	*delegator.Suite
	PreUpgradeContractCode string
	PreUpgradeContractAddr string
}

const (
	initState = `{"count": 100}`
	query     = `{"get_count":{}}`
	increment = `{"increment":{}}`

	contractFile          = "testdata/contract.wasm"
	inRangeContractFile   = "testdata/contract_in_range.wasm"
	oversizedContractFile = "testdata/contract_oversized.wasm"

	proposalQueryContractFile  = "testdata/proposal_query.wasm"
	validatorQueryContractFile = "testdata/validator_query.wasm"
)

func (s *CosmWasmSuite) SetupSuite() {
	s.Suite.SetupSuite()

	code, contractAddr := s.storeAndInstantiate(contractFile, initState)
	s.PreUpgradeContractCode = code
	s.PreUpgradeContractAddr = contractAddr

	// Pre-upgrade, the chain still enforces the 800KB cap, so the in-range contract must be rejected.
	s.assertStoreRejected(inRangeContractFile)

	s.UpgradeChain()
}

func (s *CosmWasmSuite) TestPreUpgradeContract() {
	count := s.getContractCount(s.PreUpgradeContractAddr)
	s.Require().Equal(int64(100), count)

	s.executeContractByTx(s.PreUpgradeContractAddr)

	count = s.getContractCount(s.PreUpgradeContractAddr)
	s.Require().Equal(int64(101), count)

	s.executeContractByTx(s.PreUpgradeContractAddr)

	count = s.getContractCount(s.PreUpgradeContractAddr)
	s.Require().Equal(int64(102), count)
}

func (s *CosmWasmSuite) TestWasmSizeCapRaised() {
	// A contract between the old 800KB cap and the new 1.6MiB cap is rejected pre-upgrade
	// (asserted in SetupSuite) and must be accepted, and fully functional, post-upgrade.
	_, contractAddr := s.storeAndInstantiate(inRangeContractFile, initState)

	count := s.getContractCount(contractAddr)
	s.Require().Equal(int64(100), count)

	s.executeContractByTx(contractAddr)

	count = s.getContractCount(contractAddr)
	s.Require().Equal(int64(101), count)
}

func (s *CosmWasmSuite) TestWasmSizeCapStillEnforced() {
	// A contract above the new 1.6MiB cap must still be rejected post-upgrade.
	s.assertStoreRejected(oversizedContractFile)
}

func (s *CosmWasmSuite) TestContractCanQueryProposal() {
	// Submit a proposal with a deposit above the minimum so it enters voting period
	// immediately, giving us a known proposal ID for the contract to query.
	prop, err := s.Chain.BuildProposal(nil, "Query Plugin Test Proposal", "Exercises the wasm Grpc query plugin",
		"ipfs://CID", chainsuite.GovDepositAmount, s.DelegatorWallet.FormattedAddress(), false)
	s.Require().NoError(err)
	result, err := s.Chain.SubmitProposal(s.GetContext(), s.DelegatorWallet.KeyName(), prop)
	s.Require().NoError(err)

	_, contractAddr := s.storeAndInstantiate(proposalQueryContractFile, "{}")

	queryMsg := fmt.Sprintf(`{"proposal":{"proposal_id":%s}}`, result.ProposalID)
	title, err := s.Chain.QueryJSON(s.GetContext(), "data.title", "wasm", "contract-state", "smart", contractAddr, queryMsg)
	s.Require().NoError(err)
	s.Require().Equal("Query Plugin Test Proposal", title.String())
}

func (s *CosmWasmSuite) TestContractCanQueryValidator() {
	_, contractAddr := s.storeAndInstantiate(validatorQueryContractFile, "{}")

	valoperAddr := s.Chain.ValidatorWallets[0].ValoperAddress
	queryMsg := fmt.Sprintf(`{"validator":{"validator_addr":"%s"}}`, valoperAddr)
	operatorAddr, err := s.Chain.QueryJSON(s.GetContext(), "data.operator_address", "wasm", "contract-state", "smart", contractAddr, queryMsg)
	s.Require().NoError(err)
	s.Require().Equal(valoperAddr, operatorAddr.String())
}

func (s *CosmWasmSuite) TestCreateNewContract() {
	_, contractAddr := s.storeAndInstantiate(contractFile, initState)

	count := s.getContractCount(contractAddr)
	s.Require().Equal(int64(100), count)

	s.executeContractByTx(contractAddr)

	count = s.getContractCount(contractAddr)
	s.Require().Equal(int64(101), count)

	s.executeContractByTx(contractAddr)

	count = s.getContractCount(contractAddr)
	s.Require().Equal(int64(102), count)
}

// executeContractByTx uses a plain ExecTx rather than the interchaintest ExecuteContract helper,
// since that helper decodes the tx via an interface registry that doesn't know about wasmd's
// message types on this test chain, which fails with "unable to resolve type URL ...".
func (s *CosmWasmSuite) executeContractByTx(contractAddr string) {
	_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet.FormattedAddress(),
		"wasm", "execute", contractAddr, increment,
	)
	s.Require().NoError(err)
}

func (s *CosmWasmSuite) getContractCount(contractAddr string) int64 {
	countJSON, err := s.Chain.QueryJSON(s.GetContext(), "data.count", "wasm", "contract-state", "smart", contractAddr, query)
	s.Require().NoError(err)
	count := countJSON.Int()
	return count
}

// storeAndInstantiate uses the interchaintest StoreContract helper (safe: it never decodes the
// tx via the interface registry) but instantiates via a plain ExecTx + query, for the same reason
// executeContractByTx avoids the InstantiateContract helper.
func (s *CosmWasmSuite) storeAndInstantiate(filePath, initState string) (string, string) {
	codeID, err := s.Chain.StoreContract(s.GetContext(), s.DelegatorWallet.FormattedAddress(), filePath)
	s.Require().NoError(err)

	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet.FormattedAddress(),
		"wasm", "instantiate", codeID, initState, "--label", "my-contract", "--no-admin",
	)
	s.Require().NoError(err)

	// @reverse.0 takes the most recently instantiated contract for this code id.
	contractAddrJSON, err := s.Chain.QueryJSON(s.GetContext(), "contracts.@reverse.0", "wasm", "list-contract-by-code", codeID)
	s.Require().NoError(err)
	contractAddr := contractAddrJSON.String()

	return codeID, contractAddr
}

// assertStoreRejected submits a direct wasm store tx for a contract that is expected to fail
// wasmd's MaxWasmSize validation, unlike storeAndInstantiate which expects it to succeed.
func (s *CosmWasmSuite) assertStoreRejected(filePath string) {
	_, err := s.Chain.StoreContract(s.GetContext(), s.DelegatorWallet.FormattedAddress(), filePath)
	s.Require().Error(err)
}

func TestCosmWasm(t *testing.T) {
	// Wasm code upload/instantiation is permissionless on the Cosmos Hub.
	wasmGenesis := append(chainsuite.DefaultGenesis(),
		cosmos.NewGenesisKV("app_state.wasm.params.code_upload_access.permission", "Everybody"),
		cosmos.NewGenesisKV("app_state.wasm.params.instantiate_default_permission", "Everybody"),
	)
	chainSpec := chainsuite.DefaultChainSpec(chainsuite.GetEnvironment())
	chainSpec.ChainConfig.ModifyGenesis = cosmos.ModifyGenesis(wasmGenesis)

	s := &CosmWasmSuite{
		Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			ChainSpec:      chainSpec,
			UpgradeOnSetup: false,
		})},
	}
	suite.Run(t, s)
}
