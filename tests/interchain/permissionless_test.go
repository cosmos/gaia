package interchain_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/gaia/v20/tests/interchain/chainsuite"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ccvclient "github.com/cosmos/interchain-security/v5/x/ccv/provider/client"
	providertypes "github.com/cosmos/interchain-security/v5/x/ccv/provider/types"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/sjson"
)

const (
	permissionlessDepositPeriod = 7 * time.Minute
)

type ConsumerPropMigrationSuite struct {
	*chainsuite.Suite
	consumerCfg chainsuite.ConsumerConfig
}

func (s *ConsumerPropMigrationSuite) addConsumer() *chainsuite.Chain {
	consumer, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, s.consumerCfg)
	s.Require().NoError(err)
	err = s.Chain.CheckCCV(s.GetContext(), consumer, s.Relayer, 1_000_000, 0, 1)
	s.Require().NoError(err)
	return consumer
}

func (s *ConsumerPropMigrationSuite) TestConsumerAddition() {
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

func (s *ConsumerPropMigrationSuite) TestConsumerRemoval() {
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

func (s *ConsumerPropMigrationSuite) TestConsumerModification() {
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

func (s *ConsumerPropMigrationSuite) TestChangeRewardDenom() {
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

func (s *ConsumerPropMigrationSuite) depositAndPass(propID string) {
	_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Moniker, "gov", "deposit", propID, chainsuite.GovDepositAmount)
	s.Require().NoError(err)
	s.Require().NoError(s.Chain.PassProposal(s.GetContext(), propID))
	s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 2, s.Chain))
}

func (s *ConsumerPropMigrationSuite) TestPassedProposalsDontChange() {
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

func TestConsumerPropMigration(t *testing.T) {
	genesis := chainsuite.DefaultGenesis()
	genesis = append(genesis,
		cosmos.NewGenesisKV("app_state.gov.params.max_deposit_period", permissionlessDepositPeriod.String()),
	)
	s := &ConsumerPropMigrationSuite{
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

func (s *ConsumerPropMigrationSuite) submitChangeRewardDenoms(consumer *chainsuite.Chain) (string, string) {
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

func (s *ConsumerPropMigrationSuite) submitConsumerModification(consumer *chainsuite.Chain) string {
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

func (s *ConsumerPropMigrationSuite) submitConsumerRemoval(consumer *chainsuite.Chain, stopTime time.Time) string {
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
