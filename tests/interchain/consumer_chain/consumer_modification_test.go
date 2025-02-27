package consumer_chain_test

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	providertypes "github.com/cosmos/interchain-security/v7/x/ccv/provider/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
)

const (
	permissionlessDepositPeriod = 7 * time.Minute
)

type ConsumerModificationSuite struct {
	*chainsuite.Suite
	consumerCfg chainsuite.ConsumerConfig
}

func (s *ConsumerModificationSuite) TestChangeOwner() {
	cfg := s.consumerCfg
	cfg.TopN = 0
	cfg.BeforeSpawnTime = func(ctx context.Context, consumer *cosmos.CosmosChain) {
		consumerID, err := s.Chain.GetConsumerID(s.GetContext(), consumer.Config().ChainID)
		s.Require().NoError(err)
		eg := errgroup.Group{}
		for i := 0; i < 3; i++ {
			i := i
			eg.Go(func() error {
				_, err := s.Chain.Validators[i].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[i].Moniker, "provider", "opt-in", consumerID)
				return err
			})
		}
		s.Require().NoError(eg.Wait())
	}
	consumer, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, cfg)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), consumer, s.Relayer, 1_000_000, 0, 1))

	govAddress, err := s.Chain.GetGovernanceAddress(s.GetContext())
	s.Require().NoError(err)
	consumerID, err := s.Chain.GetConsumerID(s.GetContext(), consumer.Config().ChainID)
	s.Require().NoError(err)
	update := &providertypes.MsgUpdateConsumer{
		ConsumerId:      consumerID,
		NewOwnerAddress: govAddress,
		Metadata: &providertypes.ConsumerMetadata{
			Name:        consumer.Config().Name,
			Description: "Consumer chain",
			Metadata:    "ipfs://",
		},
	}
	updateBz, err := json.Marshal(update)
	s.Require().NoError(err)
	err = s.Chain.GetNode().WriteFile(s.GetContext(), updateBz, "consumer-update.json")
	s.Require().NoError(err)
	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), interchaintest.FaucetAccountKeyName,
		"provider", "update-consumer", path.Join(s.Chain.GetNode().HomeDir(), "consumer-update.json"))
	s.Require().NoError(err)

	update.Owner = govAddress
	update.NewOwnerAddress = s.Chain.ValidatorWallets[0].Address
	prop, err := s.Chain.BuildProposal([]cosmos.ProtoMessage{update},
		"update consumer", "update consumer", "",
		chainsuite.GovDepositAmount, "", false)
	s.Require().NoError(err)
	txhash, err := s.Chain.GetNode().SubmitProposal(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, prop)
	s.Require().NoError(err)
	propID, err := s.Chain.GetProposalID(s.GetContext(), txhash)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.PassProposal(s.GetContext(), propID))
}

func (s *ConsumerModificationSuite) TestChangePowerShaping() {
	cfg := s.consumerCfg
	cfg.TopN = 0
	const (
		oldValidatorCount = 4
		newValidatorCount = 3
	)
	cfg.BeforeSpawnTime = func(ctx context.Context, consumer *cosmos.CosmosChain) {
		consumerID, err := s.Chain.GetConsumerID(s.GetContext(), consumer.Config().ChainID)
		s.Require().NoError(err)
		eg := errgroup.Group{}
		for i := 0; i < oldValidatorCount; i++ {
			i := i
			eg.Go(func() error {
				_, err := s.Chain.Validators[i].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[i].Moniker, "provider", "opt-in", consumerID)
				return err
			})
		}
		s.Require().NoError(eg.Wait())
	}
	consumer, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, cfg)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), consumer, s.Relayer, 1_000_000, 0, 1))

	consumerID, err := s.Chain.GetConsumerID(s.GetContext(), consumer.Config().ChainID)
	s.Require().NoError(err)
	update := &providertypes.MsgUpdateConsumer{
		ConsumerId: consumerID,
		Metadata: &providertypes.ConsumerMetadata{
			Name:        consumer.Config().Name,
			Description: "Consumer chain",
			Metadata:    "ipfs://",
		},
		PowerShapingParameters: &providertypes.PowerShapingParameters{
			ValidatorSetCap: newValidatorCount,
		},
	}
	updateBz, err := json.Marshal(update)
	s.Require().NoError(err)
	err = s.Chain.GetNode().WriteFile(s.GetContext(), updateBz, "consumer-update.json")
	s.Require().NoError(err)
	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), interchaintest.FaucetAccountKeyName,
		"provider", "update-consumer", path.Join(s.Chain.GetNode().HomeDir(), "consumer-update.json"))
	s.Require().NoError(err)

	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), consumer, s.Relayer, 1_000_000, 0, 1))

	vals, err := consumer.QueryJSON(s.GetContext(), "validators", "tendermint-validator-set")
	s.Require().NoError(err)
	s.Require().Equal(newValidatorCount, len(vals.Array()), vals)
	for i := 0; i < newValidatorCount; i++ {
		valCons := vals.Array()[i].Get("address").String()
		s.Require().NoError(err)
		s.Require().Equal(consumer.ValidatorWallets[i].ValConsAddress, valCons)
	}
}

