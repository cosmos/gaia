package e2e

import (
	"fmt"
	"strconv"
	"time"

	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	proposalAddRateLimitFilename    = "proposal_add_rate_limit.json"
	proposalRemoveRateLimitFilename = "proposal_remove_rate_limit.json"
)

func (s *IntegrationTestSuite) writeAddRateLimitProposal(c *chain) {
	template := `
	{
		"messages": [
		 {
		  "@type": "/ratelimit.v1.MsgAddRateLimit",
		  "authority": "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
		  "denom": "%s",
		  "channel_id": "%s",
		  "max_percent_send": "%s",
		  "max_percent_recv": "%s",
		  "duration_hours": "%d"
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "Add Rate Limit",
		"summary": "e2e-test adding an IBC rate limit"
	   }`
	propMsgBody := fmt.Sprintf(template,
		uatomDenom,             // denom: uatom
		transferChannel,        // channel_id: channel-0
		sdk.NewInt(1).String(), // max_percent_send: 1%
		sdk.NewInt(1).String(), // max_percent_recv: 1%
		24,                     // duration_hours: 24
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalAddRateLimitFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) testAddRateLimit() {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	validatorAAddr, _ := validatorA.keyInfo.GetAddress()

	s.writeAddRateLimitProposal(s.chainA)
	proposalCounter++
	submitGovFlags := []string{configFile(proposalAddRateLimitFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Add IBC rate limit")
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, "ratelimittypes.MsgAddRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	// query the proposal status
	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(chainEndpoint, proposalCounter)
			s.Require().NoError(err)
			return proposal.GetProposal().Status == govv1beta1.StatusPassed
		},
		15*time.Second,
		5*time.Second,
	)
}
