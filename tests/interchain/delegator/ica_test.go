package delegator_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v26/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v26/tests/interchain/delegator"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ICAControllerSuite struct {
	*delegator.Suite
	Host          *chainsuite.Chain
	icaAddress    string
	srcChannel    *ibc.ChannelOutput
	srcAddress    string
	ibcStakeDenom string
}

const icaAcctFunds = int64(3_300_000_000)

func (s *ICAControllerSuite) SetupSuite() {
	s.Suite.SetupSuite()
	// Use upgraded chain spec for Host so it has the custom gov module with stake validation
	hostSpec := upgradedChainSpec(s.Env)
	host, err := s.Chain.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, hostSpec)
	s.Require().NoError(err)
	s.Host = host
	s.srcAddress = s.DelegatorWallet.FormattedAddress()
	s.srcChannel, err = s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.Host)
	s.Require().NoError(err)

	s.icaAddress, err = s.Chain.SetupICAAccount(s.GetContext(), s.Host, s.Relayer, s.srcAddress, 0, icaAcctFunds)
	s.Require().NoError(err)

	_, err = s.Chain.SendIBCTransfer(s.GetContext(), s.srcChannel.ChannelID, s.srcAddress, ibc.WalletAmount{
		Address: s.icaAddress,
		Amount:  sdkmath.NewInt(icaAcctFunds),
		Denom:   s.Chain.Config().Denom,
	}, ibc.TransferOptions{})
	s.Require().NoError(err)
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balances, err := s.Host.BankQueryAllBalances(s.GetContext(), s.icaAddress)
		s.Require().NoError(err)
		s.Require().NotEmpty(balances)
		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc") {
				s.ibcStakeDenom = c.Denom
				break
			}
		}
		assert.NotEmpty(c, s.ibcStakeDenom)
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
}

func (s *ICAControllerSuite) TestICABankSend() {
	wallets := s.Host.ValidatorWallets
	dstAddress := wallets[0].Address

	recipientBalanceBefore, err := s.Host.GetBalance(s.GetContext(), dstAddress, s.ibcStakeDenom)
	s.Require().NoError(err)

	icaAmount := int64(icaAcctFunds / 10)
	srcConnection := s.srcChannel.ConnectionHops[0]

	// Create the bank send transaction JSON
	jsonBankSend := fmt.Sprintf(`{"@type": "/cosmos.bank.v1beta1.MsgSend", "from_address":"%s","to_address":"%s","amount":[{"denom":"%s","amount":"%d"}]}`,
		s.icaAddress, dstAddress, s.ibcStakeDenom, icaAmount)

	s.Require().NoError(s.sendICATx(s.GetContext(), s.srcAddress, srcConnection, jsonBankSend))
	s.Relayer.ClearTransferChannel(s.GetContext(), s.Chain, s.Host)

	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		recipientBalanceAfter, err := s.Host.GetBalance(s.GetContext(), dstAddress, s.ibcStakeDenom)
		assert.NoError(c, err)

		assert.Equal(c, recipientBalanceBefore.Add(sdkmath.NewInt(icaAmount)), recipientBalanceAfter)
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
}

func (s *ICAControllerSuite) TestICADelegate() {
	const delegateAmount = int64(1000000)
	// Get validator address from host chain
	validator := s.Host.ValidatorWallets[0]

	// Query validator's voting power before delegation
	votingPowerBefore, err := s.Host.QueryJSON(s.GetContext(), "validator.tokens", "staking", "validator", validator.ValoperAddress)
	s.Require().NoError(err)
	votingPowerBeforeInt, err := chainsuite.StrToSDKInt(votingPowerBefore.String())
	s.Require().NoError(err)

	// Create the delegation transaction JSON
	jsonDelegate := fmt.Sprintf(`{
		"@type": "/cosmos.staking.v1beta1.MsgDelegate",
		"delegator_address": "%s",
		"validator_address": "%s",
		"amount": {
			"denom": "%s",
			"amount": "%d"
		}
	}`, s.icaAddress, validator.ValoperAddress, s.Host.Config().Denom, delegateAmount)

	// Send the ICA transaction
	srcConnection := s.srcChannel.ConnectionHops[0]
	s.Require().NoError(s.sendICATx(s.GetContext(), s.srcAddress, srcConnection, jsonDelegate))

	// Make sure changes are propagated through the relayer
	s.Relayer.ClearTransferChannel(s.GetContext(), s.Chain, s.Host)

	// Verify voting power has increased
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		votingPowerAfter, err := s.Host.QueryJSON(s.GetContext(), "validator.tokens", "staking", "validator", validator.ValoperAddress)
		assert.NoError(c, err)

		votingPowerAfterInt, err := chainsuite.StrToSDKInt(votingPowerAfter.String())
		assert.NoError(c, err)

		// Verify that voting power increased by the delegated amount
		expectedVotingPower := votingPowerBeforeInt.Add(sdkmath.NewInt(delegateAmount))
		assert.True(c, votingPowerAfterInt.Sub(expectedVotingPower).Abs().LTE(sdkmath.NewInt(1)),
			"voting power after: %s, expected: %s", votingPowerAfterInt, expectedVotingPower)

		delegations, _, err := s.Host.GetNode().ExecQuery(s.GetContext(), "staking",
			"delegation", s.icaAddress, validator.ValoperAddress)
		assert.NoError(c, err)
		assert.Contains(c, string(delegations), s.icaAddress)
		assert.Contains(c, string(delegations), validator.ValoperAddress)
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
}

