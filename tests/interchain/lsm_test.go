package interchain_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v20/tests/interchain/chainsuite"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sync/errgroup"
)

const (
	lsmBondingMoniker = "bonding"
	lsmLiquid1Moniker = "liquid_1"
	lsmLiquid2Moniker = "liquid_2"
	lsmLiquid3Moniker = "liquid_3"
	lsmOwnerMoniker   = "owner"
)

type LSMSuite struct {
	*chainsuite.Suite
	Stride     *chainsuite.Chain
	ICAAddr    string
	LSMWallets map[string]ibc.Wallet
}

func (s *LSMSuite) checkAMinusBEqualsX(a, b string, x sdkmath.Int) {
	intA, err := chainsuite.StrToSDKInt(a)
	s.Require().NoError(err)
	intB, err := chainsuite.StrToSDKInt(b)
	s.Require().NoError(err)
	s.Require().True(intA.Sub(intB).Equal(x), "a - b = %s, expected %s", intA.Sub(intB).String(), x.String())
}

func (s *LSMSuite) TestLSMHappyPath() {
	const (
		delegation    = 100000000
		tokenize      = 50000000
		bankSend      = 20000000
		ibcTransfer   = 10000000
		liquid1Redeem = 20000000
	)
	shareFactor, ok := sdkmath.NewIntFromString("1000000000000000000")
	s.Require().True(ok)
	providerWallet := s.Chain.ValidatorWallets[0]

	strideWallet := s.Stride.ValidatorWallets[0]

	s.Run("Validator Bond", func() {
		delegatorShares1, err := s.Chain.QueryJSON(s.GetContext(), "validator.delegator_shares", "staking", "validator", providerWallet.ValoperAddress)
		s.Require().NoError(err)
		validatorBondShares1, err := s.Chain.QueryJSON(s.GetContext(), "validator.validator_bond_shares", "staking", "validator", providerWallet.ValoperAddress)
		s.Require().NoError(err)

		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmBondingMoniker].FormattedAddress(),
			"staking", "delegate", providerWallet.ValoperAddress, fmt.Sprintf("%d%s", delegation, s.Chain.Config().Denom))
		s.Require().NoError(err)
		delegatorShares2, err := s.Chain.QueryJSON(s.GetContext(), "validator.delegator_shares", "staking", "validator", providerWallet.ValoperAddress)
		s.Require().NoError(err)
		s.checkAMinusBEqualsX(delegatorShares2.String(), delegatorShares1.String(), sdkmath.NewInt(delegation).Mul(shareFactor))

		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmBondingMoniker].FormattedAddress(),
			"staking", "validator-bond", providerWallet.ValoperAddress)
		s.Require().NoError(err)
		validatorBondShares2, err := s.Chain.QueryJSON(s.GetContext(), "validator.validator_bond_shares", "staking", "validator", providerWallet.ValoperAddress)
		s.Require().NoError(err)
		s.checkAMinusBEqualsX(validatorBondShares2.String(), validatorBondShares1.String(), sdkmath.NewInt(delegation).Mul(shareFactor))
	})

	var tokenizedDenom string
	s.Run("Tokenize", func() {
		delegatorShares1, err := s.Chain.QueryJSON(s.GetContext(), "validator.delegator_shares", "staking", "validator", providerWallet.ValoperAddress)
		s.Require().NoError(err)
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmLiquid1Moniker].FormattedAddress(),
			"staking", "delegate", providerWallet.ValoperAddress, fmt.Sprintf("%d%s", delegation, s.Chain.Config().Denom))
		s.Require().NoError(err)
		delegatorShares2, err := s.Chain.QueryJSON(s.GetContext(), "validator.delegator_shares", "staking", "validator", providerWallet.ValoperAddress)
		s.Require().NoError(err)
		s.checkAMinusBEqualsX(delegatorShares2.String(), delegatorShares1.String(), sdkmath.NewInt(delegation).Mul(shareFactor))

		sharesPreTokenize, err := s.Chain.QueryJSON(s.GetContext(), "validator.liquid_shares", "staking", "validator", providerWallet.ValoperAddress)
		s.Require().NoError(err)
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmLiquid1Moniker].FormattedAddress(),
			"staking", "tokenize-share",
			providerWallet.ValoperAddress, fmt.Sprintf("%d%s", tokenize, s.Chain.Config().Denom), s.LSMWallets[lsmLiquid1Moniker].FormattedAddress(),
			"--gas", "auto")
		s.Require().NoError(err)
		sharesPostTokenize, err := s.Chain.QueryJSON(s.GetContext(), "validator.liquid_shares", "staking", "validator", providerWallet.ValoperAddress)
		s.Require().NoError(err)
		s.checkAMinusBEqualsX(sharesPostTokenize.String(), sharesPreTokenize.String(), sdkmath.NewInt(tokenize).Mul(shareFactor))

		balances, err := s.Chain.BankQueryAllBalances(s.GetContext(), s.LSMWallets[lsmLiquid1Moniker].FormattedAddress())
		s.Require().NoError(err)
		for _, balance := range balances {
			if balance.Amount.Int64() == tokenize {
				tokenizedDenom = balance.Denom
			}
		}
		s.Require().NotEmpty(tokenizedDenom)
	})

	s.Run("Transfer Ownership", func() {
		recordIDResult, err := s.Chain.QueryJSON(s.GetContext(), "record.id", "staking", "tokenize-share-record-by-denom", tokenizedDenom)
		s.Require().NoError(err)
		recordID := recordIDResult.String()

		ownerResult, err := s.Chain.QueryJSON(s.GetContext(), "record.owner", "staking", "tokenize-share-record-by-denom", tokenizedDenom)
		s.Require().NoError(err)
		owner := ownerResult.String()

		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), owner,
			"staking", "transfer-tokenize-share-record", recordID, s.LSMWallets[lsmOwnerMoniker].FormattedAddress())
		s.Require().NoError(err)

		ownerResult, err = s.Chain.QueryJSON(s.GetContext(), "record.owner", "staking", "tokenize-share-record-by-denom", tokenizedDenom)
		s.Require().NoError(err)
		owner = ownerResult.String()
		s.Require().Equal(s.LSMWallets[lsmOwnerMoniker].FormattedAddress(), owner)

		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), owner,
			"staking", "transfer-tokenize-share-record", recordID, s.LSMWallets[lsmLiquid1Moniker].FormattedAddress())
		s.Require().NoError(err)

		ownerResult, err = s.Chain.QueryJSON(s.GetContext(), "record.owner", "staking", "tokenize-share-record-by-denom", tokenizedDenom)
		s.Require().NoError(err)
		owner = ownerResult.String()
		s.Require().Equal(s.LSMWallets[lsmLiquid1Moniker].FormattedAddress(), owner)
	})

	var happyLiquid1Delegations1 string
	var ibcDenom string

	ibcChannelProvider, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.Stride)
	s.Require().NoError(err)
	ibcChannelStride, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Stride, s.Chain)
	s.Require().NoError(err)

	s.Run("Transfer Tokens", func() {
		happyLiquid1Delegations1Result, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("delegation_responses.#(delegation.validator_address==\"%s\").delegation.shares", providerWallet.ValoperAddress), "staking", "delegations", s.LSMWallets[lsmLiquid1Moniker].FormattedAddress())
		s.Require().NoError(err)
		happyLiquid1Delegations1 = happyLiquid1Delegations1Result.String()

		err = s.Chain.SendFunds(s.GetContext(), s.LSMWallets[lsmLiquid1Moniker].FormattedAddress(), ibc.WalletAmount{
			Amount:  sdkmath.NewInt(bankSend),
			Denom:   tokenizedDenom,
			Address: s.LSMWallets[lsmLiquid2Moniker].FormattedAddress(),
		})
		s.Require().NoError(err)

		_, err = s.Chain.SendIBCTransfer(s.GetContext(), ibcChannelProvider.ChannelID, s.LSMWallets[lsmLiquid1Moniker].FormattedAddress(), ibc.WalletAmount{
			Amount:  sdkmath.NewInt(ibcTransfer),
			Denom:   tokenizedDenom,
			Address: strideWallet.Address,
		}, ibc.TransferOptions{})
		s.Require().NoError(err)
		s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 5, s.Stride))
		balances, err := s.Stride.BankQueryAllBalances(s.GetContext(), strideWallet.Address)
		s.Require().NoError(err)
		for _, balance := range balances {
			if balance.Amount.Int64() == ibcTransfer {
				ibcDenom = balance.Denom
			}
		}
		s.Require().NotEmpty(ibcDenom)
	})

	var happyLiquid1DelegationBalance string
	s.Run("Redeem Tokens", func() {
		_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmLiquid1Moniker].FormattedAddress(),
			"staking", "redeem-tokens", fmt.Sprintf("%d%s", liquid1Redeem, tokenizedDenom),
			"--gas", "auto")
		s.Require().NoError(err)

		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmLiquid2Moniker].FormattedAddress(),
			"staking", "redeem-tokens", fmt.Sprintf("%d%s", bankSend, tokenizedDenom),
			"--gas", "auto")
		s.Require().NoError(err)

		_, err = s.Stride.SendIBCTransfer(s.GetContext(), ibcChannelStride.ChannelID, strideWallet.Address, ibc.WalletAmount{
			Amount:  sdkmath.NewInt(ibcTransfer),
			Denom:   ibcDenom,
			Address: s.LSMWallets[lsmLiquid3Moniker].FormattedAddress(),
		}, ibc.TransferOptions{})
		s.Require().NoError(err)
		// wait for the transfer to be reflected
		s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 5, s.Chain))

		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmLiquid3Moniker].FormattedAddress(),
			"staking", "redeem-tokens", fmt.Sprintf("%d%s", ibcTransfer, tokenizedDenom),
			"--gas", "auto")
		s.Require().NoError(err)

		happyLiquid1Delegations2Result, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("delegation_responses.#(delegation.validator_address==\"%s\").delegation.shares", providerWallet.ValoperAddress), "staking", "delegations", s.LSMWallets[lsmLiquid1Moniker].FormattedAddress())
		s.Require().NoError(err)
		happyLiquid1Delegations2 := happyLiquid1Delegations2Result.String()
		s.checkAMinusBEqualsX(happyLiquid1Delegations2, happyLiquid1Delegations1, sdkmath.NewInt(liquid1Redeem).Mul(shareFactor))

		happyLiquid2DelegationsResult, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("delegation_responses.#(delegation.validator_address==\"%s\").delegation.shares", providerWallet.ValoperAddress), "staking", "delegations", s.LSMWallets[lsmLiquid2Moniker].FormattedAddress())
		s.Require().NoError(err)
		happyLiquid2Delegations := happyLiquid2DelegationsResult.String()
		// LOL there are better ways of doing this
		s.checkAMinusBEqualsX(happyLiquid2Delegations, "0", sdkmath.NewInt(bankSend).Mul(shareFactor))

		happyLiquid3DelegationsResult, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("delegation_responses.#(delegation.validator_address==\"%s\").delegation.shares", providerWallet.ValoperAddress), "staking", "delegations", s.LSMWallets[lsmLiquid3Moniker].FormattedAddress())
		s.Require().NoError(err)
		happyLiquid3Delegations := happyLiquid3DelegationsResult.String()
		s.checkAMinusBEqualsX(happyLiquid3Delegations, "0", sdkmath.NewInt(ibcTransfer).Mul(shareFactor))

		happyLiquid1DelegationBalanceResult, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("delegation_responses.#(delegation.validator_address==\"%s\").balance.amount", providerWallet.ValoperAddress), "staking", "delegations", s.LSMWallets[lsmLiquid1Moniker].FormattedAddress())
		s.Require().NoError(err)
		happyLiquid1DelegationBalance = happyLiquid1DelegationBalanceResult.String()

		happyLiquid2DelegationBalanceResult, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("delegation_responses.#(delegation.validator_address==\"%s\").balance.amount", providerWallet.ValoperAddress), "staking", "delegations", s.LSMWallets[lsmLiquid2Moniker].FormattedAddress())
		s.Require().NoError(err)
		happyLiquid2DelegationBalance := happyLiquid2DelegationBalanceResult.String()

		happyLiquid3DelegationBalanceResult, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("delegation_responses.#(delegation.validator_address==\"%s\").balance.amount", providerWallet.ValoperAddress), "staking", "delegations", s.LSMWallets[lsmLiquid3Moniker].FormattedAddress())
		s.Require().NoError(err)
		happyLiquid3DelegationBalance := happyLiquid3DelegationBalanceResult.String()

		s.checkAMinusBEqualsX(happyLiquid1DelegationBalance, "0", sdkmath.NewInt(70000000))
		s.checkAMinusBEqualsX(happyLiquid2DelegationBalance, "0", sdkmath.NewInt(bankSend))
		s.checkAMinusBEqualsX(happyLiquid3DelegationBalance, "0", sdkmath.NewInt(ibcTransfer))
	})
	s.Run("Cleanup", func() {
		_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmBondingMoniker].FormattedAddress(),
			"staking", "unbond", providerWallet.ValoperAddress, fmt.Sprintf("%d%s", delegation, s.Chain.Config().Denom))
		s.Require().NoError(err)

		validatorBondSharesResult, err := s.Chain.QueryJSON(s.GetContext(), "validator.validator_bond_shares", "staking", "validator", providerWallet.ValoperAddress)
		s.Require().NoError(err)
		validatorBondShares := validatorBondSharesResult.String()
		s.checkAMinusBEqualsX(validatorBondShares, "0", sdkmath.NewInt(0).Mul(shareFactor))

		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmLiquid1Moniker].FormattedAddress(),
			"staking", "unbond", providerWallet.ValoperAddress, fmt.Sprintf("%s%s", happyLiquid1DelegationBalance, s.Chain.Config().Denom))
		s.Require().NoError(err)
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmLiquid2Moniker].FormattedAddress(),
			"staking", "unbond", providerWallet.ValoperAddress, fmt.Sprintf("%d%s", bankSend, s.Chain.Config().Denom))
		s.Require().NoError(err)
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmLiquid3Moniker].FormattedAddress(),
			"staking", "unbond", providerWallet.ValoperAddress, fmt.Sprintf("%d%s", ibcTransfer, s.Chain.Config().Denom))
		s.Require().NoError(err)
	})
}

