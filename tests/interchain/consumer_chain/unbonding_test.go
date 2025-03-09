package consumer_chain_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/suite"
	"golang.org/x/mod/semver"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UnbondingSuite struct {
	*chainsuite.Suite
	Consumer *chainsuite.Chain
}

const (
	vscTimeoutPeriod  = 450 * time.Second
	unbondingTime     = 420 * time.Second
	initTimeoutPeriod = 180 * time.Second
)

func (s *UnbondingSuite) SetupSuite() {
	s.Suite.SetupSuite()
	cfg := chainsuite.ConsumerConfig{
		ChainName:             "ics-consumer",
		Version:               selectConsumerVersion("v6.0.0", "v6.2.1"),
		ShouldCopyProviderKey: allProviderKeysCopied(),
		Denom:                 chainsuite.Ucon,
		TopN:                  100,
		Spec: &interchaintest.ChainSpec{
			ChainConfig: ibc.ChainConfig{
				Images: []ibc.DockerImage{
					{
						Repository: chainsuite.HyphaICSRepo,
						Version:    selectConsumerVersion("v6.0.0", "v6.2.1"),
						UIDGID:     chainsuite.ICSUidGuid,
					},
				},
			},
		},
	}
	consumer, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, cfg)
	s.Require().NoError(err)
	s.Consumer = consumer
	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), s.Consumer, s.Relayer, 1_000_000, 0, 1))
	s.UpgradeChain()
}

func (s *UnbondingSuite) TestChainNotRemoved() {
	s.Require().NoError(s.Relayer.PauseRelayer(s.GetContext()))
	defer s.Relayer.ResumeRelayer(s.GetContext())

	s.Require().NoError(s.Chain.GetNode().StakingDelegate(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, s.Chain.ValidatorWallets[0].ValoperAddress, "1000000"+s.Chain.Config().Denom))
	time.Sleep(vscTimeoutPeriod)
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain))

	// check that the chain is still around
	chain, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("chains.#(chain_id=%q)", s.Consumer.Config().ChainID), "provider", "list-consumer-chains")
	s.Require().NoError(err)
	s.Require().True(chain.Exists())
}

func (s *UnbondingSuite) TestNoDelayForUnbonding() {
	s.Require().NoError(s.Relayer.PauseRelayer(s.GetContext()))
	defer s.Relayer.ResumeRelayer(s.GetContext())

	amount := int64(1_000_000)
	s.Require().NoError(s.Chain.GetNode().StakingUnbond(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, s.Chain.ValidatorWallets[0].ValoperAddress, fmt.Sprintf("%d%s", amount, s.Chain.Config().Denom)))

	unbonding, err := s.Chain.StakingQueryUnbondingDelegation(s.GetContext(), s.Chain.ValidatorWallets[0].Address, s.Chain.ValidatorWallets[0].ValoperAddress)
	s.Require().NoError(err)

	s.Require().Equal(1, len(unbonding.Entries))
	s.Require().Equal(int64(0), unbonding.Entries[0].UnbondingOnHoldRefCount)

	time.Sleep(unbondingTime)
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain))

	_, err = s.Chain.StakingQueryUnbondingDelegation(s.GetContext(), s.Chain.ValidatorWallets[0].Address, s.Chain.ValidatorWallets[0].ValoperAddress)
	s.Require().Error(err)
	s.Require().Equal(status.Code(err), codes.NotFound)
}

func (s *UnbondingSuite) TestCanLaunchAfterInitTimeout() {
	cfg := chainsuite.ConsumerConfig{
		ChainName:             "ics-consumer",
		Version:               "v5.0.0",
		ShouldCopyProviderKey: allProviderKeysCopied(),
		Denom:                 chainsuite.Ucon,
		TopN:                  0,
		BeforeSpawnTime: func(_ context.Context, _ *cosmos.CosmosChain) {
			time.Sleep(initTimeoutPeriod)
			s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain))
		},
	}
	chainID := cfg.ChainName + "-timeout"
	spawnTime := time.Now().Add(2 * time.Minute)
	err := s.Chain.CreateConsumerPermissionless(s.GetContext(), chainID, cfg, spawnTime)
	s.Require().NoError(err)

	time.Sleep(time.Until(spawnTime))
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain))
	time.Sleep(initTimeoutPeriod)
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain))

	chain, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("chains.#(chain_id=%q)", chainID), "provider", "list-consumer-chains")
	s.Require().NoError(err)
	s.Require().True(chain.Exists())
}

func TestUnbonding(t *testing.T) {
	genesis := chainsuite.DefaultGenesis()
	env := chainsuite.GetEnvironment()
	if semver.Compare(env.OldGaiaImageVersion, "v20.0.0") < 0 {
		genesis = append(genesis,
			cosmos.NewGenesisKV("app_state.provider.params.vsc_timeout_period", vscTimeoutPeriod.String()),
			cosmos.NewGenesisKV("app_state.provider.params.init_timeout_period", initTimeoutPeriod.String()),
		)
	}
	genesis = append(genesis,
		cosmos.NewGenesisKV("app_state.staking.params.unbonding_time", unbondingTime.String()),
	)
	s := &UnbondingSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			CreateRelayer: true,
			ChainSpec: &interchaintest.ChainSpec{
				ChainConfig: ibc.ChainConfig{
					ModifyGenesis: cosmos.ModifyGenesis(genesis),
				},
			},
		}),
	}
	suite.Run(t, s)
}
