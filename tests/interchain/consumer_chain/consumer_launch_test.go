package consumer_chain_test

import (
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/suite"
	"golang.org/x/mod/semver"

	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
)

type ConsumerLaunchSuite struct {
	*chainsuite.Suite
	OtherChain                   string
	OtherChainVersionPreUpgrade  string
	OtherChainVersionPostUpgrade string
	ShouldCopyProviderKey        []bool
}

func noProviderKeysCopied() []bool {
	return []bool{false, false, false, false, false, false}
}

func allProviderKeysCopied() []bool {
	return []bool{true, true, true, true, true, true}
}

func someProviderKeysCopied() []bool {
	return []bool{true, false, true, false, true, false}
}

func (s *ConsumerLaunchSuite) TestChainLaunch() {
	cfg := chainsuite.ConsumerConfig{
		ChainName:             s.OtherChain,
		Version:               s.OtherChainVersionPreUpgrade,
		ShouldCopyProviderKey: s.ShouldCopyProviderKey,
		Denom:                 chainsuite.Ucon,
		TopN:                  94,
		Spec: &interchaintest.ChainSpec{
			ChainConfig: ibc.ChainConfig{
				Images: []ibc.DockerImage{
					{
						Repository: chainsuite.HyphaICSRepo,
						Version:    s.OtherChainVersionPreUpgrade,
						UIDGID:     chainsuite.ICSUidGuid,
					},
				},
			},
		},
	}
	consumer, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, cfg)
	s.Require().NoError(err)
	err = s.Chain.CheckCCV(s.GetContext(), consumer, s.Relayer, 1_000_000, 0, 1)
	s.Require().NoError(err)

	s.UpgradeChain()

	err = s.Chain.CheckCCV(s.GetContext(), consumer, s.Relayer, 1_000_000, 0, 1)
	s.Require().NoError(err)
	s.Require().NoError(chainsuite.SendSimpleIBCTx(s.GetContext(), s.Chain, consumer, s.Relayer))

	jailed, err := s.Chain.IsValidatorJailedForConsumerDowntime(s.GetContext(), s.Relayer, consumer, 1)
	s.Require().NoError(err)
	s.Require().True(jailed, "validator 1 should be jailed for downtime")
	jailed, err = s.Chain.IsValidatorJailedForConsumerDowntime(s.GetContext(), s.Relayer, consumer, 5)
	s.Require().NoError(err)
	s.Require().False(jailed, "validator 5 should not be jailed for downtime")

	cfg.Version = s.OtherChainVersionPostUpgrade
	cfg.Spec.ChainConfig.Images[0].Version = s.OtherChainVersionPostUpgrade
	consumer2, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, cfg)
	s.Require().NoError(err)
	err = s.Chain.CheckCCV(s.GetContext(), consumer2, s.Relayer, 1_000_000, 0, 1)
	s.Require().NoError(err)
	s.Require().NoError(chainsuite.SendSimpleIBCTx(s.GetContext(), s.Chain, consumer2, s.Relayer))

	jailed, err = s.Chain.IsValidatorJailedForConsumerDowntime(s.GetContext(), s.Relayer, consumer2, 1)
	s.Require().NoError(err)
	s.Require().True(jailed, "validator 1 should be jailed for downtime")
	jailed, err = s.Chain.IsValidatorJailedForConsumerDowntime(s.GetContext(), s.Relayer, consumer2, 5)
	s.Require().NoError(err)
	s.Require().False(jailed, "validator 5 should not be jailed for downtime")
}

func selectConsumerVersion(preV21, postV21 string) string {
	if semver.Compare(semver.Major(chainsuite.GetEnvironment().OldGaiaImageVersion), "v21") >= 0 {
		return postV21
	}
	return preV21
}

func TestICS4ChainLaunch(t *testing.T) {
	s := &ConsumerLaunchSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			CreateRelayer: true,
			ChainSpec: &interchaintest.ChainSpec{
				NumValidators: &chainsuite.SixValidators,
			},
		}),
		OtherChain:                   "ics-consumer",
		OtherChainVersionPreUpgrade:  selectConsumerVersion("v4.4.1", "v4.5.0"),
		OtherChainVersionPostUpgrade: "v4.5.0",
		ShouldCopyProviderKey:        noProviderKeysCopied(),
	}
	suite.Run(t, s)
}

func TestICS6ConsumerAllKeysChainLaunch(t *testing.T) {
	s := &ConsumerLaunchSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			CreateRelayer: true,
			ChainSpec: &interchaintest.ChainSpec{
				NumValidators: &chainsuite.SixValidators,
			},
		}),
		OtherChain:                   "ics-consumer",
		OtherChainVersionPreUpgrade:  selectConsumerVersion("v6.0.0", "v6.2.1"),
		OtherChainVersionPostUpgrade: "v6.2.1",
		ShouldCopyProviderKey:        allProviderKeysCopied(),
	}
	suite.Run(t, s)
}

func TestICS6ConsumerSomeKeysChainLaunch(t *testing.T) {
	s := &ConsumerLaunchSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			CreateRelayer: true,
			ChainSpec: &interchaintest.ChainSpec{
				NumValidators: &chainsuite.SixValidators,
			},
		}),
		OtherChain:                   "ics-consumer",
		OtherChainVersionPreUpgrade:  selectConsumerVersion("v6.0.0", "v6.2.1"),
		OtherChainVersionPostUpgrade: "v6.2.1",
		ShouldCopyProviderKey:        someProviderKeysCopied(),
	}
	suite.Run(t, s)
}

func TestICS6ConsumerNoKeysChainLaunch(t *testing.T) {
	s := &ConsumerLaunchSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			CreateRelayer: true,
			ChainSpec: &interchaintest.ChainSpec{
				NumValidators: &chainsuite.SixValidators,
			},
		}),
		OtherChain:                   "ics-consumer",
		OtherChainVersionPreUpgrade:  selectConsumerVersion("v6.0.0", "v6.2.1"),
		OtherChainVersionPostUpgrade: "v6.2.1",
		ShouldCopyProviderKey:        noProviderKeysCopied(),
	}
	suite.Run(t, s)
}
