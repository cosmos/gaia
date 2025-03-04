package e2e

import (
	"context"
	"fmt"
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/msg"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
)

func (s *IntegrationTestSuite) testAddRateLimits(v2 bool) {
	chainEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))

	validatorA := s.Resources.ChainA.Validators[0]
	validatorAAddr, _ := validatorA.KeyInfo.GetAddress()

	err := msg.WriteAddRateLimitAtomProposal(s.Resources.ChainA, v2)
	s.Require().NoError(err)
	s.TestCounters.ProposalCounter++
	submitGovFlags := []string{configFile(common.ProposalAddRateLimitAtomFilename)}
	depositGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes"}

	channel := common.TransferChannel
	if v2 {
		channel = common.V2TransferClient
	}

	s.T().Logf("Proposal number: %d", s.TestCounters.ProposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Add IBC rate limit for (%s, %s)", channel, common.UAtomDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.TestCounters.ProposalCounter, "ratelimittypes.MsgAddRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			channel := common.TransferChannel
			if v2 {
				channel = common.V2TransferClient
			}
			s.T().Logf("After AddRateLimit proposal (channel-0, uatom)")

			rateLimits, err := query.AllRateLimits(chainEndpoint)
			s.Require().NoError(err)
			s.Require().Len(rateLimits, 1)
			s.Require().Equal(channel, rateLimits[0].Path.ChannelOrClientId)
			s.Require().Equal(common.UAtomDenom, rateLimits[0].Path.Denom)
			s.Require().Equal(uint64(24), rateLimits[0].Quota.DurationHours)
			s.Require().Equal(sdkmath.NewInt(1), rateLimits[0].Quota.MaxPercentRecv)
			s.Require().Equal(sdkmath.NewInt(1), rateLimits[0].Quota.MaxPercentSend)

			res, err := query.RateLimit(chainEndpoint, channel, common.UAtomDenom)
			s.Require().NoError(err)
			s.Require().NotNil(res.RateLimit)
			s.Require().Equal(*rateLimits[0].Path, *res.RateLimit.Path)
			s.Require().Equal(*rateLimits[0].Quota, *res.RateLimit.Quota)

			if !v2 {
				rateLimitsByChainID, err := query.RateLimitsByChainID(chainEndpoint, s.Resources.ChainB.ID)
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

	err = msg.WriteAddRateLimitStakeProposal(s.Resources.ChainA, v2)
	s.Require().NoError(err)
	s.TestCounters.ProposalCounter++
	submitGovFlags = []string{configFile(common.ProposalAddRateLimitStakeFilename)}
	depositGovFlags = []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags = []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes"}

	s.T().Logf("Proposal number: %d", s.TestCounters.ProposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Add IBC rate limit for (%s, %s)", channel, common.StakeDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.TestCounters.ProposalCounter, "ratelimittypes.MsgAddRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			channel := common.TransferChannel
			if v2 {
				channel = common.V2TransferClient
			}
			s.T().Logf("After AddRateLimit proposal (channel-0, stake)")

			rateLimits, err := query.AllRateLimits(chainEndpoint)
			s.Require().NoError(err)
			s.Require().Len(rateLimits, 2)
			// Note: the rate limits are ordered lexicographically by denom
			s.Require().Equal(channel, rateLimits[0].Path.ChannelOrClientId)
			s.Require().Equal(common.StakeDenom, rateLimits[0].Path.Denom)
			s.Require().Equal(uint64(6), rateLimits[0].Quota.DurationHours)
			s.Require().Equal(sdkmath.NewInt(5), rateLimits[0].Quota.MaxPercentRecv)
			s.Require().Equal(sdkmath.NewInt(10), rateLimits[0].Quota.MaxPercentSend)

			res, err := query.RateLimit(chainEndpoint, channel, common.StakeDenom)
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
	chainEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))

	validatorA := s.Resources.ChainA.Validators[0]
	validatorAAddr, _ := validatorA.KeyInfo.GetAddress()

	err := msg.WriteUpdateRateLimitAtomProposal(s.Resources.ChainA, v2)
	s.Require().NoError(err)
	s.TestCounters.ProposalCounter++
	submitGovFlags := []string{configFile(common.ProposalUpdateRateLimitAtomFilename)}
	depositGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes"}

	channel := common.TransferChannel
	if v2 {
		channel = common.V2TransferClient
	}

	s.T().Logf("Proposal number: %d", s.TestCounters.ProposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Update IBC rate limit for (%s, %s)", channel, common.UAtomDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.TestCounters.ProposalCounter, "ratelimittypes.MsgUpdateRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			s.T().Logf("After UpdateRateLimit proposal")

			res, err := query.RateLimit(chainEndpoint, channel, common.UAtomDenom)
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
	chainEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))

	validatorA := s.Resources.ChainA.Validators[0]
	validatorAAddr, _ := validatorA.KeyInfo.GetAddress()

	err := msg.WriteResetRateLimitAtomProposal(s.Resources.ChainA, v2)
	s.Require().NoError(err)
	s.TestCounters.ProposalCounter++
	submitGovFlags := []string{configFile(common.ProposalResetRateLimitAtomFilename)}
	depositGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes"}

	channel := common.TransferChannel
	if v2 {
		channel = common.V2TransferClient
	}

	s.T().Logf("Proposal number: %d", s.TestCounters.ProposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Reset IBC rate limit for (%s, %s)", channel, common.UAtomDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.TestCounters.ProposalCounter, "ratelimittypes.MsgResetRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			s.T().Logf("After ResetRateLimit proposal")

			res, err := query.RateLimit(chainEndpoint, channel, common.UAtomDenom)
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
	chainEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))

	validatorA := s.Resources.ChainA.Validators[0]
	validatorAAddr, _ := validatorA.KeyInfo.GetAddress()

	err := msg.WriteRemoveRateLimitAtomProposal(s.Resources.ChainA, v2)
	s.Require().NoError(err)
	s.TestCounters.ProposalCounter++
	submitGovFlags := []string{configFile(common.ProposalRemoveRateLimitAtomFilename)}
	depositGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes"}

	channel := common.TransferChannel
	if v2 {
		channel = common.V2TransferClient
	}

	s.T().Logf("Proposal number: %d", s.TestCounters.ProposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Remove IBC rate limit for (%s, %s)", channel, common.UAtomDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.TestCounters.ProposalCounter, "ratelimittypes.MsgRemoveRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			s.T().Logf("After RemoveRateLimit proposal")

			rateLimits, err := query.AllRateLimits(chainEndpoint)
			s.Require().NoError(err)
			s.Require().Len(rateLimits, 1)

			res, err := query.RateLimit(chainEndpoint, channel, common.UAtomDenom)
			s.Require().NoError(err)
			s.Require().Nil(res.RateLimit)

			return true
		},
		15*time.Second,
		5*time.Second,
	)

	err = msg.WriteRemoveRateLimitStakeProposal(s.Resources.ChainA, v2)
	s.Require().NoError(err)
	s.TestCounters.ProposalCounter++
	submitGovFlags = []string{configFile(common.ProposalRemoveRateLimitStakeFilename)}
	depositGovFlags = []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags = []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes"}

	s.T().Logf("Proposal number: %d", s.TestCounters.ProposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Remove IBC rate limit for (%s, %s)", channel, common.StakeDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.TestCounters.ProposalCounter, "ratelimittypes.MsgRemoveRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	s.Require().Eventually(
		func() bool {
			s.T().Logf("After RemoveRateLimit proposal")

			rateLimits, err := query.AllRateLimits(chainEndpoint)
			s.Require().NoError(err)
			s.Require().Len(rateLimits, 0)

			res, err := query.RateLimit(chainEndpoint, channel, common.StakeDenom)
			s.Require().NoError(err)
			s.Require().Nil(res.RateLimit)

			return true
		},
		15*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) testIBCTransfer(expToFail bool, v2 bool) {
	chainEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))

	address, _ := s.Resources.ChainA.Validators[0].KeyInfo.GetAddress()
	sender := address.String()

	address, _ = s.Resources.ChainB.Validators[0].KeyInfo.GetAddress()
	recipient := address.String()

	totalAmount, err := query.SupplyOf(chainEndpoint, common.UAtomDenom)
	s.Require().NoError(err)

	threshold := totalAmount.Amount.Mul(sdkmath.NewInt(1)).Quo(sdkmath.NewInt(100))
	tokenAmt := threshold.Add(sdkmath.NewInt(10)).String()

	channel := common.TransferChannel
	if v2 {
		channel = common.V2TransferClient
	}

	var absoluteTimeout *int64
	if v2 {
		timeout := time.Now().Unix() + 10000
		absoluteTimeout = &timeout
	}

	s.SendIBC(s.Resources.ChainA, 0, sender, recipient, tokenAmt+common.UAtomDenom, common.StandardFees.String(), "", channel, absoluteTimeout, expToFail)

	if !expToFail {
		s.T().Logf("After successful IBC transfer")

		res, err := query.RateLimit(chainEndpoint, channel, common.UAtomDenom)
		s.Require().NoError(err)
		s.Require().NotNil(res.RateLimit)
		s.Require().Equal(sdkmath.NewInt(0), res.RateLimit.Flow.Inflow)
		s.Require().NotEqual(sdkmath.NewInt(0), res.RateLimit.Flow.Outflow)
	}
}