func (s *ICAControllerSuite) TestICAGovVoteStakeValidation() {
	// Test that ICA gov votes require sufficient stake (validates the fix for the ICA bypass vector)
	const insufficientStake = int64(2)     // 2 uatom - insufficient for voting
	const sufficientStake = int64(1000000) // 1 ATOM - sufficient for voting

	validator := s.Host.ValidatorWallets[0]
	srcConnection := s.srcChannel.ConnectionHops[0]

	// 1. First delegate minimal tokens to ICA account (insufficient for voting)
	jsonDelegate := fmt.Sprintf(`{
		"@type": "/cosmos.staking.v1beta1.MsgDelegate",
		"delegator_address": "%s",
		"validator_address": "%s",
		"amount": {
			"denom": "%s",
			"amount": "%d"
		}
	}`, s.icaAddress, validator.ValoperAddress, s.Host.Config().Denom, insufficientStake)
	s.Require().NoError(s.sendICATx(s.GetContext(), s.srcAddress, srcConnection, jsonDelegate))
	s.Relayer.ClearTransferChannel(s.GetContext(), s.Chain, s.Host)

	// Verify delegation was created with insufficient stake
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		delegations, _, err := s.Host.GetNode().ExecQuery(s.GetContext(), "staking",
			"delegation", s.icaAddress, validator.ValoperAddress)
		assert.NoError(c, err)
		assert.Contains(c, string(delegations), s.icaAddress)
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// 2. Create and submit a proposal on the host chain
	prop, err := s.Host.BuildProposal(nil, "ICA Vote Stake Test", "Test", "ipfs://CID",
		chainsuite.GovDepositAmount, "", false)
	s.Require().NoError(err)
	result, err := s.Host.SubmitProposal(s.GetContext(), validator.Moniker, prop)
	s.Require().NoError(err)
	proposalId := result.ProposalID
	fmt.Println("Proposal ID:", proposalId)

	// 3. Attempt to vote via ICA with insufficient stake - should FAIL
	// Note: sendICATx only sends the packet; failure happens on host chain via ack
	jsonVote := fmt.Sprintf(`{
		"@type": "/cosmos.gov.v1.MsgVote",
		"proposal_id": "%s",
		"voter": "%s",
		"option": "VOTE_OPTION_YES"
	}`, proposalId, s.icaAddress)
	// Send the ICA tx - packet submission succeeds but execution on host should fail
	s.Require().NoError(s.sendICATx(s.GetContext(), s.srcAddress, srcConnection, jsonVote))
	s.Relayer.ClearTransferChannel(s.GetContext(), s.Chain, s.Host)

	// Wait one minute to ensure vote is not recorded
	time.Sleep(1 * time.Minute)

	// Print tally in host chain
	tally, err := s.Host.QueryJSON(s.GetContext(), "tally", "gov", "tally", proposalId)
	s.Require().NoError(err)
	fmt.Println("Tally after insufficient stake vote attempt:", tally.String())

	// Wait and verify the vote was NOT recorded (rejected due to insufficient stake)
	s.Require().Never(func() bool {
		vote, err := s.Host.QueryJSON(s.GetContext(), "vote", "gov", "vote", proposalId, s.icaAddress)
		if err != nil {
			return false // Query error means vote doesn't exist - expected
		}
		// If we got a result, check if it has actual vote options
		return vote.Get("options").Exists() && len(vote.Get("options").Array()) > 0
	}, 5*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "vote should NOT be recorded with insufficient stake")

	vote, err := s.Host.QueryJSON(s.GetContext(), "vote", "gov", "vote", proposalId, s.icaAddress)
	s.Require().Error(err)

	// 4. Delegate more tokens to meet stake requirement
	jsonDelegate = fmt.Sprintf(`{
		"@type": "/cosmos.staking.v1beta1.MsgDelegate",
		"delegator_address": "%s",
		"validator_address": "%s",
		"amount": {
			"denom": "%s",
			"amount": "%d"
		}
	}`, s.icaAddress, validator.ValoperAddress, s.Host.Config().Denom, sufficientStake)
	s.Require().NoError(s.sendICATx(s.GetContext(), s.srcAddress, srcConnection, jsonDelegate))
	s.Relayer.ClearTransferChannel(s.GetContext(), s.Chain, s.Host)

	// Wait for delegation to be confirmed
	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		response, err := s.Host.StakingQueryDelegations(s.GetContext(), s.icaAddress)
		assert.NoError(c, err)
		assert.NotEmpty(c, response)
		// Verify total delegation is now sufficient (insufficientStake + sufficientStake)
		totalStake := insufficientStake + sufficientStake
		assert.True(c, response[0].Balance.Amount.GTE(sdkmath.NewInt(totalStake)),
			"expected delegation >= %d, got %s", totalStake, response[0].Balance.Amount)
	}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Submit a new proposal to reset voting period
	prop2, err := s.Host.BuildProposal(nil, "ICA Vote Stake Test 2", "Test", "ipfs://CID2",
		chainsuite.GovDepositAmount, "", false)
	s.Require().NoError(err)
	result2, err := s.Host.SubmitProposal(s.GetContext(), validator.Moniker, prop2)
	s.Require().NoError(err)
	proposalId = result2.ProposalID
	fmt.Println("Proposal ID for second proposal:", proposalId)

	// 5. Attempt to vote via ICA with sufficient stake - should SUCCEED
	jsonVote = fmt.Sprintf(`{
		"@type": "/cosmos.gov.v1.MsgVote",
		"proposal_id": "%s",
		"voter": "%s",
		"option": "VOTE_OPTION_YES"
	}`, proposalId, s.icaAddress)
	// Send the ICA tx
	s.Require().NoError(s.sendICATx(s.GetContext(), s.srcAddress, srcConnection, jsonVote))
	s.Relayer.ClearTransferChannel(s.GetContext(), s.Chain, s.Host)

	// Wait one minute to ensure vote is recorded
	time.Sleep(1 * time.Minute)

	// Print tally in host chain
	tally, err = s.Host.QueryJSON(s.GetContext(), "tally", "gov", "tally", proposalId)
	s.Require().NoError(err)
	fmt.Println("Tally after sufficient stake vote attempt:", tally.String())

	// 6. Test vote was counted
	vote, err = s.Host.QueryJSON(s.GetContext(), "vote", "gov", "vote", proposalId, s.icaAddress)
	s.Require().NoError(err)
	// Print vote
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("ICA Vote: %s", vote.String())
	actual_yes_weight := vote.Get("options.#(option==\"VOTE_OPTION_YES\").weight")
	s.Require().Equal(float64(1.0), actual_yes_weight.Float())
}

