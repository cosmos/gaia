package delegator_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gaia/v25/tests/interchain/chainsuite"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
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
	LinkedChain *chainsuite.Chain
	LSMWallets  map[string]ibc.Wallet
	ShareFactor sdkmath.Int
}

type ProposalJSON struct {
	Messages       []json.RawMessage `json:"messages"`
	InitialDeposit string            `json:"deposit"`
	Title          string            `json:"title"`
	Summary        string            `json:"summary"`
	Metadata       string            `json:"metadata"`
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
	providerWallet := s.Chain.ValidatorWallets[0]
	ibcWallet := s.LinkedChain.ValidatorWallets[0]

	var tokenizedDenom string
	s.Run("Tokenize", func() {
		delegatorShares1, err := s.Chain.QueryJSON(s.GetContext(), "validator.delegator_shares", "staking", "validator", providerWallet.ValoperAddress)
		s.Require().NoError(err)
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmLiquid1Moniker].FormattedAddress(),
			"staking", "delegate", providerWallet.ValoperAddress, fmt.Sprintf("%d%s", delegation, s.Chain.Config().Denom))
		s.Require().NoError(err)
		delegatorShares2, err := s.Chain.QueryJSON(s.GetContext(), "validator.delegator_shares", "staking", "validator", providerWallet.ValoperAddress)
		s.Require().NoError(err)
		s.checkAMinusBEqualsX(delegatorShares2.String(), delegatorShares1.String(), sdkmath.NewInt(delegation))

		sharesPreTokenize, err := s.Chain.QueryJSON(s.GetContext(), "liquid_validator.liquid_shares", "liquid",
			"liquid-validator",
			providerWallet.ValoperAddress)
		s.Require().NoError(err)
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmLiquid1Moniker].FormattedAddress(),
			"liquid", "tokenize-share",
			providerWallet.ValoperAddress, fmt.Sprintf("%d%s", tokenize, s.Chain.Config().Denom), s.LSMWallets[lsmLiquid1Moniker].FormattedAddress(),
			"--gas", "auto")
		s.Require().NoError(err)
		sharesPostTokenize, err := s.Chain.QueryJSON(s.GetContext(), "liquid_validator.liquid_shares", "liquid",
			"liquid-validator", providerWallet.ValoperAddress)
		s.Require().NoError(err)
		s.checkAMinusBEqualsX(sharesPostTokenize.String(), sharesPreTokenize.String(), sdkmath.NewInt(tokenize).Mul(s.ShareFactor))

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
		recordIDResult, err := s.Chain.QueryJSON(s.GetContext(), "record.id", "liquid",
			"tokenize-share-record-by-denom", tokenizedDenom)
		s.Require().NoError(err)
		recordID := recordIDResult.String()

		ownerResult, err := s.Chain.QueryJSON(s.GetContext(), "record.owner", "liquid",
			"tokenize-share-record-by-denom", tokenizedDenom)
		s.Require().NoError(err)
		owner := ownerResult.String()

		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), owner,
			"liquid", "transfer-tokenize-share-record", recordID, s.LSMWallets[lsmOwnerMoniker].FormattedAddress())
		s.Require().NoError(err)

		ownerResult, err = s.Chain.QueryJSON(s.GetContext(), "record.owner", "liquid",
			"tokenize-share-record-by-denom", tokenizedDenom)
		s.Require().NoError(err)
		owner = ownerResult.String()
		s.Require().Equal(s.LSMWallets[lsmOwnerMoniker].FormattedAddress(), owner)

		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), owner,
			"liquid", "transfer-tokenize-share-record", recordID, s.LSMWallets[lsmLiquid1Moniker].FormattedAddress())
		s.Require().NoError(err)

		ownerResult, err = s.Chain.QueryJSON(s.GetContext(), "record.owner", "liquid",
			"tokenize-share-record-by-denom", tokenizedDenom)
		s.Require().NoError(err)
		owner = ownerResult.String()
		s.Require().Equal(s.LSMWallets[lsmLiquid1Moniker].FormattedAddress(), owner)
	})

	var happyLiquid1Delegations1 string
	var ibcDenom string

	ibcChannelProvider, err := s.Relayer.GetTransferChannel(s.GetContext(), s.Chain, s.LinkedChain)
	s.Require().NoError(err)
	ibcChannel, err := s.Relayer.GetTransferChannel(s.GetContext(), s.LinkedChain, s.Chain)
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
			Address: ibcWallet.Address,
		}, ibc.TransferOptions{})
		s.Require().NoError(err)
		s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 5, s.LinkedChain))
		balances, err := s.LinkedChain.BankQueryAllBalances(s.GetContext(), ibcWallet.Address)
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
			"liquid", "redeem-tokens", fmt.Sprintf("%d%s", liquid1Redeem, tokenizedDenom),
			"--gas", "auto")
		s.Require().NoError(err)

		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmLiquid2Moniker].FormattedAddress(),
			"liquid", "redeem-tokens", fmt.Sprintf("%d%s", bankSend, tokenizedDenom),
			"--gas", "auto")
		s.Require().NoError(err)

		_, err = s.LinkedChain.SendIBCTransfer(s.GetContext(), ibcChannel.ChannelID, ibcWallet.Address, ibc.WalletAmount{
			Amount:  sdkmath.NewInt(ibcTransfer),
			Denom:   ibcDenom,
			Address: s.LSMWallets[lsmLiquid3Moniker].FormattedAddress(),
		}, ibc.TransferOptions{})
		s.Require().NoError(err)
		// wait for the transfer to be reflected
		s.Require().NoError(testutil.WaitForBlocks(s.GetContext(), 5, s.Chain))

		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.LSMWallets[lsmLiquid3Moniker].FormattedAddress(),
			"liquid", "redeem-tokens", fmt.Sprintf("%d%s", ibcTransfer, tokenizedDenom),
			"--gas", "auto")
		s.Require().NoError(err)

		happyLiquid1Delegations2Result, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("delegation_responses.#(delegation.validator_address==\"%s\").delegation.shares", providerWallet.ValoperAddress), "staking", "delegations", s.LSMWallets[lsmLiquid1Moniker].FormattedAddress())
		s.Require().NoError(err)
		happyLiquid1Delegations2 := happyLiquid1Delegations2Result.String()
		s.checkAMinusBEqualsX(happyLiquid1Delegations2, happyLiquid1Delegations1, sdkmath.NewInt(liquid1Redeem))

		happyLiquid2DelegationsResult, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("delegation_responses.#(delegation.validator_address==\"%s\").delegation.shares", providerWallet.ValoperAddress), "staking", "delegations", s.LSMWallets[lsmLiquid2Moniker].FormattedAddress())
		s.Require().NoError(err)
		happyLiquid2Delegations := happyLiquid2DelegationsResult.String()
		// LOL there are better ways of doing this
		s.checkAMinusBEqualsX(happyLiquid2Delegations, "0", sdkmath.NewInt(bankSend))

		happyLiquid3DelegationsResult, err := s.Chain.QueryJSON(s.GetContext(), fmt.Sprintf("delegation_responses.#(delegation.validator_address==\"%s\").delegation.shares", providerWallet.ValoperAddress), "staking", "delegations", s.LSMWallets[lsmLiquid3Moniker].FormattedAddress())
		s.Require().NoError(err)
		happyLiquid3Delegations := happyLiquid3DelegationsResult.String()
		s.checkAMinusBEqualsX(happyLiquid3Delegations, "0", sdkmath.NewInt(ibcTransfer))

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
}

