package delegator_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/gaia/v29/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v29/tests/interchain/delegator"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// rateLimitTransferAmount is the amount of chain B's native denom transferred
// to chain A to mint the IBC voucher denom under test. Since this is the
// first (and only) transfer of this denom trace, it becomes that denom's
// entire on-chain supply, making rate-limit quota percentages exact.
const rateLimitTransferAmount = int64(1_000_000)

// RateLimitUpgradeSuite verifies that a rate limit configured on an IBC
// denom BEFORE a chain software upgrade is still present and enforced AFTER
// the upgrade completes.
type RateLimitUpgradeSuite struct {
	*delegator.Suite
	ChainB       *chainsuite.Chain
	ChainBWallet ibc.Wallet

	// Recorded pre-upgrade so the post-upgrade tests can assert nothing drifted.
	RateLimitDenom         string
	RateLimitChannelID     string
	RateLimitTotalSupply   int64
	RateLimitSendPercent   int64
	RateLimitRecvPercent   int64
	RateLimitDurationHours uint64
}

func (s *RateLimitUpgradeSuite) SetupSuite() {
	s.Suite.SetupSuite() // chain created on the old version, not yet upgraded (UpgradeOnSetup: false)

	ctx := s.GetContext()

	// Voting power for the gov proposal below.
	err := s.Chain.GetNode().StakingDelegate(ctx, s.DelegatorWallet.KeyName(),
		s.Chain.ValidatorWallets[0].ValoperAddress, "10000000"+s.Chain.Config().Denom)
	s.Require().NoError(err)

	// Second chain + real transfer channel for the rate limit's channel_or_client_id.
	chainB, err := s.Chain.AddLinkedChain(ctx, s.T(), s.Relayer, chainsuite.DefaultChainSpec(s.Env))
	s.Require().NoError(err)
	s.ChainB = chainB

	wallet, err := chainB.BuildWallet(ctx, "delegator", "")
	s.Require().NoError(err)
	s.ChainBWallet = wallet

	s.Require().NoError(chainB.SendFunds(ctx, interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Address: s.ChainBWallet.FormattedAddress(),
		Amount:  sdkmath.NewInt(100_000_000_000),
		Denom:   chainB.Config().Denom,
	}))

	transferCh, err := s.Relayer.GetTransferChannel(ctx, s.Chain, s.ChainB)
	s.Require().NoError(err)
	s.RateLimitChannelID = transferCh.ChannelID

	// Send tokens from chain B to chain A, minting a fresh IBC voucher denom
	// on chain A whose entire supply equals this one transfer.
	err = s.ibcTransfer(ctx, s.ChainB, s.ChainBWallet, s.DelegatorWallet.FormattedAddress(),
		fmt.Sprintf("%d", rateLimitTransferAmount), s.ChainB.Config().Denom, transferCh.Counterparty.ChannelID)
	s.Require().NoError(err)

	ibcDenom := transfertypes.NewDenom(
		s.ChainB.Config().Denom,
		transfertypes.NewHop(transferCh.PortID, transferCh.ChannelID),
	).IBCDenom()
	s.RateLimitDenom = ibcDenom
	s.RateLimitTotalSupply = rateLimitTransferAmount

	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		balance, err := s.Chain.GetBalance(ctx, s.DelegatorWallet.FormattedAddress(), ibcDenom)
		assert.NoError(c, err)
		assert.True(c, balance.Equal(sdkmath.NewInt(rateLimitTransferAmount)), "chain A should receive the funding transfer")
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

	// Add the rate limit via governance, before the upgrade.
	s.RateLimitSendPercent = 10
	s.RateLimitRecvPercent = 10
	s.RateLimitDurationHours = 24
	s.addRateLimit(ctx, s.RateLimitDenom, s.RateLimitChannelID,
		s.RateLimitSendPercent, s.RateLimitRecvPercent, s.RateLimitDurationHours)

	// Verify the rate limit is actually operational before the upgrade. These
	// use a small slice of the 10% quota (2% then an attempted 15% more) so
	// there's still comfortable headroom left for TestRateLimitEnforcedAfterUpgrade,
	// which reuses this same rate limit's cumulative flow post-upgrade.
	s.assertTransferWithinQuota(ctx, 2, "in-quota transfer should succeed before the upgrade")
	s.assertTransferBlockedByQuota(ctx, 15, "over-quota transfer should be blocked before the upgrade")

	// Now upgrade the chain — the rate limit above must survive this.
	s.UpgradeChain()
}

