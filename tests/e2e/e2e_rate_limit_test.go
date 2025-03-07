package e2e

import (
	"fmt"
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
)

func (s *IntegrationTestSuite) testAddRateLimits(v2 bool) {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	validatorAAddr, _ := validatorA.keyInfo.GetAddress()

	s.writeAddRateLimitAtomProposal(s.chainA, v2)
	s.testCounters.proposalCounter++
	submitGovFlags := []string{configFile(proposalAddRateLimitAtomFilename)}
	depositGovFlags := []string{strconv.Itoa(s.testCounters.proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.testCounters.proposalCounter), "yes"}

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}

	s.T().Logf("Proposal number: %d", s.testCounters.proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Add IBC rate limit for (%s, %s)", channel, uatomDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.testCounters.proposalCounter, "ratelimittypes.MsgAddRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			channel := transferChannel
			if v2 {
				channel = v2TransferClient
			}
			s.T().Logf("After AddRateLimit proposal (channel-0, uatom)")

			rateLimits, err := queryAllRateLimits(chainEndpoint)
			s.Require().NoError(err)
			s.Require().Len(rateLimits, 1)
			s.Require().Equal(channel, rateLimits[0].Path.ChannelOrClientId)
			s.Require().Equal(uatomDenom, rateLimits[0].Path.Denom)
			s.Require().Equal(uint64(24), rateLimits[0].Quota.DurationHours)
			s.Require().Equal(sdkmath.NewInt(1), rateLimits[0].Quota.MaxPercentRecv)
			s.Require().Equal(sdkmath.NewInt(1), rateLimits[0].Quota.MaxPercentSend)

			res, err := queryRateLimit(chainEndpoint, channel, uatomDenom)
			s.Require().NoError(err)
			s.Require().NotNil(res.RateLimit)
			s.Require().Equal(*rateLimits[0].Path, *res.RateLimit.Path)
			s.Require().Equal(*rateLimits[0].Quota, *res.RateLimit.Quota)

			if !v2 {
				rateLimitsByChainID, err := queryRateLimitsByChainID(chainEndpoint, s.chainB.id)
				s.Require().NoError(err)
				s.Require().Len(rateLimits, 1)
				s.Require().Equal(*rateLimits[0].Path, *rateLimitsByChainID[0].Path)
				s.Require().Equal(*rateLimits[0].Quota, *rateLimitsByChainID[0].Quota)
			}

			return true
		},
		15*time.Second,
		5*time.Second,
	)

	s.writeAddRateLimitStakeProposal(s.chainA, v2)
	s.testCounters.proposalCounter++
	submitGovFlags = []string{configFile(proposalAddRateLimitStakeFilename)}
	depositGovFlags = []string{strconv.Itoa(s.testCounters.proposalCounter), depositAmount.String()}
	voteGovFlags = []string{strconv.Itoa(s.testCounters.proposalCounter), "yes"}

	s.T().Logf("Proposal number: %d", s.testCounters.proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Add IBC rate limit for (%s, %s)", channel, stakeDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.testCounters.proposalCounter, "ratelimittypes.MsgAddRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			channel := transferChannel
			if v2 {
				channel = v2TransferClient
			}
			s.T().Logf("After AddRateLimit proposal (channel-0, stake)")

			rateLimits, err := queryAllRateLimits(chainEndpoint)
			s.Require().NoError(err)
			s.Require().Len(rateLimits, 2)
			// Note: the rate limits are ordered lexicographically by denom
			s.Require().Equal(channel, rateLimits[0].Path.ChannelOrClientId)
			s.Require().Equal(stakeDenom, rateLimits[0].Path.Denom)
			s.Require().Equal(uint64(6), rateLimits[0].Quota.DurationHours)
			s.Require().Equal(sdkmath.NewInt(5), rateLimits[0].Quota.MaxPercentRecv)
			s.Require().Equal(sdkmath.NewInt(10), rateLimits[0].Quota.MaxPercentSend)

			res, err := queryRateLimit(chainEndpoint, channel, stakeDenom)
			s.Require().NoError(err)
			s.Require().NotNil(res.RateLimit)
			s.Require().Equal(*rateLimits[0].Path, *res.RateLimit.Path)
			s.Require().Equal(*rateLimits[0].Quota, *res.RateLimit.Quota)

			return true
		},
		15*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) testUpdateRateLimit(v2 bool) {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	validatorAAddr, _ := validatorA.keyInfo.GetAddress()

	s.writeUpdateRateLimitAtomProposal(s.chainA, v2)
	s.testCounters.proposalCounter++
	submitGovFlags := []string{configFile(proposalUpdateRateLimitAtomFilename)}
	depositGovFlags := []string{strconv.Itoa(s.testCounters.proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.testCounters.proposalCounter), "yes"}

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}

	s.T().Logf("Proposal number: %d", s.testCounters.proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Update IBC rate limit for (%s, %s)", channel, uatomDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.testCounters.proposalCounter, "ratelimittypes.MsgUpdateRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			s.T().Logf("After UpdateRateLimit proposal")

			res, err := queryRateLimit(chainEndpoint, channel, uatomDenom)
			s.Require().NoError(err)
			s.Require().NotNil(res.RateLimit)
			s.Require().Equal(sdkmath.NewInt(2), res.RateLimit.Quota.MaxPercentSend)
			s.Require().Equal(uint64(6), res.RateLimit.Quota.DurationHours)

			return true
		},
		15*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) testResetRateLimit(v2 bool) {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	validatorAAddr, _ := validatorA.keyInfo.GetAddress()

	s.writeResetRateLimitAtomProposal(s.chainA, v2)
	s.testCounters.proposalCounter++
	submitGovFlags := []string{configFile(proposalResetRateLimitAtomFilename)}
	depositGovFlags := []string{strconv.Itoa(s.testCounters.proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.testCounters.proposalCounter), "yes"}

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}

	s.T().Logf("Proposal number: %d", s.testCounters.proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Reset IBC rate limit for (%s, %s)", channel, uatomDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.testCounters.proposalCounter, "ratelimittypes.MsgResetRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			s.T().Logf("After ResetRateLimit proposal")

			res, err := queryRateLimit(chainEndpoint, channel, uatomDenom)
			s.Require().NoError(err)
			s.Require().NotNil(res.RateLimit)
			s.Require().Equal(sdkmath.NewInt(0), res.RateLimit.Flow.Inflow)
			s.Require().Equal(sdkmath.NewInt(0), res.RateLimit.Flow.Outflow)

			return true
		},
		15*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) testRemoveRateLimit(v2 bool) {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	validatorAAddr, _ := validatorA.keyInfo.GetAddress()

	s.writeRemoveRateLimitAtomProposal(s.chainA, v2)
	s.testCounters.proposalCounter++
	submitGovFlags := []string{configFile(proposalRemoveRateLimitAtomFilename)}
	depositGovFlags := []string{strconv.Itoa(s.testCounters.proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.testCounters.proposalCounter), "yes"}

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}

	s.T().Logf("Proposal number: %d", s.testCounters.proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Remove IBC rate limit for (%s, %s)", channel, uatomDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.testCounters.proposalCounter, "ratelimittypes.MsgRemoveRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			s.T().Logf("After RemoveRateLimit proposal")

			rateLimits, err := queryAllRateLimits(chainEndpoint)
			s.Require().NoError(err)
			s.Require().Len(rateLimits, 1)

			res, err := queryRateLimit(chainEndpoint, channel, uatomDenom)
			s.Require().NoError(err)
			s.Require().Nil(res.RateLimit)

			return true
		},
		15*time.Second,
		5*time.Second,
	)

	s.writeRemoveRateLimitStakeProposal(s.chainA, v2)
	s.testCounters.proposalCounter++
	submitGovFlags = []string{configFile(proposalRemoveRateLimitStakeFilename)}
	depositGovFlags = []string{strconv.Itoa(s.testCounters.proposalCounter), depositAmount.String()}
	voteGovFlags = []string{strconv.Itoa(s.testCounters.proposalCounter), "yes"}

	s.T().Logf("Proposal number: %d", s.testCounters.proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Remove IBC rate limit for (%s, %s)", channel, stakeDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.testCounters.proposalCounter, "ratelimittypes.MsgRemoveRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			s.T().Logf("After RemoveRateLimit proposal")

			rateLimits, err := queryAllRateLimits(chainEndpoint)
			s.Require().NoError(err)
			s.Require().Len(rateLimits, 0)

			res, err := queryRateLimit(chainEndpoint, channel, stakeDenom)
			s.Require().NoError(err)
			s.Require().Nil(res.RateLimit)

			return true
		},
		15*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) testIBCTransfer(expToFail bool, v2 bool) {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	address, _ := s.chainA.validators[0].keyInfo.GetAddress()
	sender := address.String()

	address, _ = s.chainB.validators[0].keyInfo.GetAddress()
	recipient := address.String()

	totalAmount, err := querySupplyOf(chainEndpoint, uatomDenom)
	s.Require().NoError(err)

	threshold := totalAmount.Amount.Mul(sdkmath.NewInt(1)).Quo(sdkmath.NewInt(100))
	tokenAmt := threshold.Add(sdkmath.NewInt(10)).String()

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}

	var absoluteTimeout *int64
	if v2 {
		timeout := time.Now().Unix() + 10000
		absoluteTimeout = &timeout
	}

	s.sendIBC(s.chainA, 0, sender, recipient, tokenAmt+uatomDenom, standardFees.String(), "", channel, absoluteTimeout, expToFail)

	if !expToFail {
		s.T().Logf("After successful IBC transfer")

		res, err := queryRateLimit(chainEndpoint, channel, uatomDenom)
		s.Require().NoError(err)
		s.Require().NotNil(res.RateLimit)
		s.Require().Equal(sdkmath.NewInt(0), res.RateLimit.Flow.Inflow)
		s.Require().NotEqual(sdkmath.NewInt(0), res.RateLimit.Flow.Outflow)
	}
}
