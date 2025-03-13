package e2e

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/msg"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
)

func (s *IntegrationTestSuite) testLSM() {
	chainEndpoint := fmt.Sprintf("http://%s", s.commonHelper.Resources.ValResources[s.commonHelper.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))

	validatorA := s.commonHelper.Resources.ChainA.Validators[0]
	validatorAAddr, _ := validatorA.KeyInfo.GetAddress()

	validatorAddressA := sdk.ValAddress(validatorAAddr).String()

	oldStakingParams, err := query.StakingParams(chainEndpoint)
	s.Require().NoError(err)
	err = msg.WriteLiquidStakingParamsUpdateProposal(s.commonHelper.Resources.ChainA, oldStakingParams.Params)
	s.Require().NoError(err)
	s.commonHelper.TestCounters.ProposalCounter++
	submitGovFlags := []string{configFile(common.ProposalLSMParamUpdateFilename)}
	depositGovFlags := []string{strconv.Itoa(s.commonHelper.TestCounters.ProposalCounter), common.DepositAmount.String()}
	voteGovFlags := []string{strconv.Itoa(s.commonHelper.TestCounters.ProposalCounter), "yes"}

	// gov proposing LSM parameters (global liquid staking cap, validator liquid staking cap, validator bond factor)
	s.T().Logf("Proposal number: %d", s.commonHelper.TestCounters.ProposalCounter)
	s.T().Logf("Submitting, deposit and vote legacy Gov Proposal: Set parameters (global liquid staking cap, validator liquid staking cap, validator bond factor)")
	s.submitGovProposal(chainEndpoint, validatorAAddr.String(), s.commonHelper.TestCounters.ProposalCounter, "stakingtypes.MsgUpdateProposal", submitGovFlags, depositGovFlags, voteGovFlags, "vote")

	// query the proposal status and new fee
	s.Require().Eventually(
		func() bool {
			proposal, err := query.GovProposal(chainEndpoint, s.commonHelper.TestCounters.ProposalCounter)
			s.Require().NoError(err)
			return proposal.GetProposal().Status == govv1beta1.StatusPassed
		},
		15*time.Second,
		5*time.Second,
	)

	s.Require().Eventually(
		func() bool {
			stakingParams, err := query.StakingParams(chainEndpoint)
			s.T().Logf("After LSM parameters update proposal")
			s.Require().NoError(err)

			s.Require().Equal(stakingParams.Params.GlobalLiquidStakingCap, math.LegacyNewDecWithPrec(25, 2))
			s.Require().Equal(stakingParams.Params.ValidatorLiquidStakingCap, math.LegacyNewDecWithPrec(50, 2))
			s.Require().Equal(stakingParams.Params.ValidatorBondFactor, math.LegacyNewDec(250))

			return true
		},
		15*time.Second,
		5*time.Second,
	)
	delegatorAddress, _ := s.commonHelper.Resources.ChainA.GenesisAccounts[2].KeyInfo.GetAddress()

	fees := sdk.NewCoin(common.UAtomDenom, math.NewInt(1))

	// Validator bond
	s.tx.ExecuteValidatorBond(s.commonHelper.Resources.ChainA, 0, validatorAddressA, validatorAAddr.String(), common.GaiaHomePath, fees.String())

	// Validate validator bond successful
	selfBondedShares := math.LegacyZeroDec()
	s.Require().Eventually(
		func() bool {
			res, err := query.Delegation(chainEndpoint, validatorAddressA, validatorAAddr.String())
			delegation := res.GetDelegationResponse().GetDelegation()
			selfBondedShares = delegation.Shares
			isValidatorBond := delegation.ValidatorBond
			s.Require().NoError(err)

			return isValidatorBond == true
		},
		20*time.Second,
		5*time.Second,
	)

	delegationAmount := math.NewInt(500000000)
	delegation := sdk.NewCoin(common.UAtomDenom, delegationAmount) // 500 atom

	// Alice delegate uatom to Validator A
	s.tx.ExecDelegate(s.commonHelper.Resources.ChainA, 0, delegation.String(), validatorAddressA, delegatorAddress.String(), common.GaiaHomePath, fees.String())

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
	s.tx.ExecuteTokenizeShares(s.commonHelper.Resources.ChainA, 0, tokenize.String(), validatorAddressA, delegatorAddress.String(), common.GaiaHomePath, fees.String())

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

	// Bank send LSM token
	sendAmount := sdk.NewCoin(shareDenom, tokenizeAmount)
	s.tx.ExecBankSend(s.commonHelper.Resources.ChainA, 0, delegatorAddress.String(), validatorAAddr.String(), sendAmount.String(), common.StandardFees.String(), false)

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
	s.tx.ExecuteTransferTokenizeShareRecord(s.commonHelper.Resources.ChainA, 0, strconv.Itoa(recordID), delegatorAddress.String(), validatorAAddr.String(), common.GaiaHomePath, common.StandardFees.String())
	tokenizeShareRecord := stakingtypes.TokenizeShareRecord{}
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

	// IBC transfer LSM token
	ibcTransferAmount := sdk.NewCoin(shareDenom, math.NewInt(100000000))
	sendRecipientAddr, _ := s.commonHelper.Resources.ChainB.Validators[0].KeyInfo.GetAddress()
	s.tx.SendIBC(s.commonHelper.Resources.ChainA, 0, validatorAAddr.String(), sendRecipientAddr.String(), ibcTransferAmount.String(), common.StandardFees.String(), "memo", common.TransferChannel, nil, false)

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
	s.tx.ExecuteRedeemShares(s.commonHelper.Resources.ChainA, 0, redeemAmount.String(), validatorAAddr.String(), common.GaiaHomePath, fees.String())

	// check redeem success
	s.Require().Eventually(
		func() bool {
			balanceRes, err := query.SpecificBalance(chainEndpoint, validatorAAddr.String(), shareDenom)
			s.Require().NoError(err)
			if !balanceRes.Amount.IsNil() && balanceRes.Amount.IsZero() {
				return false
			}

			delegationRes, err := query.Delegation(chainEndpoint, validatorAddressA, validatorAAddr.String())
			delegation := delegationRes.GetDelegationResponse().GetDelegation()
			s.Require().NoError(err)

			if !delegation.Shares.Equal(selfBondedShares.Add(math.LegacyNewDecFromInt(redeemAmount.Amount))) {
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
}
