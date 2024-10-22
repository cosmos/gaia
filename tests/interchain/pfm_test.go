package interchain_test

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/gaia/v20/tests/interchain/chainsuite"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PFMSuite struct {
	*chainsuite.Suite
	Chains []*chainsuite.Chain
}

func (s *PFMSuite) SetupSuite() {
	s.Suite.SetupSuite()
	chainB, err := s.Chain.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, chainsuite.DefaultChainSpec(s.Env))
	s.Require().NoError(err)
	chainC, err := chainB.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, chainsuite.DefaultChainSpec(s.Env))
	s.Require().NoError(err)
	chainD, err := chainC.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, chainsuite.DefaultChainSpec(s.Env))
	s.Require().NoError(err)

	s.Chains = []*chainsuite.Chain{s.Chain, chainB, chainC, chainD}
}

func (s *PFMSuite) TestPFMHappyPath() {
	var forwardChannels []*ibc.ChannelOutput
	targetDenomAD := s.Chains[0].Config().Denom
	for i := 0; i < len(s.Chains)-1; i++ {
		transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chains[i], s.Chains[i+1])
		s.Require().NoError(err)
		forwardChannels = append(forwardChannels, transferCh)
		targetDenomAD = transfertypes.GetPrefixedDenom(transferCh.PortID, transferCh.Counterparty.ChannelID, targetDenomAD)
	}
	targetDenomAD = transfertypes.ParseDenomTrace(targetDenomAD).IBCDenom()

	// backwardChannels[2] = chain3 -> chain2, backwardChannels[1] = chain2 -> chain1, backwardChannels[0] = chain1 -> chain0
	backwardChannels := make([]*ibc.ChannelOutput, len(forwardChannels))
	targetDenomDA := s.Chains[3].Config().Denom
	for i := len(s.Chains) - 1; i > 0; i-- {
		transferCh, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chains[i], s.Chains[i-1])
		s.Require().NoError(err)
		backwardChannels[i-1] = transferCh
		targetDenomDA = transfertypes.GetPrefixedDenom(transferCh.PortID, transferCh.Counterparty.ChannelID, targetDenomDA)
	}
	targetDenomDA = transfertypes.ParseDenomTrace(targetDenomDA).IBCDenom()

	dWallet1 := s.Chains[3].ValidatorWallets[0]

	aWallet1 := s.Chains[0].ValidatorWallets[0]

	dStartBalance, err := s.Chains[3].GetBalance(s.GetContext(), dWallet1.Address, targetDenomAD)
	s.Require().NoError(err)

	timeout := "10m"
	memo := map[string]interface{}{
		"forward": map[string]interface{}{
			"receiver": "pfm",
			"port":     "transfer",
			"channel":  forwardChannels[1].ChannelID,
			"timeout":  timeout,
			"next": map[string]interface{}{
				"forward": map[string]interface{}{
					"receiver": dWallet1.Address,
					"port":     "transfer",
					"channel":  forwardChannels[2].ChannelID,
					"timeout":  timeout,
				},
			},
		},
	}
	memoBytes, err := json.Marshal(memo)
	s.Require().NoError(err)
	_, err = s.Chains[0].GetNode().ExecTx(s.GetContext(), aWallet1.Address,
		"ibc-transfer", "transfer", "transfer", forwardChannels[0].ChannelID, "pfm", "1000000"+s.Chains[0].Config().Denom,
		"--memo", string(memoBytes))
	s.Require().NoError(err)

	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		dEndBalance, err := s.Chains[3].GetBalance(s.GetContext(), dWallet1.Address, targetDenomAD)
		assert.NoError(c, err)
		balances, err := s.Chains[3].BankQueryAllBalances(s.GetContext(), dWallet1.Address)
		assert.NoError(c, err)
		assert.Truef(c, dEndBalance.GT(dStartBalance), "expected %d > %d in %s; balances are: %+v",
			dEndBalance, dStartBalance, targetDenomAD, balances)
	}, 15*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "chain D balance has not increased")

	aStartBalance, err := s.Chains[0].GetBalance(s.GetContext(), aWallet1.Address, targetDenomDA)
	s.Require().NoError(err)

	memo = map[string]interface{}{
		"forward": map[string]interface{}{
			"receiver": "pfm",
			"port":     "transfer",
			"channel":  backwardChannels[1].ChannelID,
			"timeout":  timeout,
			"next": map[string]interface{}{
				"forward": map[string]interface{}{
					"receiver": aWallet1.Address,
					"port":     "transfer",
					"channel":  backwardChannels[0].ChannelID,
					"timeout":  timeout,
				},
			},
		},
	}
	memoBytes, err = json.Marshal(memo)
	s.Require().NoError(err)
	_, err = s.Chains[3].GetNode().ExecTx(s.GetContext(), dWallet1.Address,
		"ibc-transfer", "transfer", "transfer", backwardChannels[2].ChannelID, "pfm", "1000000"+s.Chains[3].Config().Denom,
		"--memo", string(memoBytes))
	s.Require().NoError(err)

	s.Require().EventuallyWithT(func(c *assert.CollectT) {
		aEndBalance, err := s.Chains[0].GetBalance(s.GetContext(), aWallet1.Address, targetDenomDA)
		assert.NoError(c, err)
		balances, err := s.Chains[0].BankQueryAllBalances(s.GetContext(), aWallet1.Address)
		assert.NoError(c, err)
		assert.Truef(c, aEndBalance.GT(aStartBalance), "expected %d > %d in %s; balances are: %+v",
			aEndBalance, aStartBalance, targetDenomDA, balances)
	}, 15*chainsuite.CommitTimeout, chainsuite.CommitTimeout, "chain A balance has not increased")

}

func TestPFM(t *testing.T) {
	s := &PFMSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			UpgradeOnSetup: true,
			CreateRelayer:  true,
		})}
	suite.Run(t, s)
}
