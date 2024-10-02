package interchain_test

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	govtypes "github.cogaia/v21/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/gaia/v21/tests/interchain/chainsuite"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ccvclient "github.com/cosmos/interchain-security/v5/x/ccv/provider/client"
	providertypes "github.com/cosmos/interchain-security/v5/x/ccv/provider/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/sjson"
	"golang.org/x/mod/semver"
	"golang.org/x/sync/errgroup"
)

const (
	permissionlessDepositPeriod = 7 * time.Minute
)

type PermissionlessConsumersSuite struct {
	*chainsuite.Suite
	consumerCfg chainsuite.ConsumerConfig
}

func (s *PermissionlessConsumersSuite) addConsumer() *chainsuite.Chain {
	consumer, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, s.consumerCfg)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), consumer, s.Relayer, 1_000_000, 0, 1))
	return consumer
}

func (s *PermissionlessConsumersSuite) isOverV19() bool {
	return semver.Compare(s.Env.OldGaiaImageVersion, "v19.0.0") > 0
}

func (s *PermissionlessConsumersSuite) TestConsumerAdditionMigration() {
	if s.isOverV19() {
		s.T().Skip("Migration test for v19 -> v20")
	}
	consumer := s.addConsumer()
	json, _, err := s.Chain.GetNode().ExecQuery(s.GetContext(), "gov", "proposals")
	s.Require().NoError(err)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("%s", string(json))

	proposals, err := s.Chain.GovQueryProposalsV1(s.GetContext(), govtypes.ProposalStatus_PROPOSAL_STATUS_PASSED)
	s.Require().NoError(err)
	s.Require().Len(proposals, 1)
	oldProposalCh1 := proposals[0]

	chainIDCh2 := s.consumerCfg.ChainName + "-2"
	propWaiter, errCh, err := s.Chain.SubmitConsumerAdditionProposal(s.GetContext(), chainIDCh2, s.consumerCfg, time.Now().Add(permissionlessDepositPeriod+2*time.Minute))
	s.Require().NoError(err)

	s.UpgradeChain()

	proposals, err = s.Chain.GovQueryProposalsV1(s.GetContext(), govtypes.ProposalStatus_PROPOSAL_STATUS_PASSED)
	s.Require().NoError(err)
	s.Require().Len(proposals, 2)
	newProposalCh1 := proposals[0]
	s.Require().Equal(oldProposalCh1, newProposalCh1)

	proposals, err = s.Chain.GovQueryProposalsV1(s.GetContext(), govtypes.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD)
	s.Require().NoError(err)
	s.Require().Len(proposals, 1)
	newProposalCh2 := proposals[0]
	s.Require().Contains(newProposalCh2.Messages[0].TypeUrl, "MsgUpdateConsumer")

	// check that the new chain is around
	chain, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("chains.#(chain_id=%q)", consumer.Config().ChainID), "provider", "list-consumer-chains")
	s.Require().NoError(err)
	s.Require().True(chain.Exists())

	chain2, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("chains.#(chain_id=%q)", chainIDCh2), "provider", "list-consumer-chains")
	s.Require().NoError(err)
	s.Require().True(chain2.Exists())
	s.Require().Equal(uint64(0), chain2.Get("top_N").Uint())

	propWaiter.AllowDeposit()
	propWaiter.WaitForVotingPeriod()
	propWaiter.AllowVote()
	propWaiter.WaitForPassed()
	s.Require().NoError(<-errCh)

	testutil.WaitForBlocks(s.GetContext(), 2, s.Chain)

	chain2, err = s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("chains.#(chain_id=%q)", chainIDCh2), "provider", "list-consumer-chains")
	s.Require().NoError(err)
	s.Require().True(chain2.Exists())
	s.Require().Equal(uint64(100), chain2.Get("top_N").Uint())
}

