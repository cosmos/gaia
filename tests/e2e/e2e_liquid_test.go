package e2e

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/cosmos/gaia/v27/tests/e2e/common"
	"github.com/cosmos/gaia/v27/tests/e2e/msg"
	"github.com/cosmos/gaia/v27/tests/e2e/query"
	liquidtypes "github.com/cosmos/gaia/v27/x/liquid/types"
)

// underBuffer is the number of tokens under the limit to tokenize
var underBuffer = math.NewInt(5_000_000)

func (s *IntegrationTestSuite) testLiquid() {
	chainEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))

	validatorA := s.Resources.ChainA.Validators[0]
	validatorAAddr, _ := validatorA.KeyInfo.GetAddress()

	validatorAddressA := sdk.ValAddress(validatorAAddr).String()

	err := msg.WriteLiquidStakingParamsUpdateProposal(s.Resources.ChainA, 25, 50)
	s.Require().NoError(err)
	s.TestCounters.ProposalCounter++
	submitGovFlags := []string{configFile(common.ProposalLiquidParamUpdateFilename)}
	depositGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes"}

	// gov proposing Liquid parameters (global liquid staking cap, validator liquid staking cap, validator bond factor)
	s.T().Logf("Proposal number: %d", s.TestCounters.ProposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: Set parameters (global liquid staking cap, validator liquid staking cap, validator bond factor)")
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.TestCounters.ProposalCounter, "liquidtypes.MsgUpdateProposal",
		submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	// query the proposal status and new fee
	s.Require().Eventually(
		func() bool {
			proposal, err := query.GovProposal(chainEndpoint, s.TestCounters.ProposalCounter)
			s.Require().NoError(err)
			return proposal.GetProposal().Status == govv1beta1.StatusPassed
		},
		15*time.Second,
		5*time.Second,
	)

	s.Require().Eventually(
		func() bool {
			liquidParams, err := query.LiquidParams(chainEndpoint)
			s.T().Logf("After Liquid parameters update proposal")
			s.Require().NoError(err)

			s.Require().Equal(liquidParams.Params.GlobalLiquidStakingCap, math.LegacyNewDecWithPrec(25, 2))
			s.Require().Equal(liquidParams.Params.ValidatorLiquidStakingCap, math.LegacyNewDecWithPrec(50, 2))

			return true
		},
		15*time.Second,
		5*time.Second,
	)
	delegatorAddress, _ := s.Resources.ChainA.GenesisAccounts[2].KeyInfo.GetAddress()

	fees := sdk.NewCoin(common.UAtomDenom, math.NewInt(1))

	delegationAmount := math.NewInt(500000000)
	delegation := sdk.NewCoin(common.UAtomDenom, delegationAmount) // 500 atom

	// Alice delegate uatom to Validator A
	s.ExecDelegate(s.Resources.ChainA, 0, delegation.String(), validatorAddressA, delegatorAddress.String(), common.GaiaHomePath, fees.String())

	// Validate delegation successful
	s.Require().Eventually(
		func() bool {
			res, err := query.Delegation(chainEndpoint, validatorAddressA, delegatorAddress.String())
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(math.LegacyNewDecFromInt(delegationAmount))
		},
		20*time.Second,
		5*time.Second,
	)

	// Tokenize shares
	tokenizeAmount := math.NewInt(200000000)
	tokenize := sdk.NewCoin(common.UAtomDenom, tokenizeAmount) // 200 atom
	s.executeTokenizeShares(s.Resources.ChainA, 0, tokenize.String(), validatorAddressA, delegatorAddress.String(), common.GaiaHomePath, fees.String())

	// Validate delegation reduced
	s.Require().Eventually(
		func() bool {
			res, err := query.Delegation(chainEndpoint, validatorAddressA, delegatorAddress.String())
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(math.LegacyNewDecFromInt(delegationAmount.Sub(tokenizeAmount)))
		},
		20*time.Second,
		5*time.Second,
	)

	// Validate balance increased
	recordID := int(1)
	shareDenom := fmt.Sprintf("%s/%s", strings.ToLower(validatorAddressA), strconv.Itoa(recordID))
	s.Require().Eventually(
		func() bool {
			res, err := query.SpecificBalance(chainEndpoint, delegatorAddress.String(), shareDenom)
			s.Require().NoError(err)
			return res.Amount.Equal(tokenizeAmount)
		},
		20*time.Second,
		5*time.Second,
	)

	// Validate liquid shares increased
	s.Require().Eventually(
		func() bool {
			res, err := query.LiquidValidator(chainEndpoint, validatorAddressA)
			s.Require().NoError(err)
			return res.LiquidValidator.LiquidShares.TruncateInt().Equal(tokenizeAmount)
		},
		20*time.Second,
		5*time.Second,
	)

	// Bank send Liquid token
	sendAmount := sdk.NewCoin(shareDenom, tokenizeAmount)
	s.ExecBankSend(s.Resources.ChainA, 0, delegatorAddress.String(), validatorAAddr.String(), sendAmount.String(), common.StandardFees.String(), false)

	// Validate tokens are sent properly
	s.Require().Eventually(
		func() bool {
			afterSenderShareDenomBalance, err := query.SpecificBalance(chainEndpoint, delegatorAddress.String(), shareDenom)
			s.Require().NoError(err)

			afterRecipientShareDenomBalance, err := query.SpecificBalance(chainEndpoint, validatorAAddr.String(), shareDenom)
			s.Require().NoError(err)

			decremented := afterSenderShareDenomBalance.IsNil() || afterSenderShareDenomBalance.IsZero()
			incremented := afterRecipientShareDenomBalance.IsEqual(sendAmount)

			return decremented && incremented
		},
		time.Minute,
		5*time.Second,
	)

	// transfer reward ownership
	s.executeTransferTokenizeShareRecord(s.Resources.ChainA, 0, strconv.Itoa(recordID), delegatorAddress.String(), validatorAAddr.String(), common.GaiaHomePath, common.StandardFees.String())
	tokenizeShareRecord := liquidtypes.TokenizeShareRecord{}
	// Validate ownership transferred correctly
	s.Require().Eventually(
		func() bool {
			record, err := query.TokenizeShareRecordByID(chainEndpoint, recordID)
			s.Require().NoError(err)
			tokenizeShareRecord = record
			return record.Owner == validatorAAddr.String()
		},
		time.Minute,
		5*time.Second,
	)
	_ = tokenizeShareRecord

	// IBC transfer Liquid token
	ibcTransferAmount := sdk.NewCoin(shareDenom, math.NewInt(100000000))
	sendRecipientAddr, _ := s.Resources.ChainA.Validators[0].KeyInfo.GetAddress()
	s.SendIBC(s.Resources.ChainA, 0, validatorAAddr.String(), sendRecipientAddr.String(), ibcTransferAmount.String(), common.StandardFees.String(), "memo", common.TransferChannel, nil, false)

	s.Require().Eventually(
		func() bool {
			afterSenderShareBalance, err := query.SpecificBalance(chainEndpoint, validatorAAddr.String(), shareDenom)
			s.Require().NoError(err)

			decremented := afterSenderShareBalance.Add(ibcTransferAmount).IsEqual(sendAmount)
			return decremented
		},
		1*time.Minute,
		5*time.Second,
	)

	// Redeem tokens for shares
	redeemAmount := sendAmount.Sub(ibcTransferAmount)
	s.executeRedeemShares(s.Resources.ChainA, 0, redeemAmount.String(), validatorAAddr.String(), common.GaiaHomePath, fees.String())

	// check redeem success
	s.Require().Eventually(
		func() bool {
			balanceRes, err := query.SpecificBalance(chainEndpoint, validatorAAddr.String(), shareDenom)
			s.Require().NoError(err)
			if !balanceRes.Amount.IsNil() && balanceRes.Amount.IsZero() {
				return false
			}

			// check that tokenize share record module account received some rewards, since it unbonded during redeem tx execution
			balanceRes, err = query.SpecificBalance(chainEndpoint, tokenizeShareRecord.GetModuleAddress().String(), common.UAtomDenom)
			s.Require().NoError(err)
			if balanceRes.Amount.IsNil() || balanceRes.Amount.IsZero() {
				return false
			}
			return true
		},
		20*time.Second,
		5*time.Second,
	)

	// Validate liquid shares decreased
	s.Require().Eventually(
		func() bool {
			res, err := query.LiquidValidator(chainEndpoint, validatorAddressA)
			s.Require().NoError(err)
			return res.LiquidValidator.LiquidShares.TruncateInt().Equal(tokenizeAmount.Sub(redeemAmount.Amount))
		},
		20*time.Second,
		5*time.Second,
	)
}

