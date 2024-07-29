package interchain_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/cosmos/gaia/v20/tests/interchain/chainsuite"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/suite"
)

type CosmWasmSuite struct {
	*chainsuite.Suite
}

func (s *CosmWasmSuite) TestPermissionedCosmWasm() {
	const (
		initState = `{"count": 100}`
		query     = `{"get_count":{}}`
		increment = `{"increment":{}}`
	)

	contractWasm, err := os.ReadFile("testdata/contract.wasm")
	s.Require().NoError(err)

	s.Require().NoError(s.Chain.GetNode().WriteFile(s.GetContext(), contractWasm, "contract.wasm"))
	contractPath := path.Join(s.Chain.GetNode().HomeDir(), "contract.wasm")

	govAddr, err := s.Chain.GetGovernanceAddress(s.GetContext())
	s.Require().NoError(err)

	infos, err := s.Chain.QueryJSON(s.GetContext(), "code_infos", "wasm", "list-code")
	s.Require().NoError(err)
	codeCountBefore := len(infos.Array())

	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), interchaintest.FaucetAccountKeyName,
		"wasm", "store", contractPath,
	)
	s.Require().Error(err)

	infos, err = s.Chain.QueryJSON(s.GetContext(), "code_infos", "wasm", "list-code")
	s.Require().NoError(err)
	codeCountAfter := len(infos.Array())
	s.Require().Equal(codeCountBefore, codeCountAfter)

	txhash, err := s.Chain.GetNode().ExecTx(s.GetContext(), interchaintest.FaucetAccountKeyName,
		"wasm", "submit-proposal", "store-instantiate",
		contractPath,
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

	codeJSON, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("code_infos.#(creator=\"%s\").code_id", govAddr), "wasm", "list-code")
	s.Require().NoError(err)
	code := codeJSON.String()

	contractAddrJSON, err := s.Chain.QueryJSON(s.GetContext(), "contracts.0", "wasm", "list-contract-by-code", code)
	s.Require().NoError(err)
	contractAddr := contractAddrJSON.String()

	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), interchaintest.FaucetAccountKeyName,
		"wasm", "instantiate", code, initState, "--label", "my-contract", "--no-admin",
	)
	s.Require().Error(err)

	countJSON, err := s.Chain.QueryJSON(s.GetContext(), "data.count", "wasm", "contract-state", "smart", contractAddr, query)
	s.Require().NoError(err)
	count := countJSON.Int()
	s.Require().Equal(int64(100), count)

	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), interchaintest.FaucetAccountKeyName,
		"wasm", "execute", contractAddr, increment,
	)
	s.Require().NoError(err)

	countJSON, err = s.Chain.QueryJSON(s.GetContext(), "data.count", "wasm", "contract-state", "smart", contractAddr, query)
	s.Require().NoError(err)
	count = countJSON.Int()
	s.Require().Equal(int64(101), count)

	txhash, err = s.Chain.GetNode().ExecTx(s.GetContext(), interchaintest.FaucetAccountKeyName,
		"wasm", "submit-proposal", "execute-contract",
		contractAddr, increment,
		"--title", "Increment count",
		"--summary", "Increment count",
		"--deposit", fmt.Sprintf("1000000%s", s.Config.ChainSpec.Denom),
	)
	s.Require().NoError(err)

	proposalId, err = s.Chain.GetProposalID(s.GetContext(), txhash)
	s.Require().NoError(err)

	err = s.Chain.PassProposal(s.GetContext(), proposalId)
	s.Require().NoError(err)

	countJSON, err = s.Chain.QueryJSON(s.GetContext(), "data.count", "wasm", "contract-state", "smart", contractAddr, query)
	s.Require().NoError(err)
	count = countJSON.Int()
	s.Require().Equal(int64(102), count)
}

func TestCosmWasm(t *testing.T) {
	s := &CosmWasmSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{UpgradeOnSetup: true}),
	}
	suite.Run(t, s)
}
