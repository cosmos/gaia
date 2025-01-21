package delegator_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v23/tests/interchain/delegator"
	"github.com/stretchr/testify/suite"
)

type CosmWasmSuite struct {
	*delegator.Suite
	ContractWasm           []byte
	ContractPath           string
	PreUpgradeContractCode string
	PreUpgradeContractAddr string
}

const (
	initState = `{"count": 100}`
	query     = `{"get_count":{}}`
	increment = `{"increment":{}}`
)

func (s *CosmWasmSuite) SetupSuite() {
	s.Suite.SetupSuite()

	contractWasm, err := os.ReadFile("testdata/contract.wasm")
	s.Require().NoError(err)
	s.ContractWasm = contractWasm
	s.Require().NoError(s.Chain.GetNode().WriteFile(s.GetContext(), s.ContractWasm, "contract.wasm"))
	s.ContractPath = path.Join(s.Chain.GetNode().HomeDir(), "contract.wasm")

	code, contractAddr := s.storeInstantiateProposal(initState)
	s.PreUpgradeContractCode = code
	s.PreUpgradeContractAddr = contractAddr

	s.UpgradeChain()
}

func (s *CosmWasmSuite) TestPreUpgradeContract() {
	count := s.getContractCount(s.PreUpgradeContractAddr)
	s.Require().Equal(int64(100), count)

	s.executeContractByTx(s.PreUpgradeContractAddr)

	count = s.getContractCount(s.PreUpgradeContractAddr)
	s.Require().Equal(int64(101), count)

	s.executeContractByProposal(s.PreUpgradeContractAddr)

	count = s.getContractCount(s.PreUpgradeContractAddr)
	s.Require().Equal(int64(102), count)
}

func (s *CosmWasmSuite) TestCantStoreWithoutProp() {
	infos, err := s.Chain.QueryJSON(s.GetContext(), "code_infos", "wasm", "list-code")
	s.Require().NoError(err)
	codeCountBefore := len(infos.Array())

	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet.FormattedAddress(),
		"wasm", "store", s.ContractPath,
	)
	s.Require().Error(err)

	infos, err = s.Chain.QueryJSON(s.GetContext(), "code_infos", "wasm", "list-code")
	s.Require().NoError(err)
	codeCountAfter := len(infos.Array())
	s.Require().Equal(codeCountBefore, codeCountAfter)
}

func (s *CosmWasmSuite) TestCantInstantiateWithoutProp() {
	_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet.FormattedAddress(),
		"wasm", "instantiate", s.PreUpgradeContractCode, initState, "--label", "my-contract", "--no-admin",
	)
	s.Require().Error(err)
}

func (s *CosmWasmSuite) TestCreateNewContract() {
	_, contractAddr := s.storeInstantiateProposal(initState)

	count := s.getContractCount(contractAddr)
	s.Require().Equal(int64(100), count)

	s.executeContractByTx(contractAddr)

	count = s.getContractCount(contractAddr)
	s.Require().Equal(int64(101), count)

	s.executeContractByProposal(contractAddr)

	count = s.getContractCount(contractAddr)
	s.Require().Equal(int64(102), count)
}

func (s *CosmWasmSuite) executeContractByTx(contractAddr string) {
	_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet.FormattedAddress(),
		"wasm", "execute", contractAddr, increment,
	)
	s.Require().NoError(err)
}

func (s *CosmWasmSuite) executeContractByProposal(contractAddr string) {
	txhash, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet.FormattedAddress(),
		"wasm", "submit-proposal", "execute-contract",
		contractAddr, increment,
		"--title", "Increment count",
		"--summary", "Increment count",
		"--deposit", fmt.Sprintf("1000000%s", s.Config.ChainSpec.Denom),
	)
	s.Require().NoError(err)

	proposalId, err := s.Chain.GetProposalID(s.GetContext(), txhash)
	s.Require().NoError(err)

	err = s.Chain.PassProposal(s.GetContext(), proposalId)
	s.Require().NoError(err)
}

func (s *CosmWasmSuite) getContractCount(contractAddr string) int64 {
	countJSON, err := s.Chain.QueryJSON(s.GetContext(), "data.count", "wasm", "contract-state", "smart", contractAddr, query)
	s.Require().NoError(err)
	count := countJSON.Int()
	return count
}

func (s *CosmWasmSuite) storeInstantiateProposal(initState string) (string, string) {
	govAddr, err := s.Chain.GetGovernanceAddress(s.GetContext())
	s.Require().NoError(err)

	txhash, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.DelegatorWallet.FormattedAddress(),
		"wasm", "submit-proposal", "store-instantiate",
		s.ContractPath,
		initState, "--label", "my-contract",
		"--no-admin", "--instantiate-nobody", "true",
		"--title", "Store and instantiate template",
		"--summary", "Store and instantiate template",
		"--deposit", fmt.Sprintf("10000000%s", s.Config.ChainSpec.Denom),
	)
	s.Require().NoError(err)

	proposalId, err := s.Chain.GetProposalID(s.GetContext(), txhash)
	s.Require().NoError(err)

	err = s.Chain.PassProposal(s.GetContext(), proposalId)
	s.Require().NoError(err)

	codeJSON, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("code_infos.@reverse.#(creator=\"%s\").code_id", govAddr), "wasm", "list-code")
	s.Require().NoError(err)
	code := codeJSON.String()

	contractAddrJSON, err := s.Chain.QueryJSON(s.GetContext(), "contracts.0", "wasm", "list-contract-by-code", code)
	s.Require().NoError(err)
	contractAddr := contractAddrJSON.String()
	return code, contractAddr
}

func TestCosmWasm(t *testing.T) {
	s := &CosmWasmSuite{
		Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{UpgradeOnSetup: false})},
	}
	suite.Run(t, s)
}
