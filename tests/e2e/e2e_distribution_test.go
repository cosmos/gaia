package e2e

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) testDistribution() {

	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	validatorB := s.chainA.validators[1]
	validatorBAddr := validatorB.keyInfo.GetAddress()

	valOperAddressA := sdk.ValAddress(validatorBAddr).String()

	delegatorAddress := s.chainA.genesisAccounts[2].keyInfo.GetAddress().String()
	newWithdrawalAddress := s.chainA.genesisAccounts[3].keyInfo.GetAddress().String()

	fees := sdk.NewCoin(uatomDenom, sdk.NewInt(1000))

	beforeBalance, err := getSpecificBalance(chainEndpoint, newWithdrawalAddress, uatomDenom)
	s.Require().NoError(err)
	if beforeBalance.IsNil() {
		beforeBalance = sdk.NewCoin(uatomDenom, sdk.NewInt(0))
	}

	s.execSetWithdrawAddress(s.chainA, 0, fees.String(), delegatorAddress, newWithdrawalAddress, gaiaHomePath)

	// Verify
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegatorWithdrawalAddress(chainEndpoint, delegatorAddress)
			s.Require().NoError(err)

			return res.WithdrawAddress == newWithdrawalAddress
		},
		10*time.Second,
		5*time.Second,
	)

	s.execWithdrawReward(s.chainA, 0, delegatorAddress, valOperAddressA, gaiaHomePath)
	s.Require().Eventually(
		func() bool {
			afterBalance, err := getSpecificBalance(chainEndpoint, newWithdrawalAddress, uatomDenom)
			s.Require().NoError(err)

			return afterBalance.IsGTE(beforeBalance)
		},
		10*time.Second,
		5*time.Second,
	)
}
