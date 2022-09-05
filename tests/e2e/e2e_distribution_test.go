package e2e

import (
	"time"
	
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) testDistribution(
	chainEndpoint,
	delegatorAddress,
	newWithrawalAddress,
	valOperAddressA,
	homePath string,
) {
	fees = sdk.NewCoin(uatomDenom, math.NewInt(1000))

	beforeBalance, err := getSpecificBalance(chainEndpoint, newWithrawalAddress, uatomDenom)
	s.Require().NoError(err)
	if beforeBalance.IsNil() {
		beforeBalance = sdk.NewCoin(uatomDenom, math.NewInt(0))
	}

	s.execSetWithrawAddress(s.chainA, 0, chainEndpoint, fees.String(), delegatorAddress, newWithrawalAddress, homePath)

	// Verify
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegatorWithdrawalAddress(chainEndpoint, delegatorAddress)
			s.Require().NoError(err)

			return res.WithdrawAddress == newWithrawalAddress
		},
		10*time.Second,
		5*time.Second,
	)

	s.execWithdrawReward(s.chainA, 0, chainEndpoint, fees.String(), delegatorAddress, valOperAddressA, homePath)
	s.Require().Eventually(
		func() bool {
			afterBalance, err := getSpecificBalance(chainEndpoint, newWithrawalAddress, uatomDenom)
			s.Require().NoError(err)

			return afterBalance.IsGTE(beforeBalance)
		},
		10*time.Second,
		5*time.Second,
	)
}
