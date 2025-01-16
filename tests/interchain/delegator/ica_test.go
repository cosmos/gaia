package delegator_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v23/tests/interchain/chainsuite"
	"github.com/cosmos/gaia/v23/tests/interchain/delegator"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ICAControllerSuite struct {
	*delegator.Suite
	Host *chainsuite.Chain
}

func (s *ICAControllerSuite) SetupSuite() {
	s.Suite.SetupSuite()
	host, err := s.Chain.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, chainsuite.DefaultChainSpec(s.Env))
	s.Require().NoError(err)
	s.Host = host
}

func (s *ICAControllerSuite) TestNonGovICA() {
	const amountToSend = int64(3_300_000_000)

	srcAddress := s.DelegatorWallet.FormattedAddress()
	srcChannel, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.Host)
	s.Require().NoError(err)

	var icaAddress string
	s.Run("Register ICA account", func() {
		icaAddress, err = s.Chain.SetupICAAccount(s.GetContext(), s.Host, s.Relayer, srcAddress, 0, amountToSend)
		s.Require().NoError(err)

		_, err = s.Chain.SendIBCTransfer(s.GetContext(), srcChannel.ChannelID, srcAddress, ibc.WalletAmount{
			Address: icaAddress,
			Amount:  sdkmath.NewInt(amountToSend),
			Denom:   s.Chain.Config().Denom,
		}, ibc.TransferOptions{})
		s.Require().NoError(err)
	})

	s.Run("Generate and send ICA transaction", func() {
		wallets := s.Host.ValidatorWallets
		s.Require().NoError(err)
		dstAddress := wallets[0].Address

		var ibcStakeDenom string
		s.Require().EventuallyWithT(func(c *assert.CollectT) {
			balances, err := s.Host.BankQueryAllBalances(s.GetContext(), icaAddress)
			s.Require().NoError(err)
			s.Require().NotEmpty(balances)
			for _, c := range balances {
				if strings.Contains(c.Denom, "ibc") {
					ibcStakeDenom = c.Denom
					break
				}
			}
			assert.NotEmpty(c, ibcStakeDenom)
		}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)

		recipientBalanceBefore, err := s.Host.GetBalance(s.GetContext(), dstAddress, ibcStakeDenom)
		s.Require().NoError(err)

		icaAmount := int64(amountToSend / 3)
		srcConnection := srcChannel.ConnectionHops[0]

		s.Require().NoError(s.sendICATx(s.GetContext(), 0, srcAddress, dstAddress, icaAddress, srcConnection, icaAmount, ibcStakeDenom))
		s.Relayer.ClearTransferChannel(s.GetContext(), s.Chain, s.Host)

		s.Require().EventuallyWithT(func(c *assert.CollectT) {
			recipientBalanceAfter, err := s.Host.GetBalance(s.GetContext(), dstAddress, ibcStakeDenom)
			assert.NoError(c, err)

			assert.Equal(c, recipientBalanceBefore.Add(sdkmath.NewInt(icaAmount)), recipientBalanceAfter)
		}, 10*chainsuite.CommitTimeout, chainsuite.CommitTimeout)
	})
}

func TestDelegatorICA(t *testing.T) {
	s := &ICAControllerSuite{Suite: &delegator.Suite{Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
		UpgradeOnSetup: true,
		CreateRelayer:  true,
	})}}
	suite.Run(t, s)
}

func (s *ICAControllerSuite) sendICATx(ctx context.Context, valIdx int, srcAddress string, dstAddress string, icaAddress string, srcConnection string, amount int64, denom string) error {
	jsonBankSend := fmt.Sprintf(`{"@type": "/cosmos.bank.v1beta1.MsgSend", "from_address":"%s","to_address":"%s","amount":[{"denom":"%s","amount":"%d"}]}`, icaAddress, dstAddress, denom, amount)

	msgBz, _, err := s.Chain.GetNode().Exec(ctx, []string{"gaiad", "tx", "ica", "host", "generate-packet-data", string(jsonBankSend), "--encoding", "proto3"}, nil)
	if err != nil {
		return err
	}

	msgPath := "msg.json"
	if err := s.Chain.Validators[valIdx].WriteFile(ctx, msgBz, msgPath); err != nil {
		return err
	}
	msgPath = s.Chain.Validators[valIdx].HomeDir() + "/" + msgPath
	_, err = s.Chain.Validators[valIdx].ExecTx(ctx, srcAddress,
		"interchain-accounts", "controller", "send-tx",
		srcConnection, msgPath,
	)
	if err != nil {
		return err
	}
	return nil
}
