package e2e

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	providertypes "github.com/cosmos/interchain-security/v5/x/ccv/provider/types"
)

const (
	proposalAddConsumerChainFilename    = "proposal_add_consumer.json"
	proposalRemoveConsumerChainFilename = "proposal_remove_consumer.json"
)

type ConsumerAdditionProposalWithDeposit struct {
	providertypes.ConsumerAdditionProposal
	Deposit string `json:"deposit"`
	Summary string `json:"summary"` // required on legacy proposals
}

type ConsumerRemovalProposalWithDeposit struct {
	providertypes.ConsumerRemovalProposal
	Deposit string `json:"deposit"`
	Summary string `json:"summary"` // required on legacy proposals
}

func (s *IntegrationTestSuite) writeAddRemoveConsumerProposals(c *chain, consumerChainID string) {
	hash, _ := json.Marshal("Z2VuX2hhc2g=")
	addProp := &providertypes.ConsumerAdditionProposal{
		Title:       "Create consumer chain",
		Description: "First consumer chain",
		ChainId:     consumerChainID,
		InitialHeight: ibcclienttypes.Height{
			RevisionHeight: 1,
		},
		GenesisHash:                       hash,
		BinaryHash:                        hash,
		SpawnTime:                         time.Now(),
		UnbondingPeriod:                   time.Duration(100000000000),
		CcvTimeoutPeriod:                  time.Duration(100000000000),
		TransferTimeoutPeriod:             time.Duration(100000000000),
		ConsumerRedistributionFraction:    "0.75",
		BlocksPerDistributionTransmission: 10,
		HistoricalEntries:                 10000,
		Top_N:                             95,
	}
	addPropWithDeposit := ConsumerAdditionProposalWithDeposit{
		ConsumerAdditionProposal: *addProp,
		Deposit:                  "1000uatom",
		// Summary is
		Summary: "Summary for the First consumer chain addition proposal",
	}

	removeProp := &providertypes.ConsumerRemovalProposal{
		Title:       "Remove consumer chain",
		Description: "Removing consumer chain",
		ChainId:     consumerChainID,
		StopTime:    time.Now(),
	}

	removePropWithDeposit := ConsumerRemovalProposalWithDeposit{
		ConsumerRemovalProposal: *removeProp,
		Summary:                 "Summary for the First consumer chain removal proposal",
		Deposit:                 "1000uatom",
	}

	consumerAddBody, err := json.MarshalIndent(addPropWithDeposit, "", " ")
	s.Require().NoError(err)

	consumerRemoveBody, err := json.MarshalIndent(removePropWithDeposit, "", " ")
	s.Require().NoError(err)

	err = writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalAddConsumerChainFilename), consumerAddBody)
	s.Require().NoError(err)
	err = writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalRemoveConsumerChainFilename), consumerRemoveBody)
	s.Require().NoError(err)
}

/*
AddRemoveConsumerChain tests adding and subsequently removing a new consumer chain to Gaia.
Test Benchmarks:
1. Submit and pass proposal to add consumer chain
2. Validation that consumer chain was added
3. Submit and pass proposal to remove consumer chain
4. Validation that consumer chain was removed
*/
func (s *IntegrationTestSuite) AddRemoveConsumerChain() {
	s.fundCommunityPool()
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	proposerAddress, _ := s.chainA.validators[0].keyInfo.GetAddress()
	sender := proposerAddress.String()
	consumerChainID := "consumer"
	s.writeAddRemoveConsumerProposals(s.chainA, consumerChainID)

	// Gov tests may be run in arbitrary order, each test must increment proposalCounter to have the correct proposal id to submit and query
	// Add Consumer Chain
	proposalCounter++
	submitGovFlags := []string{"consumer-addition", configFile(proposalAddConsumerChainFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}
	s.submitLegacyGovProposal(chainAAPIEndpoint, sender, proposalCounter, providertypes.ProposalTypeConsumerAddition, submitGovFlags, depositGovFlags, voteGovFlags, "vote", false)

	// Query and assert consumer has been added
	s.execQueryConsumerChains(s.chainA, 0, gaiaHomePath, validateConsumerAddition, consumerChainID)

	// Remove Consumer Chain
	proposalCounter++
	submitGovFlags = []string{"consumer-removal", configFile(proposalRemoveConsumerChainFilename)}
	depositGovFlags = []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags = []string{strconv.Itoa(proposalCounter), "yes"}
	s.submitLegacyGovProposal(chainAAPIEndpoint, sender, proposalCounter, providertypes.ProposalTypeConsumerRemoval, submitGovFlags, depositGovFlags, voteGovFlags, "vote", false)
	// Query and assert consumer has been removed
	s.execQueryConsumerChains(s.chainA, 0, gaiaHomePath, validateConsumerRemoval, consumerChainID)
}

func validateConsumerAddition(res providertypes.QueryConsumerChainsResponse, consumerChainID string) bool {
	if res.Size() == 0 {
		return false
	}
	for _, chain := range res.GetChains() {
		return strings.Compare(chain.ChainId, consumerChainID) == 0
	}
	return false
}

func validateConsumerRemoval(res providertypes.QueryConsumerChainsResponse, consumerChainID string) bool {
	if res.Size() > 0 {
		for _, chain := range res.GetChains() {
			if strings.Compare(chain.ChainId, consumerChainID) == 0 {
				return false
			}
		}
	}
	return true
}