// testLiquidGlobalLimit validates the global liquid staking cap
func (s *IntegrationTestSuite) testLiquidGlobalLimit() {
	chainEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))

	validatorA := s.Resources.ChainA.Validators[0]
	validatorAAddr, _ := validatorA.KeyInfo.GetAddress()
	validatorB := s.Resources.ChainA.Validators[1]
	validatorBAddr, _ := validatorB.KeyInfo.GetAddress()

	validatorAddressB := sdk.ValAddress(validatorBAddr).String()

	err := msg.WriteLiquidStakingParamsUpdateProposal(s.Resources.ChainA, 25, 100)
	s.Require().NoError(err)
	s.TestCounters.ProposalCounter++
	submitGovFlags := []string{configFile(common.ProposalLiquidParamUpdateFilename)}
	depositGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes"}

	// gov proposing Liquid parameters (global liquid staking cap, validator liquid staking cap, validator bond factor)
	s.T().Logf("Proposal number: %d", s.TestCounters.ProposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: Set parameters (global liquid staking cap, validator liquid staking cap)")
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.TestCounters.ProposalCounter, "liquidtypes.MsgUpdateProposal",
		submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	// query the proposal status and new fee
	s.Require().Eventually(
		func() bool {
			proposal, err := query.GovProposal(chainEndpoint, s.TestCounters.ProposalCounter)
			s.Require().NoError(err)
			return proposal.GetProposal().Status == govv1beta1.StatusPassed
		},
		15*time.Second,
		5*time.Second,
	)

	s.Require().Eventually(
		func() bool {
			liquidParams, err := query.LiquidParams(chainEndpoint)
			s.T().Logf("After Liquid parameters update proposal")
			s.Require().NoError(err)

			s.Require().Equal(liquidParams.Params.GlobalLiquidStakingCap, math.LegacyNewDecWithPrec(25, 2))
			s.Require().Equal(liquidParams.Params.ValidatorLiquidStakingCap, math.LegacyNewDecWithPrec(100, 2))

			return true
		},
		15*time.Second,
		5*time.Second,
	)
	delegatorAddress, _ := s.Resources.ChainA.GenesisAccounts[3].KeyInfo.GetAddress()

	fees := sdk.NewCoin(common.UAtomDenom, math.NewInt(1))

	pool, err := query.Pool(chainEndpoint)
	s.Require().NoError(err)
	tokens, err := query.TotalLiquidStaked(chainEndpoint)
	s.Require().NoError(err)
	liquidStakedAmount, err := strconv.ParseInt(tokens, 10, 64)
	s.Require().NoError(err)

	delegationAmount := math.LegacyNewDec(pool.BondedTokens.Int64()).QuoInt(math.NewInt(4)).Sub(math.LegacyNewDec(
		liquidStakedAmount)).TruncateInt().Sub(underBuffer)
	delegation := sdk.NewCoin(common.UAtomDenom, delegationAmount)

	// Alice delegate uatom to Validator A
	s.ExecDelegate(s.Resources.ChainA, 0, delegation.String(), validatorAddressB, delegatorAddress.String(), common.GaiaHomePath, fees.String())

	// Validate delegation successful
	s.Require().Eventually(
		func() bool {
			res, err := query.Delegation(chainEndpoint, validatorAddressB, delegatorAddress.String())
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(math.LegacyNewDecFromInt(delegationAmount))
		},
		20*time.Second,
		5*time.Second,
	)

	// Tokenize shares
	tokenizeAmount := delegationAmount
	tokenize := sdk.NewCoin(common.UAtomDenom, tokenizeAmount)
	s.executeTokenizeShares(s.Resources.ChainA, 0, tokenize.String(), validatorAddressB, delegatorAddress.String(), common.GaiaHomePath, fees.String())

	recordID, err := query.LastTokenizeShareRecordID(chainEndpoint)
	s.Require().NoError(err)

	// Validate balance increased
	shareDenom := fmt.Sprintf("%s/%s", strings.ToLower(validatorAddressB), strconv.Itoa(int(recordID))) //nolint:gosec
	s.Require().Eventually(
		func() bool {
			res, err := query.SpecificBalance(chainEndpoint, delegatorAddress.String(), shareDenom)
			s.Require().NoError(err)
			return res.Amount.Equal(tokenizeAmount)
		},
		20*time.Second,
		5*time.Second,
	)

	// Try to tokenize over the limit
	secondTokenizeAmount := math.NewInt(5)
	secondTokenize := sdk.NewCoin(common.UAtomDenom, secondTokenizeAmount)
	s.executeTokenizeSharesFailure(s.Resources.ChainA, 0, secondTokenize.String(), validatorAddressB, delegatorAddress.String(),
		common.GaiaHomePath, fees.String())

	// Redeem tokens for shares
	redeemAmount := sdk.NewCoin(shareDenom, tokenizeAmount)
	s.executeRedeemShares(s.Resources.ChainA, 0, redeemAmount.String(), delegatorAddress.String(), common.GaiaHomePath, fees.String())

	// check redeem success
	s.Require().Eventually(
		func() bool {
			balanceRes, err := query.SpecificBalance(chainEndpoint, delegatorAddress.String(), shareDenom)
			s.Require().NoError(err)
			if !balanceRes.Amount.IsNil() && balanceRes.Amount.IsZero() {
				return false
			}
			return true
		},
		20*time.Second,
		5*time.Second,
	)
}