func (s *LSMSuite) TestICADelegate() {
	s.T().Skip("not implemented")
}

func (s *LSMSuite) setupLSMWallets() {
	names := []string{lsmBondingMoniker, lsmLiquid1Moniker, lsmLiquid2Moniker, lsmLiquid3Moniker, lsmOwnerMoniker}
	wallets := make(map[string]ibc.Wallet)
	eg := new(errgroup.Group)
	for _, name := range names {
		keyName := "happy_" + name
		wallet, err := s.Chain.BuildWallet(s.GetContext(), keyName, "")
		s.Require().NoError(err)
		wallets[name] = wallet
		amount := 500_000_000
		if name == "owner" {
			amount = 10_000_000
		}
		eg.Go(func() error {
			return s.Chain.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
				Amount:  sdkmath.NewInt(int64(amount)),
				Denom:   s.Chain.Config().Denom,
				Address: wallet.FormattedAddress(),
			})
		})
	}
	s.Require().NoError(eg.Wait())
	s.LSMWallets = wallets
}

func (s *LSMSuite) SetupSuite() {
	s.Suite.SetupSuite()
	stride, err := s.Chain.AddConsumerChain(s.GetContext(), s.Relayer, chainsuite.ConsumerConfig{
		ChainName: "stride",
		Version:   chainsuite.StrideVersion,
		Denom:     chainsuite.StrideDenom,
		TopN:      100,
	})
	s.Require().NoError(err)
	s.Stride = stride
	err = s.Chain.CheckCCV(s.GetContext(), stride, s.Relayer, 1_000_000, 0, 1)
	s.Require().NoError(err)
	s.setupLSMWallets()
	s.UpgradeChain()
}

func TestLSM(t *testing.T) {
	s := &LSMSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			CreateRelayer: true,
		}),
	}
	suite.Run(t, s)
}