func (s *LSMSuite) TestTokenizeVested() {
	const amount = 100_000_000_000
	const vestingPeriod = 100 * time.Second
	vestedByTimestamp := time.Now().Add(vestingPeriod).Unix()
	vestingAccount, err := s.Chain.BuildWallet(s.GetContext(), fmt.Sprintf("vesting-%d", vestedByTimestamp), "")
	s.Require().NoError(err)
	validatorWallet := s.Chain.ValidatorWallets[0]

	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), interchaintest.FaucetAccountKeyName,
		"vesting", "create-vesting-account", vestingAccount.FormattedAddress(),
		fmt.Sprintf("%d%s", amount, s.Chain.Config().Denom),
		fmt.Sprintf("%d", vestedByTimestamp))
	s.Require().NoError(err)

	// give the vesting account a little cash for gas fees
	err = s.Chain.SendFunds(s.GetContext(), interchaintest.FaucetAccountKeyName, ibc.WalletAmount{
		Amount:  sdkmath.NewInt(5_000),
		Denom:   s.Chain.Config().Denom,
		Address: vestingAccount.FormattedAddress(),
	})
	s.Require().NoError(err)

	vestingAmount := int64(amount - 5000)
	// delegate the vesting account to the validator
	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), vestingAccount.FormattedAddress(),
		"staking", "delegate", validatorWallet.ValoperAddress, fmt.Sprintf("%d%s", vestingAmount, s.Chain.Config().Denom))
	s.Require().NoError(err)

	// wait for half the vesting period
	time.Sleep(vestingPeriod / 2)

	// try to tokenize full amount. Should fail.
	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), vestingAccount.FormattedAddress(),
		"liquid", "tokenize-share", validatorWallet.ValoperAddress, fmt.Sprintf("%d%s", vestingAmount,
			s.Chain.Config().Denom), vestingAccount.FormattedAddress(),
		"--gas", "auto")
	s.Require().Error(err)

	sharesPreTokenizeResult, err := s.Chain.QueryJSON(s.GetContext(), "liquid_validator.liquid_shares", "liquid",
		"liquid-validator", validatorWallet.ValoperAddress)
	s.Require().NoError(err)
	sharesPreTokenize := sharesPreTokenizeResult.String()

	// try to tokenize vested amount (i.e. half) should succeed if upgraded
	tokenizeAmount := vestingAmount / 2
	_, err = s.Chain.GetNode().ExecTx(s.GetContext(), vestingAccount.FormattedAddress(),
		"liquid", "tokenize-share", validatorWallet.ValoperAddress, fmt.Sprintf("%d%s", tokenizeAmount,
			s.Chain.Config().Denom), vestingAccount.FormattedAddress(),
		"--gas", "auto")
	s.Require().NoError(err)
	sharesPostTokenizeResult, err := s.Chain.QueryJSON(s.GetContext(), "liquid_validator.liquid_shares", "liquid",
		"liquid-validator", validatorWallet.ValoperAddress)
	s.Require().NoError(err)
	sharesPostTokenize := sharesPostTokenizeResult.String()
	s.checkAMinusBEqualsX(sharesPostTokenize, sharesPreTokenize, sdkmath.NewInt(tokenizeAmount).Mul(s.ShareFactor))
}

