package e2e

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) testStaking(chainEndpoint string, delegatorAddress string, validatorAddressA string, validatorAddressB string, fees sdk.Coin, homePath string) {
	delegationAmount := math.NewInt(500000000)
	delegation := sdk.NewCoin(uatomDenom, delegationAmount) // 500 atom

	// Alice delegate uatom to Validator A
	s.executeDelegate(s.chainA, 0, delegation.String(), validatorAddressA, delegatorAddress, homePath, fees.String())

	// Validate delegation successful
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressA, delegatorAddress)
			amt := res.DelegationResponse.Delegation.GetShares()
			s.Require().NoError(err)

			return amt.Equal(sdk.NewDecFromInt(delegationAmount))
		},
		20*time.Second,
		5*time.Second,
	)

	// Alice re-delegate uatom from Validator A to Validator B
	s.executeRedelegate(s.chainA, 0, delegation.String(), validatorAddressA, validatorAddressB, delegatorAddress, homePath, fees.String())

	// Validate re-delegation successful
	s.Require().Eventually(
		func() bool {
			res, err := queryDelegation(chainEndpoint, validatorAddressB, delegatorAddress)
			amt := res.DelegationResponse.Delegation.GetShares()
			s.Require().NoError(err)

			return amt.Equal(sdk.NewDecFromInt(delegationAmount))
		},
		20*time.Second,
		5*time.Second,
	)
}