// testLiquidValidatorLimit validates the validator liquid staking cap
func (s *IntegrationTestSuite) testLiquidValidatorLimit() {
	chainEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))

	validatorA := s.Resources.ChainA.Validators[0]
	validatorAAddr, _ := validatorA.KeyInfo.GetAddress()
	validatorB := s.Resources.ChainA.Validators[1]
	validatorBAddr, _ := validatorB.KeyInfo.GetAddress()

	validatorAddressB := sdk.ValAddress(validatorBAddr).String()

	err := msg.WriteLiquidStakingParamsUpdateProposal(s.Resources.ChainA, 100, 10)
	s.Require().NoError(err)
	s.TestCounters.ProposalCounter++
	submitGovFlags := []string{configFile(common.ProposalLiquidParamUpdateFilename)}
	depositGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.TestCounters.ProposalCounter), "yes"}

	// gov proposing Liquid parameters (global liquid staking cap, validator liquid staking cap, validator bond factor)
	s.T().Logf("Proposal number: %d", s.TestCounters.ProposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: Set parameters (global liquid staking cap, validator liquid staking cap)")
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.TestCounters.ProposalCounter, "liquidtypes.MsgUpdateProposal",
		submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	// query the proposal status and new fee
	s.Require().Eventually(
		func() bool {
			proposal, err := query.GovProposal(chainEndpoint, s.TestCounters.ProposalCounter)
			s.Require().NoError(err)
			return proposal.GetProposal().Status == govv1beta1.StatusPassed
		},
		15*time.Second,
		5*time.Second,
	)

	s.Require().Eventually(
		func() bool {
			liquidParams, err := query.LiquidParams(chainEndpoint)
			s.T().Logf("After Liquid parameters update proposal")
			s.Require().NoError(err)

			s.Require().Equal(liquidParams.Params.GlobalLiquidStakingCap, math.LegacyNewDecWithPrec(100, 2))
			s.Require().Equal(liquidParams.Params.ValidatorLiquidStakingCap, math.LegacyNewDecWithPrec(10, 2))

			return true
		},
		15*time.Second,
		5*time.Second,
	)
	delegatorAddress, _ := s.Resources.ChainA.GenesisAccounts[4].KeyInfo.GetAddress()

	fees := sdk.NewCoin(common.UAtomDenom, math.NewInt(1))

	valB, err := query.Validator(chainEndpoint, validatorAddressB)
	s.Require().NoError(err)

	// delegationAmount should put us at exactly 10% of validator's delegator shares
	// delegationAmount := valB.DelegatorShares.TruncateInt().Quo(math.NewInt(9))
	delegationAmount := valB.DelegatorShares.Quo(math.LegacyNewDecFromInt(math.NewInt(9))).TruncateInt().Sub(underBuffer)
	delegation := sdk.NewCoin(common.UAtomDenom, delegationAmount)

	// Alice delegate uatom to Validator B
	s.ExecDelegate(s.Resources.ChainA, 0, delegation.String(), validatorAddressB, delegatorAddress.String(), common.GaiaHomePath, fees.String())

	// Validate delegation successful
	s.Require().Eventually(
		func() bool {
			res, err := query.Delegation(chainEndpoint, validatorAddressB, delegatorAddress.String())
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(math.LegacyNewDecFromInt(delegationAmount))
		},
		20*time.Second,
		5*time.Second,
	)

	// Tokenize shares
	tokenizeAmount := delegationAmount
	tokenize := sdk.NewCoin(common.UAtomDenom, tokenizeAmount)
	s.executeTokenizeShares(s.Resources.ChainA, 0, tokenize.String(), validatorAddressB, delegatorAddress.String(), common.GaiaHomePath, fees.String())

	recordID, err := query.LastTokenizeShareRecordID(chainEndpoint)
	s.Require().NoError(err)

	// Validate balance increased
	shareDenom := fmt.Sprintf("%s/%s", strings.ToLower(validatorAddressB), strconv.Itoa(int(recordID))) //nolint:gosec
	s.Require().Eventually(
		func() bool {
			res, err := query.SpecificBalance(chainEndpoint, delegatorAddress.String(), shareDenom)
			s.Require().NoError(err)
			return res.Amount.Equal(tokenizeAmount)
		},
		20*time.Second,
		5*time.Second,
	)

	// Try to tokenize over the limit
	secondTokenizeAmount := math.NewInt(5)
	secondTokenize := sdk.NewCoin(common.UAtomDenom, secondTokenizeAmount)
	s.executeTokenizeSharesFailure(s.Resources.ChainA, 0, secondTokenize.String(), validatorAddressB, delegatorAddress.String(),
		common.GaiaHomePath, fees.String())

	// Redeem tokens for shares
	redeemAmount := sdk.NewCoin(shareDenom, tokenizeAmount)
	s.executeRedeemShares(s.Resources.ChainA, 0, redeemAmount.String(), delegatorAddress.String(), common.GaiaHomePath, fees.String())

	// check redeem success
	s.Require().Eventually(
		func() bool {
			balanceRes, err := query.SpecificBalance(chainEndpoint, delegatorAddress.String(), shareDenom)
			s.Require().NoError(err)
			if !balanceRes.Amount.IsNil() && balanceRes.Amount.IsZero() {
				return false
			}
			return true
		},
		20*time.Second,
		5*time.Second,
	)
}