func TestDelegatorICA(t *testing.T) {
	s := &ICAControllerSuite{Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
		CreateRelayer:  true,
	})}}
	suite.Run(t, s)
}

func (s *ICAControllerSuite) sendICATx(ctx context.Context, srcAddress string, srcConnection string, txJSON string) error {
	msgBz, _, err := s.Chain.GetNode().Exec(ctx, []string{"gaiad", "tx", "ica", "host", "generate-packet-data", txJSON, "--encoding", "proto3"}, nil)
	if err != nil {
		return err
	}

	msgPath := "msg.json"
	if err := s.Chain.GetNode().WriteFile(ctx, msgBz, msgPath); err != nil {
		return err
	}
	msgPath = s.Chain.GetNode().HomeDir() + "/" + msgPath
	_, err = s.Chain.GetNode().ExecTx(ctx, srcAddress,
		"interchain-accounts", "controller", "send-tx",
		srcConnection, msgPath,
	)
	if err != nil {
		return err
	}
	return nil
}

// upgradedChainSpec returns a ChainSpec using the new Gaia version (with custom gov module)
func upgradedChainSpec(env chainsuite.Environment) *interchaintest.ChainSpec {
	fullNodes := 0
	var repository string
	if env.DockerRegistry == "" {
		repository = env.GaiaImageName
	} else {
		repository = fmt.Sprintf("%s/%s", env.DockerRegistry, env.GaiaImageName)
	}
	return &interchaintest.ChainSpec{
		Name:          "gaia",
		NumFullNodes:  &fullNodes,
		NumValidators: &chainsuite.OneValidator,
		Version:       env.NewGaiaImageVersion, // Use NEW version with custom gov module
		ChainConfig: ibc.ChainConfig{
			Denom:         chainsuite.Uatom,
			GasPrices:     chainsuite.GasPrices,
			GasAdjustment: 2.0,
			ConfigFileOverrides: map[string]any{
				"config/config.toml": chainsuite.DefaultConfigToml(),
			},
			Images: []ibc.DockerImage{{
				Repository: repository,
				UIDGID:     "1025:1025",
			}},
			ModifyGenesis:        cosmos.ModifyGenesis(chainsuite.DefaultGenesis()),
			ModifyGenesisAmounts: chainsuite.DefaultGenesisAmounts(chainsuite.Uatom),
		},
	}
}