func (s *PermissionlessConsumersSuite) TestConsumerRemovalMigration() {
	if s.isOverV19() {
		s.T().Skip("Migration test for v19 -> v20")
	}

	consumer := s.addConsumer()

	stopTime := time.Now().Add(permissionlessDepositPeriod + 2*time.Minute)

	propID := s.submitConsumerRemoval(consumer, stopTime)

	s.UpgradeChain()

	proposals, err := s.Chain.GovQueryProposalsV1(s.GetContext(), govtypes.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD)
	s.Require().NoError(err)
	s.Require().Len(proposals, 1)
	newProposalCh2 := proposals[0]
	s.Require().Contains(newProposalCh2.Messages[0].TypeUrl, "MsgRemoveConsumer")

	chain, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("chains.#(chain_id=%q)", consumer.Config().ChainID), "provider", "list-consumer-chains")
	s.Require().NoError(err)
	s.Require().True(chain.Exists())

	s.depositAndPass(propID)

	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("waiting for stop time %s", stopTime)
	time.Sleep(time.Until(stopTime))
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain))

	chain, err = s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("chains.#(chain_id=%q)", consumer.Config().ChainID), "provider", "list-consumer-chains")
	s.Require().NoError(err)
	s.Require().True(chain.Exists())
	s.Require().Equal("CONSUMER_PHASE_STOPPED", chain.Get("phase").String())
}

func (s *PermissionlessConsumersSuite) TestConsumerModificationMigration() {
	if s.isOverV19() {
		s.T().Skip("Migration test for v19 -> v20")
	}

	consumer := s.addConsumer()

	propID := s.submitConsumerModification(consumer)

	s.UpgradeChain()

	proposals, err := s.Chain.GovQueryProposalsV1(s.GetContext(), govtypes.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD)
	s.Require().NoError(err)
	s.Require().Len(proposals, 1)
	newProposalCh2 := proposals[0]
	s.Require().Contains(newProposalCh2.Messages[0].TypeUrl, "MsgUpdateConsumer")

	s.depositAndPass(propID)

	chain, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("chains.#(chain_id=%q)", consumer.Config().ChainID), "provider", "list-consumer-chains")
	s.Require().NoError(err)
	s.Require().True(chain.Exists())
	s.Require().Equal(uint64(80), chain.Get("top_N").Uint())
}

func (s *PermissionlessConsumersSuite) TestChangeRewardDenomMigration() {
	if s.isOverV19() {
		s.T().Skip("Migration test for v19 -> v20")
	}

	consumer := s.addConsumer()

	denom, propID := s.submitChangeRewardDenoms(consumer)

	s.UpgradeChain()

	proposals, err := s.Chain.GovQueryProposalsV1(s.GetContext(), govtypes.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD)
	s.Require().NoError(err)
	s.Require().Len(proposals, 1)
	newProposalCh2 := proposals[0]
	s.Require().Contains(newProposalCh2.Messages[0].TypeUrl, "MsgChangeRewardDenoms")

	s.depositAndPass(propID)

	denoms, err := s.Chain.QueryJSON(s.GetContext(), "denoms", "provider", "registered-consumer-reward-denoms")
	s.Require().NoError(err)
	s.Require().Contains(denoms.String(), denom)
}

func (s *PermissionlessConsumersSuite) depositAndPass(propID string) {
	_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, "gov", "deposit", propID, chainsuite.GovDepositAmount)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.PassProposal(s.GetContext(), propID))
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain))
}

