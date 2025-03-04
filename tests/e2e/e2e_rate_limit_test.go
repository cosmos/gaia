package e2e

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

const (
	proposalAddRateLimitAtomFilename     = "proposal_add_rate_limit_atom.json"
	proposalAddRateLimitStakeFilename    = "proposal_add_rate_limit_stake.json"
	proposalUpdateRateLimitAtomFilename  = "proposal_update_rate_limit_atom.json"
	proposalResetRateLimitAtomFilename   = "proposal_reset_rate_limit_atom.json"
	proposalRemoveRateLimitAtomFilename  = "proposal_remove_rate_limit_atom.json"
	proposalRemoveRateLimitStakeFilename = "proposal_remove_rate_limit_stake.json"
)

func (s *IntegrationTestSuite) writeAddRateLimitAtomProposal(c *chain, v2 bool) {
	template := `
	{
		"messages": [
		 {
		  "@type": "/ratelimit.v1.MsgAddRateLimit",
		  "authority": "%s",
		  "denom": "%s",
		  "channel_or_client_id": "%s",
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

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		uatomDenom,                 // denom: uatom
		channel,                    // channel_or_client_id: channel-0 / 08-wasm-1
		sdkmath.NewInt(1).String(), // max_percent_send: 1%
		sdkmath.NewInt(1).String(), // max_percent_recv: 1%
		24,                         // duration_hours: 24
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalAddRateLimitAtomFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeAddRateLimitStakeProposal(c *chain, v2 bool) {
	template := `
	{
		"messages": [
		 {
		  "@type": "/ratelimit.v1.MsgAddRateLimit",
		  "authority": "%s",
		  "denom": "%s",
		  "channel_or_client_id": "%s",
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

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		stakeDenom,                  // denom: stake
		channel,                     // channel_or_client_id: channel-0 / 08-wasm-1
		sdkmath.NewInt(10).String(), // max_percent_send: 10%
		sdkmath.NewInt(5).String(),  // max_percent_recv: 5%
		6,                           // duration_hours: 6
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalAddRateLimitStakeFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeUpdateRateLimitAtomProposal(c *chain, v2 bool) {
	template := `
	{
		"messages": [
		 {
		  "@type": "/ratelimit.v1.MsgUpdateRateLimit",
		  "authority": "%s",
		  "denom": "%s",
		  "channel_or_client_id": "%s",
		  "max_percent_send": "%s",
		  "max_percent_recv": "%s",
		  "duration_hours": "%d"
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "Update Rate Limit on (channel-0, uatom)",
		"summary": "e2e-test updating an IBC rate limit"
	   }`

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		uatomDenom,                 // denom: uatom
		channel,                    // channel_or_client_id: channel-0 / 08-wasm-1
		sdkmath.NewInt(2).String(), // max_percent_send: 2%
		sdkmath.NewInt(1).String(), // max_percent_recv: 1%
		6,                          // duration_hours: 6
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalUpdateRateLimitAtomFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeResetRateLimitAtomProposal(c *chain, v2 bool) {
	template := `
	{
		"messages": [
		 {
		  "@type": "/ratelimit.v1.MsgResetRateLimit",
		  "authority": "%s",
		  "denom": "%s",
		  "channel_or_client_id": "%s"
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "Reset Rate Limit on (channel-0, uatom)",
		"summary": "e2e-test resetting an IBC rate limit"
	   }`

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		uatomDenom, // denom: uatom
		channel,    // channel_or_client_id: channel-0 / 08-wasm-1
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalResetRateLimitAtomFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeRemoveRateLimitAtomProposal(c *chain, v2 bool) {
	template := `
	{
		"messages": [
		 {
		  "@type": "/ratelimit.v1.MsgRemoveRateLimit",
		  "authority": "%s",
		  "denom": "%s",
		  "channel_or_client_id": "%s"
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "Remove Rate Limit (channel-0, uatom)",
		"summary": "e2e-test removing an IBC rate limit"
	   }`

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		uatomDenom, // denom: uatom
		channel,    // channel_or_client_id: channel-0 / 08-wasm-1
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalRemoveRateLimitAtomFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) writeRemoveRateLimitStakeProposal(c *chain, v2 bool) {
	template := `
	{
		"messages": [
		 {
		  "@type": "/ratelimit.v1.MsgRemoveRateLimit",
		  "authority": "%s",
		  "denom": "%s",
		  "channel_or_client_id": "%s"
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100uatom",
		"title": "Remove Rate Limit (channel-0, stake)",
		"summary": "e2e-test removing an IBC rate limit"
	   }`

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}
	propMsgBody := fmt.Sprintf(template,
		govAuthority,
		stakeDenom, // denom: stake
		channel,    // channel_or_client_id: channel-0 / 08-wasm-1
	)

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalRemoveRateLimitStakeFilename), []byte(propMsgBody))
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) testAddRateLimits(v2 bool) {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	validatorAAddr, _ := validatorA.keyInfo.GetAddress()

	s.writeAddRateLimitAtomProposal(s.chainA, v2)
	proposalCounter++
	submitGovFlags := []string{configFile(proposalAddRateLimitAtomFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}

	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Add IBC rate limit for (%s, %s)", channel, uatomDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, "ratelimittypes.MsgAddRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

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
	proposalCounter++
	submitGovFlags = []string{configFile(proposalAddRateLimitStakeFilename)}
	depositGovFlags = []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags = []string{strconv.Itoa(proposalCounter), "yes"}

	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Add IBC rate limit for (%s, %s)", channel, stakeDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, "ratelimittypes.MsgAddRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

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
	proposalCounter++
	submitGovFlags := []string{configFile(proposalUpdateRateLimitAtomFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}

	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Update IBC rate limit for (%s, %s)", channel, uatomDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, "ratelimittypes.MsgUpdateRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

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
	proposalCounter++
	submitGovFlags := []string{configFile(proposalResetRateLimitAtomFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}

	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Reset IBC rate limit for (%s, %s)", channel, uatomDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, "ratelimittypes.MsgResetRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

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
	proposalCounter++
	submitGovFlags := []string{configFile(proposalRemoveRateLimitAtomFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	channel := transferChannel
	if v2 {
		channel = v2TransferClient
	}

	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Remove IBC rate limit for (%s, %s)", channel, uatomDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, "ratelimittypes.MsgRemoveRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

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
	proposalCounter++
	submitGovFlags = []string{configFile(proposalRemoveRateLimitStakeFilename)}
	depositGovFlags = []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags = []string{strconv.Itoa(proposalCounter), "yes"}

	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote Gov Proposal: Remove IBC rate limit for (%s, %s)", channel, stakeDenom)
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, "ratelimittypes.MsgRemoveRateLimit", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

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