// Helper functions

// addRateLimit creates a governance proposal to add a rate limit.
func (s *RateLimitUpgradeSuite) addRateLimit(
	ctx context.Context,
	denom, channel string,
	sendPercent, recvPercent int64,
	durationHours uint64,
) string {
	authority, err := s.Chain.GetGovernanceAddress(ctx)
	s.Require().NoError(err)

	rateLimitMessage := fmt.Sprintf(`{
		"@type": "/ratelimit.v1.MsgAddRateLimit",
		"authority": "%s",
		"denom": "%s",
		"channel_or_client_id": "%s",
		"max_percent_send": "%d",
		"max_percent_recv": "%d",
		"duration_hours": "%d"
	}`, authority, denom, channel, sendPercent, recvPercent, durationHours)

	proposal := ProposalJSON{
		Messages:       []json.RawMessage{json.RawMessage(rateLimitMessage)},
		InitialDeposit: fmt.Sprintf("%duatom", chainsuite.GovMinDepositAmount),
		Title:          "Add Rate Limit for " + denom,
		Summary:        fmt.Sprintf("Add %d%% send, %d%% recv quota", sendPercent, recvPercent),
		Metadata:       "ipfs://CID",
	}

	proposalBytes, err := json.MarshalIndent(proposal, "", "  ")
	s.Require().NoError(err)

	err = s.Chain.GetNode().WriteFile(ctx, proposalBytes, "ratelimit-upgrade-proposal.json")
	s.Require().NoError(err)

	proposalPath := s.Chain.GetNode().HomeDir() + "/ratelimit-upgrade-proposal.json"

	_, err = s.Chain.GetNode().ExecTx(ctx, s.DelegatorWallet.FormattedAddress(),
		"gov", "submit-proposal", proposalPath)
	s.Require().NoError(err)

	lastProposal, err := s.Chain.QueryJSON(ctx, "proposals.@reverse.0.id", "gov", "proposals")
	s.Require().NoError(err)
	proposalID := lastProposal.String()

	err = s.Chain.PassProposal(ctx, proposalID)
	s.Require().NoError(err)

	s.Require().Eventually(func() bool {
		proposal, err := s.Chain.GovQueryProposalV1(ctx, mustParseUint(proposalID))
		if err != nil {
			return false
		}
		return proposal.Status == govv1.StatusPassed
	}, 30*chainsuite.CommitTimeout, chainsuite.CommitTimeout,
		"rate limit proposal did not pass")

	return proposalID
}

// ibcTransfer sends amount+denom from sender (on the `from` chain) to
// recipient over channelID (the channel ID as seen by `from`).
func (s *RateLimitUpgradeSuite) ibcTransfer(
	ctx context.Context,
	from *chainsuite.Chain,
	sender ibc.Wallet,
	recipient, amount, denom, channelID string,
) error {
	_, err := from.GetNode().ExecTx(ctx, sender.KeyName(),
		"ibc-transfer", "transfer", "transfer", channelID, recipient, amount+denom,
	)
	return err
}

// assertTransferWithinQuota sends percent% of the rate limit's total supply
// from DelegatorWallet to ChainB and asserts it succeeds.
func (s *RateLimitUpgradeSuite) assertTransferWithinQuota(ctx context.Context, percent int64, msgAndArgs ...interface{}) {
	amount := sdkmath.NewInt(s.RateLimitTotalSupply).MulRaw(percent).QuoRaw(100)
	err := s.ibcTransfer(ctx, s.Chain, s.DelegatorWallet, s.ChainBWallet.FormattedAddress(),
		amount.String(), s.RateLimitDenom, s.RateLimitChannelID)
	s.Require().NoError(err, msgAndArgs...)
}