func (s *ConsumerModificationSuite) TestConsumerCommissionRate() {
	cfg := s.consumerCfg
	cfg.TopN = 0
	cfg.BeforeSpawnTime = func(ctx context.Context, consumer *cosmos.CosmosChain) {
		consumerID, err := s.Chain.GetConsumerID(s.GetContext(), consumer.Config().ChainID)
		s.Require().NoError(err)
		_, err = s.Chain.Validators[0].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, "provider", "opt-in", consumerID)
		s.Require().NoError(err)
	}

	images := []ibc.DockerImage{
		{
			Repository: "ghcr.io/hyphacoop/ics",
			Version:    "v4.5.0",
			UIDGID:     "1025:1025",
		},
	}
	chainID := fmt.Sprintf("%s-test-%d", cfg.ChainName, len(s.Chain.Consumers)+1)
	spawnTime := time.Now().Add(chainsuite.ChainSpawnWait)
	cfg.Spec = s.Chain.DefaultConsumerChainSpec(s.GetContext(), chainID, cfg, spawnTime, nil)
	cfg.Spec.Version = "v4.5.0"
	cfg.Spec.Images = images
	consumer1, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, cfg)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), consumer1, s.Relayer, 1_000_000, 0, 1))

	chainID = fmt.Sprintf("%s-test-%d", cfg.ChainName, len(s.Chain.Consumers)+1)
	spawnTime = time.Now().Add(chainsuite.ChainSpawnWait)
	cfg.Spec = s.Chain.DefaultConsumerChainSpec(s.GetContext(), chainID, cfg, spawnTime, nil)
	cfg.Spec.Version = "v4.5.0"
	cfg.Spec.Images = images
	consumer2, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, cfg)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), consumer2, s.Relayer, 1_000_000, 0, 1))

	for i := 1; i < len(consumer1.Validators); i++ {
		s.Require().NoError(consumer1.Validators[i].StopContainer(s.GetContext()))
		s.Require().NoError(consumer2.Validators[i].StopContainer(s.GetContext()))
	}

	consumer1Ch, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, consumer1)
	s.Require().NoError(err)
	consumer2Ch, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, consumer2)
	s.Require().NoError(err)
	denom1 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", consumer1Ch.ChannelID, consumer1.Config().Denom)).IBCDenom()
	denom2 := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", consumer2Ch.ChannelID, consumer2.Config().Denom)).IBCDenom()

	s.Require().NotEqual(denom1, denom2, "denom1: %s, denom2: %s; channel1: %s, channel2: %s", denom1, denom2, consumer1Ch.Counterparty.ChannelID, consumer2Ch.Counterparty.ChannelID)

	govAuthority, err := s.Chain.GetGovernanceAddress(s.GetContext())
	s.Require().NoError(err)
	rewardDenomsProp := providertypes.MsgChangeRewardDenoms{
		DenomsToAdd: []string{denom1, denom2},
		Authority:   govAuthority,
	}
	prop, err := s.Chain.BuildProposal([]cosmos.ProtoMessage{&rewardDenomsProp},
		"add denoms to list of registered reward denoms",
		"add denoms to list of registered reward denoms",
		"", chainsuite.GovDepositAmount, "", false)
	s.Require().NoError(err)
	propResult, err := s.Chain.SubmitProposal(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, prop)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.PassProposal(s.GetContext(), propResult.ProposalID))

	eg := errgroup.Group{}

	_, err = s.Chain.Validators[0].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, "distribution", "withdraw-all-rewards")
	s.Require().NoError(err)

	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, "distribution", "withdraw-rewards", s.Chain.ValidatorWallets[0].ValoperAddress, "--commission")
	s.Require().NoError(err)

	consumerID1, err := s.Chain.GetConsumerID(s.GetContext(), consumer1.Config().ChainID)
	s.Require().NoError(err)
	consumerID2, err := s.Chain.GetConsumerID(s.GetContext(), consumer2.Config().ChainID)
	s.Require().NoError(err)

	eg.Go(func() error {
		_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, "provider", "set-consumer-commission-rate", consumerID1, "0.5")
		return err
	})
	eg.Go(func() error {
		_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, "provider", "set-consumer-commission-rate", consumerID2, "0.5")
		return err
	})
	s.Require().NoError(eg.Wait())

	_, err = s.Chain.Validators[0].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, "distribution", "withdraw-rewards", s.Chain.ValidatorWallets[0].ValoperAddress, "--commission")
	s.Require().NoError(err)

	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 1, consumer1, consumer2))

	eg.Go(func() error {
		_, err := consumer1.Validators[0].ExecTx(s.GetContext(), consumer1.ValidatorWallets[0].Moniker, "bank", "send", consumer1.ValidatorWallets[0].Address, consumer1.ValidatorWallets[1].Address, "1"+consumer1.Config().Denom, "--fees", "100000000"+consumer1.Config().Denom)
		return err
	})
	eg.Go(func() error {
		_, err := consumer2.Validators[0].ExecTx(s.GetContext(), consumer2.ValidatorWallets[0].Moniker, "bank", "send", consumer2.ValidatorWallets[0].Address, consumer2.ValidatorWallets[1].Address, "1"+consumer2.Config().Denom, "--fees", "100000000"+consumer2.Config().Denom)
		return err
	})
	s.Require().NoError(eg.Wait())

	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), chainsuite.BlocksPerDistribution+2, s.Chain, consumer1, consumer2))
	s.Require().NoError(s.Relayer.ClearTransferChannel(s.GetContext(), s.Chain, consumer1))
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain, consumer1, consumer2))

	rewardStr, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("total.#(%%\"*%s\")", denom1), "distribution", "rewards", s.Chain.ValidatorWallets[0].Address)
	s.Require().NoError(err)
	rewardsDenom1, err := chainsuite.StrToSDKInt(rewardStr.String())
	s.Require().NoError(err)
	rewardStr, err = s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("total.#(%%\"*%s\")", denom2), "distribution", "rewards", s.Chain.ValidatorWallets[0].Address)
	s.Require().NoError(err)
	rewardsDenom2, err := chainsuite.StrToSDKInt(rewardStr.String())
	s.Require().NoError(err)

	s.Require().NotEmpty(rewardsDenom1)
	s.Require().NotEmpty(rewardsDenom2)
	s.Require().True(rewardsDenom1.Sub(rewardsDenom2).Abs().LT(sdkmath.NewInt(1000)), "rewards1Int: %s, rewards2Int: %s", rewardsDenom1.String(), rewardsDenom2.String())

	_, err = s.Chain.Validators[0].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, "distribution", "withdraw-rewards", s.Chain.ValidatorWallets[0].ValoperAddress, "--commission")
	s.Require().NoError(err)

	eg.Go(func() error {
		_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, "provider", "set-consumer-commission-rate", consumerID1, "0.25")
		return err
	})
	eg.Go(func() error {
		_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, "provider", "set-consumer-commission-rate", consumerID2, "0.5")
		return err
	})
	s.Require().NoError(eg.Wait())

	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, "distribution", "withdraw-rewards", s.Chain.ValidatorWallets[0].ValoperAddress, "--commission")
	s.Require().NoError(err)

	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 1, consumer1, consumer2))

	eg.Go(func() error {
		_, err := consumer1.Validators[0].ExecTx(s.GetContext(), consumer1.ValidatorWallets[0].Moniker, "bank", "send", consumer1.ValidatorWallets[0].Address, consumer1.ValidatorWallets[1].Address, "1"+consumer1.Config().Denom, "--fees", "100000000"+consumer1.Config().Denom)
		return err
	})
	eg.Go(func() error {
		_, err := consumer2.Validators[0].ExecTx(s.GetContext(), consumer2.ValidatorWallets[0].Moniker, "bank", "send", consumer2.ValidatorWallets[0].Address, consumer2.ValidatorWallets[1].Address, "1"+consumer2.Config().Denom, "--fees", "100000000"+consumer2.Config().Denom)
		return err
	})
	s.Require().NoError(eg.Wait())

	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), chainsuite.BlocksPerDistribution+2, s.Chain, consumer1, consumer2))
	s.Require().NoError(s.Relayer.ClearTransferChannel(s.GetContext(), s.Chain, consumer1))
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain, consumer1, consumer2))

	rewardStr, err = s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("total.#(%%\"*%s\")", denom1), "distribution", "rewards", s.Chain.ValidatorWallets[0].Address)
	s.Require().NoError(err)
	rewardsDenom1, err = chainsuite.StrToSDKInt(rewardStr.String())
	s.Require().NoError(err)
	rewardStr, err = s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("total.#(%%\"*%s\")", denom2), "distribution", "rewards", s.Chain.ValidatorWallets[0].Address)
	s.Require().NoError(err)
	rewardsDenom2, err = chainsuite.StrToSDKInt(rewardStr.String())
	s.Require().NoError(err)

	s.Require().True(rewardsDenom1.GT(rewardsDenom2), "rewards1Int: %s, rewards2Int: %s", rewardsDenom1.String(), rewardsDenom2.String())
	s.Require().False(rewardsDenom1.Sub(rewardsDenom2).Abs().LT(sdkmath.NewInt(1000)), "rewards1Int: %s, rewards2Int: %s", rewardsDenom1.String(), rewardsDenom2.String())
}

