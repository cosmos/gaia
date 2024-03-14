package e2e

import (
	"fmt"
	"strconv"
	"time"

	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	proposalAddRateLimitAtomFilename  = "proposal_add_rate_limit_atom.json"
	proposalAddRateLimitStakeFilename = "proposal_add_rate_limit_stake.json"
	proposalRemoveRateLimitFilename   = "proposal_remove_rate_limit.json"
)

func (s *IntegrationTestSuite) writeAddRateLimitAtomProposal(c *chain) {
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
		"title": "Add Rate Limit on (channel-0, uatom)",
		"summary": "e2e-test adding an IBC rate limit"
	   }`
	propMsgBody := fmt.Sprintf(template,
		uatomDenom,             // denom: uatom
		transferChannel,        // channel_id: channel-0
		sdk.NewInt(1).String(), // max_percent_send: 1%
		sdk.NewInt(1).String(), // max_percent_recv: 1%
		24,                     // duration_hours: 24
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalAddRateLimitAtomFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeAddRateLimitStakeProposal(c *chain) {
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
		"title": "Add Rate Limit on (channel-0, stake)",
		"summary": "e2e-test adding an IBC rate limit"
	   }`
	propMsgBody := fmt.Sprintf(template,
		stakeDenom,              // denom: stake
		transferChannel,         // channel_id: channel-0
		sdk.NewInt(10).String(), // max_percent_send: 10%
		sdk.NewInt(5).String(),  // max_percent_recv: 5%
		6,                       // duration_hours: 6
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalAddRateLimitStakeFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) testAddRateLimits() {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	validatorAAddr, _ := validatorA.keyInfo.GetAddress()

	s.writeAddRateLimitAtomProposal(s.chainA)
	proposalCounter++
	submitGovFlags := []string{configFile(proposalAddRateLimitAtomFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Add IBC rate limit for (channel-0, uatom)")
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, "ratelimittypes.MsgAddRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			rateLimits, err := queryAllRateLimits(chainEndpoint)
			s.T().Logf("After AddRateLimit proposal")
			s.Require().NoError(err)
			s.Require().Len(rateLimits, 1)
			s.Require().Equal(transferChannel, rateLimits[0].Path.ChannelId)
			s.Require().Equal(uatomDenom, rateLimits[0].Path.Denom)
			s.Require().Equal(uint64(24), rateLimits[0].Quota.DurationHours)
			s.Require().Equal(sdk.NewInt(1), rateLimits[0].Quota.MaxPercentRecv)
			s.Require().Equal(sdk.NewInt(1), rateLimits[0].Quota.MaxPercentSend)

			res, err := queryRateLimit(chainEndpoint, transferChannel, uatomDenom)
			s.Require().NoError(err)
			s.Require().NotNil(res)
			s.Require().Equal(*rateLimits[0].Path, *res.RateLimit.Path)
			s.Require().Equal(*rateLimits[0].Quota, *res.RateLimit.Quota)

			rateLimitsByChainId, err := queryRateLimitsByChainId(chainEndpoint, s.chainB.id)
			s.Require().NoError(err)
			s.Require().Len(rateLimits, 1)
			s.Require().Equal(*rateLimits[0].Path, *rateLimitsByChainId[0].Path)
			s.Require().Equal(*rateLimits[0].Quota, *rateLimitsByChainId[0].Quota)

			return true
		},
		15*time.Second,
		5*time.Second,
	)

	s.writeAddRateLimitStakeProposal(s.chainA)
	proposalCounter++
	submitGovFlags = []string{configFile(proposalAddRateLimitStakeFilename)}
	depositGovFlags = []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags = []string{strconv.Itoa(proposalCounter), "yes"}

	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Add IBC rate limit for (channel-0, stake)")
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, "ratelimittypes.MsgAddRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			rateLimits, err := queryAllRateLimits(chainEndpoint)
			s.T().Logf("After AddRateLimit proposal")
			s.Require().NoError(err)
			s.Require().Len(rateLimits, 2)
			// Note: the rate limits are ordered lexicographically by denom
			s.Require().Equal(transferChannel, rateLimits[0].Path.ChannelId)
			s.Require().Equal(stakeDenom, rateLimits[0].Path.Denom)
			s.Require().Equal(uint64(6), rateLimits[0].Quota.DurationHours)
			s.Require().Equal(sdk.NewInt(5), rateLimits[0].Quota.MaxPercentRecv)
			s.Require().Equal(sdk.NewInt(10), rateLimits[0].Quota.MaxPercentSend)

			res, err := queryRateLimit(chainEndpoint, transferChannel, stakeDenom)
			s.Require().NoError(err)
			s.Require().NotNil(res)
			s.Require().Equal(*rateLimits[0].Path, *res.RateLimit.Path)
			s.Require().Equal(*rateLimits[0].Quota, *res.RateLimit.Quota)

			return true
		},
		15*time.Second,
		5*time.Second,
	)
}