// assertTransferBlockedByQuota sends percent% of the rate limit's total
// supply from DelegatorWallet to ChainB and asserts it's rejected for
// exceeding the quota, and that DelegatorWallet's balance is unaffected.
func (s *RateLimitUpgradeSuite) assertTransferBlockedByQuota(ctx context.Context, percent int64, msgAndArgs ...interface{}) {
	balanceBefore, err := s.Chain.GetBalance(ctx, s.DelegatorWallet.FormattedAddress(), s.RateLimitDenom)
	s.Require().NoError(err)

	amount := sdkmath.NewInt(s.RateLimitTotalSupply).MulRaw(percent).QuoRaw(100)
	err = s.ibcTransfer(ctx, s.Chain, s.DelegatorWallet, s.ChainBWallet.FormattedAddress(),
		amount.String(), s.RateLimitDenom, s.RateLimitChannelID)
	s.Require().Error(err, msgAndArgs...)
	s.Require().Contains(err.Error(), "quota", "error should mention quota/rate limit")

	balanceAfter, err := s.Chain.GetBalance(ctx, s.DelegatorWallet.FormattedAddress(), s.RateLimitDenom)
	s.Require().NoError(err)
	s.Require().Equal(balanceBefore, balanceAfter, "balance should be unchanged after blocked transfer")
}

// TestRateLimitPersistsAfterUpgrade verifies the rate limit's configuration
// (path and quota) is unchanged after the software upgrade.
func (s *RateLimitUpgradeSuite) TestRateLimitPersistsAfterUpgrade() {
	ctx := s.GetContext()

	// The CLI prints the RateLimit object itself (path/quota/flow at the top
	// level), not wrapped in the QueryRateLimitResponse's `rate_limit` field
	// (see ratelimit's client/cli/query.go: it calls PrintProto(res.RateLimit),
	// not PrintProto(res)) — so fetch the whole document with gjson's @this.
	result, err := s.Chain.QueryJSON(ctx, "@this",
		"ratelimit", "rate-limit", s.RateLimitChannelID, "--denom", s.RateLimitDenom)
	s.Require().NoError(err, "rate limit should still be queryable after upgrade")

	s.Require().Equal(s.RateLimitDenom, result.Get("path.denom").String())
	s.Require().Equal(s.RateLimitChannelID, result.Get("path.channel_or_client_id").String())
	s.Require().Equal(fmt.Sprintf("%d", s.RateLimitSendPercent), result.Get("quota.max_percent_send").String())
	s.Require().Equal(fmt.Sprintf("%d", s.RateLimitRecvPercent), result.Get("quota.max_percent_recv").String())
	s.Require().Equal(fmt.Sprintf("%d", s.RateLimitDurationHours), result.Get("quota.duration_hours").String())

	// flow.channel_value is a snapshot of the denom's total supply taken when
	// the rate limit was created, frozen for the duration of the quota window
	// (see keeper.ResetRateLimit / BeginBlocker) — it should survive the
	// upgrade unchanged too, regardless of any burns from transfers since.
	s.Require().Equal(fmt.Sprintf("%d", s.RateLimitTotalSupply), result.Get("flow.channel_value").String())
}

// TestRateLimitEnforcedAfterUpgrade verifies the rate limit still blocks
// over-quota transfers after the software upgrade. This builds on the
// cumulative flow already recorded by SetupSuite's pre-upgrade operational
// check (2% used, well under the 10% quota), confirming that flow state
// survived the upgrade too: 2%+5%=7% stays within quota, and a further 10%
// would push the cumulative total to 17%, over quota.
func (s *RateLimitUpgradeSuite) TestRateLimitEnforcedAfterUpgrade() {
	ctx := s.GetContext()
	s.assertTransferWithinQuota(ctx, 5, "in-quota transfer should still succeed after upgrade")
	s.assertTransferBlockedByQuota(ctx, 10, "over-quota transfer should still be blocked after upgrade")
}

func TestRateLimit(t *testing.T) {
	s := &RateLimitUpgradeSuite{
		Suite: &delegator.Suite{
			Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
				UpgradeOnSetup: false, // upgrade is triggered manually in SetupSuite, after pre-upgrade state exists
				CreateRelayer:  true,
			}),
		},
	}
	suite.Run(t, s)
}
