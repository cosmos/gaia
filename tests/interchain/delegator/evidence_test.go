package delegator_test

import (
	"testing"
	"time"

	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v23/tests/interchain/delegator"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/suite"
)

type EvidenceSuite struct {
	*delegator.Suite
}

func (s *EvidenceSuite) TestDoubleSigning() {
	from, to := 0, 3
	privkey, err := s.Chain.Validators[from].PrivValFileContent(s.GetContext())
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.Validators[to].OverwritePrivValFile(s.GetContext(), privkey))

	s.Require().NoError(s.Chain.StopAllNodes(s.GetContext()))

	s.Require().NoError(s.Chain.Validators[to].CreateNodeContainer(s.GetContext()))
	s.Require().NoError(s.Chain.Validators[to].StartContainer(s.GetContext()))
	time.Sleep(10 * time.Second)
	s.Require().NoError(s.Chain.Validators[from].CreateNodeContainer(s.GetContext()))
	s.Require().NoError(s.Chain.Validators[from].StartContainer(s.GetContext()))
	time.Sleep(10 * time.Second)

	for i := 0; i < len(s.Chain.Validators); i++ {
		if i == from || i == to {
			continue
		}
		s.Require().NoError(s.Chain.Validators[i].CreateNodeContainer(s.GetContext()))
		s.Require().NoError(s.Chain.Validators[i].StartContainer(s.GetContext()))
	}
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 5, s.Chain))

	evidence, err := s.Chain.QueryJSON(s.GetContext(), "evidence", "evidence", "list")
	s.Require().NoError(err)
	s.Require().NotEmpty(evidence.Array())
	valcons := evidence.Get("0.value.consensus_address").String()
	s.Require().NotEmpty(valcons)
	s.Require().Equal(s.Chain.ValidatorWallets[from].ValConsAddress, valcons)
}

func TestEvidence(t *testing.T) {
	nodes := 4
	s := &EvidenceSuite{
		Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			UpgradeOnSetup: true,
			ChainSpec: &interchaintest.ChainSpec{
				NumValidators: &nodes,
			},
		})},
	}
	suite.Run(t, s)
}