func (s *PermissionlessConsumersSuite) TestPassedProposalsDontChange() {
	if s.isOverV19() {
		s.T().Skip("Migration test for v19 -> v20")
	}
	consumer := s.addConsumer()

	_, denomPropID := s.submitChangeRewardDenoms(consumer)
	s.depositAndPass(denomPropID)

	denomPropIDInt, err := strconv.Atoi(denomPropID)
	s.Require().NoError(err)
	denomProposal, err := s.Chain.GovQueryProposalV1(s.GetContext(), uint64(denomPropIDInt))
	s.Require().NoError(err)

	modificationPropID := s.submitConsumerModification(consumer)
	s.depositAndPass(modificationPropID)

	modificationPropIDInt, err := strconv.Atoi(modificationPropID)
	s.Require().NoError(err)
	modificationProposal, err := s.Chain.GovQueryProposalV1(s.GetContext(), uint64(modificationPropIDInt))
	s.Require().NoError(err)

	stopTime := time.Now().Add(permissionlessDepositPeriod + 2*time.Minute)
	removalPropID := s.submitConsumerRemoval(consumer, stopTime)
	s.depositAndPass(removalPropID)

	removalPropIDInt, err := strconv.Atoi(removalPropID)
	s.Require().NoError(err)
	removalProposal, err := s.Chain.GovQueryProposalV1(s.GetContext(), uint64(removalPropIDInt))
	s.Require().NoError(err)

	s.UpgradeChain()

	denomProposalAfter, err := s.Chain.GovQueryProposalV1(s.GetContext(), uint64(denomPropIDInt))
	s.Require().NoError(err)
	s.Require().Equal(denomProposal, denomProposalAfter)

	modificationProposalAfter, err := s.Chain.GovQueryProposalV1(s.GetContext(), uint64(modificationPropIDInt))
	s.Require().NoError(err)
	s.Require().Equal(modificationProposal, modificationProposalAfter)

	removalProposalAfter, err := s.Chain.GovQueryProposalV1(s.GetContext(), uint64(removalPropIDInt))
	s.Require().NoError(err)
	s.Require().Equal(removalProposal, removalProposalAfter)

	time.Sleep(time.Until(stopTime))
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain))

	chain, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("chains.#(chain_id=%q)", consumer.Config().ChainID), "provider", "list-consumer-chains")
	s.Require().NoError(err)
	s.Require().True(chain.Exists())
	s.Require().Equal("CONSUMER_PHASE_STOPPED", chain.Get("phase").String())
}

func (s *PermissionlessConsumersSuite) TestChangeOwner() {
	s.UpgradeChain()

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

func (s *PermissionlessConsumersSuite) TestChangePowerShaping() {
	s.UpgradeChain()

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

	vals, err := consumer.QueryJSON(s.GetContext(), "validators", "comet-validator-set")
	s.Require().NoError(err)
	s.Require().Equal(newValidatorCount, len(vals.Array()), vals)
	for i := 0; i < newValidatorCount; i++ {
		valCons := vals.Array()[i].Get("address").String()
		s.Require().NoError(err)
		s.Require().Equal(consumer.ValidatorWallets[i].ValConsAddress, valCons)
	}

}
func (s *PermissionlessConsumersSuite) TestConsumerCommissionRate() {
	s.UpgradeChain()
	cfg := s.consumerCfg

	cfg.TopN = 0
	cfg.BeforeSpawnTime = func(ctx context.Context, consumer *cosmos.CosmosChain) {
		consumerID, err := s.Chain.GetConsumerID(s.GetContext(), consumer.Config().ChainID)
		s.Require().NoError(err)
		_, err = s.Chain.Validators[0].ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, "provider", "opt-in", consumerID)
		s.Require().NoError(err)
	}
	consumer1, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, cfg)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), consumer1, s.Relayer, 1_000_000, 0, 1))

	consumer2, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, cfg)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.CheckCCV(s.GetContext(), consumer2, s.Relayer, 1_000_000, 0, 1))

	for i := 1; i < chainsuite.ValidatorCount; i++ {
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

	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), chainsuite.BlocksPerDistribution+3, s.Chain, consumer1, consumer2))

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

	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), chainsuite.BlocksPerDistribution+3, s.Chain, consumer1, consumer2))

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

func (s *PermissionlessConsumersSuite) TestLaunchWithAllowListThenModify() {
	s.UpgradeChain()

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

func TestPermissionlessConsumers(t *testing.T) {
	genesis := chainsuite.DefaultGenesis()
	genesis = append(genesis,
		cosmos.NewGenesisKV("app_state.gov.params.max_deposit_period", permissionlessDepositPeriod.String()),
	)
	s := &PermissionlessConsumersSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			CreateRelayer: true,
			Scope:         chainsuite.ChainScopeTest,
			ChainSpec: &interchaintest.ChainSpec{
				ChainConfig: ibc.ChainConfig{
					ModifyGenesis: cosmos.ModifyGenesis(genesis),
				},
			},
		}),
		consumerCfg: chainsuite.ConsumerConfig{
			ChainName:             "ics-consumer",
			Version:               "v5.0.0",
			ShouldCopyProviderKey: allProviderKeysCopied(),
			Denom:                 chainsuite.Ucon,
			TopN:                  100,
			AllowInactiveVals:     true,
			MinStake:              1_000_000,
		},
	}
	suite.Run(t, s)
}

