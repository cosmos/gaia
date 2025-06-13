package delegator_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v25/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v25/tests/interchain/delegator"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
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
	host, err := s.Chain.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, chainsuite.DefaultChainSpec(s.Env))
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