func (s *ConsumerModificationSuite) TestLaunchWithAllowListThenModify() {
	consumerConfig := s.consumerCfg
	consumerConfig.Allowlist = []string{
		s.Chain.ValidatorWallets[0].ValConsAddress,
		s.Chain.ValidatorWallets[1].ValConsAddress,
		s.Chain.ValidatorWallets[2].ValConsAddress,
	}
	consumerConfig.TopN = 0
	consumerConfig.BeforeSpawnTime = func(ctx context.Context, consumer *cosmos.CosmosChain) {
		consumerID, err := s.Chain.GetConsumerID(s.GetContext(), consumer.Config().ChainID)
		s.Require().NoError(err)
		eg := errgroup.Group{}
		for i := 0; i < 3; i++ {
			i := i
			eg.Go(func() error {
				_, err := s.Chain.Validators[i].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[i].Moniker, "provider", "opt-in", consumerID)
				return err
			})
		}
		s.Require().NoError(eg.Wait())
	}

	consumer, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, consumerConfig)
	s.Require().NoError(err)

	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), consumer, s.Relayer, 1_000_000, 0, 1))

	consumerID, err := s.Chain.GetConsumerID(s.GetContext(), consumer.Config().ChainID)
	s.Require().NoError(err)

	// ensure we can't opt in a non-allowlisted validator
	_, err = s.Chain.Validators[3].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[3].Moniker,
		"provider", "opt-in", consumerID)
	s.Require().NoError(err)

	validators, err := consumer.QueryJSON(s.GetContext(), "validators", "tendermint-validator-set")
	s.Require().NoError(err)
	s.Require().Equal(3, len(validators.Array()))

	update := &providertypes.MsgUpdateConsumer{
		ConsumerId: consumerID,
		PowerShapingParameters: &providertypes.PowerShapingParameters{
			Allowlist: []string{},
		},
	}
	updateBz, err := json.Marshal(update)
	s.Require().NoError(err)
	err = s.Chain.GetNode().WriteFile(s.GetContext(), updateBz, "consumer-update.json")
	s.Require().NoError(err)
	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), interchaintest.FaucetAccountKeyName,
		"provider", "update-consumer", path.Join(s.Chain.GetNode().HomeDir(), "consumer-update.json"))
	s.Require().NoError(err)

	// // ensure we can opt in a non-allowlisted validator after the modification
	_, err = s.Chain.Validators[3].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[3].Moniker,
		"provider", "opt-in", consumerID)
	s.Require().NoError(err)
	validators, err = consumer.QueryJSON(s.GetContext(), "validators", "tendermint-validator-set")
	s.Require().NoError(err)
	s.Require().Equal(4, len(validators.Array()))
}