func (s *PermissionlessConsumersSuite) submitChangeRewardDenoms(consumer *chainsuite.Chain) (string, string) {
	consumerCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, consumer)
	s.Require().NoError(err)
	denom := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", consumerCh.ChannelID, consumer.Config().Denom)).IBCDenom()

	denomProp := &ccvclient.ChangeRewardDenomsProposalJSON{
		ChangeRewardDenomsProposal: providertypes.ChangeRewardDenomsProposal{
			Title:          "change reward denoms",
			Description:    "change reward denoms",
			DenomsToAdd:    []string{denom},
			DenomsToRemove: []string{},
		},
		Deposit: fmt.Sprintf("%d%s", chainsuite.GovMinDepositAmount/2, s.Chain.Config().Denom),
		Summary: "change reward denoms",
	}
	propBz, err := json.Marshal(denomProp)
	s.Require().NoError(err)

	fileName := "proposal_consumer_denoms.json"

	s.Require().NoError(s.Chain.GetNode().WriteFile(s.GetContext(), propBz, fileName))

	filePath := filepath.Join(s.Chain.GetNode().HomeDir(), fileName)

	txhash, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker,
		"gov", "submit-legacy-proposal", "change-reward-denoms", filePath,
		"--gas", "auto",
	)
	s.Require().NoError(err)

	propID, err := s.Chain.GetProposalID(s.GetContext(), txhash)
	s.Require().NoError(err)
	return denom, propID
}

func (s *PermissionlessConsumersSuite) submitConsumerModification(consumer *chainsuite.Chain) string {
	modifyProp := &ccvclient.ConsumerModificationProposalJSON{
		Title:   "modify consumer",
		Summary: "modify consumer",
		ChainId: consumer.Config().ChainID,
		TopN:    80,
		Deposit: fmt.Sprintf("%d%s", chainsuite.GovMinDepositAmount/2, s.Chain.Config().Denom),
	}

	propBz, err := json.Marshal(modifyProp)
	s.Require().NoError(err)

	propBz, err = sjson.DeleteBytes(propBz, "allow_inactive_vals")
	s.Require().NoError(err)
	propBz, err = sjson.DeleteBytes(propBz, "min_stake")
	s.Require().NoError(err)

	fileName := "proposal_consumer_modification.json"

	s.Require().NoError(s.Chain.GetNode().WriteFile(s.GetContext(), propBz, fileName))

	filePath := filepath.Join(s.Chain.GetNode().HomeDir(), fileName)

	txhash, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker,
		"gov", "submit-legacy-proposal", "consumer-modification", filePath,
		"--gas", "auto",
	)
	s.Require().NoError(err)

	propID, err := s.Chain.GetProposalID(s.GetContext(), txhash)
	s.Require().NoError(err)
	return propID
}

func (s *PermissionlessConsumersSuite) submitConsumerRemoval(consumer *chainsuite.Chain, stopTime time.Time) string {
	removalProp := &ccvclient.ConsumerRemovalProposalJSON{
		Title:    "remove consumer",
		Summary:  "remove consumer",
		ChainId:  consumer.Config().ChainID,
		StopTime: stopTime,
		Deposit:  fmt.Sprintf("%d%s", chainsuite.GovMinDepositAmount/2, s.Chain.Config().Denom),
	}

	propBz, err := json.Marshal(removalProp)
	s.Require().NoError(err)

	fileName := "proposal_consumer_removal.json"

	s.Require().NoError(s.Chain.GetNode().WriteFile(s.GetContext(), propBz, fileName))

	filePath := filepath.Join(s.Chain.GetNode().HomeDir(), fileName)

	txhash, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker,
		"gov", "submit-legacy-proposal", "consumer-removal", filePath,
		"--gas", "auto",
	)
	s.Require().NoError(err)

	propID, err := s.Chain.GetProposalID(s.GetContext(), txhash)
	s.Require().NoError(err)
	return propID
}
