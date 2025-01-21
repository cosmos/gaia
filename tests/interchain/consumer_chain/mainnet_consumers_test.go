package consumer_chain_test

import (
	"testing"

	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/suite"
	"golang.org/x/mod/semver"
)

type MainnetConsumerChainsSuite struct {
	*chainsuite.Suite
}

func (s *MainnetConsumerChainsSuite) TestMainnetConsumerChainsAfterUpgrade() {
	// We can't do these consumer launches yet because the chains aren't compatible with launching on v21 yet
	if semver.Major(s.Env.OldGaiaImageVersion) == s.Env.UpgradeName && s.Env.UpgradeName == "v21" {
		s.T().Skip("Skipping Consumer Launch tests when going from v21 -> v21")
	}
	neutron, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, chainsuite.ConsumerConfig{
		ChainName:             "neutron",
		Version:               chainsuite.NeutronVersion,
		ShouldCopyProviderKey: allProviderKeysCopied(),
		Denom:                 chainsuite.NeutronDenom,
		TopN:                  95,
	})
	s.Require().NoError(err)
	stride, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, chainsuite.ConsumerConfig{
		ChainName:             "stride",
		Version:               chainsuite.StrideVersion,
		ShouldCopyProviderKey: allProviderKeysCopied(),
		Denom:                 chainsuite.StrideDenom,
		TopN:                  95,
	})
	s.Require().NoError(err)

	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), neutron, s.Relayer, 1_000_000, 0, 1))
	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), stride, s.Relayer, 1_000_000, 0, 1))

	s.UpgradeChain()

	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), neutron, s.Relayer, 1_000_000, 0, 1))
	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), stride, s.Relayer, 1_000_000, 0, 1))
	s.Require().NoError(chainsuite.SendSimpleIBCTx(s.GetContext(), s.Chain, neutron, s.Relayer))
	s.Require().NoError(chainsuite.SendSimpleIBCTx(s.GetContext(), s.Chain, stride, s.Relayer))
}

func TestMainnetConsumerChainsAfterUpgrade(t *testing.T) {
	s := &MainnetConsumerChainsSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			CreateRelayer: true,
			ChainSpec: &interchaintest.ChainSpec{
				NumValidators: &chainsuite.SixValidators,
			},
		}),
	}
	suite.Run(t, s)
}