func (s *LSMSuite) TestLSMParams() {
	const (
		delegation   = 100000000
		globalCap    = 0.1
		validatorCap = 0.05
	)

	providerWallet := s.Chain.ValidatorWallets[0]

	liquidParams, err := s.Chain.QueryJSON(s.GetContext(), "params", "liquid", "params")
	s.Require().NoError(err)

	startingValidatorCap := gjson.Get(liquidParams.String(), "validator_liquid_staking_cap").String()
	startingGlobalCap := gjson.Get(liquidParams.String(), "global_liquid_staking_cap").String()
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Starting validator liquid staking cap: %s", startingValidatorCap)
	chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Starting global liquid staking cap: %s", startingGlobalCap)

	s.Run("Update params", func() {
		authority, err := s.Chain.GetGovernanceAddress(s.GetContext())
		s.Require().NoError(err)

		updatedParams, err := sjson.Set(liquidParams.String(), "global_liquid_staking_cap", fmt.Sprintf("%f", globalCap))
		s.Require().NoError(err)
		updatedParams, err = sjson.Set(updatedParams, "validator_liquid_staking_cap", fmt.Sprintf("%f", validatorCap))
		s.Require().NoError(err)
		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Updated params: %s", updatedParams)

		paramChangeMessage := fmt.Sprintf(`{
		"@type": "/gaia.liquid.v1beta1.MsgUpdateParams",
		"authority": "%s",
		"params": %s
	}`, authority, updatedParams)

		proposal := ProposalJSON{
			Messages:       []json.RawMessage{json.RawMessage(paramChangeMessage)},
			InitialDeposit: "5000000uatom",
			Title:          "Liquid Param Change Proposal",
			Summary:        "Test Proposal",
			Metadata:       "ipfs://CID",
		}

		// Marshal to JSON
		proposalBytes, err := json.MarshalIndent(proposal, "", "  ")
		s.Require().NoError(err)

		// Get the home directory of the node
		homeDir := s.Chain.GetNode().HomeDir()
		proposalPath := homeDir + "/params-proposal.json"
		// Write to file
		err = s.Chain.GetNode().WriteFile(s.GetContext(), proposalBytes, "params-proposal.json")
		s.Require().NoError(err)

		// out, _, _ := s.Chain.GetNode().Exec(s.GetContext(), []string{"ls", "-l", string(homeDir)}, []string{homeDir})
		// fmt.Println("Files in home dir:", string(out))
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), providerWallet.Address,
			"gov", "submit-proposal", proposalPath)
		s.Require().NoError(err)

		lastProposalId, err := s.Chain.QueryJSON(s.GetContext(), "proposals.@reverse.0.id", "gov", "proposals")
		s.Require().NoError(err)

		// Pass proposal
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), providerWallet.Address,
			"gov", "vote", lastProposalId.String(), "yes")
		s.Require().NoError(err)
		time.Sleep(80 * time.Second)

		// Check proposal
		proposalStatus, err := s.Chain.QueryJSON(s.GetContext(), "proposal", "gov", "proposal", lastProposalId.String())
		s.Require().NoError(err)
		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Proposal status: %s", proposalStatus)
		// Check params
		liquidParams, err = s.Chain.QueryJSON(s.GetContext(), "params", "liquid", "params")
		s.Require().NoError(err)
		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Current params: %s", liquidParams)
		validatorLiquidCapParam := gjson.Get(liquidParams.String(), "validator_liquid_staking_cap")
		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Validator liquid cap: %s", validatorLiquidCapParam)
		globalLiquidCapParam := gjson.Get(liquidParams.String(), "global_liquid_staking_cap")
		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Global liquid cap: %s", globalLiquidCapParam)
		validatorLiquidCapFloat := validatorLiquidCapParam.Float()
		globalLiquidCapFloat := globalLiquidCapParam.Float()
		s.Require().Equal(validatorCap, validatorLiquidCapFloat)
		s.Require().Equal(globalCap, globalLiquidCapFloat)
	})

	s.Run("Test liquid caps", func() {
		_, err := s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Address, "staking", "delegate", s.Chain.ValidatorWallets[1].ValoperAddress, fmt.Sprintf("%d%s", delegation, chainsuite.Uatom))
		s.Require().NoError(err)
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Address, "staking", "delegate", s.Chain.ValidatorWallets[2].ValoperAddress, fmt.Sprintf("%d%s", delegation, chainsuite.Uatom))
		s.Require().NoError(err)

		liquidParams, err := s.Chain.QueryJSON(s.GetContext(), "params", "liquid", "params")
		s.Require().NoError(err)
		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Starting params: %s", liquidParams)
		validatorLiquidCapParam := gjson.Get(liquidParams.String(), "validator_liquid_staking_cap")
		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Validator liquid cap: %s", validatorLiquidCapParam)
		globalLiquidCapParam := gjson.Get(liquidParams.String(), "global_liquid_staking_cap")
		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Global liquid cap: %s", globalLiquidCapParam)

		val0Tokens, err := s.Chain.QueryJSON(s.GetContext(), "validator.tokens", "staking", "validator", s.Chain.ValidatorWallets[0].ValoperAddress)
		s.Require().NoError(err)
		val1Tokens, err := s.Chain.QueryJSON(s.GetContext(), "validator.tokens", "staking", "validator", s.Chain.ValidatorWallets[1].ValoperAddress)
		s.Require().NoError(err)
		val2Tokens, err := s.Chain.QueryJSON(s.GetContext(), "validator.tokens", "staking", "validator", s.Chain.ValidatorWallets[2].ValoperAddress)
		s.Require().NoError(err)
		totalBondedTokens, err := s.Chain.QueryJSON(s.GetContext(), "pool.bonded_tokens", "staking", "pool")
		s.Require().NoError(err)
		totalLiquidStaked, err := s.Chain.QueryJSON(s.GetContext(), "tokens", "liquid", "total-liquid-staked")
		s.Require().NoError(err)
		val0TokensFloat := val0Tokens.Float()
		val1TokensFloat := val1Tokens.Float()
		val2TokensFloat := val2Tokens.Float()
		totalBondedTokensFloat := totalBondedTokens.Float()
		totalLiquidStakedFloat := totalLiquidStaked.Float()

		validatorCap := validatorLiquidCapParam.Float()
		globalCap := globalLiquidCapParam.Float()

		val0Cap := validatorCap * val0TokensFloat
		val1Cap := validatorCap * val1TokensFloat
		val2Cap := validatorCap * val2TokensFloat
		globalCapShares := globalCap * totalBondedTokensFloat

		val0LiquidShares, err := s.Chain.QueryJSON(s.GetContext(), "liquid_validator.liquid_shares", "liquid", "liquid-validator", s.Chain.ValidatorWallets[0].ValoperAddress)
		s.Require().NoError(err)
		val1LiquidShares, err := s.Chain.QueryJSON(s.GetContext(), "liquid_validator.liquid_shares", "liquid", "liquid-validator", s.Chain.ValidatorWallets[1].ValoperAddress)
		s.Require().NoError(err)
		val2LiquidShares, err := s.Chain.QueryJSON(s.GetContext(), "liquid_validator.liquid_shares", "liquid", "liquid-validator", s.Chain.ValidatorWallets[2].ValoperAddress)
		s.Require().NoError(err)
		val0LiquidSharesFloat := val0LiquidShares.Float()
		val1LiquidSharesFloat := val1LiquidShares.Float()
		val2LiquidSharesFloat := val2LiquidShares.Float()

		val0CapAvailable := val0Cap - val0LiquidSharesFloat
		val1CapAvailable := val1Cap - val1LiquidSharesFloat
		val2CapAvailable := val2Cap - val2LiquidSharesFloat
		globalCapSharesAvailable := globalCapShares - totalLiquidStakedFloat

		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("val0 cap: %f, available shares: %f", val0Cap, val0CapAvailable)
		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("val1 cap: %f, available shares: %f", val1Cap, val1CapAvailable)
		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Val2 cap: %f, available shares: %f", val2Cap, val2CapAvailable)
		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Global cap: %f, available shares: %f", globalCapShares, globalCapSharesAvailable)

		// Validator 1 must have a lower cap than the global amount
		s.Require().Less(val1CapAvailable, globalCapSharesAvailable)

		// Try to tokenize more than the global cap
		globalFailAmount := int(globalCapSharesAvailable + 1000000)
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Address, "liquid", "tokenize-share",
			s.Chain.ValidatorWallets[1].ValoperAddress, fmt.Sprintf("%d%s", globalFailAmount, s.Chain.Config().Denom), s.Chain.ValidatorWallets[0].Address)
		s.Require().Error(err)

		// Try to tokenize more than the validator cap
		validatorFailAmount := int(val1CapAvailable + 1000000)
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Address, "liquid", "tokenize-share",
			s.Chain.ValidatorWallets[1].ValoperAddress, fmt.Sprintf("%d%s", validatorFailAmount, s.Chain.Config().Denom), s.Chain.ValidatorWallets[0].Address)
		s.Require().Error(err)

		// Tokenize less than the validator cap
		validatorSuccessAmount := int(val1CapAvailable - 1000000)
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Address, "liquid", "tokenize-share",
			s.Chain.ValidatorWallets[1].ValoperAddress, fmt.Sprintf("%d%s", validatorSuccessAmount, s.Chain.Config().Denom), s.Chain.ValidatorWallets[0].Address)
		s.Require().NoError(err)

		tokenizedDenom, err := s.Chain.QueryJSON(s.GetContext(), "balances.@reverse.1.denom", "bank", "balances", s.Chain.ValidatorWallets[0].Address)
		s.Require().NoError(err)
		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Tokenized denom: %s", tokenizedDenom)
		// Redeem tokenized amount
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Address, "liquid", "redeem-tokens",
			fmt.Sprintf("%d%s", validatorSuccessAmount, tokenizedDenom))
		s.Require().NoError(err)

		// Unbond
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Address, "staking", "unbond", s.Chain.ValidatorWallets[1].ValoperAddress, fmt.Sprintf("%d%s", delegation, chainsuite.Uatom))
		s.Require().NoError(err)
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), s.Chain.ValidatorWallets[0].Address, "staking", "unbond", s.Chain.ValidatorWallets[2].ValoperAddress, fmt.Sprintf("%d%s", delegation, chainsuite.Uatom))
		s.Require().NoError(err)
	})

	s.Run("Restore params", func() {
		authority, err := s.Chain.GetGovernanceAddress(s.GetContext())
		s.Require().NoError(err)

		updatedParams, err := sjson.Set(liquidParams.String(), "global_liquid_staking_cap", startingGlobalCap)
		s.Require().NoError(err)
		updatedParams, err = sjson.Set(updatedParams, "validator_liquid_staking_cap", startingValidatorCap)
		s.Require().NoError(err)
		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Updated params: %s", updatedParams)

		paramChangeMessage := fmt.Sprintf(`{
				"@type": "/gaia.liquid.v1beta1.MsgUpdateParams",
				"authority": "%s",
				"params": %s
			}`, authority, updatedParams)

		proposal := ProposalJSON{
			Messages:       []json.RawMessage{json.RawMessage(paramChangeMessage)},
			InitialDeposit: "5000000uatom",
			Title:          "Liquid Param Restore Proposal",
			Summary:        "Test Proposal",
			Metadata:       "ipfs://CID",
		}

		// Marshal to JSON
		proposalBytes, err := json.MarshalIndent(proposal, "", "  ")
		s.Require().NoError(err)

		// Get the home directory of the node
		homeDir := s.Chain.GetNode().HomeDir()
		proposalPath := homeDir + "/restore-params-proposal.json"
		// Write to file
		err = s.Chain.GetNode().WriteFile(s.GetContext(), proposalBytes, "restore-params-proposal.json")
		s.Require().NoError(err)

		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), providerWallet.Address,
			"gov", "submit-proposal", proposalPath)
		s.Require().NoError(err)

		lastProposalId, err := s.Chain.QueryJSON(s.GetContext(), "proposals.@reverse.0.id", "gov", "proposals")
		s.Require().NoError(err)
		chainsuite.GetLogger(s.GetContext()).Sugar().Infof("Last Proposal ID: %s", lastProposalId)

		// Pass proposal
		_, err = s.Chain.GetNode().ExecTx(s.GetContext(), providerWallet.Address,
			"gov", "vote", lastProposalId.String(), "yes")
		s.Require().NoError(err)
		time.Sleep(80 * time.Second)

		// Check params
		liquidParams, err = s.Chain.QueryJSON(s.GetContext(), "params", "liquid", "params")
		s.Require().NoError(err)
		validatorLiquidCapParam := gjson.Get(liquidParams.String(), "validator_liquid_staking_cap").Float()
		globalLiquidCapParam := gjson.Get(liquidParams.String(), "global_liquid_staking_cap").Float()
		startingValidatorCapFloat, err := strconv.ParseFloat(startingValidatorCap, 64)
		s.Require().NoError(err)
		startingGlobalCapFloat, err := strconv.ParseFloat(startingGlobalCap, 64)
		s.Require().NoError(err)
		s.Require().Equal(startingValidatorCapFloat, validatorLiquidCapParam)
		s.Require().Equal(startingGlobalCapFloat, globalLiquidCapParam)
	})
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
	secondChain, err := s.Chain.AddLinkedChain(s.GetContext(), s.T(), s.Relayer, chainsuite.DefaultChainSpec(s.Env))
	s.Require().NoError(err)
	s.LinkedChain = secondChain

	shareFactor, ok := sdkmath.NewIntFromString("1000000000000000000")
	s.Require().True(ok)
	s.ShareFactor = shareFactor

	s.setupLSMWallets()
}

func TestLSM(t *testing.T) {
	s := &LSMSuite{
		Suite: chainsuite.NewSuite(chainsuite.SuiteConfig{
			CreateRelayer:  true,
			UpgradeOnSetup: true,
			ChainSpec: &interchaintest.ChainSpec{
				NumValidators: &chainsuite.SixValidators,
			},
		}),
	}
	suite.Run(t, s)
}
