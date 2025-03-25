package e2e

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	liquidtypes "github.com/cosmos/gaia/v23/x/liquid/types"
)

var (
	// underBuffer is the number of tokens under the limit to tokenize
	underBuffer = math.NewInt(5_000_000)
)

func (s *IntegrationTestSuite) testLiquid() {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	validatorAAddr, _ := validatorA.keyInfo.GetAddress()

	validatorAddressA := sdk.ValAddress(validatorAAddr).String()

	s.writeLiquidStakingParamsUpdateProposal(s.chainA, 25, 50)
	proposalCounter++
	submitGovFlags := []string{configFile(proposalLiquidParamUpdateFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	// gov proposing Liquid parameters (global liquid staking cap, validator liquid staking cap, validator bond factor)
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: Set parameters (global liquid staking cap, validator liquid staking cap, validator bond factor)")
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, "liquidtypes.MsgUpdateProposal",
		submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	// query the proposal status and new fee
	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(chainEndpoint, proposalCounter)
			s.Require().NoError(err)
			return proposal.GetProposal().Status == govv1beta1.StatusPassed
		},
		15*time.Second,
		5*time.Second,
	)

	s.Require().Eventually(
		func() bool {
			liquidParams, err := queryLiquidParams(chainEndpoint)
			s.T().Logf("After Liquid parameters update proposal")
			s.Require().NoError(err)

			s.Require().Equal(liquidParams.Params.GlobalLiquidStakingCap, math.LegacyNewDecWithPrec(25, 2))
			s.Require().Equal(liquidParams.Params.ValidatorLiquidStakingCap, math.LegacyNewDecWithPrec(50, 2))

			return true
		},
		15*time.Second,
		5*time.Second,
	)
	delegatorAddress, _ := s.chainA.genesisAccounts[2].keyInfo.GetAddress()

	fees := sdk.NewCoin(uatomDenom, math.NewInt(1))

	delegationAmount := math.NewInt(500000000)
	delegation := sdk.NewCoin(uatomDenom, delegationAmount) // 500 atom

	// Alice delegate uatom to Validator A
	s.execDelegate(s.chainA, 0, delegation.String(), validatorAddressA, delegatorAddress.String(), gaiaHomePath, fees.String())

	// Validate delegation successful
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressA, delegatorAddress.String())
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(math.LegacyNewDecFromInt(delegationAmount))
		},
		20*time.Second,
		5*time.Second,
	)

	// Tokenize shares
	tokenizeAmount := math.NewInt(200000000)
	tokenize := sdk.NewCoin(uatomDenom, tokenizeAmount) // 200 atom
	s.executeTokenizeShares(s.chainA, 0, tokenize.String(), validatorAddressA, delegatorAddress.String(), gaiaHomePath, fees.String())

	// Validate delegation reduced
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressA, delegatorAddress.String())
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
			res, err := getSpecificBalance(chainEndpoint, delegatorAddress.String(), shareDenom)
			s.Require().NoError(err)
			return res.Amount.Equal(tokenizeAmount)
		},
		20*time.Second,
		5*time.Second,
	)

	// Bank send Liquid token
	sendAmount := sdk.NewCoin(shareDenom, tokenizeAmount)
	s.execBankSend(s.chainA, 0, delegatorAddress.String(), validatorAAddr.String(), sendAmount.String(), standardFees.String(), false)

	// Validate tokens are sent properly
	s.Require().Eventually(
		func() bool {
			afterSenderShareDenomBalance, err := getSpecificBalance(chainEndpoint, delegatorAddress.String(), shareDenom)
			s.Require().NoError(err)

			afterRecipientShareDenomBalance, err := getSpecificBalance(chainEndpoint, validatorAAddr.String(), shareDenom)
			s.Require().NoError(err)

			decremented := afterSenderShareDenomBalance.IsNil() || afterSenderShareDenomBalance.IsZero()
			incremented := afterRecipientShareDenomBalance.IsEqual(sendAmount)

			return decremented && incremented
		},
		time.Minute,
		5*time.Second,
	)

	// transfer reward ownership
	s.executeTransferTokenizeShareRecord(s.chainA, 0, strconv.Itoa(recordID), delegatorAddress.String(), validatorAAddr.String(), gaiaHomePath, standardFees.String())
	tokenizeShareRecord := liquidtypes.TokenizeShareRecord{}
	// Validate ownership transferred correctly
	s.Require().Eventually(
		func() bool {
			record, err := queryTokenizeShareRecordByID(chainEndpoint, recordID)
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
	sendRecipientAddr, _ := s.chainB.validators[0].keyInfo.GetAddress()
	s.sendIBC(s.chainA, 0, validatorAAddr.String(), sendRecipientAddr.String(), ibcTransferAmount.String(), standardFees.String(), "memo", false)

	s.Require().Eventually(
		func() bool {
			afterSenderShareBalance, err := getSpecificBalance(chainEndpoint, validatorAAddr.String(), shareDenom)
			s.Require().NoError(err)

			decremented := afterSenderShareBalance.Add(ibcTransferAmount).IsEqual(sendAmount)
			return decremented
		},
		1*time.Minute,
		5*time.Second,
	)

	// Redeem tokens for shares
	redeemAmount := sendAmount.Sub(ibcTransferAmount)
	s.executeRedeemShares(s.chainA, 0, redeemAmount.String(), validatorAAddr.String(), gaiaHomePath, fees.String())

	// check redeem success
	s.Require().Eventually(
		func() bool {
			balanceRes, err := getSpecificBalance(chainEndpoint, validatorAAddr.String(), shareDenom)
			s.Require().NoError(err)
			if !balanceRes.Amount.IsNil() && balanceRes.Amount.IsZero() {
				return false
			}

			// check that tokenize share record module account received some rewards, since it unbonded during redeem tx execution
			balanceRes, err = getSpecificBalance(chainEndpoint, tokenizeShareRecord.GetModuleAddress().String(), uatomDenom)
			s.Require().NoError(err)
			if balanceRes.Amount.IsNil() || balanceRes.Amount.IsZero() {
				return false
			}
			return true
		},
		20*time.Second,
		5*time.Second,
	)
}

// testLiquidGlobalLimit validates the global liquid staking cap
func (s *IntegrationTestSuite) testLiquidGlobalLimit() {
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	validatorAAddr, _ := validatorA.keyInfo.GetAddress()
	validatorB := s.chainA.validators[1]
	validatorBAddr, _ := validatorB.keyInfo.GetAddress()

	validatorAddressB := sdk.ValAddress(validatorBAddr).String()

	s.writeLiquidStakingParamsUpdateProposal(s.chainA, 25, 100)
	proposalCounter++
	submitGovFlags := []string{configFile(proposalLiquidParamUpdateFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	// gov proposing Liquid parameters (global liquid staking cap, validator liquid staking cap, validator bond factor)
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: Set parameters (global liquid staking cap, validator liquid staking cap)")
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, "liquidtypes.MsgUpdateProposal",
		submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	// query the proposal status and new fee
	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(chainEndpoint, proposalCounter)
			s.Require().NoError(err)
			return proposal.GetProposal().Status == govv1beta1.StatusPassed
		},
		15*time.Second,
		5*time.Second,
	)

	s.Require().Eventually(
		func() bool {
			liquidParams, err := queryLiquidParams(chainEndpoint)
			s.T().Logf("After Liquid parameters update proposal")
			s.Require().NoError(err)

			s.Require().Equal(liquidParams.Params.GlobalLiquidStakingCap, math.LegacyNewDecWithPrec(25, 2))
			s.Require().Equal(liquidParams.Params.ValidatorLiquidStakingCap, math.LegacyNewDecWithPrec(100, 2))

			return true
		},
		15*time.Second,
		5*time.Second,
	)
	delegatorAddress, _ := s.chainA.genesisAccounts[3].keyInfo.GetAddress()

	fees := sdk.NewCoin(uatomDenom, math.NewInt(1))

	pool, err := queryPool(chainEndpoint)
	s.Require().NoError(err)
	tokens, err := queryTotalLiquidStaked(chainEndpoint)
	s.Require().NoError(err)
	liquidStakedAmount, err := strconv.ParseInt(tokens, 10, 64)
	s.Require().NoError(err)

	delegationAmount := math.LegacyNewDec(pool.BondedTokens.Int64()).QuoInt(math.NewInt(4)).Sub(math.LegacyNewDec(
		liquidStakedAmount)).TruncateInt().Sub(underBuffer)
	delegation := sdk.NewCoin(uatomDenom, delegationAmount)

	// Alice delegate uatom to Validator A
	s.execDelegate(s.chainA, 0, delegation.String(), validatorAddressB, delegatorAddress.String(), gaiaHomePath, fees.String())

	// Validate delegation successful
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressB, delegatorAddress.String())
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(math.LegacyNewDecFromInt(delegationAmount))
		},
		20*time.Second,
		5*time.Second,
	)

	// Tokenize shares
	tokenizeAmount := delegationAmount
	tokenize := sdk.NewCoin(uatomDenom, tokenizeAmount)
	s.executeTokenizeShares(s.chainA, 0, tokenize.String(), validatorAddressB, delegatorAddress.String(), gaiaHomePath, fees.String())

	recordID, err := queryLastTokenizeShareRecordID(chainEndpoint)
	s.Require().NoError(err)

	// Validate balance increased
	shareDenom := fmt.Sprintf("%s/%s", strings.ToLower(validatorAddressB), strconv.Itoa(int(recordID)))
	s.Require().Eventually(
		func() bool {
			res, err := getSpecificBalance(chainEndpoint, delegatorAddress.String(), shareDenom)
			s.Require().NoError(err)
			return res.Amount.Equal(tokenizeAmount)
		},
		20*time.Second,
		5*time.Second,
	)

	// Try to tokenize over the limit
	secondTokenizeAmount := math.NewInt(5)
	secondTokenize := sdk.NewCoin(uatomDenom, secondTokenizeAmount)
	s.executeTokenizeSharesFailure(s.chainA, 0, secondTokenize.String(), validatorAddressB, delegatorAddress.String(),
		gaiaHomePath, fees.String())

	// Redeem tokens for shares
	redeemAmount := sdk.NewCoin(shareDenom, tokenizeAmount)
	s.executeRedeemShares(s.chainA, 0, redeemAmount.String(), delegatorAddress.String(), gaiaHomePath, fees.String())

	// check redeem success
	s.Require().Eventually(
		func() bool {
			balanceRes, err := getSpecificBalance(chainEndpoint, delegatorAddress.String(), shareDenom)
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
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorA := s.chainA.validators[0]
	validatorAAddr, _ := validatorA.keyInfo.GetAddress()
	validatorB := s.chainA.validators[1]
	validatorBAddr, _ := validatorB.keyInfo.GetAddress()

	validatorAddressB := sdk.ValAddress(validatorBAddr).String()

	s.writeLiquidStakingParamsUpdateProposal(s.chainA, 100, 10)
	proposalCounter++
	submitGovFlags := []string{configFile(proposalLiquidParamUpdateFilename)}
	depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}

	// gov proposing Liquid parameters (global liquid staking cap, validator liquid staking cap, validator bond factor)
	s.T().Logf("Proposal number: %d", proposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: Set parameters (global liquid staking cap, validator liquid staking cap)")
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), proposalCounter, "liquidtypes.MsgUpdateProposal",
		submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	// query the proposal status and new fee
	s.Require().Eventually(
		func() bool {
			proposal, err := queryGovProposal(chainEndpoint, proposalCounter)
			s.Require().NoError(err)
			return proposal.GetProposal().Status == govv1beta1.StatusPassed
		},
		15*time.Second,
		5*time.Second,
	)

	s.Require().Eventually(
		func() bool {
			liquidParams, err := queryLiquidParams(chainEndpoint)
			s.T().Logf("After Liquid parameters update proposal")
			s.Require().NoError(err)

			s.Require().Equal(liquidParams.Params.GlobalLiquidStakingCap, math.LegacyNewDecWithPrec(100, 2))
			s.Require().Equal(liquidParams.Params.ValidatorLiquidStakingCap, math.LegacyNewDecWithPrec(10, 2))

			return true
		},
		15*time.Second,
		5*time.Second,
	)
	delegatorAddress, _ := s.chainA.genesisAccounts[4].keyInfo.GetAddress()

	fees := sdk.NewCoin(uatomDenom, math.NewInt(1))

	valB, err := queryValidator(chainEndpoint, validatorAddressB)
	s.Require().NoError(err)

	// delegationAmount should put us at exactly 10% of validator's delegator shares
	// delegationAmount := valB.DelegatorShares.TruncateInt().Quo(math.NewInt(9))
	delegationAmount := valB.DelegatorShares.Quo(math.LegacyNewDecFromInt(math.NewInt(9))).TruncateInt().Sub(underBuffer)
	delegation := sdk.NewCoin(uatomDenom, delegationAmount)

	// Alice delegate uatom to Validator B
	s.execDelegate(s.chainA, 0, delegation.String(), validatorAddressB, delegatorAddress.String(), gaiaHomePath, fees.String())

	// Validate delegation successful
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressB, delegatorAddress.String())
			amt := res.GetDelegationResponse().GetDelegation().GetShares()
			s.Require().NoError(err)

			return amt.Equal(math.LegacyNewDecFromInt(delegationAmount))
		},
		20*time.Second,
		5*time.Second,
	)

	// Tokenize shares
	tokenizeAmount := delegationAmount
	tokenize := sdk.NewCoin(uatomDenom, tokenizeAmount)
	s.executeTokenizeShares(s.chainA, 0, tokenize.String(), validatorAddressB, delegatorAddress.String(), gaiaHomePath, fees.String())

	recordID, err := queryLastTokenizeShareRecordID(chainEndpoint)
	s.Require().NoError(err)

	// Validate balance increased
	shareDenom := fmt.Sprintf("%s/%s", strings.ToLower(validatorAddressB), strconv.Itoa(int(recordID)))
	s.Require().Eventually(
		func() bool {
			res, err := getSpecificBalance(chainEndpoint, delegatorAddress.String(), shareDenom)
			s.Require().NoError(err)
			return res.Amount.Equal(tokenizeAmount)
		},
		20*time.Second,
		5*time.Second,
	)

	// Try to tokenize over the limit
	secondTokenizeAmount := math.NewInt(5)
	secondTokenize := sdk.NewCoin(uatomDenom, secondTokenizeAmount)
	s.executeTokenizeSharesFailure(s.chainA, 0, secondTokenize.String(), validatorAddressB, delegatorAddress.String(),
		gaiaHomePath, fees.String())

	// Redeem tokens for shares
	redeemAmount := sdk.NewCoin(shareDenom, tokenizeAmount)
	s.executeRedeemShares(s.chainA, 0, redeemAmount.String(), delegatorAddress.String(), gaiaHomePath, fees.String())

	// check redeem success
	s.Require().Eventually(
		func() bool {
			balanceRes, err := getSpecificBalance(chainEndpoint, delegatorAddress.String(), shareDenom)
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
