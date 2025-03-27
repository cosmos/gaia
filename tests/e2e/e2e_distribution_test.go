package e2e

import (
	"fmt"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/gaia/v23/tests/e2e/common"
	"github.com/cosmos/gaia/v23/tests/e2e/query"
)

func (s *IntegrationTestSuite) testDistribution() {
	chainEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))

	validatorB := s.Resources.ChainA.Validators[1]
	validatorBAddr, _ := validatorB.KeyInfo.GetAddress()

	valOperAddressA := sdk.ValAddress(validatorBAddr).String()

	delegatorAddress, _ := s.Resources.ChainA.GenesisAccounts[2].KeyInfo.GetAddress()

	newWithdrawalAddress, _ := s.Resources.ChainA.GenesisAccounts[3].KeyInfo.GetAddress()
	fees := sdk.NewCoin(common.UAtomDenom, math.NewInt(1000))

	beforeBalance, err := query.SpecificBalance(chainEndpoint, newWithdrawalAddress.String(), common.UAtomDenom)
	s.Require().NoError(err)
	if beforeBalance.IsNil() {
		beforeBalance = sdk.NewCoin(common.UAtomDenom, math.NewInt(0))
	}

	s.ExecSetWithdrawAddress(s.Resources.ChainA, 0, fees.String(), delegatorAddress.String(), newWithdrawalAddress.String(), common.GaiaHomePath)

	// Verify
	s.Require().Eventually(
		func() bool {
			res, err := query.DelegatorWithdrawalAddress(chainEndpoint, delegatorAddress.String())
			s.Require().NoError(err)

			return res.WithdrawAddress == newWithdrawalAddress.String()
		},
		10*time.Second,
		5*time.Second,
	)

	s.ExecWithdrawReward(s.Resources.ChainA, 0, delegatorAddress.String(), valOperAddressA, common.GaiaHomePath)
	s.Require().Eventually(
		func() bool {
			afterBalance, err := query.SpecificBalance(chainEndpoint, newWithdrawalAddress.String(), common.UAtomDenom)
			s.Require().NoError(err)

			return afterBalance.IsGTE(beforeBalance)
		},
		10*time.Second,
		5*time.Second,
	)
}

/*
fundCommunityPool tests the funding of the community pool on behalf of the distribution module.
Test Benchmarks:
1. Validation that balance of the distribution module account before funding
2. Execution funding the community pool
3. Verification that correct funds have been deposited to distribution module account
*/
func (s *IntegrationTestSuite) fundCommunityPool() {
	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.Resources.ValResources[s.Resources.ChainA.ID][0].GetHostPort("1317/tcp"))
	sender, _ := s.Resources.ChainA.Validators[0].KeyInfo.GetAddress()

	beforeDistUatomBalance, _ := query.SpecificBalance(chainAAPIEndpoint, common.DistModuleAddress, common.TokenAmount.Denom)
	if beforeDistUatomBalance.IsNil() {
		// Set balance to 0 if previous balance does not exist
		beforeDistUatomBalance = sdk.NewInt64Coin(common.UAtomDenom, 0)
	}

	s.ExecDistributionFundCommunityPool(s.Resources.ChainA, 0, sender.String(), common.TokenAmount.String(), common.StandardFees.String())

	s.Require().Eventually(
		func() bool {
			afterDistUatomBalance, err := query.SpecificBalance(chainAAPIEndpoint, common.DistModuleAddress, common.TokenAmount.Denom)
			s.Require().NoErrorf(err, "Error getting balance: %s", afterDistUatomBalance)

			// check if the balance is increased by the TokenAmount and at least some portion of
			// the fees (some amount of the fees will be given to the proposer)
			return beforeDistUatomBalance.Add(common.TokenAmount).IsLT(afterDistUatomBalance) &&
				afterDistUatomBalance.IsLT(beforeDistUatomBalance.Add(common.TokenAmount).Add(common.StandardFees))
		},
		15*time.Second,
		5*time.Second,
	)
}