func (s *IntegrationTestSuite) createV2LightClient() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	valIdx := 0
	val := s.chainA.validators[valIdx]
	address, _ := val.keyInfo.GetAddress()
	sender := address.String()

	clientState := `{"@type":"/ibc.lightclients.wasm.v1.ClientState","data":"ZG9lc250IG1hdHRlcg==","checksum":"O45STPnbLLar4DtFwDx0dE6tuXQW5XTKPHpbjaugun4=","latest_height":{"revision_number":"0","revision_height":"7795583"}}`
	consensusState := `{"@type":"/ibc.lightclients.wasm.v1.ConsensusState","data":"ZG9lc250IG1hdHRlcg=="}`

	s.T().Logf("sender: %s", sender)

	cmd := []string{
		gaiadBinary,
		txCommand,
		"ibc",
		"client",
		"create",
		clientState,
		consensusState,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, standardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, s.chainA.id),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.T().Logf("Creating light client on chain %s", s.chainA.id)
	s.executeGaiaTxCommand(ctx, s.chainA, cmd, valIdx, s.defaultExecValidation(s.chainA, valIdx))
	s.T().Log("successfully created light client")

	s.T().Logf("sender: %s", sender)

	cmd2 := []string{
		gaiadBinary,
		txCommand,
		"ibc",
		"client",
		"add-counterparty",
		v2TransferClient,
		"client-0",
		"aWJj",
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, standardFees.String()),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, s.chainA.id),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}

	s.T().Logf("Adding client counterparty on chain %s", s.chainA.id)
	s.executeGaiaTxCommand(ctx, s.chainA, cmd2, valIdx, s.defaultExecValidation(s.chainA, valIdx))
	s.T().Log("successfully added client counterparty")
}
