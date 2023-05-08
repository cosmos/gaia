package e2e

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) testByPassMinFeeWithdrawReward(endpoint string) {
	paidFeeAmt := math.LegacyMustNewDecFromStr(minGasPrice).Mul(math.LegacyNewDec(gas)).String()
	payee := s.chainA.validators[0].keyInfo.GetAddress()

	testCases := []struct {
		name   string
		fee    string
		expErr bool
	}{
		{
			"bypass-msg with fee in the denom of global fee, pass",
			paidFeeAmt + uatomDenom,
			false,
		},
		{
			"bypass-msg with zero coin in the denom of global fee, pass",
			"0" + uatomDenom,
			false,
		},
		{
			"bypass-msg with zero coin not in the denom of global fee, pass",
			"0" + photonDenom,
			false,
		},
		{
			"bypass-msg with non-zero coin not in the denom of global fee, fail",
			paidFeeAmt + photonDenom,
			true,
		},
	}

	for _, tc := range testCases {

		// get delegator rewards
		rewards, err := queryDelegatorTotalRewards(endpoint, payee.String())
		s.Require().NoError(err)

		// get current delegator stake balance
		oldBalance, err := getSpecificBalance(endpoint, payee.String(), stakeDenom)
		s.Require().NoError(err)

		// withdraw rewards
		s.Run(tc.name, func() {
			s.execWithdrawAllRewards(s.chainA, 0, payee.String(), tc.fee, tc.expErr)
		})

		if !tc.expErr {
			// get updated balance
			balance, err := getSpecificBalance(endpoint, payee.String(), stakeDenom)
			s.Require().NoError(err)

			// compute sum of old balance and rewards
			oldBalancePlusRewards := rewards.GetTotal().Add(sdk.NewDecCoinFromCoin(oldBalance))
			s.Require().Equal(oldBalancePlusRewards[0].Denom, stakeDenom)

			// check that the updated balance is GTE than the sum of old balance and rewards
			s.Require().True(sdk.NewDecCoinFromCoin(balance).IsGTE(oldBalancePlusRewards[0]))
		}
	}
}
