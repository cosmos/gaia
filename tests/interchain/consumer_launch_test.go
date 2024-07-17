package interchain_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/gaia/v20/tests/interchain/chainsuite"
)

type ConsumerLaunchSuite struct {
	*chainsuite.Suite
	OtherChain            string
	OtherChainVersion     string
	ShouldCopyProviderKey [chainsuite.ValidatorCount]bool
}

func noProviderKeysCopied() [chainsuite.ValidatorCount]bool {
	return [chainsuite.ValidatorCount]bool{false, false, false, false, false, false}
}

func allProviderKeysCopied() [chainsuite.ValidatorCount]bool {
	return [chainsuite.ValidatorCount]bool{true, true, true, true, true, true}
}

func someProviderKeysCopied() [chainsuite.ValidatorCount]bool {
	return [chainsuite.ValidatorCount]bool{true, false, true, false, true, false}
}

func (s *ConsumerLaunchSuite) TestChainLaunch() {
	cfg := chainsuite.ConsumerConfig{
		ChainName:             s.OtherChain,
		Version:               s.OtherChainVersion,
		ShouldCopyProviderKey: s.ShouldCopyProviderKey,
		Denom:                 chainsuite.Ucon,
		TopN:                  95,
	}
	consumer, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, cfg)
	s.Require().NoError(err)
	err = s.Chain.CheckCCV(s.GetContext(), consumer, s.Relayer, 1_000_000, 0, 1)
	s.Require().NoError(err)

	s.UpgradeChain()

	err = s.Chain.CheckCCV(s.GetContext(), consumer, s.Relayer, 1_000_000, 0, 1)
	s.Require().NoError(err)
	consumer2, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, cfg)
	s.Require().NoError(err)
	err = s.Chain.CheckCCV(s.GetContext(), consumer2, s.Relayer, 1_000_000, 0, 1)
	s.Require().NoError(err)
}

func TestICS40ChainLaunch(t *testing.T) {
	s := &ConsumerLaunchSuite{
		Suite:                 chainsuite.NewSuite(chainsuite.SuiteConfig{CreateRelayer: true}),
		OtherChain:            "ics-consumer",
		OtherChainVersion:     "v4.0.0",
		ShouldCopyProviderKey: noProviderKeysCopied(),
	}
	suite.Run(t, s)
}

func TestICS33ConsumerAllKeysChainLaunch(t *testing.T) {
	s := &ConsumerLaunchSuite{
		Suite:                 chainsuite.NewSuite(chainsuite.SuiteConfig{CreateRelayer: true}),
		OtherChain:            "ics-consumer",
		OtherChainVersion:     "v3.3.0",
		ShouldCopyProviderKey: allProviderKeysCopied(),
	}
	suite.Run(t, s)
}

func TestICS33ConsumerSomeKeysChainLaunch(t *testing.T) {
	s := &ConsumerLaunchSuite{
		Suite:                 chainsuite.NewSuite(chainsuite.SuiteConfig{CreateRelayer: true}),
		OtherChain:            "ics-consumer",
		OtherChainVersion:     "v3.3.0",
		ShouldCopyProviderKey: someProviderKeysCopied(),
	}
	suite.Run(t, s)
}

func TestICS33ConsumerNoKeysChainLaunch(t *testing.T) {
	s := &ConsumerLaunchSuite{
		Suite:                 chainsuite.NewSuite(chainsuite.SuiteConfig{CreateRelayer: true}),
		OtherChain:            "ics-consumer",
		OtherChainVersion:     "v3.3.0",
		ShouldCopyProviderKey: noProviderKeysCopied(),
	}
	suite.Run(t, s)
}

type MainnetConsumerChainsSuite struct {
	*chainsuite.Suite
}

func (s *MainnetConsumerChainsSuite) TestMainnetConsumerChainsAfterUpgrade() {
	const neutronVersion = "v3.0.2"
	const strideVersion = "v22.0.0"

	neutron, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, chainsuite.ConsumerConfig{
		ChainName:             "neutron",
		Version:               neutronVersion,
		ShouldCopyProviderKey: allProviderKeysCopied(),
		Denom:                 chainsuite.NeutronDenom,
		TopN:                  95,
	})
	s.Require().NoError(err)
	stride, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, chainsuite.ConsumerConfig{
		ChainName:             "stride",
		Version:               strideVersion,
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
}

func TestMainnetConsumerChainsAfterUpgrade(t *testing.T) {
	s := &MainnetConsumerChainsSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{CreateRelayer: true}),
	}
	suite.Run(t, s)
}