func TestConsumerModification(t *testing.T) {
	genesis := chainsuite.DefaultGenesis()
	genesis = append(genesis,
		cosmos.NewGenesisKV("app_state.gov.params.max_deposit_period", permissionlessDepositPeriod.String()),
	)
	s := &ConsumerModificationSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			CreateRelayer:  true,
			Scope:          chainsuite.ChainScopeTest,
			UpgradeOnSetup: true,
			ChainSpec: &interchaintest.ChainSpec{
				NumValidators: &chainsuite.SixValidators,
				ChainConfig: ibc.ChainConfig{
					ModifyGenesis: cosmos.ModifyGenesis(genesis),
				},
			},
		}),
		consumerCfg: chainsuite.ConsumerConfig{
			ChainName:             "ics-consumer",
			Version:               "v4.5.0",
			ShouldCopyProviderKey: allProviderKeysCopied(),
			Denom:                 chainsuite.Ucon,
			TopN:                  100,
			AllowInactiveVals:     true,
			MinStake:              1_000_000,
			Spec: &interchaintest.ChainSpec{
				NumValidators: &chainsuite.SixValidators,
				ChainConfig: ibc.ChainConfig{
					Images: []ibc.DockerImage{
						{
							Repository: chainsuite.HyphaICSRepo,
							Version:    "v4.5.0",
							UIDGID:     chainsuite.ICSUidGuid,
						},
					},
				},
			},
		},
	}
	suite.Run(t, s)
}