func (s *IntegrationTestSuite) executeTokenizeSharesFailure(c *common.Chain, valIdx int, amount, valOperAddress,
	delegatorAddr, home, delegateFees string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx liquid tokenize-share %s", c.ID)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		liquidtypes.ModuleName,
		"tokenize-share",
		valOperAddress,
		amount,
		delegatorAddr,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 1000000),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.ExpectErrExecValidation(c, valIdx, true))
	s.T().Logf("%s expected failure on execution of tokenize share tx from %s", delegatorAddr, valOperAddress)
}

func (s *IntegrationTestSuite) executeTokenizeShares(c *common.Chain, valIdx int, amount, valOperAddress, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx liquid tokenize-share %s", c.ID)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		liquidtypes.ModuleName,
		"tokenize-share",
		valOperAddress,
		amount,
		delegatorAddr,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 1000000),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.DefaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully executed tokenize share tx from %s", delegatorAddr, valOperAddress)
}

func (s *IntegrationTestSuite) executeRedeemShares(c *common.Chain, valIdx int, amount, delegatorAddr, home, delegateFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx liquid redeem-tokens %s", c.ID)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		liquidtypes.ModuleName,
		"redeem-tokens",
		amount,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, delegatorAddr),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, delegateFees),
		fmt.Sprintf("--%s=%d", flags.FlagGas, 1000000),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.DefaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully executed redeem share tx for %s", delegatorAddr, amount)
}

func (s *IntegrationTestSuite) executeTransferTokenizeShareRecord(c *common.Chain, valIdx int, recordID, owner, newOwner, home, txFees string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	s.T().Logf("Executing gaiad tx liquid transfer-tokenize-share-record %s", c.ID)

	gaiaCommand := []string{
		common.GaiadBinary,
		common.TxCommand,
		liquidtypes.ModuleName,
		"transfer-tokenize-share-record",
		recordID,
		newOwner,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, owner),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.ID),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, txFees),
		"--keyring-backend=test",
		fmt.Sprintf("--%s=%s", flags.FlagHome, home),
		"--output=json",
		"-y",
	}

	s.ExecuteGaiaTxCommand(ctx, c, gaiaCommand, valIdx, s.DefaultExecValidation(c, valIdx))
	s.T().Logf("%s successfully executed transfer tokenize share record for %s", owner, recordID)
}
